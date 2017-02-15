package gcp

import (
	compute "google.golang.org/api/compute/v1"
)

type ClientAPI interface {
	CreateVM(instanceName string, sourceImageTarballUrl string) error
	DeleteVM(instanceName string) error
	GetVMInfo(filter Filter) (*compute.Instance, error)
	StopVM(instanceName string) error
}

type GCPClientAPI struct {
	project string
	zone    string
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
