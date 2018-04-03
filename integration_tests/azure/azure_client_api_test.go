package azure_test

import (
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cliaas/iaas/azure"

	"fmt"
	"time"
	"strconv"
)

const (
	resourceManagerEndpoint = "https://management.azure.com/"
	controlOpsManVMDiskURL  = "https://opsmanagereastus.blob.core.windows.net/images/ops-manager-1.10.3.vhd"
	controlOpsManVMDiskSize = int64(150)
)

const (
	imageURL = "https://opsmanagerwestus.blob.core.windows.net/images/ops-manager-1.11.11.vhd"
)

var (
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	clientID       = os.Getenv("AZURE_CLIENT_ID")
	clientSecret   = os.Getenv("AZURE_CLIENT_SECRET")
	tenantID       = os.Getenv("AZURE_TENANT_ID")
	prefix         = os.Getenv("PREFIX")
	location       = os.Getenv("LOCATION")

	storageAccountName = strings.Replace(prefix, "-", "", -1) + strconv.Itoa(int(time.Now().Unix()))
	containerName      = "cliaas"

	testAzureClient   azureTestClient
	subnet            network.Subnet
	newImageURL       string
	storageAccountKey string
)

var _ = Describe("Azure API Client", func() {
	var identifier string
	var storageAccountKey string
	var azureClient *azure.Client

	BeforeEach(func() {
		testAzureClient.deleteResourceGroup(prefix)

		testAzureClient = azureTestClient{
			SubscriptionID: subscriptionID,
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			TenantID:       tenantID,
			Location:       location,
		}

		testAzureClient.createResourceGroup(prefix)
		storageAccountKey = testAzureClient.createStorageAccount(prefix, storageAccountName)
		createContainer(storageAccountName, storageAccountKey, containerName)
		newImageURL, _ := copyBlob(storageAccountName, storageAccountKey, containerName, imageURL, "image.vhd")

		subnet = testAzureClient.createVirtualNetwork(prefix, fmt.Sprintf("%s-network", prefix))

		identifier = fmt.Sprintf("%s-vm", prefix)
		testAzureClient.createVM(prefix, identifier, newImageURL, storageAccountName, containerName, &subnet)

		var err error
		azureClient, err = azure.NewClient(subscriptionID, clientID, clientSecret, tenantID, prefix, resourceManagerEndpoint)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		err := testAzureClient.deleteResourceGroup(prefix)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Delete", func() {
		JustBeforeEach(func() {
			Expect(testAzureClient.vmExists(identifier, prefix)).Should(BeTrue())
			err := azureClient.Delete(identifier)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("when called on a vm with a name matching the given regex", func() {
			It("should delete the matching VM from Azure", func() {
				Expect(testAzureClient.vmExists(identifier, prefix)).Should(BeFalse())
			})
		})
	})

	Describe("Replace", func() {
		var VMID string

		JustBeforeEach(func() {
			Expect(testAzureClient.vmExists(identifier, prefix)).Should(BeTrue())

			vmListResults, err := testAzureClient.getVirtualMachinesClient().List(prefix)
			VMID = *(*vmListResults.Value)[0].VMID

			azureClient.SetStorageContainerName(containerName)
			azureClient.SetStorageAccountName(storageAccountName)
			azureClient.SetStorageBaseURL("core.windows.net")
			err = azureClient.SetBlobServiceClient(storageAccountName, storageAccountKey, "core.windows.net")
			Expect(err).ShouldNot(HaveOccurred())

			err = azureClient.Replace(identifier, controlOpsManVMDiskURL, controlOpsManVMDiskSize)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("when called on a vm with a name matching the given regex", func() {
			It("should delete the matching VM and spin up a the new VM in its place", func() {
				vmListResults, err := testAzureClient.getVirtualMachinesClient().List(prefix)
				Expect(err).ShouldNot(HaveOccurred())
				newVMID := *(*vmListResults.Value)[0].VMID
				Expect(testAzureClient.vmExists(identifier, prefix)).Should(BeTrue())
				Expect(newVMID).NotTo(Equal(VMID))

				By("the disk size should be the original disk size")
				disk, err := azureClient.GetDisk(newVMID)
				Expect(err).ToNot(HaveOccurred())
				Expect(disk).ShouldNot(BeNil())
				Expect(disk.SizeGB).Should(Equal(controlOpsManVMDiskSize))
			})
		})
	})
})
