package gcp_test

import (
	. "github.com/c0-ops/cliaas/iaas/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPClientAPI", func() {

	Describe("given a NewGCPCLIentAPI()", func() {

		Context("when passed a valid set of configs", func() {

			var client *GCPClientAPI
			var err error
			var controlZone = "zone"
			var controlProject = "prj"
			var controlCreds = "/tmp/some.json"
			BeforeEach(func() {
				client, err = NewGCPClientAPI(
					ConfigZoneName(controlZone),
					ConfigProjectName(controlProject),
					ConfigCredPath(controlCreds),
				)
			})
			It("then it should provide a properly initialized GCPCLientAPI object", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(client).ShouldNot(BeNil())
			})
		})
	})
})
