package gcp

import (
	"fmt"
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

	if om.credPath == "" {
		return nil, fmt.Errorf("You have an incomplete OpsManagerGCP.credPath")
	}
	if om.projectName == "" {
		return nil, fmt.Errorf("You have an incomplete OpsManagerGCP.projectName")
	}
	if om.zoneName == "" {
		return nil, fmt.Errorf("You have an incomplete OpsManagerGCP.zoneName")
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

func ConfigZoneName(value string) func(*OpsManagerGCP) error {
	return func(om *OpsManagerGCP) error {
		om.zoneName = value
		return nil
	}
}

func ConfigProjectName(value string) func(*OpsManagerGCP) error {
	return func(om *OpsManagerGCP) error {
		om.projectName = value
		return nil
	}
}

func ConfigCredPath(value string) func(*OpsManagerGCP) error {
	return func(om *OpsManagerGCP) error {
		om.credPath = value
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
