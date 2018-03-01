package azure_test

import (
	"fmt"
	"log"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	armstorage "github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"

	. "github.com/onsi/gomega"
)

type azureTestClient struct {
	TenantID       string
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	Location       string
}

func (t azureTestClient) createServicePrincipalToken() *azure.ServicePrincipalToken {
	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(t.TenantID)
	Expect(err).NotTo(HaveOccurred())

	token, err := azure.NewServicePrincipalToken(*oauthConfig, t.ClientID, t.ClientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	Expect(err).NotTo(HaveOccurred())

	return token
}

func (t azureTestClient) getVirtualMachinesClient() compute.VirtualMachinesClient {
	vmClient := compute.NewVirtualMachinesClient(t.SubscriptionID)
	vmClient.Authorizer = t.createServicePrincipalToken()

	return vmClient
}

func (t azureTestClient) createResourceGroup(resourceGroupName string) resources.Group {
	log.Printf("Creating resource group %s...\n", resourceGroupName)
	groupClient := resources.NewGroupsClient(t.SubscriptionID)
	groupClient.Authorizer = t.createServicePrincipalToken()

	group, err := groupClient.CreateOrUpdate(resourceGroupName, resources.Group{
		Location: &t.Location,
	})
	Expect(err).NotTo(HaveOccurred())

	return group
}

func (t azureTestClient) deleteResourceGroup(resourceGroupName string)  error{
	log.Printf("Deleting resource group '%s' ...\n", resourceGroupName)
	groupClient := resources.NewGroupsClient(t.SubscriptionID)
	groupClient.Authorizer = t.createServicePrincipalToken()

	_, err := groupClient.Delete(resourceGroupName, nil)
	return err
}

func (t azureTestClient) createStorageAccount(resourceGroupName, accountName string) string {
	log.Printf("Creating storage account %s...\n", accountName)
	accountClient := armstorage.NewAccountsClient(t.SubscriptionID)
	accountClient.Authorizer = t.createServicePrincipalToken()

	_, err := accountClient.Create(resourceGroupName, accountName, armstorage.AccountCreateParameters{
		Sku: &armstorage.Sku{
			Name: armstorage.StandardLRS,
		},
		Location: &t.Location,
		AccountPropertiesCreateParameters: &armstorage.AccountPropertiesCreateParameters{},
	}, nil)
	Expect(err).NotTo(HaveOccurred())

	result, err := accountClient.ListKeys(resourceGroupName, accountName)
	Expect(err).NotTo(HaveOccurred())

	return *(*result.Keys)[0].Value
}

func (t azureTestClient) createVirtualNetwork(resourceGroupName, networkName string) network.Subnet {
	log.Printf("Creating virtual network %s...\n", networkName)
	vNetClient := network.NewVirtualNetworksClient(t.SubscriptionID)
	subnetClient := network.NewSubnetsClient(t.SubscriptionID)

	token := t.createServicePrincipalToken()
	vNetClient.Authorizer = token
	subnetClient.Authorizer = token

	_, err := vNetClient.CreateOrUpdate(resourceGroupName, networkName, network.VirtualNetwork{
		Location: &t.Location,
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{"10.0.0.0/16"},
			},
		},
	}, nil)
	Expect(err).NotTo(HaveOccurred())

	log.Printf("Creating default subnet...\n")
	_, err = subnetClient.CreateOrUpdate(resourceGroupName, networkName, "default", network.Subnet{
		SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
			AddressPrefix: to.StringPtr("10.0.0.0/24"),
		},
	}, nil)
	Expect(err).NotTo(HaveOccurred())

	defaultSubnet, err := subnetClient.Get(resourceGroupName, networkName, "default", "")
	Expect(err).NotTo(HaveOccurred())

	return defaultSubnet
}

