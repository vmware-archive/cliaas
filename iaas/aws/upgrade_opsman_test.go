package aws_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/c0-ops/cliaas/iaas/aws"
	"github.com/c0-ops/cliaas/iaas/aws/awsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	iaasaws "github.com/aws/aws-sdk-go/aws"
)

var _ = Describe("UpgradeOpsMan", func() {
	var (
		fakeClient *awsfakes.FakeClient
		upgrade    *UpgradeOpsMan
	)

	BeforeEach(func() {
		fakeClient = new(awsfakes.FakeClient)
		upgrade = NewUpgradeOpsMan(fakeClient)
	})

	Describe("Upgrade", func() {
		var (
			controlInstance *ec2.Instance
			newInstance     *ec2.Instance
		)

		BeforeEach(func() {
			controlInstance = &ec2.Instance{
				InstanceId:   iaasaws.String("some-instance-id"),
				ImageId:      iaasaws.String("some-ami"),
				InstanceType: iaasaws.String("some-instance-type"),
				KeyName:      iaasaws.String("some-key-name"),
				SubnetId:     iaasaws.String("some-subnet-id"),
				SecurityGroups: []*ec2.GroupIdentifier{
					&ec2.GroupIdentifier{
						GroupId: iaasaws.String("some-security-group-id"),
					},
				},
			}

			newInstance = &ec2.Instance{InstanceId: iaasaws.String("bar")}

			// These are to test the order of the calls
			fakeClient.GetVMInfoStub = func(name string) (*ec2.Instance, error) {
				Expect(fakeClient.StopVMCallCount()).To(BeZero())
				Expect(fakeClient.CreateVMCallCount()).To(BeZero())
				Expect(fakeClient.WaitForStartedVMCallCount()).To(BeZero())
				Expect(fakeClient.AssignPublicIPCallCount()).To(BeZero())
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
				return controlInstance, nil
			}

			fakeClient.StopVMStub = func(instance ec2.Instance) error {
				Expect(fakeClient.CreateVMCallCount()).To(BeZero())
				Expect(fakeClient.WaitForStartedVMCallCount()).To(BeZero())
				Expect(fakeClient.AssignPublicIPCallCount()).To(BeZero())
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
				return nil
			}

			fakeClient.CreateVMStub = func(name string, ami string, instanceType string, keyName string, subnetID string, securityGroupID string) (*ec2.Instance, error) {
				Expect(fakeClient.WaitForStartedVMCallCount()).To(BeZero())
				Expect(fakeClient.AssignPublicIPCallCount()).To(BeZero())
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
				return newInstance, nil
			}

			fakeClient.WaitForStartedVMStub = func(instanceName string) error {
				Expect(fakeClient.AssignPublicIPCallCount()).To(BeZero())
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
				return nil
			}

			fakeClient.AssignPublicIPStub = func(instance ec2.Instance, ip string) error {
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
				return nil
			}
		})

		It("tries to upgrade the Ops Mgr VM", func() {
			err := upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeClient.GetVMInfoCallCount()).To(Equal(1))
			vmName := fakeClient.GetVMInfoArgsForCall(0)
			Expect(vmName).To(Equal("some-name*"))

			Expect(fakeClient.StopVMCallCount()).To(Equal(1))
			stopVMInstance := fakeClient.StopVMArgsForCall(0)
			Expect(stopVMInstance).To(Equal(*controlInstance))

			Expect(fakeClient.CreateVMCallCount()).To(Equal(1))
			createAMI, createVMType, createName, createKeyName, createSubnetID, createSecurityGroupID := fakeClient.CreateVMArgsForCall(0)
			Expect(createAMI).To(Equal("some-ami"))
			Expect(createVMType).To(Equal("some-instance-type"))
			Expect(createName).To(ContainSubstring("some-name - "))
			Expect(createKeyName).To(Equal("some-key-name"))
			Expect(createSubnetID).To(Equal("some-subnet-id"))
			Expect(createSecurityGroupID).To(Equal("some-security-group-id"))

			Expect(fakeClient.WaitForStartedVMCallCount()).To(Equal(1))
			waitVMName := fakeClient.WaitForStartedVMArgsForCall(0)
			Expect(waitVMName).To(Equal(createName))

			Expect(fakeClient.AssignPublicIPCallCount()).To(Equal(1))
			assignIPInstance, assignIP := fakeClient.AssignPublicIPArgsForCall(0)
			Expect(assignIPInstance).To(Equal(*newInstance))
			Expect(assignIP).To(Equal("some-ip"))

			Expect(fakeClient.DeleteVMCallCount()).To(Equal(1))
			deleteInstanceID := fakeClient.DeleteVMArgsForCall(0)
			Expect(deleteInstanceID).To(Equal(*controlInstance.InstanceId))
		})

		Context("when GetVMInfo fails", func() {
			BeforeEach(func() {
				fakeClient.GetVMInfoReturns(nil, errors.New("an error"))
			})

			It("returns an error", func() {
				err := upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(err).To(HaveOccurred())
			})

			It("does not try to stop the vm", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.StopVMCallCount()).To(BeZero())
			})

			It("does not try to create a new vm", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.CreateVMCallCount()).To(BeZero())
			})

			It("does not assign any public IPs", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.AssignPublicIPCallCount()).To(BeZero())
			})

			It("does not delete any VMs", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
			})
		})

		Context("when stopping the old VM fails", func() {
			BeforeEach(func() {
				fakeClient.StopVMReturns(errors.New("an error"))
			})

			It("returns an error", func() {
				err := upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(err).To(HaveOccurred())
			})

			It("does not try to create a new vm", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.CreateVMCallCount()).To(BeZero())
			})

			It("does not assign any public IPs", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.AssignPublicIPCallCount()).To(BeZero())
			})

			It("does not delete any VMs", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
			})
		})

		Context("when creating the new VM fails", func() {
			BeforeEach(func() {
				fakeClient.CreateVMReturns(nil, errors.New("an error"))
			})

			It("returns an error", func() {
				err := upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(err).To(HaveOccurred())
			})

			It("does not assign any public IPs", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.AssignPublicIPCallCount()).To(BeZero())
			})

			It("does not delete any VMs", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
			})

			XIt("starts the old VM again", func() {
			})
		})

		Context("when waiting for the new VM to come up fails", func() {
			BeforeEach(func() {
				fakeClient.WaitForStartedVMReturns(errors.New("an error"))
			})

			It("returns an error", func() {
				err := upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(err).To(HaveOccurred())
			})

			It("does not assign any public IPs", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.AssignPublicIPCallCount()).To(BeZero())
			})

			It("does not delete any VMs", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
			})

			XIt("starts the old VM again", func() {
			})
		})

		Context("when assigning the old IP to the new VM fails", func() {
			BeforeEach(func() {
				fakeClient.AssignPublicIPReturns(errors.New("an error"))
			})

			It("returns an error", func() {
				err := upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(err).To(HaveOccurred())
			})

			It("does not delete any VMs", func() {
				upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(fakeClient.DeleteVMCallCount()).To(BeZero())
			})

			XIt("starts the old VM again", func() {
			})
		})

		Context("when deleting the old VM fails", func() {
			BeforeEach(func() {
				fakeClient.AssignPublicIPReturns(errors.New("an error"))
			})

			It("returns an error", func() {
				err := upgrade.Upgrade("some-name", "some-ami", "some-instance-type", "some-ip")
				Expect(err).To(HaveOccurred())
			})

			XIt("starts the old VM again", func() {
			})
		})
	})
})
