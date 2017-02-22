package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/c0-ops/cliaas/iaas"
)

type ClientAPI interface {
	CreateVM(instance ec2.Instance) error
	DeleteVM(instanceName string) error
	GetVMInfo(filter iaas.Filter) (*ec2.Instance, error)
	StopVM(instanceName string) error
}

type AWSClientAPI struct {
}

func (s *AWSClientAPI) CreateVM(instance ec2.Instance) error {

	return nil
}

func (s *AWSClientAPI) DeleteVM(instanceName string) error {
	return nil
}

//StopVM - will try to stop the VM with the given name
func (s *AWSClientAPI) StopVM(instanceName string) error {
	return nil
}

//GetVMInfo - gets the information on the first VM to match the given filter argument
// currently filter will only do a regex on teh tag||name regex fields against
// the List's result set
func (s *AWSClientAPI) GetVMInfo(filter iaas.Filter) (*ec2.Instance, error) {
	return nil, fmt.Errorf("No instance matches found")
}
