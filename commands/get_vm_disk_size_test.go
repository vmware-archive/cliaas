package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/cliaas/commands"
	"github.com/jessevdk/go-flags"
)

var _ = Describe("GetVmDiskSize", func() {
	It("errors if the identifier is not provided", func() {
		r := commands.GetVMDiskSizeCommand{}
		_, err := flags.ParseArgs(&r, []string{})
		Expect(err).To(HaveOccurred())
	})
})
