package commands

type DeleteVMCommand struct {
	Identifier string `short:"i" long:"identifier" required:"true" description:"Identifier of the VM to delete"`
}

func (c *DeleteVMCommand) Execute([]string) error {
	deleter, err := Cliaas.Config.NewVMDeleter()
	if err != nil {
		return err
	}

	return deleter.Delete(c.Identifier)
}