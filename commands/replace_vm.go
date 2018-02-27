package commands

import "strconv"

type ReplaceVMCommand struct {
	Identifier string `long:"identifier" required:"true" description:"Identifier of the VM that is being replaced"`
	DiskSizeGB string `long:"disk-size-gb" required:"true" description:"Disk size of the VM that is being replaced"`
}

func (r *ReplaceVMCommand) Execute([]string) error {
	client, err := Cliaas.Config.NewClient()
	if err != nil {
		return err
	}

	size, err := strconv.ParseInt(r.DiskSizeGB, 10, 64)
	if err != nil {
		return err
	}

	return client.Replace(r.Identifier, Cliaas.Config.Image(), size)
}
