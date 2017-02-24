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
	Describe("UpgradeOpsMan", func() {
		var upgrade *UpgradeOpsMan
		var err error

		Describe("given a Upgrade method", func() {
			Context("when called", func() {
				var fakeClient *awsfakes.FakeClientAPI
				controlInstance := &ec2.Instance{InstanceId: iaasaws.String("foo")}
				newInstance := &ec2.Instance{InstanceId: iaasaws.String("bar")}
				BeforeEach(func() {
					fakeClient = new(awsfakes.FakeClientAPI)
					upgrade, err = NewUpgradeOpsMan(ConfigClient(fakeClient))
					Expect(err).ShouldNot(HaveOccurred())
					Expect(upgrade).ShouldNot(BeNil())
				})

				It("then should complete with no errors", func() {

					fakeClient.GetVMInfoReturns(controlInstance, nil)
					fakeClient.CreateVMReturns(newInstance, nil)

					name := "originalVM"
					ami := "testAMI"
					instanceType := "testType"
					ip := "1.1.1.1"
					err := upgrade.Upgrade(name, ami, instanceType, ip)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(fakeClient.GetVMInfoCallCount()).To(BeEquivalentTo(1))
					vmName := fakeClient.GetVMInfoArgsForCall(0)
					Expect(vmName).Should(BeEquivalentTo("originalVM*"))

					Expect(fakeClient.StopVMCallCount()).To(BeEquivalentTo(1))
					stopVMInstance := fakeClient.StopVMArgsForCall(0)
					Expect(stopVMInstance).Should(BeEquivalentTo(*controlInstance))

					Expect(fakeClient.CreateVMCallCount()).To(BeEquivalentTo(1))
					createVMInstance, createAMI, createInstanceType, createNewName := fakeClient.CreateVMArgsForCall(0)
					Expect(createVMInstance).Should(BeEquivalentTo(*controlInstance))
					Expect(createAMI).Should(BeEquivalentTo(ami))
					Expect(createInstanceType).Should(BeEquivalentTo(instanceType))
					Expect(createNewName).Should(ContainSubstring(name))

					Expect(fakeClient.WaitForStartedVMCallCount()).To(BeEquivalentTo(1))
					waitVMName := fakeClient.WaitForStartedVMArgsForCall(0)
					Expect(waitVMName).Should(BeEquivalentTo(createNewName))

					Expect(fakeClient.AssignPublicIPCallCount()).To(BeEquivalentTo(1))
					assignIPInstance, assignIP := fakeClient.AssignPublicIPArgsForCall(0)
					Expect(assignIPInstance).Should(BeEquivalentTo(*newInstance))
					Expect(assignIP).Should(BeEquivalentTo(ip))

					Expect(fakeClient.DeleteVMCallCount()).To(BeEquivalentTo(1))
					deleteInstance := fakeClient.DeleteVMArgsForCall(0)
					Expect(deleteInstance).Should(BeEquivalentTo(*controlInstance))
				})
				It("then should error on GetVMInfo", func() {
					fakeClient.GetVMInfoReturns(nil, errors.New("got an error"))
					name := "originalVM"
					ami := "testAMI"

					instanceType := "testType"
					ip := "1.1.1.1"
					err := upgrade.Upgrade(name, ami, instanceType, ip)
					Expect(err).Should(HaveOccurred())
					Expect(fakeClient.GetVMInfoCallCount()).To(BeEquivalentTo(1))
					vmName := fakeClient.GetVMInfoArgsForCall(0)
					Expect(vmName).Should(BeEquivalentTo("originalVM*"))
				})

				It("then error on StopVM", func() {

					fakeClient.GetVMInfoReturns(controlInstance, nil)
					fakeClient.StopVMReturns(errors.New("got an error"))

					name := "originalVM"
					ami := "testAMI"

					instanceType := "testType"
					ip := "1.1.1.1"
					err := upgrade.Upgrade(name, ami, instanceType, ip)
					Expect(err).Should(HaveOccurred())
					Expect(fakeClient.GetVMInfoCallCount()).To(BeEquivalentTo(1))
					vmName := fakeClient.GetVMInfoArgsForCall(0)
					Expect(vmName).Should(BeEquivalentTo("originalVM*"))

					Expect(fakeClient.StopVMCallCount()).To(BeEquivalentTo(1))
					stopVMInstance := fakeClient.StopVMArgsForCall(0)
					Expect(stopVMInstance).Should(BeEquivalentTo(*controlInstance))

					Expect(fakeClient.CreateVMCallCount()).To(BeEquivalentTo(0))
					Expect(fakeClient.WaitForStartedVMCallCount()).To(BeEquivalentTo(0))
					Expect(fakeClient.AssignPublicIPCallCount()).To(BeEquivalentTo(0))
					Expect(fakeClient.DeleteVMCallCount()).To(BeEquivalentTo(0))
				})

				It("then should error on create VM", func() {
					fakeClient.GetVMInfoReturns(controlInstance, nil)
					fakeClient.CreateVMReturns(nil, errors.New("got an error"))

					name := "originalVM"
					ami := "testAMI"

					instanceType := "testType"
					ip := "1.1.1.1"
					err := upgrade.Upgrade(name, ami, instanceType, ip)
					Expect(err).Should(HaveOccurred())
					Expect(fakeClient.GetVMInfoCallCount()).To(BeEquivalentTo(1))
					vmName := fakeClient.GetVMInfoArgsForCall(0)
					Expect(vmName).Should(BeEquivalentTo("originalVM*"))

					Expect(fakeClient.StopVMCallCount()).To(BeEquivalentTo(1))
					stopVMInstance := fakeClient.StopVMArgsForCall(0)
					Expect(stopVMInstance).Should(BeEquivalentTo(*controlInstance))

					Expect(fakeClient.CreateVMCallCount()).To(BeEquivalentTo(1))
					createVMInstance, createAMI, createInstanceType, createNewName := fakeClient.CreateVMArgsForCall(0)
					Expect(createVMInstance).Should(BeEquivalentTo(*controlInstance))
					Expect(createAMI).Should(BeEquivalentTo(ami))
					Expect(createInstanceType).Should(BeEquivalentTo(instanceType))
					Expect(createNewName).Should(ContainSubstring(name))

					Expect(fakeClient.WaitForStartedVMCallCount()).To(BeEquivalentTo(0))
					Expect(fakeClient.AssignPublicIPCallCount()).To(BeEquivalentTo(0))
					Expect(fakeClient.DeleteVMCallCount()).To(BeEquivalentTo(0))
				})

				It("then it should error on wait for vm", func() {
					fakeClient.GetVMInfoReturns(controlInstance, nil)
					fakeClient.CreateVMReturns(newInstance, nil)
					fakeClient.WaitForStartedVMReturns(errors.New("got an error"))

					name := "originalVM"
					ami := "testAMI"

					instanceType := "testType"
					ip := "1.1.1.1"
					err := upgrade.Upgrade(name, ami, instanceType, ip)
					Expect(err).Should(HaveOccurred())
					Expect(fakeClient.GetVMInfoCallCount()).To(BeEquivalentTo(1))
					vmName := fakeClient.GetVMInfoArgsForCall(0)
					Expect(vmName).Should(BeEquivalentTo("originalVM*"))

					Expect(fakeClient.StopVMCallCount()).To(BeEquivalentTo(1))
					stopVMInstance := fakeClient.StopVMArgsForCall(0)
					Expect(stopVMInstance).Should(BeEquivalentTo(*controlInstance))

					Expect(fakeClient.CreateVMCallCount()).To(BeEquivalentTo(1))
					createVMInstance, createAMI, createInstanceType, createNewName := fakeClient.CreateVMArgsForCall(0)
					Expect(createVMInstance).Should(BeEquivalentTo(*controlInstance))
					Expect(createAMI).Should(BeEquivalentTo(ami))
					Expect(createInstanceType).Should(BeEquivalentTo(instanceType))
					Expect(createNewName).Should(ContainSubstring(name))

					Expect(fakeClient.WaitForStartedVMCallCount()).To(BeEquivalentTo(1))
					waitVMName := fakeClient.WaitForStartedVMArgsForCall(0)
					Expect(waitVMName).Should(BeEquivalentTo(createNewName))

					Expect(fakeClient.AssignPublicIPCallCount()).To(BeEquivalentTo(0))
					Expect(fakeClient.DeleteVMCallCount()).To(BeEquivalentTo(0))

				})

				It("then it should error on assign ip", func() {
					fakeClient.GetVMInfoReturns(controlInstance, nil)
					fakeClient.CreateVMReturns(newInstance, nil)
					fakeClient.AssignPublicIPReturns(errors.New("got an error"))

					name := "originalVM"
					ami := "testAMI"

					instanceType := "testType"
					ip := "1.1.1.1"
					err := upgrade.Upgrade(name, ami, instanceType, ip)
					Expect(err).Should(HaveOccurred())
					Expect(fakeClient.GetVMInfoCallCount()).To(BeEquivalentTo(1))
					vmName := fakeClient.GetVMInfoArgsForCall(0)
					Expect(vmName).Should(BeEquivalentTo("originalVM*"))

					Expect(fakeClient.StopVMCallCount()).To(BeEquivalentTo(1))
					stopVMInstance := fakeClient.StopVMArgsForCall(0)
					Expect(stopVMInstance).Should(BeEquivalentTo(*controlInstance))

					Expect(fakeClient.CreateVMCallCount()).To(BeEquivalentTo(1))
					createVMInstance, createAMI, createInstanceType, createNewName := fakeClient.CreateVMArgsForCall(0)
					Expect(createVMInstance).Should(BeEquivalentTo(*controlInstance))
					Expect(createAMI).Should(BeEquivalentTo(ami))
					Expect(createInstanceType).Should(BeEquivalentTo(instanceType))
					Expect(createNewName).Should(ContainSubstring(name))

					Expect(fakeClient.WaitForStartedVMCallCount()).To(BeEquivalentTo(1))
					waitVMName := fakeClient.WaitForStartedVMArgsForCall(0)
					Expect(waitVMName).Should(BeEquivalentTo(createNewName))

					Expect(fakeClient.AssignPublicIPCallCount()).To(BeEquivalentTo(1))
					assignIPInstance, assignIP := fakeClient.AssignPublicIPArgsForCall(0)
					Expect(assignIPInstance).Should(BeEquivalentTo(*newInstance))
					Expect(assignIP).Should(BeEquivalentTo(ip))

					Expect(fakeClient.DeleteVMCallCount()).To(BeEquivalentTo(0))
				})

				It("then should error on DeleteVM", func() {
					fakeClient.GetVMInfoReturns(controlInstance, nil)
					fakeClient.CreateVMReturns(newInstance, nil)
					fakeClient.DeleteVMReturns(errors.New("got an error"))

					name := "originalVM"
					ami := "testAMI"

					instanceType := "testType"
					ip := "1.1.1.1"
					err := upgrade.Upgrade(name, ami, instanceType, ip)
					Expect(err).Should(HaveOccurred())
					Expect(fakeClient.GetVMInfoCallCount()).To(BeEquivalentTo(1))
					vmName := fakeClient.GetVMInfoArgsForCall(0)
					Expect(vmName).Should(BeEquivalentTo("originalVM*"))

					Expect(fakeClient.StopVMCallCount()).To(BeEquivalentTo(1))
					stopVMInstance := fakeClient.StopVMArgsForCall(0)
					Expect(stopVMInstance).Should(BeEquivalentTo(*controlInstance))

					Expect(fakeClient.CreateVMCallCount()).To(BeEquivalentTo(1))
					createVMInstance, createAMI, createInstanceType, createNewName := fakeClient.CreateVMArgsForCall(0)
					Expect(createVMInstance).Should(BeEquivalentTo(*controlInstance))
					Expect(createAMI).Should(BeEquivalentTo(ami))
					Expect(createInstanceType).Should(BeEquivalentTo(instanceType))
					Expect(createNewName).Should(ContainSubstring(name))

					Expect(fakeClient.WaitForStartedVMCallCount()).To(BeEquivalentTo(1))
					waitVMName := fakeClient.WaitForStartedVMArgsForCall(0)
					Expect(waitVMName).Should(BeEquivalentTo(createNewName))

					Expect(fakeClient.AssignPublicIPCallCount()).To(BeEquivalentTo(1))
					assignIPInstance, assignIP := fakeClient.AssignPublicIPArgsForCall(0)
					Expect(assignIPInstance).Should(BeEquivalentTo(*newInstance))
					Expect(assignIP).Should(BeEquivalentTo(ip))

					Expect(fakeClient.DeleteVMCallCount()).To(BeEquivalentTo(1))
					deleteInstance := fakeClient.DeleteVMArgsForCall(0)
					Expect(deleteInstance).Should(BeEquivalentTo(*controlInstance))
				})
			})
		})
	})
})
