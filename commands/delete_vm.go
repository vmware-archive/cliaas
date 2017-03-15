package commands

type DeleteVMCommand struct {
	Identifier string `short:"i" long:"identifier" required:"true" description:"Identifier of the VM to delete"`
}

func (c *DeleteVMCommand) Execute([]string) error {
	client, err := Cliaas.Config.NewClient()
	if err != nil {
		return err
	}

	return client.Delete(c.Identifier)
}
