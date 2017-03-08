package commands

import (
	"github.com/pivotal-cf/cliaas"
)

type ReplaceVMCommand struct {
	Identifier string `short:"i" long:"identifier" required:"true" description:"Identifier of the VM that is being replaced"`
}

func (r *ReplaceVMCommand) Execute([]string) error {
	replacer, err := cliaas.NewVMReplacer(Cliaas.Config)
	if err != nil {
		return err
	}

	return replacer.Replace(r.Identifier)
}
