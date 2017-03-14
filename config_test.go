package cliaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cliaas"
)

var _ = Describe("Config", func() {
	Describe("ConfigParser", func() {
		Describe("NewVMDeleter", func() {
			var vmDeleter VMDeleter
			var err error
			var configParser ConfigParser
			var validAWSConfig AWS
			var validGCPConfig GCP

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
				vmDeleter, err = configParser.NewVMDeleter()
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
				It("should return an error", func() {
					Expect(err).Should(HaveOccurred())
					Expect(vmDeleter).Should(BeNil(), "parse error should prevent any objects being returned")
				})
			})

			Context("when called on a config with a single complete and valid iaas config", func() {
				BeforeEach(func() {
					configParser = ConfigParser{
						Config: Config{
							GCP: validGCPConfig,
						},
					}
				})
				It("should always return the valid iaas config as a ValidDeleter", func() {
					Expect(err).ShouldNot(HaveOccurred())
					Expect(vmDeleter).ShouldNot(BeNil())
					_, ok := vmDeleter.(GCPVMDeleter)
					Expect(ok).Should(BeTrue(), "our vm deleter should be for gcp")
				})
			})

			Context("when called on a config with a single complete and valid iaas config as well as a partially configured iaas", func() {
				BeforeEach(func() {
					configParser = ConfigParser{
						Config: Config{
							GCP: validGCPConfig,
							AWS: AWS{AMI: "ami-something"},
						},
					}
				})
				It("should always return the valid iaas config as a ValidDeleter and ignore incomplete iaas blocks", func() {
					Expect(err).ShouldNot(HaveOccurred())
					Expect(vmDeleter).ShouldNot(BeNil())
					_, ok := vmDeleter.(GCPVMDeleter)
					Expect(ok).Should(BeTrue(), "our vm deleter should be for gcp")
				})
			})

			Context("when called on a config with a no complete and valid iaas configs", func() {
				BeforeEach(func() {
					configParser = ConfigParser{}
				})
				It("should return an error", func() {
					Expect(err).Should(HaveOccurred())
					Expect(vmDeleter).Should(BeNil(), "parse error should prevent any objects being returned")
				})
			})
		})
	})
})

type GCPVMDeleter interface {
	GCPVMDeleter()
}

type AWSVMDeleter interface {
	AWSVMDeleter()
}