func (t azureTestClient) createVM(resourceGroupName, name, imageURL, accountName, containerName string, subnet *network.Subnet) {
	log.Printf("Creating vm %s...\n", name)
	vmClient := compute.NewVirtualMachinesClient(t.SubscriptionID)
	vmClient.Authorizer = t.createServicePrincipalToken()

	nicParameters := t.createNIC(resourceGroupName, name, subnet)
	osDiskImageURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s.vhd", accountName, containerName, name)

	_, err := vmClient.CreateOrUpdate(resourceGroupName, name, compute.VirtualMachine{
		Location: &t.Location,
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: &compute.HardwareProfile{
				VMSize: compute.StandardDS1,
			},
			StorageProfile: &compute.StorageProfile{
				OsDisk: &compute.OSDisk{
					Name: to.StringPtr("osDisk"),
					Vhd: &compute.VirtualHardDisk{
						URI: &osDiskImageURL,
					},
					CreateOption: compute.FromImage,
					OsType:       compute.Linux,
					Image: &compute.VirtualHardDisk{
						URI: &imageURL,
					},
				},
			},
			OsProfile: &compute.OSProfile{
				ComputerName:  &name,
				AdminUsername: to.StringPtr("notadmin"),
				AdminPassword: to.StringPtr("Pa$$w0rd1975"),
			},
			NetworkProfile: &compute.NetworkProfile{
				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					{
						ID: nicParameters.ID,
						NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
							Primary: to.BoolPtr(true),
						},
					},
				},
			},
		},
	}, nil)
	Expect(err).NotTo(HaveOccurred())
}

func (t azureTestClient) createNIC(resourceGroupName, machineName string, subnet *network.Subnet) *network.Interface {
	interfacesClient := network.NewInterfacesClient(t.SubscriptionID)
	interfacesClient.Authorizer = t.createServicePrincipalToken()

	_, err := interfacesClient.CreateOrUpdate(resourceGroupName, fmt.Sprintf("nic-%s", machineName), network.Interface{
		Location: &t.Location,
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					Name: to.StringPtr(fmt.Sprintf("ip-config-%s", machineName)),
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: network.Dynamic,
						Subnet: subnet,
					},
				},
			},
		},
	}, nil)
	Expect(err).NotTo(HaveOccurred())

	nicParameters, err := interfacesClient.Get(resourceGroupName, fmt.Sprintf("nic-%s", machineName), "")
	Expect(err).NotTo(HaveOccurred())

	return &nicParameters
}

func (t azureTestClient) vmExists(identifier, resourceGroupName string) bool {
	vmListResults, err := t.getVirtualMachinesClient().List(resourceGroupName)
	if err != nil {
		return false
	}

	var matchingInstances = make([]compute.VirtualMachine, 0)
	var vmNameFilter = regexp.MustCompile(identifier)

	for vmListResults.Value != nil && len(*vmListResults.Value) > 0 {
		matchingInstances = getMatchingInstances(*vmListResults.Value, vmNameFilter, matchingInstances)
		vmListResults, err = t.getVirtualMachinesClient().ListAllNextResults(vmListResults)
		if err != nil {
			return false
		}
	}

	if len(matchingInstances) == 0 {
		return false
	}

	return true
}

func createContainer(storageAccountName, storageAccountKey, containerName string) {
	log.Printf("Creating container '%s' with account name '%s' and account key '%s' ...\n", containerName, storageAccountName, storageAccountKey)
	client, err := storage.NewBasicClient(storageAccountName, storageAccountKey)
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(containerName)
	log.Printf("Got container '%s' with properties '%#v'", cnt.Name, cnt.Properties)
	err = cnt.Create()
	Expect(err).NotTo(HaveOccurred())
}

func copyBlob(storageAccountName, storageAccountKey, containerName, sourceBlobURL, destinationBlobName string) (string, storage.BlobStorageClient) {
	log.Printf("Copying blob '%s' to container '%s' with name '%s' ...\n", sourceBlobURL, containerName, destinationBlobName)
	client, err := storage.NewBasicClient(storageAccountName, storageAccountKey)
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	err = blobClient.CopyBlob(containerName, destinationBlobName, sourceBlobURL)
	Expect(err).NotTo(HaveOccurred())

	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", storageAccountName, containerName, destinationBlobName), blobClient
}

func deleteBlob(containerName, destinationBlobName string, blobClient storage.BlobStorageClient) error {
	return blobClient.DeleteBlob(containerName, destinationBlobName, nil)
}

func getMatchingInstances(vmList []compute.VirtualMachine, identifierRegex *regexp.Regexp, matchingInstances []compute.VirtualMachine) []compute.VirtualMachine {
	for _, instance := range vmList {
		if identifierRegex.MatchString(*instance.Name) {
			matchingInstances = append(matchingInstances, instance)
		}
	}
	return matchingInstances
}
