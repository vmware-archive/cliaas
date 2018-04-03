package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/gbytes"

	"testing"
	"os/exec"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = Describe("Main", func() {
	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})

	It("builds", func() {
		_, err := gexec.Build("github.com/pivotal-cf/cliaas/cmd/cliaas")
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns the current version string and exits", func() {
		bin, err := gexec.Build("github.com/pivotal-cf/cliaas/cmd/cliaas", "-ldflags", `-X github.com/pivotal-cf/cliaas.Version=1.2.3-dev`)

		Expect(err).NotTo(HaveOccurred())

		command := exec.Command(bin, "-c", "../../testdata/fake_aws_config.yml","version")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say("1.2.3-dev"))
		Eventually(session).Should(gexec.Exit(0))
	})
})
