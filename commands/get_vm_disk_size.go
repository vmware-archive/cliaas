package commands

import "fmt"

type GetVMDiskSizeCommand struct {
	Identifier string `short:"i" long:"identifier" required:"true" description:"Identifier of the VM to delete"`
}

func (c *GetVMDiskSizeCommand) Execute([]string) error {
	client, err := Cliaas.Config.NewClient()
	if err != nil {
		return err
	}

	disk, err := client.GetDisk(c.Identifier)
	if err != nil {
		return err
	}

	fmt.Printf("%d", disk.SizeGB)
	return nil
}
