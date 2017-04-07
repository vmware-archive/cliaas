package azure_test

import (
	"os"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cliaas/iaas/azure"

	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
)

const (
	resourceManagerEndpoint = "https://management.azure.com/"
	controlOpsManVMDiskURL  = "https://opsmanagereastus.blob.core.windows.net/images/ops-manager-1.10.3.vhd"
)

var (
	controlVMBaseURL     = "https://azureintstore.blob.core.windows.net/opsmanager/"
	controlVMImageURL    = controlVMBaseURL + "opsmanagerimage193.vhd"
	subscriptionID       = os.Getenv("AZURE_SUBSCRIPTION_ID")
	clientID             = os.Getenv("AZURE_CLIENT_ID")
	clientSecret         = os.Getenv("AZURE_CLIENT_SECRET")
	tenantID             = os.Getenv("AZURE_TENANT_ID")
	resourceGroupName    = os.Getenv("AZURE_RESOURCE_GROUP_NAME")
	nicID                = os.Getenv("AZURE_NIC_ID")
	storageAccountName   = os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	storageAccountKey    = os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	storageContainerName = os.Getenv("AZURE_STORAGE_CONTAINER_NAME")
	storageBaseURL       = os.Getenv("AZURE_STORAGE_BASE_URL")
	adminVMPassword      = os.Getenv("AZURE_VM_ADMIN_PASSWORD")
	osDiskName           = "linux"
	vmLocation           = "eastus"
	vmUser               = "opsmanuser"
	vmPass               = getGUID()
)

var _ = Describe("Azure API Client", func() {
	var client compute.VirtualMachinesClient
	var identifier string
	var err error
	var azureClient *azure.Client

	BeforeEach(func() {
		client = newClient()
		identifier = "ops-man-rotation-test-" + getGUID()
		azureClient, err = azure.NewClient(subscriptionID, clientID, clientSecret, tenantID, resourceGroupName, resourceManagerEndpoint)
		Expect(err).ShouldNot(HaveOccurred())
		osDiskImageURL := fmt.Sprintf("%svm-%s.vhd", controlVMBaseURL, getGUID())
		err = createVM(client, identifier, controlVMImageURL, osDiskImageURL)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		azureClient.Delete(identifier)
		Expect(vmExists(client, identifier)).Should(BeFalse(), "this vm should have been removed")
	})

	Describe("Delete", func() {
		JustBeforeEach(func() {
			Expect(vmExists(client, identifier)).Should(BeTrue(), "does the control test VM exist?")
			err = azureClient.Delete(identifier)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("when called on a vm with a name matching the given regex", func() {
			It("should delete the matching VM from Azure", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmExists(client, identifier)).Should(BeFalse(), "was the control test VM removed?")
			})
		})
	})

	Describe("Replace", func() {
		JustBeforeEach(func() {
			Expect(vmExists(client, identifier+"_....*")).Should(BeFalse(), "was the control test VM removed?")
			azureClient.SetStorageContainerName(storageContainerName)
			azureClient.SetStorageAccountName(storageAccountName)
			azureClient.SetStorageBaseURL(storageBaseURL)
			err = azureClient.SetBlobServiceClient(storageAccountName, storageAccountKey, storageBaseURL)
			Expect(err).ShouldNot(HaveOccurred())
			err = azureClient.Replace(identifier, controlOpsManVMDiskURL)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("when called on a vm with a name matching the given regex", func() {
			It("should delete the matching VM and spin up a the new VM in its place", func() {
				Expect(vmExists(client, identifier+"_....*")).Should(BeTrue(), "was the control test VM removed?")
			})
		})
	})
})

func createVM(client compute.VirtualMachinesClient, name string, imageURL string, osDiskURL string) error {
	instance := newVirtualMachine(name, imageURL, osDiskURL)
	_, err := client.CreateOrUpdate(resourceGroupName, *instance.Name, instance, nil)
	return err
}

func newVirtualMachine(name string, vmImageURL string, osDiskURL string) compute.VirtualMachine {
	tmpName := name
	tmpImageURL := vmImageURL
	tmpOsDiskURL := osDiskURL
	vm := compute.VirtualMachine{
		Location: &vmLocation,
		Name:     &tmpName,
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: &compute.HardwareProfile{
				VMSize: compute.BasicA0,
			},
			OsProfile: &compute.OSProfile{
				ComputerName:  &vmUser,
				AdminUsername: &vmUser,
				AdminPassword: &vmPass,
			},
			NetworkProfile: &compute.NetworkProfile{
				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					compute.NetworkInterfaceReference{
						ID: &nicID,
					},
				},
			},
			StorageProfile: &compute.StorageProfile{
				OsDisk: &compute.OSDisk{
					Name: &osDiskName,
					Vhd: &compute.VirtualHardDisk{
						URI: &tmpOsDiskURL,
					},
					CreateOption: compute.FromImage,
					OsType:       compute.Linux,
					Image: &compute.VirtualHardDisk{
						URI: &tmpImageURL,
					},
				},
			},
		},
	}
	return vm
}

func vmExists(client compute.VirtualMachinesClient, identifier string) bool {
	vmListResults, err := client.List(resourceGroupName)
	if err != nil {
		return false
	}

	var matchingInstances = make([]compute.VirtualMachine, 0)
	var vmNameFilter = regexp.MustCompile(identifier)

	for vmListResults.Value != nil && len(*vmListResults.Value) > 0 {
		matchingInstances = getMatchingInstances(*vmListResults.Value, vmNameFilter, matchingInstances)
		vmListResults, err = client.ListAllNextResults(vmListResults)
		if err != nil {
			return false
		}
	}

	if len(matchingInstances) == 0 {
		return false
	}

	return true
}

func getMatchingInstances(vmList []compute.VirtualMachine, identifierRegex *regexp.Regexp, matchingInstances []compute.VirtualMachine) []compute.VirtualMachine {

	for _, instance := range vmList {
		if identifierRegex.MatchString(*instance.Name) {
			matchingInstances = append(matchingInstances, instance)
		}
	}
	return matchingInstances
}

func newClient() compute.VirtualMachinesClient {
	c := map[string]string{
		"AZURE_CLIENT_ID":                 clientID,
		"AZURE_CLIENT_SECRET":             clientSecret,
		"AZURE_SUBSCRIPTION_ID":           subscriptionID,
		"AZURE_TENANT_ID":                 tenantID,
		"AZURE_RESOURCE_GROUP_NAME":       resourceGroupName,
		"AZURE_RESOURCE_MANAGER_ENDPOINT": resourceManagerEndpoint,
	}
	if err := checkEnvVar(&c); err != nil {
		Expect(err).ShouldNot(HaveOccurred())
	}

	spt, err := helpers.NewServicePrincipalTokenFromCredentials(c, resourceManagerEndpoint)
	if err != nil {
		Expect(err).ShouldNot(HaveOccurred())
	}
	client := compute.NewVirtualMachinesClient(subscriptionID)
	client.Authorizer = spt
	return client
}

func checkEnvVar(envVars *map[string]string) error {
	var missingVars []string
	for varName, value := range *envVars {
		if value == "" {
			missingVars = append(missingVars, varName)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("Missing environment variables %v", missingVars)
	}
	return nil
}

func getGUID() string {
	uuid, _ := uuid.NewRandom()
	return uuid.String()
}
