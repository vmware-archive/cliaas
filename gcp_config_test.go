package cliaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cliaas"
)

var _ = Describe("GCP", func() {
	var validGCPConfig GCP

	BeforeEach(func() {
		validGCPConfig = GCP{
			CredfilePath: "testdata/fake_gcp_creds.json",
			Project:      "--",
			Zone:         "--",
			DiskImageURL: "--",
		}
	})

	Describe("NewDeleter", func() {
		var vmDeleter VMDeleter
		var err error

		JustBeforeEach(func() {
			vmDeleter, err = validGCPConfig.NewDeleter()
		})

		Context("when called on a complete config with a valid cred file", func() {

			It("should return a valid VMDeleter", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmDeleter).ShouldNot(BeNil())
			})
		})

		Context("when called on an incomplete config", func() {

			BeforeEach(func() {
				validGCPConfig.Zone = ""
			})

			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
				Expect(vmDeleter).Should(BeNil())
			})
		})

		Context("when using an invalid cred file", func() {

			BeforeEach(func() {
				validGCPConfig.CredfilePath = ""
			})

			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
				Expect(vmDeleter).Should(BeNil())
			})
		})
	})

	Describe("IsValid", func() {
		var isValidResult bool
		JustBeforeEach(func() {
			isValidResult = validGCPConfig.IsValid()
		})

		Context("when we have a valid config", func() {
			It("should pass the is valid test", func() {
				Expect(isValidResult).Should(BeTrue())
			})
		})

		Context("when not passed a valid zone", func() {
			BeforeEach(func() {
				validGCPConfig.Zone = ""
			})
			It("should not pass the is valid test", func() {
				Expect(isValidResult).Should(BeFalse())
			})
		})

		Context("when not passed a valid project", func() {
			BeforeEach(func() {
				validGCPConfig.Project = ""
			})
			It("should not pass the is valid test", func() {
				Expect(isValidResult).Should(BeFalse())
			})
		})

		Context("when not passed a valid cred file path", func() {
			BeforeEach(func() {
				validGCPConfig.CredfilePath = ""
			})
			It("should not pass the is valid test", func() {
				Expect(isValidResult).Should(BeFalse())
			})
		})

		Context("when not passed a valid diskimage url", func() {
			BeforeEach(func() {
				validGCPConfig.DiskImageURL = ""
			})
			It("should not pass the is valid test", func() {
				Expect(isValidResult).Should(BeFalse())
			})
		})
	})
})
