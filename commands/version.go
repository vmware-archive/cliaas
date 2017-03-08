package commands

import (
	"fmt"
	"os"

	"github.com/pivotal-cf/cliaas"
)

type VersionCommand struct {
}

func (c *VersionCommand) Execute([]string) error {
	fmt.Println(cliaas.Version)
	os.Exit(0)
	return nil
}
