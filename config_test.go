package cliaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cliaas"
)

var _ = Describe("Config", func() {
	var err error
	var config Config
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
	Describe("NewVMReplacer", func() {
		var vmReplacer VMReplacer

		JustBeforeEach(func() {
			vmReplacer, err = config.NewVMReplacer()
		})

		Context("when called on a config with multiple complete and valid iaas configs", func() {
			BeforeEach(func() {
				config = Config{
					AWS: validAWSConfig,
					GCP: validGCPConfig,
				}
			})
			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
				Expect(vmReplacer).Should(BeNil(), "parse error should prevent any objects being returned")
			})
		})

		Context("when called on a config with a single complete and valid iaas config", func() {
			BeforeEach(func() {
				config = Config{
					GCP: validGCPConfig,
				}
			})
			It("should always return the valid iaas config as a ReplacerCreator", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmReplacer).ShouldNot(BeNil())
			})
		})

		Context("when called on a config with a single complete GCP config as well as a partially configured iaas", func() {
			BeforeEach(func() {
				config = Config{
					GCP: validGCPConfig,
					AWS: AWS{AMI: "ami-something"},
				}
			})
			It("should always return the valid iaas config as a ReplacerCreator and ignore incomplete iaas blocks", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmReplacer).ShouldNot(BeNil())
			})
		})

		Context("when called on a config with a single complete AWS config as well as a partially configured iaas", func() {
			BeforeEach(func() {
				config = Config{
					GCP: GCP{Zone: "asdf"},
					AWS: validAWSConfig,
				}
			})
			It("should always return the valid iaas config as a ReplacerCreator and ignore incomplete iaas blocks", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmReplacer).ShouldNot(BeNil())
			})
		})

		Context("when called on a config with a no complete and valid iaas configs", func() {
			BeforeEach(func() {
				config = Config{}
			})
			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
				Expect(vmReplacer).Should(BeNil(), "parse error should prevent any objects being returned")
			})
		})
	})
	Describe("NewVMDeleter", func() {
		var vmDeleter VMDeleter

		JustBeforeEach(func() {
			vmDeleter, err = config.NewVMDeleter()
		})

		Context("when called on a config with multiple complete and valid iaas configs", func() {
			BeforeEach(func() {
				config = Config{
					AWS: validAWSConfig,
					GCP: validGCPConfig,
				}
			})
			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
				Expect(vmDeleter).Should(BeNil(), "parse error should prevent any objects being returned")
			})
		})

		Context("when called on a config with a single complete and valid iaas config", func() {
			BeforeEach(func() {
				config = Config{
					GCP: validGCPConfig,
				}
			})
			It("should always return the valid iaas config as a ValidDeleter", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmDeleter).ShouldNot(BeNil())
			})
		})

		Context("when called on a config with a single complete AWS config as well as a partially configured iaas", func() {
			BeforeEach(func() {
				config = Config{
					GCP: GCP{Zone: "dasfas"},
					AWS: validAWSConfig,
				}
			})
			It("should always return the valid iaas config as a ValidDeleter and ignore incomplete iaas blocks", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmDeleter).ShouldNot(BeNil())
			})
		})

		Context("when called on a config with a single complete GCP config as well as a partially configured iaas", func() {
			BeforeEach(func() {
				config = Config{
					GCP: validGCPConfig,
					AWS: AWS{AMI: "ami-something"},
				}
			})
			It("should always return the valid iaas config as a ValidDeleter and ignore incomplete iaas blocks", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmDeleter).ShouldNot(BeNil())
			})
		})

		Context("when called on a config with a no complete and valid iaas configs", func() {
			BeforeEach(func() {
				config = Config{}
			})
			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
				Expect(vmDeleter).Should(BeNil(), "parse error should prevent any objects being returned")
			})
		})
	})
})
