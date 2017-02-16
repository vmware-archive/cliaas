package gcp

import (
	"fmt"
	"time"

	compute "google.golang.org/api/compute/v1"

	errwrap "github.com/pkg/errors"
)

const defaultTimeoutSeconds int = 300

type OpsManager interface {
	RunBlueGreen(filter Filter, imageURL string) error
}

type OpsManagerGCP struct {
	client               ClientAPI
	clientTimeoutSeconds int
}

type Filter struct {
	TagRegexString  string
	NameRegexString string
}

func NewOpsManager(configs ...func(*OpsManagerGCP) error) (*OpsManagerGCP, error) {
	om := &OpsManagerGCP{
		clientTimeoutSeconds: defaultTimeoutSeconds,
	}

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

func ConfigClientTimeoutSeconds(value int) func(*OpsManagerGCP) error {
	return func(om *OpsManagerGCP) error {
		om.clientTimeoutSeconds = value
		return nil
	}
}

func (s *OpsManagerGCP) Deploy(vmInstance *compute.Instance) error {
	err := s.createVM(vmInstance)
	if err != nil {
		return errwrap.Wrap(err, "createVM failed")
	}
	return nil
}

func (s *OpsManagerGCP) SpinDown(filter Filter) (*compute.Instance, error) {
	vmInfo, err := s.client.GetVMInfo(filter)

	if err != nil {
		return nil, errwrap.Wrap(err, "GetVMInfo failed")
	}
	err = s.stopVM(vmInfo.Name)

	if err != nil {
		return nil, errwrap.Wrap(err, "stopVM failed")
	}
	return vmInfo, nil
}

func (s *OpsManagerGCP) CleanUp(filter Filter, imageURL string) error {
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

	err = s.pollVMStatus("STOPPED", vmName)
	if err != nil {
		return errwrap.Wrap(err, "polling VM Status failed")
	}

	return nil
}

func (s *OpsManagerGCP) createVM(vmInstance *compute.Instance) error {
	err := s.client.CreateVM(*vmInstance)
	if err != nil {
		return errwrap.Wrap(err, "CreateVM call failed")
	}

	err = s.pollVMStatus("RUNNING", vmInstance.Name)
	if err != nil {
		return errwrap.Wrap(err, "polling VM Status failed")
	}
	return nil
}

func (s *OpsManagerGCP) pollVMStatus(desiredStatus string, vmName string) error {
	errChannel := make(chan error)
	go func() {
		for {
			vmInfo, err := s.client.GetVMInfo(Filter{NameRegexString: vmName})
			if err != nil {
				errChannel <- errwrap.Wrap(err, "GetVMInfo call failed")
			}

			if vmInfo.Status == desiredStatus {
				errChannel <- nil
			}
		}
	}()
	select {
	case res := <-errChannel:
		return res
	case <-time.After(time.Second * time.Duration(s.clientTimeoutSeconds)):
		return fmt.Errorf("polling for status timed out")
	}
	return fmt.Errorf("polling for status failed")
}
