package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/cliaas/commands"
	"github.com/jessevdk/go-flags"
)

var _ = Describe("ReplaceVm", func() {
	It("errors if the identifier is not provided", func() {
		r := commands.ReplaceVMCommand{}
		_, err := flags.ParseArgs(&r, []string{})
		Expect(err).To(HaveOccurred())
	})

	It("defaults to 100GB for disk size", func() {
		r := commands.ReplaceVMCommand{}
		_, err := flags.ParseArgs(&r, []string{"--identifier", "an-identifier"})
		Expect(err).ToNot(HaveOccurred())

		Expect(r.Identifier).To(Equal("an-identifier"))
		Expect(r.DiskSizeGB).To(Equal(int64(100)))
	})

	It("allows a custom disk size", func() {
		r := commands.ReplaceVMCommand{}
		_, err := flags.ParseArgs(&r, []string{"--identifier", "an-identifier", "--disk-size-gb", "153"})
		Expect(err).ToNot(HaveOccurred())

		Expect(r.DiskSizeGB).To(Equal(int64(153)))
	})

	It("errors if the disk size is not an integer", func() {
		r := commands.ReplaceVMCommand{}
		_, err := flags.ParseArgs(&r, []string{"--identifier", "an-identifier", "--disk-size-gb", "256MB"})
		Expect(err).To(HaveOccurred())
	})
})
