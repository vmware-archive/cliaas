package gcp

import (
	"fmt"
	"time"

	"github.com/c0-ops/cliaas/iaas"
	errwrap "github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
)

const (
	defaultTimeoutSeconds int = 300
	InstanceStatusRunning     = "RUNNING"
	InstanceStatusStopped     = "STOPPED"
)

type OpsManagerGCP struct {
	client               ClientAPI
	clientTimeoutSeconds int
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

//Deploy - this should take a compute.Instance that is a copy of your existing
//Ops manager instance object, with 2 modifications.
// 1) you should swap the `Name` with a unique name you wish to use for the new opsmanager vm instnace
// 2) you should swap the `Instance.Disks` to match the latest instance image tarball for ops manager (found on network.pivotal.io)
func (s *OpsManagerGCP) Deploy(vmInstance *compute.Instance) error {
	err := s.createVM(vmInstance)
	if err != nil {
		return errwrap.Wrap(err, "createVM failed")
	}
	return nil
}

func (s *OpsManagerGCP) SpinDown(filter iaas.Filter) (*compute.Instance, error) {
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

func (s *OpsManagerGCP) CleanUp(filter iaas.Filter) error {
	vmInfo, err := s.client.GetVMInfo(filter)
	if err != nil {
		return errwrap.Wrap(err, "GetVMInfo failed")
	}

	err = s.deleteVM(vmInfo.Name)
	if err != nil {
		return errwrap.Wrap(err, "GetVMInfo failed")
	}

	return nil
}

func (s *OpsManagerGCP) deleteVM(instanceName string) error {
	err := s.client.DeleteVM(instanceName)
	if err != nil {
		return errwrap.Wrap(err, "DeleteVM failed")
	}

	return nil
}

func (s *OpsManagerGCP) stopVM(vmName string) error {
	err := s.client.StopVM(vmName)
	if err != nil {
		return errwrap.Wrap(err, "StopVM failed")
	}

	err = s.pollVMStatus(InstanceStatusStopped, vmName)
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

	err = s.pollVMStatus(InstanceStatusRunning, vmInstance.Name)
	if err != nil {
		return errwrap.Wrap(err, "polling VM Status failed")
	}
	return nil
}

func (s *OpsManagerGCP) pollVMStatus(desiredStatus string, vmName string) error {
	errChannel := make(chan error)
	go func() {
		for {
			vmInfo, err := s.client.GetVMInfo(iaas.Filter{NameRegexString: vmName})
			if err != nil {
				errChannel <- errwrap.Wrap(err, "GetVMInfo call failed")
				return
			}

			if vmInfo.Status == desiredStatus {
				errChannel <- nil
				return
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
