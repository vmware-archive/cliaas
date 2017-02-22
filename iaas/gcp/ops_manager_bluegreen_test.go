package gcp_test

import (
	"fmt"

	"github.com/c0-ops/cliaas/iaas"
	. "github.com/c0-ops/cliaas/iaas/gcp"
	"github.com/c0-ops/cliaas/iaas/gcp/gcpfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	compute "google.golang.org/api/compute/v1"
)

var _ = Describe("OpsManager struct and a valid client", func() {
	var opsManager *OpsManagerGCP
	var (
		controlFilter = iaas.Filter{
			TagRegexString:  "ops",
			NameRegexString: "ops-manager",
		}
		controlDiskImageURL        = "some/good/version.img"
		fakeClient                 *gcpfakes.FakeClientAPI
		controlGetVMInfoInstance   = createFakeInstance(InstanceStatusStopped, controlDiskImageURL)
		controlStartVMInfoInstance = createFakeInstance(InstanceStatusRunning, controlDiskImageURL)
		controlDeployInstance      = createFakeInstance(InstanceStatusRunning, controlDiskImageURL)
	)

	Context("when calling SpinDown() on running vms", func() {
		var vmInstance *compute.Instance
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
			vmInstance, err = opsManager.SpinDown(controlFilter)
			Expect(err).ToNot(HaveOccurred())
			close(done)
		}, 5)

		It("should spin down the existing ops manager", func() {
			Expect(fakeClient.GetVMInfoCallCount()).Should(BeNumerically(">", 1), "we should call getVM a few times")
			Expect(fakeClient.GetVMInfoArgsForCall(0)).Should(Equal(controlFilter), "the getvm calls should use the correct filter for the running ops manager")
			Expect(fakeClient.StopVMCallCount()).Should(Equal(1), "this should only ever be called once")
			Expect(fakeClient.StopVMArgsForCall(0)).Should(Equal(controlGetVMInfoInstance.Name), "the name of the found running instance should be used for the stop call")
			Expect(vmInstance.Status).Should(Equal("STOPPED"))
		})

		Context("when polling for proper SpinDown status hits timeout ", func() {

			BeforeEach(func() {
				var err error
				opsManager, err = NewOpsManager(
					ConfigClient(fakeClient),
					ConfigClientTimeoutSeconds(1),
				)
				Expect(err).ToNot(HaveOccurred())
				fakeClient.GetVMInfoReturns(&controlStartVMInfoInstance, fmt.Errorf("I FAILED"))
			})
			It("then it should timeout and give a error", func(done Done) {
				vmInstance, err := opsManager.SpinDown(controlFilter)
				Expect(err).Should(HaveOccurred())
				Expect(vmInstance).Should(BeNil())
				close(done)
			}, 5)
		})
	})
	Context("when calling Deploy()", func() {
		BeforeEach(func(done Done) {
			fakeClient = new(gcpfakes.FakeClientAPI)
			var err error
			opsManager, err = NewOpsManager(
				ConfigClient(fakeClient),
				ConfigClientTimeoutSeconds(1),
			)
			Expect(err).ToNot(HaveOccurred())
			fakeClient.GetVMInfoReturns(&controlDeployInstance, nil)
			fakeClient.StopVMReturns(nil)
			err = opsManager.Deploy(&controlDeployInstance)
			Expect(err).ToNot(HaveOccurred())
			close(done)
		}, 5)
		It("should spin up a new ops manager successfully", func() {
			Expect(fakeClient.CreateVMCallCount()).Should(Equal(1), "we should call createVM once")
			instance := fakeClient.CreateVMArgsForCall(0)
			Expect(instance.Name).Should(Equal(controlDeployInstance.Name))
			Expect(instance.Disks).Should(HaveLen(1))
			Expect(instance.Disks[0].Source).Should(Equal(controlDiskImageURL))
		})

		Context("when polling for proper RUNNING status hits timeout ", func() {
			failingInstance := controlGetVMInfoInstance

			BeforeEach(func() {
				var err error
				opsManager, err = NewOpsManager(
					ConfigClient(fakeClient),
					ConfigClientTimeoutSeconds(1),
				)
				Expect(err).ToNot(HaveOccurred())
				failingInstance.Status = "NOT_RUNNING"
				fakeClient.GetVMInfoReturns(&failingInstance, fmt.Errorf("I FAILED"))
			})
			It("then it should timeout and give a error", func(done Done) {
				err := opsManager.Deploy(&failingInstance)
				Expect(err).Should(HaveOccurred())
				close(done)
			}, 5)
		})
	})
	Context("when calling CleanUp on venerable VM", func() {
		var controlCleanUpFilter iaas.Filter
		BeforeEach(func(done Done) {
			fakeClient = new(gcpfakes.FakeClientAPI)
			var err error
			controlCleanUpFilter = iaas.Filter{
				NameRegexString: controlGetVMInfoInstance.Name,
				Status:          InstanceStatusStopped,
			}
			opsManager, err = NewOpsManager(
				ConfigClient(fakeClient),
				ConfigClientTimeoutSeconds(1),
			)
			Expect(err).ToNot(HaveOccurred())
			fakeClient.DeleteVMReturns(nil)
			fakeClient.GetVMInfoReturns(&controlGetVMInfoInstance, nil)
			err = opsManager.CleanUp(controlCleanUpFilter)
			Expect(err).ToNot(HaveOccurred())
			close(done)
		}, 5)
		It("should destroy the old ops manager", func() {
			Expect(fakeClient.DeleteVMCallCount()).Should(Equal(1), "we should call deleteVM once")
			instanceName := fakeClient.DeleteVMArgsForCall(0)
			Expect(instanceName).To(Equal(controlCleanUpFilter.NameRegexString))
		})
	})
})

func createFakeInstance(status string, imageURL string) compute.Instance {
	return compute.Instance{
		Name: "ops-manager",
		Tags: &compute.Tags{
			Items: []string{
				"ops-manager",
			},
		},
		Disks: []*compute.AttachedDisk{
			&compute.AttachedDisk{
				Source: imageURL,
			},
		},
		Status: status,
	}
}
