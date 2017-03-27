package azure_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/cliaas/iaas/azure"
)

var _ = Describe("Azure", func() {
	Describe("NewClient", func() {
		var azureClient *azure.Client
		var err error
		var subID string
		var clientID string
		var clientSecret string
		var tenantID string
		var resourceGroupName string
		var resourceManagerEndpoint string

		JustBeforeEach(func() {
			azureClient, err = azure.NewClient(
				subID,
				clientID,
				clientSecret,
				tenantID,
				resourceGroupName,
				resourceManagerEndpoint,
			)
		})

		Context("when provided a set of configuration values which allows for calling the Azure API", func() {
			BeforeEach(func() {
				subID = "asdf"
				clientID = "asdf"
				clientSecret = "asdf"
				tenantID = "asdf"
				resourceGroupName = "asdf"
				resourceManagerEndpoint = ""
			})
			It("should return a azure client which is capable of calling the Azure API", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(azureClient).ShouldNot(BeNil())
			})
		})

		Context("when provided a set of configuration values which does not allow for calling the Azure API", func() {
			BeforeEach(func() {
				subID = ""
				clientID = ""
				clientSecret = ""
				tenantID = ""
				resourceGroupName = ""
			})
			It("should return an error that the client was not able to be created", func() {
				Expect(err).Should(HaveOccurred())
				Expect(azureClient).Should(BeNil())
			})
		})
	})
})
