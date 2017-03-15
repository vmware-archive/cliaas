package cliaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cliaas"

	errwrap "github.com/pkg/errors"
)

var _ = Describe("AWS", func() {
	var validAWSConfig AWS

	BeforeEach(func() {
		validAWSConfig = AWS{
			AccessKeyID:     "--",
			SecretAccessKey: "--",
			Region:          "--",
			VPCID:           "--",
			AMI:             "--",
		}
	})

	Describe("NewReplacer", func() {
		var vmReplacer VMReplacer
		var err error

		JustBeforeEach(func() {
			vmReplacer, err = validAWSConfig.NewReplacer()
		})

		Context("when called on a valid aws config", func() {

			It("should return a valid VMReplacer", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmReplacer).ShouldNot(BeNil())
			})
		})

		Context("when called on an incomplete config", func() {

			BeforeEach(func() {
				validAWSConfig.AccessKeyID = ""
			})

			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
				Expect(errwrap.Cause(err)).Should(Equal(InvalidConfigErr))
				Expect(vmReplacer).Should(BeNil())
			})
		})
	})

	Describe("NewDeleter", func() {
		var vmDeleter VMDeleter
		var err error

		JustBeforeEach(func() {
			vmDeleter, err = validAWSConfig.NewDeleter()
		})

		Context("when called on a valid aws config", func() {

			It("should return a valid VMDeleter", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(vmDeleter).ShouldNot(BeNil())
			})
		})

		Context("when called on an incomplete config", func() {

			BeforeEach(func() {
				validAWSConfig.AccessKeyID = ""
			})

			It("should return an error", func() {
				Expect(err).Should(HaveOccurred())
				Expect(errwrap.Cause(err)).Should(Equal(InvalidConfigErr))
				Expect(vmDeleter).Should(BeNil())
			})
		})
	})

	Describe("IsValid", func() {
		var isValidResult bool
		JustBeforeEach(func() {
			isValidResult = validAWSConfig.IsValid()
		})

		Context("when we have a valid config", func() {
			It("should pass the is valid test", func() {
				Expect(isValidResult).Should(BeTrue())
			})
		})

		Context("when not passed a valid access key", func() {
			BeforeEach(func() {
				validAWSConfig.AccessKeyID = ""
			})
			It("should not pass the is valid test", func() {
				Expect(isValidResult).Should(BeFalse())
			})
		})

		Context("when not passed a valid secret key", func() {
			BeforeEach(func() {
				validAWSConfig.SecretAccessKey = ""
			})
			It("should not pass the is valid test", func() {
				Expect(isValidResult).Should(BeFalse())
			})
		})

		Context("when not passed a valid region", func() {
			BeforeEach(func() {
				validAWSConfig.Region = ""
			})
			It("should not pass the is valid test", func() {
				Expect(isValidResult).Should(BeFalse())
			})
		})

		Context("when not passed a valid vpcid", func() {
			BeforeEach(func() {
				validAWSConfig.VPCID = ""
			})
			It("should not pass the is valid test", func() {
				Expect(isValidResult).Should(BeFalse())
			})
		})

		Context("when not passed a valid AMI", func() {
			BeforeEach(func() {
				validAWSConfig.AMI = ""
			})
			It("should not pass the is valid test", func() {
				Expect(isValidResult).Should(BeFalse())
			})
		})
	})
})
