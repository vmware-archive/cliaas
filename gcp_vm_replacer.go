package cliaas

import (
	"errors"

	"github.com/pivotal-cf/cliaas/iaas/gcp"
)

func NewGCPVMReplacer(gcpClientAPI *gcp.GCPClientAPI) (VMReplacer, error) {
	return &gcpVMReplacer{
		client: gcpClientAPI,
	}, nil
}

type gcpVMReplacer struct {
	client *gcp.GCPClientAPI
}

func (d *gcpVMReplacer) Replace(identifier string) error {
	//d.client.StopVM(identifier)
	//d.client.CreateVM(identifier)
	return errors.New("not yet implemented")
}
