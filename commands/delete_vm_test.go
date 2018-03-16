package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/cliaas/commands"
	"github.com/jessevdk/go-flags"
)

var _ = Describe("DeleteVm", func() {
	It("errors if the identifier is not provided", func() {
		r := commands.DeleteVMCommand{}
		_, err := flags.ParseArgs(&r, []string{})
		Expect(err).To(HaveOccurred())
	})
})
