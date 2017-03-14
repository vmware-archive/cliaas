package cliaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cliaas"
)

var _ = Describe("VMDeleter", func() {
	var config Config
	var vmDeleter VMDeleter
	var err error
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
		vmDeleter, err = NewVMDeleter(config)
	})

	Context("when passed an Invalid Config", func() {
		BeforeEach(func() {
			config = Config{}
		})
		It("should return error", func() {
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when passed a valid config for multiple IaaS'", func() {
		BeforeEach(func() {
			config = Config{
				AWS: validAWSConfig,
				GCP: validGCPConfig,
			}
		})
		It("cascade through returning a deleter based on the first valid iaas config", func() {
			Expect(err).ShouldNot(HaveOccurred())
			Expect(vmDeleter).ShouldNot(BeNil())
			_, ok := vmDeleter.(AWSVMDeleter)
			Expect(ok).Should(BeTrue(), "our vm deleter should be for aws")
		})
	})

	Context("when passed only a valid AWS Config", func() {
		BeforeEach(func() {
			config = Config{
				AWS: validAWSConfig,
			}
		})
		It("should construct a VM Deleter", func() {
			Expect(err).ShouldNot(HaveOccurred())
			Expect(vmDeleter).ShouldNot(BeNil())
			_, ok := vmDeleter.(AWSVMDeleter)
			Expect(ok).Should(BeTrue(), "our vm deleter should be for aws")
		})
	})

	Context("when passed only a valid GCP Config", func() {
		BeforeEach(func() {
			config = Config{
				GCP: validGCPConfig,
			}
		})
		It("should construct a VM Deleter", func() {
			Expect(err).ShouldNot(HaveOccurred())
			Expect(vmDeleter).ShouldNot(BeNil())
			_, ok := vmDeleter.(GCPVMDeleter)
			Expect(ok).Should(BeTrue(), "our vm deleter should be for gcp")
		})
	})
})

type GCPVMDeleter interface {
	GCPVMDeleter()
}

type AWSVMDeleter interface {
	AWSVMDeleter()
}
