package commands

type ReplaceVMCommand struct {
	Identifier string `short:"i" long:"identifier" required:"true" description:"Identifier of the VM that is being replaced"`
}

func (r *ReplaceVMCommand) Execute([]string) error {
	replacer, err := Cliaas.Config.NewVMReplacer()
	if err != nil {
		return err
	}

	return replacer.Replace(r.Identifier)
}
