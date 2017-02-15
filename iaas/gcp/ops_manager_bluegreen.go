package gcp

import (
	"fmt"

	compute "google.golang.org/api/compute/v1"

	errwrap "github.com/pkg/errors"
)

type OpsManager interface {
	RunBlueGreen(filter Filter, imageURL string) error
}
type OpsManagerGCP struct {
	credPath    string
	projectName string
	zoneName    string
	client      ClientAPI
}
type Filter struct {
	TagRegexString  string
	NameRegexString string
}

func NewOpsManager(configs ...func(*OpsManagerGCP) error) (*OpsManagerGCP, error) {
	om := new(OpsManagerGCP)

	for _, cfg := range configs {
		err := cfg(om)
		if err != nil {
			return nil, errwrap.Wrap(err, "new ops manager config loading error")
		}
	}

	if om.client == nil {
		return nil, fmt.Errorf("You have an incomplete OpsManagerGCP.client")
	}
	return om, nil
}

func ConfigClient(value ClientAPI) func(*OpsManagerGCP) error {
	return func(om *OpsManagerGCP) error {
		om.client = value
		return nil
	}
}

func (s *OpsManagerGCP) RunBlueGreen(filter Filter, imageURL string) error {
	vmInfo, err := s.client.GetVMInfo(filter)

	if err != nil {
		return errwrap.Wrap(err, "GetVMInfo failed")
	}
	err = s.stopVM(vmInfo.Name)

	if err != nil {
		return errwrap.Wrap(err, "stopVM failed")
	}

	err = s.createVM(vmInfo, imageURL)
	if err != nil {
		return errwrap.Wrap(err, "createVM failed")
	}
	return nil
}

func (s *OpsManagerGCP) stopVM(vmName string) error {

	err := s.client.StopVM(vmName)

	if err != nil {
		return errwrap.Wrap(err, "StopVM failed")
	}

	for {
		vmInfo, err := s.client.GetVMInfo(Filter{NameRegexString: vmName})
		if err != nil {
			return errwrap.Wrap(err, "GetVMInfo call failed")
		}

		if vmInfo.Status == "STOPPED" {
			return nil
		}
	}
	return fmt.Errorf("polling of vm stop failed")
}

func (s *OpsManagerGCP) createVM(vmInfo *compute.Instance, sourceImageTarballURL string) error {
	vmInfo.Disks = []*compute.AttachedDisk{
		&compute.AttachedDisk{
			Source: sourceImageTarballURL,
		},
	}
	err := s.client.CreateVM(*vmInfo)

	if err != nil {
		return errwrap.Wrap(err, "CreateVM call failed")
	}
	return nil
}
