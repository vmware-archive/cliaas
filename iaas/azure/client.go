package azure

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest"
	errwrap "github.com/pkg/errors"
)

const defaultResourceManagerEndpoint = "https://management.azure.com/"

type Client struct {
	BlobServiceClient     BlobCopier
	VirtualMachinesClient ComputeVirtualMachinesClient
	resourceGroupName     string
	storageContainerName  string
	storageAccountName    string
	storageBaseURL        string
	vmAdminPassword       string
}

type BlobCopier interface {
	CopyBlob(container, name, sourceBlob string) error
}

type ComputeVirtualMachinesClient interface {
	Get(resourceGroupName string, vmName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error)
	ListAllNextResults(lastResults compute.VirtualMachineListResult) (result compute.VirtualMachineListResult, err error)
	CreateOrUpdate(resourceGroupName string, vmName string, parameters compute.VirtualMachine, cancel <-chan struct{}) (result autorest.Response, err error)
	Delete(resourceGroupName string, vmName string, cancel <-chan struct{}) (result autorest.Response, err error)
	Deallocate(resourceGroupName string, vmName string, cancel <-chan struct{}) (result autorest.Response, err error)
	List(resourceGroupName string) (result compute.VirtualMachineListResult, err error)
}

var InvalidAzureClientErr = errors.New("invalid azure sdk client defined")
var NoMatchesErr = errors.New("no VM names match the provided prefix")
var MultipleMatchesErr = errors.New("multiple VM names match the provided prefix")

func NewClient(
	subscriptionID string,
	clientID string,
	clientSecret string,
	tenantID string,
	resourceGroupName string,
	resourceManagerEndpoint string,
) (*Client, error) {
	c := map[string]string{
		"AZURE_CLIENT_ID":       clientID,
		"AZURE_CLIENT_SECRET":   clientSecret,
		"AZURE_SUBSCRIPTION_ID": subscriptionID,
		"AZURE_TENANT_ID":       tenantID,
	}
	if err := checkEnvVar(c); err != nil {
		return nil, errwrap.Wrap(err, "failed on check of env vars")
	}
	if resourceManagerEndpoint == "" {
		resourceManagerEndpoint = defaultResourceManagerEndpoint
	}

	spt, err := helpers.NewServicePrincipalTokenFromCredentials(c, resourceManagerEndpoint)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to generate new service principal token")
	}
	client := compute.NewVirtualMachinesClient(subscriptionID)
	client.Authorizer = spt
	return &Client{
		VirtualMachinesClient: &client,
		resourceGroupName:     resourceGroupName,
	}, nil
}

func (s *Client) SetVMAdminPassword(password string) {
	s.vmAdminPassword = password
}

func (s *Client) SetStorageContainerName(name string) {
	s.storageContainerName = name
}

func (s *Client) SetStorageAccountName(name string) {
	s.storageAccountName = name
}

func (s *Client) SetStorageBaseURL(baseURL string) {
	s.storageBaseURL = baseURL
}

func (s *Client) SetBlobServiceClient(storageAccountName string, storageAccountKey string, storageURL string) error {
	blobClient, err := newBlobClient(storageAccountName, storageAccountKey, storageURL)
	if err != nil {
		return errwrap.Wrap(err, "failed creating a blob client")
	}
	s.BlobServiceClient = blobClient
	return nil
}

func newBlobClient(accountName string, accountKey string, baseURL string) (*storage.BlobStorageClient, error) {
	client, err := storage.NewClient(accountName, accountKey, baseURL, storage.DefaultAPIVersion, true)
	if err != nil {
		return nil, err
	}
	blobClient := client.GetBlobService()
	return &blobClient, nil
}

func (s *Client) Delete(identifier string) error {
	_, err := s.executeFunctionOnMatchingVM(identifier, s.VirtualMachinesClient.Delete)
	return err
}

func generateLocalImageURL(accountName string, baseURL string, containerName string, localBlobName string) string {
	return fmt.Sprintf("https://%s.blob.%s/%s/%s", accountName, baseURL, containerName, localBlobName)
}

