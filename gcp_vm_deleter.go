package cliaas

import "github.com/pivotal-cf/cliaas/iaas/gcp"

func NewGCPVMDeleter(gcpClientAPI *gcp.GCPClientAPI) (VMDeleter, error) {
	return &gcpVMDeleter{
		client: gcpClientAPI,
	}, nil
}

type gcpVMDeleter struct {
	client *gcp.GCPClientAPI
}

func (d *gcpVMDeleter) Delete(identifier string) error {
	return d.client.DeleteVM(identifier)
}

func (d *gcpVMDeleter) GCPVMDeleter() {}
