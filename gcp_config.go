package cliaas

import (
	"os"

	"github.com/pivotal-cf/cliaas/iaas/gcp"
	errwrap "github.com/pkg/errors"
)

type GCP struct {
	CredfilePath string `yaml:"credfile"`
	Zone         string `yaml:"zone"`
	Project      string `yaml:"project"`
	DiskImageURL string `yaml:"disk_image_url"`
}

func (c GCP) IsValid() bool {
	_, err := os.Stat(c.CredfilePath)
	return c.CredfilePath != "" &&
		err == nil &&
		!os.IsNotExist(err) &&
		c.Zone != "" &&
		c.Project != "" &&
		c.DiskImageURL != ""
}

func (c GCP) NewDeleter() (VMDeleter, error) {
	gcpClientAPI, err := c.newGCPClient()
	if err != nil {
		return nil, errwrap.Wrap(err, "Failed to create new GCP API client")
	}

	return NewGCPVMDeleter(gcpClientAPI)
}

func (c GCP) NewReplacer() (VMReplacer, error) {
	gcpClientAPI, err := c.newGCPClient()
	if err != nil {
		return nil, errwrap.Wrap(err, "Failed to create new GCP API client")
	}

	return NewGCPVMReplacer(gcpClientAPI, c.DiskImageURL)
}

func (c GCP) newGCPClient() (*gcp.GCPClientAPI, error) {
	gcpClient, err := gcp.NewDefaultGoogleComputeClient(c.CredfilePath)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to create gcp default client")
	}

	gcpClientAPI, err := gcp.NewGCPClientAPI(
		gcp.ConfigGoogleClient(gcpClient),
		gcp.ConfigZoneName(c.Zone),
		gcp.ConfigProjectName(c.Project),
	)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to create gcp client api")
	}
	return gcpClientAPI, nil
}
