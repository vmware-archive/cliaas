package gcp

import (
	"fmt"

	errwrap "github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
)

type ClientAPI interface {
	CreateVM(instance compute.Instance) error
	DeleteVM(instanceName string) error
	GetVMInfo(filter Filter) (*compute.Instance, error)
	StopVM(instanceName string) error
}

type GCPClientAPI struct {
	credPath    string
	projectName string
	zoneName    string
}

func NewGCPClientAPI(configs ...func(*GCPClientAPI) error) (*GCPClientAPI, error) {
	gcpClient := new(GCPClientAPI)

	for _, cfg := range configs {
		err := cfg(gcpClient)
		if err != nil {
			return nil, errwrap.Wrap(err, "new GCP Client config loading error")
		}
	}

	if gcpClient.credPath == "" {
		return nil, fmt.Errorf("You have an incomplete GCPClientAPI.credPath")
	}
	if gcpClient.projectName == "" {
		return nil, fmt.Errorf("You have an incomplete GCPClientAPI.projectName")
	}
	if gcpClient.zoneName == "" {
		return nil, fmt.Errorf("You have an incomplete GCPClientAPI.zoneName")
	}
	return gcpClient, nil
}

func ConfigZoneName(value string) func(*GCPClientAPI) error {
	return func(gcpClient *GCPClientAPI) error {
		gcpClient.zoneName = value
		return nil
	}
}

func ConfigProjectName(value string) func(*GCPClientAPI) error {
	return func(gcpClient *GCPClientAPI) error {
		gcpClient.projectName = value
		return nil
	}
}

func ConfigCredPath(value string) func(*GCPClientAPI) error {
	return func(gcpClient *GCPClientAPI) error {
		gcpClient.credPath = value
		return nil
	}
}

func (s *GCPClientAPI) CreateVM(instanceName string, sourceImageTarballUrl string) error {
	return nil
}

func (s *GCPClientAPI) DeleteVM(instanceName string) error {
	return nil
}

func (s *GCPClientAPI) GetVMInfo(filter Filter) (*compute.Instance, error) {
	return nil, nil
}

func (s *GCPClientAPI) StopVM(instanceName string) error {
	return nil
}
