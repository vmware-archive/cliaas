package cliaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/cliaas"
)

var _ = Describe("Config", func() {
	Describe("CompleteConfigs", func() {
		var multiConfig cliaas.MultiConfig

		BeforeEach(func() {
			multiConfig = cliaas.MultiConfig{}
		})

		It("returns an empty slice when no configs are set", func() {
			Expect(multiConfig.CompleteConfigs()).To(Equal([]cliaas.Config{}))
		})

		Context("when the multi config has a complete AWS config", func() {
			var awsConfig *cliaas.AWSConfig

			BeforeEach(func() {
				awsConfig = &cliaas.AWSConfig{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
					VPCID:           "some-vpc-id",
				}

				multiConfig = cliaas.MultiConfig{
					AWS: awsConfig,
				}
			})

			It("returns a slice of the AWS config", func() {
				Expect(multiConfig.CompleteConfigs()).To(Equal([]cliaas.Config{awsConfig}))
			})
		})

		Context("when the multi config has a complete GCP config", func() {
			var gcpConfig *cliaas.GCPConfig

			BeforeEach(func() {
				gcpConfig = &cliaas.GCPConfig{
					CredfilePath: "testdata/fake_gcp_creds.json",
					Zone:         "some-zone",
					Project:      "some-project",
					DiskImageURL: "some-disk-image-url",
				}

				multiConfig = cliaas.MultiConfig{
					GCP: gcpConfig,
				}
			})

			It("returns a slice of the GCP config", func() {
				Expect(multiConfig.CompleteConfigs()).To(Equal([]cliaas.Config{gcpConfig}))
			})
		})
	})
})
