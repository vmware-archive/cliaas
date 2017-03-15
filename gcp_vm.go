package cliaas

import (
	"fmt"
	"time"

	"github.com/pivotal-cf/cliaas/iaas"
	"github.com/pivotal-cf/cliaas/iaas/gcp"
	compute "google.golang.org/api/compute/v1"

	errwrap "github.com/pkg/errors"
)

type gcpVM struct {
	client                *gcp.GCPClientAPI
	sourceImageTarballURL string
}

func (s *gcpVM) Delete(identifier string) error {
	return s.client.DeleteVM(identifier)
}

func (s *gcpVM) Replace(identifier string, sourceImageTarballURL string) error {
	vmInstance, err := s.client.GetVMInfo(iaas.Filter{
		NameRegexString: identifier + "*",
	})
	if err != nil {
		return errwrap.Wrap(err, "getvminfo failed")
	}

	err = s.client.StopVM(vmInstance.Name)
	if err != nil {
		return errwrap.Wrap(err, "stopvm failed")
	}

	err = s.client.WaitForStatus(vmInstance.Name, "stopped")
	if err != nil {
		return errwrap.Wrap(err, "waitforstatus after stopvm failed")
	}

	vmInstance.Name = fmt.Sprintf("%s-%s", identifier, time.Now().Format("2006-01-02_15-04-05"))
	vmInstance.Disks = []*compute.AttachedDisk{
		&compute.AttachedDisk{
			Source: sourceImageTarballURL,
		},
	}
	err = s.client.CreateVM(*vmInstance)
	if err != nil {
		return errwrap.Wrap(err, "CreateVM call failed")
	}

	return s.client.WaitForStatus(vmInstance.Name, "running")
}
