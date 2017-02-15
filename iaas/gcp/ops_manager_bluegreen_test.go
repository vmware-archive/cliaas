package gcp_test

import (
	"fmt"

	compute "google.golang.org/api/compute/v1"

	. "github.com/c0-ops/cliaas/iaas/gcp"
	"github.com/c0-ops/cliaas/iaas/gcp/gcpfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OpsManager struct and a valid client", func() {
	var opsManager *OpsManagerGCP
	var (
		controlFilter = Filter{
			TagRegexString:  "ops",
			NameRegexString: "ops-manager",
		}
		controlDiskImageURL      = "some/good/version.img"
		fakeClient               *gcpfakes.FakeClientAPI
		controlGetVMInfoInstance = compute.Instance{
			Name: "ops-manager",
			Tags: &compute.Tags{
				Items: []string{
					"ops-manager",
				},
			},
			Status: "STOPPED",
		}
		controlStartVMInfoInstance = compute.Instance{
			Name: "ops-manager",
			Tags: &compute.Tags{
				Items: []string{
					"ops-manager",
				},
			},
			Status: "RUNNING",
		}
	)

	Context("when attempting a RunBlueGreen() with valid arguments and a running ops manager", func() {
		BeforeEach(func(done Done) {
			fakeClient = new(gcpfakes.FakeClientAPI)
			var err error
			opsManager, err = NewOpsManager(
				ConfigClient(fakeClient),
				ConfigClientTimeoutSeconds(1),
			)
			Expect(err).ToNot(HaveOccurred())
			fakeClient.GetVMInfoReturns(&controlGetVMInfoInstance, nil)
			fakeClient.StopVMReturns(nil)
			err = opsManager.RunBlueGreen(controlFilter, controlDiskImageURL)
			Expect(err).ToNot(HaveOccurred())
			close(done)
		}, 5)

		It("should spin down the existing ops manager", func() {
			Expect(fakeClient.GetVMInfoCallCount()).Should(BeNumerically(">", 1), "we should call getVM a few times")
			Expect(fakeClient.GetVMInfoArgsForCall(0)).Should(Equal(controlFilter), "the getvm calls should use the correct filter for the running ops manager")
			Expect(fakeClient.StopVMCallCount()).Should(Equal(1), "this should only ever be called once")
			Expect(fakeClient.StopVMArgsForCall(0)).Should(Equal(controlGetVMInfoInstance.Name), "the name of the found running instance should be used for the stop call")
		})

		It("should spin up a new ops manager successfully", func() {
			Expect(fakeClient.CreateVMCallCount()).Should(Equal(1), "we should call createVM once")
			instance := fakeClient.CreateVMArgsForCall(0)
			Expect(instance.Name).Should(Equal(controlGetVMInfoInstance.Name))
			Expect(instance.Disks).Should(HaveLen(1))
			Expect(instance.Disks[0].Source).Should(Equal(controlDiskImageURL))
		})

		XIt("should destroy the old ops manager", func() {
			Expect(true).To(BeFalse())
		})
	})

	Context("when attempting a RunBlueGreen() with invalid arguments", func() {

		XIt("should fail to retrieve ops manager VM", func() {
			Expect(true).To(BeFalse())
		})

	})

	Context("when stopping a vm and the vm state never reaches a stopped status", func() {

		BeforeEach(func() {
			var err error
			opsManager, err = NewOpsManager(
				ConfigClient(fakeClient),
				ConfigClientTimeoutSeconds(1),
			)
			Expect(err).ToNot(HaveOccurred())
			fakeClient.GetVMInfoReturns(&controlStartVMInfoInstance, fmt.Errorf("I FAILED"))
		})
		It("then it should eventually timeout and give a reasonable error", func(done Done) {
			err := opsManager.RunBlueGreen(controlFilter, controlDiskImageURL)
			Expect(err).Should(HaveOccurred())
			close(done)
		}, 5)
	})

	Context("when attempting a RunBlueGreen() with valid arguments and an ops manager failing to stop", func() {

		XIt("should not retrieve a STOPPED state from the ops manager VM", func() {
			Expect(true).To(BeFalse())
		})
	})

})
