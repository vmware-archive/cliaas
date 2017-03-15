package commands

type ReplaceVMCommand struct {
	Identifier string `long:"identifier" required:"true" description:"Identifier of the VM that is being replaced"`
	Image      string `long:"image" required:"true" description:"Image to use for the new VM"`
}

func (r *ReplaceVMCommand) Execute([]string) error {
	replacer, err := Cliaas.Config.NewVMReplacer()
	if err != nil {
		return err
	}

	return replacer.Replace(r.Identifier, r.Image)
}
