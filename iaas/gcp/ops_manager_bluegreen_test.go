package gcp_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cliaas/iaas"
	. "github.com/pivotal-cf/cliaas/iaas/gcp"
	"github.com/pivotal-cf/cliaas/iaas/gcp/gcpfakes"
	compute "google.golang.org/api/compute/v1"
)

var _ = Describe("OpsManager struct and a valid client", func() {
	var (
		opsManager *OpsManagerGCP
		fakeClient *gcpfakes.FakeClientAPI

		controlFilter              iaas.Filter
		controlDiskImageURL        string
		controlGetVMInfoInstance   compute.Instance
		controlStartVMInfoInstance compute.Instance
		controlDeployInstance      compute.Instance
	)

	BeforeEach(func() {
		fakeClient = new(gcpfakes.FakeClientAPI)
		controlFilter = iaas.Filter{
			TagRegexString:  "ops",
			NameRegexString: "ops-manager",
		}
		controlDiskImageURL = "some/good/version.img"
		controlGetVMInfoInstance = createFakeInstance(InstanceStatusStopped, controlDiskImageURL)
		controlStartVMInfoInstance = createFakeInstance(InstanceStatusRunning, controlDiskImageURL)
		controlDeployInstance = createFakeInstance(InstanceStatusRunning, controlDiskImageURL)

		var err error
		opsManager, err = NewOpsManager(
			ConfigClient(fakeClient),
			ConfigClientTimeoutSeconds(1),
		)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when calling SpinDown() on running vms", func() {
		BeforeEach(func() {
			fakeClient.GetVMInfoReturns(&controlGetVMInfoInstance, nil)
		})

		It("should spin down the existing ops manager", func() {
			vmInstance, err := opsManager.SpinDown(controlFilter)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeClient.GetVMInfoCallCount()).Should(BeNumerically(">", 1), "we should call getVM a few times")
			Expect(fakeClient.GetVMInfoArgsForCall(0)).Should(Equal(controlFilter), "the getvm calls should use the correct filter for the running ops manager")
			Expect(fakeClient.StopVMCallCount()).Should(Equal(1), "this should only ever be called once")
			Expect(fakeClient.StopVMArgsForCall(0)).Should(Equal(controlGetVMInfoInstance.Name), "the name of the found running instance should be used for the stop call")
			Expect(vmInstance.Status).Should(Equal("STOPPED"))
		})

		Context("when polling for proper SpinDown status hits timeout ", func() {
			BeforeEach(func() {
				fakeClient.GetVMInfoReturns(nil, errors.New("an error"))
			})

			It("returns an error", func() {
				vmInstance, err := opsManager.SpinDown(controlFilter)
				Expect(err).Should(HaveOccurred())
				Expect(vmInstance).Should(BeNil())
			})
		})
	})

	Context("when calling Deploy()", func() {
		BeforeEach(func() {
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
		})

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
			It("then it should timeout and give a error", func() {
				err := opsManager.Deploy(&failingInstance)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Context("when calling CleanUp on venerable VM", func() {
		var controlCleanUpFilter iaas.Filter

		BeforeEach(func() {
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
		})
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
