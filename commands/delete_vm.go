package commands

import "github.com/pivotal-cf/cliaas"

type DeleteVMCommand struct {
	Identifier string `short:"i" long:"identifier" required:"true" description:"Identifier of the VM to delete"`
}

func (c *DeleteVMCommand) Execute([]string) error {
	var configParser = cliaas.ConfigParser{Config: Cliaas.Config}
	deleter, err := configParser.NewVMDeleter()
	if err != nil {
		return err
	}

	return deleter.Delete(c.Identifier)
}