func (s *Client) Replace(identifier string, vhdURL string) error {
	instance, err := s.deallocate(identifier)
	if err != nil {
		return errwrap.Wrap(err, "error shutting down VM")
	}

	tmpName := generateInstanceName(*instance.Name)
	localBlobName := tmpName + "-image.vhd"
	localDiskName := tmpName + "-osdisk.vhd"
	err = s.BlobServiceClient.CopyBlob(s.storageContainerName, localBlobName, vhdURL)
	if err != nil {
		return errwrap.Wrap(err, "error copying source blob to local blob")
	}

	localImageURL := generateLocalImageURL(s.storageAccountName, s.storageBaseURL, s.storageContainerName, localBlobName)
	localDiskURL := generateLocalImageURL(s.storageAccountName, s.storageBaseURL, s.storageContainerName, localDiskName)
	newInstance, err := s.generateInstanceCopy(*instance.Name, tmpName, localImageURL, localDiskURL)
	if err != nil {
		return errwrap.Wrap(err, "failed to generate a new instance object")
	}

	err = s.Delete(identifier)
	if err != nil {
		return errwrap.Wrap(err, "failed removing original VM")
	}

	_, err = s.VirtualMachinesClient.CreateOrUpdate(s.resourceGroupName, *newInstance.Name, *newInstance, nil)
	return err
}

func (s *Client) SwapLb(identifier string, vmidentifiers []string) error {
	return nil
}

func (s *Client) generateInstanceCopy(sourceInstanceName string, newInstanceName string, localImageURL string, localOSDiskURL string) (*compute.VirtualMachine, error) {
	instance, err := s.VirtualMachinesClient.Get(s.resourceGroupName, sourceInstanceName, compute.InstanceView)
	if err != nil {
		return nil, errwrap.Wrap(err, "unable to get virtual machine instance from azure api")
	}

	instance.Name = &newInstanceName
	instance.VirtualMachineProperties.StorageProfile.OsDisk.Image.URI = &localImageURL
	instance.VirtualMachineProperties.StorageProfile.OsDisk.Vhd.URI = &localOSDiskURL
	instance.VirtualMachineProperties.VMID = nil
	instance.Resources = nil

	if s.vmAdminPassword == "" {
		s.vmAdminPassword = getGUID()
	}
	instance.VirtualMachineProperties.OsProfile.AdminPassword = &s.vmAdminPassword
	return &instance, nil
}

func (s *Client) deallocate(identifier string) (*compute.VirtualMachine, error) {
	return s.executeFunctionOnMatchingVM(identifier, s.VirtualMachinesClient.Deallocate)
}

func (s *Client) executeFunctionOnMatchingVM(identifier string, f func(resourceGroupName string, vmName string, cancel <-chan struct{}) (result autorest.Response, err error)) (*compute.VirtualMachine, error) {
	matchingInstances, err := s.getFilteredList(identifier)
	if err != nil {
		return nil, errwrap.Wrap(err, "error when attempting to get filtered vm list")
	}

	switch len(matchingInstances) {
	case 0:
		return nil, NoMatchesErr
	case 1:
		_, err = f(s.resourceGroupName, *matchingInstances[0].Name, nil)
		return &matchingInstances[0], err
	default:
		return nil, MultipleMatchesErr
	}
}

func (s *Client) getFilteredList(identifier string) ([]compute.VirtualMachine, error) {
	vmListResults, err := s.VirtualMachinesClient.List(s.resourceGroupName)
	if err != nil {
		return nil, errwrap.Wrap(err, "error in getting list of VMs from azure")
	}

	var matchingInstances = make([]compute.VirtualMachine, 0)
	var vmNameFilter = regexp.MustCompile(identifier)

	for vmListResults.Value != nil && len(*vmListResults.Value) > 0 {
		matchingInstances = getMatchingInstances(*vmListResults.Value, vmNameFilter, matchingInstances)
		vmListResults, err = s.VirtualMachinesClient.ListAllNextResults(vmListResults)
		if err != nil {
			return nil, errwrap.Wrap(err, "ListAllNextResults call failed")
		}
	}
	return matchingInstances, nil
}

func getGUID() string {
	uuid, _ := uuid.NewRandom()
	localString := uuid.String()
	return localString
}

func checkEnvVar(envVars map[string]string) error {
	var missingVars []string
	for varName, value := range envVars {
		if value == "" {
			missingVars = append(missingVars, varName)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("Missing environment variables %v", missingVars)
	}
	return nil
}

func generateInstanceName(currentName string) string {
	tstamp := time.Now().Format("20060112123456")
	splits := strings.Split(currentName, "_")
	if len(splits) == 1 {
		return currentName + "_" + tstamp
	}

	truncatedSplits := splits[:len(splits)-1]
	truncatedSplits = append(truncatedSplits, tstamp)
	return strings.Join(truncatedSplits, "_")
}

func getMatchingInstances(vmList []compute.VirtualMachine, identifierRegex *regexp.Regexp, matchingInstances []compute.VirtualMachine) []compute.VirtualMachine {

	for _, instance := range vmList {
		if identifierRegex.MatchString(*instance.Name) {
			matchingInstances = append(matchingInstances, instance)
		}
	}
	return matchingInstances
}
