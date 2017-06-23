package cliaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cliaas"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf/cliaas/cliaasfakes"
)

var _ = Describe("test for unexported features", func() {
	Describe("awsClient", func() {
		Context("when calling Replace on a running VM with valid arguments", func() {
			var client Client
			var fakeAPIClient *cliaasfakes.FakeAWSClient
			var callIndex = map[string]int{
				"old-vm-shutdown": 0,
				"new-vm-startup":  1,
			}
			BeforeEach(func() {
				fakeAPIClient = new(cliaasfakes.FakeAWSClient)
				fakeAPIClient.GetVMInfoReturns(VMInfo{
					InstanceID:       "1234",
					InstanceType:     "abc",
					KeyName:          "xyz",
					SubnetID:         "asdf",
					SecurityGroupIDs: []string{"hithere"},
				}, nil)
				fakeAPIClient.StopVMReturns(nil)
				fakeAPIClient.WaitForStatusReturns(nil)
				fakeAPIClient.CreateVMReturns("1234", nil)
				client = NewAWSAPIClientAdaptor(fakeAPIClient)

				err := client.Replace("abc", "xyz")
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("should wait on VM state changes", func() {
				Expect(fakeAPIClient.WaitForStatusCallCount()).To(Equal(2), "we wait for the intial stopvm and the following startvm")
			})

			It("should wait for vm stopping after stopping the old vm", func() {
				_, state := fakeAPIClient.WaitForStatusArgsForCall(callIndex["old-vm-shutdown"])
				Expect(state).Should(Equal(ec2.InstanceStateNameStopped))
			})

			It("should wait for vm starting after starting the new vm", func() {
				_, state := fakeAPIClient.WaitForStatusArgsForCall(callIndex["new-vm-startup"])
				Expect(state).Should(Equal(ec2.InstanceStateNameRunning))
			})
		})
	})
})
