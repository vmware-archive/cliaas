package gcp_test

import (
	. "github.com/c0-ops/cliaas/iaas/gcp"
	"github.com/c0-ops/cliaas/iaas/gcp/gcpfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	compute "google.golang.org/api/compute/v1"
)

var _ = Describe("GCPClientAPI", func() {

	Describe("GCPClientAPI", func() {
		var client *GCPClientAPI
		var err error
		var controlZone = "zone"
		var controlProject = "prj"
		var controlInstanceName = "blah"
		var controlInstanceTag = "hello"

		Describe("given a GetVMInfo method and a filter object argument", func() {

			Context("when there is a matching instance", func() {
				controlInstanceList := createInstanceList(controlInstanceName, controlInstanceTag)
				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.ListReturns(controlInstanceList, nil)

					client, err = NewGCPClientAPI(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then it should yield the matching gcp instance", func() {
					inst, err := client.GetVMInfo(Filter{NameRegexString: controlInstanceName, TagRegexString: controlInstanceTag})
					Expect(inst).ShouldNot(BeNil())
					Expect(controlInstanceList.Items).To(HaveLen(1))
					Expect(inst).Should(Equal(controlInstanceList.Items[0]))
					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			Context("when there is no matching instance", func() {

				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.ListReturns(createInstanceList("nothing-to-match", "nothing-to-match"), nil)

					client, err = NewGCPClientAPI(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then it should give an error", func() {
					inst, err := client.GetVMInfo(Filter{NameRegexString: "bbb", TagRegexString: "ddd"})
					Expect(inst).Should(BeNil())
					Expect(err).Should(HaveOccurred())
				})
			})
			Context("when there is empty instance set", func() {

				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.ListReturns(&compute.InstanceList{}, nil)

					client, err = NewGCPClientAPI(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then it should give an error", func() {
					inst, err := client.GetVMInfo(Filter{})
					Expect(inst).Should(BeNil())
					Expect(err).Should(HaveOccurred())
				})
			})
		})
	})

	Describe("given a NewGCPCLIentAPI()", func() {

		Context("when passed a valid set of configs", func() {

			var client *GCPClientAPI
			var err error
			var controlZone = "zone"
			var controlProject = "prj"
			var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
			BeforeEach(func() {
				client, err = NewGCPClientAPI(
					ConfigGoogleClient(fakeGoogleClient),
					ConfigZoneName(controlZone),
					ConfigProjectName(controlProject),
				)
			})
			It("then it should provide a properly initialized GCPCLientAPI object", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(client).ShouldNot(BeNil())
			})
		})
	})
})

func createInstanceList(name, tag string) *compute.InstanceList {
	return &compute.InstanceList{
		Items: []*compute.Instance{
			&compute.Instance{
				Name: name,
				Tags: &compute.Tags{
					Items: []string{
						tag,
					},
				},
			},
		},
	}
}
