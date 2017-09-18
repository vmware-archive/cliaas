package azure_test

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAzure(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Azure Suite")
}

const (
	imageURL = "https://opsmanagerwestus.blob.core.windows.net/images/ops-manager-1.11.11.vhd"
)

var (
	testAzureClient   azureTestClient
	subnet            network.Subnet
	newImageURL       string
	storageAccountKey string
)

var _ = BeforeSuite(func() {
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
	newImageURL = copyBlob(storageAccountName, storageAccountKey, containerName, imageURL, "image.vhd")

	subnet = testAzureClient.createVirtualNetwork(prefix, fmt.Sprintf("%s-network", prefix))
})

var _ = AfterSuite(func() {
	testAzureClient.deleteResourceGroup(prefix)
})
