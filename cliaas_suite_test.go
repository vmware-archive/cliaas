package cliaas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCliaas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cliaas Suite")
}
