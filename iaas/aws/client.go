package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	errwrap "github.com/pkg/errors"
)

//go:generate counterfeiter . Client

type Client interface {
	CreateVM(instance ec2.Instance, ami, instanceType, newName string) (*ec2.Instance, error)
	DeleteVM(instance ec2.Instance) error
	GetVMInfo(name string) (*ec2.Instance, error)
	StopVM(instance ec2.Instance) error
	AssignPublicIP(instance ec2.Instance, ip string) error
	WaitForStartedVM(instanceName string) error
}

type client struct {
	vpcName              string
	clientTimeoutSeconds int
	awsClient            AWSClient
}

func NewClient(configs ...func(*client) error) (Client, error) {
	awsClient := new(client)
	awsClient.clientTimeoutSeconds = 60
	for _, cfg := range configs {
		err := cfg(awsClient)
		if err != nil {
			return nil, errwrap.Wrap(err, "new AWS Client config loading error")
		}
	}

	if awsClient.awsClient == nil {
		return nil, errwrap.New("must configure aws client")
	}
	return awsClient, nil
}

func ConfigAWSClient(value AWSClient) func(*client) error {
	return func(awsClient *client) error {
		awsClient.awsClient = value
		return nil
	}
}

func ConfigVPC(value string) func(*client) error {
	return func(awsClient *client) error {
		awsClient.vpcName = value
		return nil
	}
}

func (s *client) WaitForStartedVM(instanceName string) error {
	errChannel := make(chan error)
	go func() {
		for {
			instance, err := s.GetVMInfo(instanceName)
			if err != nil {
				errChannel <- errwrap.Wrap(err, "GetVMInfo call failed")
			} else {
				if *instance.State.Name == "running" {
					errChannel <- nil
				}
			}
		}
	}()
	select {
	case res := <-errChannel:
		return res
	case <-time.After(time.Second * time.Duration(s.clientTimeoutSeconds)):
		return errwrap.New("polling for status timed out")
	}
}

func (s *client) AssignPublicIP(instance ec2.Instance, ip string) error {
	err := s.awsClient.AssociateElasticIP(*instance.InstanceId, ip)
	if err != nil {
		return errwrap.Wrap(err, "call associateElasticIP on aws client failed")
	}
	return nil
}

func (s *client) CreateVM(instance ec2.Instance, ami, instanceType, name string) (*ec2.Instance, error) {
	securityGroupID := ""
	if len(instance.SecurityGroups) > 0 {
		securityGroupID = *instance.SecurityGroups[0].GroupId
	}
	newInstance, err := s.awsClient.Create(ami, instanceType, name, *instance.KeyName, *instance.SubnetId, securityGroupID)
	if err != nil {
		return nil, errwrap.Wrap(err, "call create on aws client failed")
	}
	return newInstance, nil
}

func (s *client) DeleteVM(instance ec2.Instance) error {
	err := s.awsClient.Delete(*instance.InstanceId)
	if err != nil {
		return errwrap.Wrap(err, "call delete on aws client failed")
	}
	return nil
}

//StopVM - will try to stop the VM
func (s *client) StopVM(instance ec2.Instance) error {
	err := s.awsClient.Stop(*instance.InstanceId)
	if err != nil {
		return errwrap.Wrap(err, "call stop on aws client failed")
	}
	return nil
}

//GetVMInfo - gets the information on the first VM to match the given filter argument
// currently filter will only do a regex on teh tag||name regex fields against
// the List's result set
func (s *client) GetVMInfo(name string) (*ec2.Instance, error) {
	list, err := s.awsClient.List(name, s.vpcName)
	if err != nil {
		return nil, errwrap.Wrap(err, "call List on aws client failed")
	}

	if len(list) == 0 {
		return nil, errwrap.New("No instance matches found")
	}

	if len(list) > 1 {
		return nil, errwrap.New("Found more than one match")
	}
	return list[0], nil
}
