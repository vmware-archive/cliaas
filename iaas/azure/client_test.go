package azure_test

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/go-autorest/autorest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/cliaas/iaas/azure"
	"github.com/pivotal-cf/cliaas/iaas/azure/azurefakes"
	errwrap "github.com/pkg/errors"
)

var _ = Describe("Azure", func() {
	Describe("Client", func() {
		Describe("Delete()", func() {
			var azureClient *azure.Client
			var err error
			var identifier string
			var fakeVirtualMachinesClient *azurefakes.FakeComputeVirtualMachinesClient

			JustBeforeEach(func() {
				azureClient.VirtualMachinesClient = fakeVirtualMachinesClient
				err = azureClient.Delete(identifier)
			})

			Context("when given an identifier with a single match of VM name on our regex", func() {
				controlRegex := "ops*"
				BeforeEach(func() {
					fakeVirtualMachinesClient = new(azurefakes.FakeComputeVirtualMachinesClient)
					controlValue := make([]compute.VirtualMachine, 0)
					controlName := "ops-manager"
					controlValue = append(controlValue, compute.VirtualMachine{
						Name: &controlName,
					})
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &controlValue}, nil)
					fakeVirtualMachinesClient.DeallocateReturns(autorest.Response{}, nil)
					azureClient = new(azure.Client)
					identifier = controlRegex
				})
				It("should delete the VM instance", func() {
					Expect(err).ShouldNot(HaveOccurred())
					Expect(fakeVirtualMachinesClient.DeallocateCallCount()).Should(Equal(1))
					_, vmName, _ := fakeVirtualMachinesClient.DeallocateArgsForCall(0)
					Expect(vmName).Should(MatchRegexp(controlRegex))
					var deallocateErr error
					fakeVirtualMachinesClient.DeallocateReturnsOnCall(1, autorest.Response{}, deallocateErr)
					Expect(deallocateErr).ShouldNot(HaveOccurred())
				})
			})

			Context("when unable to list (failed api call) existing VMs to match against", func() {
				controlErr := errors.New("random list err")
				BeforeEach(func() {
					fakeVirtualMachinesClient = new(azurefakes.FakeComputeVirtualMachinesClient)
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{}, controlErr)
					azureClient = new(azure.Client)
					identifier = "ops-manager"
				})
				It("should not delete any VM instances and should exit unsuccessfully", func() {
					Expect(fakeVirtualMachinesClient.DeallocateCallCount()).Should(Equal(0), "the number of times deallocate gets called should be zero")
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(controlErr))
				})
			})

			Context("when given an identifier and no VMs are found in Azure (vm empty set)", func() {
				BeforeEach(func() {
					fakeVirtualMachinesClient = new(azurefakes.FakeComputeVirtualMachinesClient)
					controlValue := make([]compute.VirtualMachine, 0)
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &controlValue}, nil)
					azureClient = new(azure.Client)
					identifier = "ops-manager"
				})
				It("should not delete any VM instances and should exit unsuccessfully", func() {
					Expect(fakeVirtualMachinesClient.DeallocateCallCount()).Should(Equal(0), "the number of times deallocate gets called should be zero")
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(azure.NoMatchesErr))
				})
			})

			Context("when given an identifier with a populated VMs list from azure and no matching VM name regex", func() {
				BeforeEach(func() {
					fakeVirtualMachinesClient = new(azurefakes.FakeComputeVirtualMachinesClient)
					controlValue := make([]compute.VirtualMachine, 0)
					controlName := "blah"
					controlValue = append(controlValue, compute.VirtualMachine{
						Name: &controlName,
					})
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &controlValue}, nil)
					azureClient = new(azure.Client)
					identifier = "ops-manager"
				})
				It("should not delete any VM instances and should exit unsuccessfully", func() {
					Expect(fakeVirtualMachinesClient.DeallocateCallCount()).Should(Equal(0), "the number of times deallocate gets called should be zero")
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(azure.NoMatchesErr))
				})
			})

			Context("when given an identifier with multiple matches on VM name from our regex", func() {
				BeforeEach(func() {
					fakeVirtualMachinesClient = new(azurefakes.FakeComputeVirtualMachinesClient)
					controlValue := make([]compute.VirtualMachine, 0)
					controlName := "ops-manager"
					controlValue = append(controlValue, compute.VirtualMachine{
						Name: &controlName,
					})
					controlValue = append(controlValue, compute.VirtualMachine{
						Name: &controlName,
					})
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &controlValue}, nil)
					azureClient = new(azure.Client)
					identifier = "ops*"
				})
				It("should not delete any VM instances and should exit unsuccessfully", func() {
					Expect(fakeVirtualMachinesClient.DeallocateCallCount()).Should(Equal(0), "the number of times deallocate gets called should be zero")
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(azure.MultipleMatchesErr))
				})
			})
		})
	})

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

		Context("when provided a valid set of configuration values", func() {
			BeforeEach(func() {
				subID = "asdf"
				clientID = "asdf"
				clientSecret = "asdf"
				tenantID = "asdf"
				resourceGroupName = "asdf"
				resourceManagerEndpoint = ""
			})
			It("should return a azure client", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(azureClient).ShouldNot(BeNil())
			})
		})

		Context("when provided a invalid set of configuration values", func() {
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
