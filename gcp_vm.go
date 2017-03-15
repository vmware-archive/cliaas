package cliaas

import (
	"errors"

	"github.com/pivotal-cf/cliaas/iaas/gcp"
)

func NewGCPVMDeleter(gcpClientAPI *gcp.GCPClientAPI) (VMDeleter, error) {
	return &gcpVM{
		client: gcpClientAPI,
	}, nil
}

func NewGCPVMReplacer(gcpClientAPI *gcp.GCPClientAPI) (VMReplacer, error) {
	return &gcpVM{
		client: gcpClientAPI,
	}, nil
}

type gcpVM struct {
	client *gcp.GCPClientAPI
}

func (d *gcpVM) Delete(identifier string) error {
	return d.client.DeleteVM(identifier)
}

func (d *gcpVM) Replace(identifier string) error {
	return errors.New("not yet implemented")
}
