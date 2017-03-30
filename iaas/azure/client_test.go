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
		Describe("Replace()", func() {
			var azureClient *azure.Client
			var err error
			var identifier string
			var fakeVirtualMachinesClient *azurefakes.FakeComputeVirtualMachinesClient
			var controlNewImageURL = "some-control-new-image-url"
			var controlRegex = "ops*"
			var controlValue []compute.VirtualMachine
			var controlID = "some-id"
			var controlOldImageURL = "some-image-url"
			var controlOldName = "ops-manager"

			JustBeforeEach(func() {
				fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &controlValue}, nil)
				fakeVirtualMachinesClient.DeallocateReturns(autorest.Response{}, nil)
				azureClient = new(azure.Client)
				identifier = controlRegex
				azureClient.VirtualMachinesClient = fakeVirtualMachinesClient
				err = azureClient.Replace(identifier, controlNewImageURL)
			})

			BeforeEach(func() {
				controlValue = make([]compute.VirtualMachine, 0)
			})

			Context("when there is a single match on a identifier regex", func() {
				controlNewNameRegex := controlOldName + "_....*"
				BeforeEach(func() {
					fakeVirtualMachinesClient = new(azurefakes.FakeComputeVirtualMachinesClient)
					vm := newVirtualMachine(controlID, controlOldName, controlOldImageURL)
					controlValue = append(controlValue, vm)
				})
				It("it should not return an error", func() {
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("it should spin down the matching vm instance", func() {
					Expect(fakeVirtualMachinesClient.DeallocateCallCount()).Should(Equal(1), "we should call deallocate exactly once")
					_, vmName, _ := fakeVirtualMachinesClient.DeallocateArgsForCall(0)
					Expect(vmName).Should(MatchRegexp(controlRegex))
					var deallocateErr error
					fakeVirtualMachinesClient.DeallocateReturnsOnCall(1, autorest.Response{}, deallocateErr)
					Expect(deallocateErr).ShouldNot(HaveOccurred())
				})
				It("it should copy the existing vms config into the new vm instance's config ", func() {
					Expect(fakeVirtualMachinesClient.CreateOrUpdateCallCount()).Should(Equal(1), "we should call createorupdate exactly once")
					_, _, parameters, _ := fakeVirtualMachinesClient.CreateOrUpdateArgsForCall(0)
					Expect(*parameters.ID).Should(Equal(controlID))
				})
				It("it should replace the disk image on the new vm instance's config with the given new version", func() {
					Expect(fakeVirtualMachinesClient.CreateOrUpdateCallCount()).Should(Equal(1), "we should call createorupdate exactly once")
					_, _, parameters, _ := fakeVirtualMachinesClient.CreateOrUpdateArgsForCall(0)
					var imageURL = *parameters.VirtualMachineProperties.StorageProfile.OsDisk.Image.URI
					Expect(imageURL).ShouldNot(Equal(controlOldImageURL))
					Expect(imageURL).Should(Equal(controlNewImageURL))
				})
				It("it should apply a new unique name to the new vm instance's config", func() {
					Expect(fakeVirtualMachinesClient.CreateOrUpdateCallCount()).Should(Equal(1), "we should call createorupdate exactly once")
					_, _, parameters, _ := fakeVirtualMachinesClient.CreateOrUpdateArgsForCall(0)
					var name = *parameters.Name
					Expect(name).ShouldNot(Equal(controlOldName))
					Expect(name).Should(MatchRegexp(controlNewNameRegex))
				})
			})

			Context("when there are no matches for the identifier regex", func() {
				BeforeEach(func() {
					fakeVirtualMachinesClient = new(azurefakes.FakeComputeVirtualMachinesClient)
				})
				It("should not try to deallocate anything and exit in error", func() {
					Expect(fakeVirtualMachinesClient.DeallocateCallCount()).Should(Equal(0), "we should never call deallocate without a matching VM")
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(azure.NoMatchesErr))
				})
			})

			Context("when there are multiple matches for the identifier regex", func() {
				BeforeEach(func() {
					fakeVirtualMachinesClient = new(azurefakes.FakeComputeVirtualMachinesClient)
					vm := newVirtualMachine(controlID, controlOldName, controlOldImageURL)
					controlValue = append(controlValue, vm, vm)
				})

				It("should not try to deallocate anything and exit in error", func() {
					Expect(fakeVirtualMachinesClient.DeallocateCallCount()).Should(Equal(0), "we should never call deallocate without a matching VM")
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(azure.MultipleMatchesErr))
				})
			})
		})

		Describe("Delete()", func() {
			var azureClient *azure.Client
			var err error
			var identifier string
			var fakeVirtualMachinesClient *azurefakes.FakeComputeVirtualMachinesClient
			var controlValue []compute.VirtualMachine

			JustBeforeEach(func() {
				azureClient.VirtualMachinesClient = fakeVirtualMachinesClient
				err = azureClient.Delete(identifier)
			})

			BeforeEach(func() {
				fakeVirtualMachinesClient = new(azurefakes.FakeComputeVirtualMachinesClient)
				controlValue = make([]compute.VirtualMachine, 0)
				azureClient = new(azure.Client)
			})

			Context("when azure running VMs list returns more than a single page of results", func() {
				BeforeEach(func() {
					identifier = "testid"
					vmMatch := newVirtualMachine(identifier, identifier, "testurl")
					vmNothing := newVirtualMachine("nomatch", "nomatch", "testurl")
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &[]compute.VirtualMachine{vmNothing}}, nil)
					fakeVirtualMachinesClient.ListAllNextResultsReturnsOnCall(
						0,
						compute.VirtualMachineListResult{
							Value: &[]compute.VirtualMachine{vmMatch},
						},
						nil,
					)
					fakeVirtualMachinesClient.ListAllNextResultsReturnsOnCall(1, compute.VirtualMachineListResult{}, nil)
				})
				It("then we should properly walk through all pages to apply our regex", func() {
					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			Context("when given an identifier with a single match of VM name on our regex", func() {
				controlRegex := "ops*"
				BeforeEach(func() {
					controlName := "ops-manager"
					controlValue = append(controlValue, compute.VirtualMachine{
						Name: &controlName,
					})
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &controlValue}, nil)
					fakeVirtualMachinesClient.DeleteReturns(autorest.Response{}, nil)
					identifier = controlRegex
				})
				It("should delete the VM instance", func() {
					Expect(err).ShouldNot(HaveOccurred())
					Expect(fakeVirtualMachinesClient.DeleteCallCount()).Should(Equal(1))
					_, vmName, _ := fakeVirtualMachinesClient.DeleteArgsForCall(0)
					Expect(vmName).Should(MatchRegexp(controlRegex))
					var deleteErr error
					fakeVirtualMachinesClient.DeleteReturnsOnCall(1, autorest.Response{}, deleteErr)
					Expect(deleteErr).ShouldNot(HaveOccurred())
				})
			})

			Context("when unable to list (failed api call) existing VMs to match against", func() {
				controlErr := errors.New("random list err")
				BeforeEach(func() {
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{}, controlErr)
					identifier = "ops-manager"
				})
				It("should not delete any VM instances and should exit unsuccessfully", func() {
					Expect(fakeVirtualMachinesClient.DeleteCallCount()).Should(Equal(0), "the number of times deletes gets called should be zero")
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(controlErr))
				})
			})

			Context("when given an identifier and no VMs are found in Azure (vm empty set)", func() {
				BeforeEach(func() {
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &controlValue}, nil)
					identifier = "ops-manager"
				})
				It("should not delete any VM instances and should exit unsuccessfully", func() {
					Expect(fakeVirtualMachinesClient.DeleteCallCount()).Should(Equal(0), "the number of times deletes gets called should be zero")
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(azure.NoMatchesErr))
				})
			})

			Context("when given an identifier with a populated VMs list from azure and no matching VM name regex", func() {
				BeforeEach(func() {
					controlName := "some-name"
					controlValue = append(controlValue,
						compute.VirtualMachine{Name: &controlName})
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &controlValue}, nil)
					identifier = "ops-manager"
				})
				It("should not delete any VM instances and should exit unsuccessfully", func() {
					Expect(fakeVirtualMachinesClient.DeleteCallCount()).Should(Equal(0), "the number of times deletes gets called should be zero")
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(azure.NoMatchesErr))
				})
			})

			Context("when given an identifier with multiple matches on VM name from our regex", func() {
				BeforeEach(func() {
					controlName := "ops-manager"
					controlValue = append(controlValue,
						compute.VirtualMachine{Name: &controlName},
						compute.VirtualMachine{Name: &controlName})
					fakeVirtualMachinesClient.ListReturns(compute.VirtualMachineListResult{Value: &controlValue}, nil)
					identifier = "ops*"
				})
				It("should not delete any VM instances and should exit unsuccessfully", func() {
					Expect(fakeVirtualMachinesClient.DeleteCallCount()).Should(Equal(0), "the number of times deletes gets called should be zero")
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
				subID = "some-sub-id"
				clientID = "some-client-id"
				clientSecret = "some-client-secret"
				tenantID = "some-tenant-id"
				resourceGroupName = "some-resource-group-name"
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

func newVirtualMachine(id string, name string, vmDiskURL string) compute.VirtualMachine {
	tmpID := id
	tmpName := name
	tmpURL := vmDiskURL
	vm := compute.VirtualMachine{
		ID:   &tmpID,
		Name: &tmpName,
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			StorageProfile: &compute.StorageProfile{
				OsDisk: &compute.OSDisk{
					Image: &compute.VirtualHardDisk{
						URI: &tmpURL,
					},
				},
			},
		},
	}
	return vm
}
