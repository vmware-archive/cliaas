package cliaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cliaas"
)

var _ = Describe("Config", func() {
	Describe("ConfigParser", func() {
		Describe("GetValidDeleters", func() {
			var validDeleters []ValidDeleter
			var err error
			var configParser ConfigParser
			var validAWSConfig AWS
			var validGCPConfig GCP
			var controlConfigCount = 2

			BeforeEach(func() {
				validAWSConfig = AWS{
					AccessKeyID:     "--",
					SecretAccessKey: "--",
					Region:          "--",
					VPCID:           "--",
					AMI:             "--",
				}
				validGCPConfig = GCP{
					CredfilePath: "testdata/fake_gcp_creds.json",
					Project:      "--",
					Zone:         "--",
					DiskImageURL: "--",
				}
			})

			JustBeforeEach(func() {
				validDeleters = make([]ValidDeleter, 0)
				validDeleters, err = configParser.GetValidDeleters()
			})

			Context("when called on a config with multiple complete and valid iaas configs", func() {
				BeforeEach(func() {
					configParser = ConfigParser{
						Config: Config{
							AWS: validAWSConfig,
							GCP: validGCPConfig,
						},
					}
				})
				It("should always return all iaas configs as ValidDeleters", func() {
					Expect(err).ShouldNot(HaveOccurred())
					Expect(validDeleters).Should(HaveLen(controlConfigCount))
				})
			})

			Context("when called on a config with a single complete and valid iaas configs", func() {
				BeforeEach(func() {
					configParser = ConfigParser{
						Config: Config{
							GCP: validGCPConfig,
						},
					}
				})
				It("should always return all iaas configs as ValidDeleters", func() {
					Expect(err).ShouldNot(HaveOccurred())
					Expect(validDeleters).Should(HaveLen(controlConfigCount))
				})
			})

			Context("when called on a config with a no complete and valid iaas configs", func() {
				BeforeEach(func() {
					configParser = ConfigParser{}
				})
				It("should always return all iaas configs as ValidDeleters", func() {
					Expect(err).ShouldNot(HaveOccurred())
					Expect(validDeleters).Should(HaveLen(controlConfigCount))
				})
			})
		})
	})
})
