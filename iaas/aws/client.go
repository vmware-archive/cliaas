package aws

import (
	"time"

	"code.cloudfoundry.org/clock"

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
	awsClient AWSClient
	vpcName   string
	timeout   time.Duration
	clock     clock.Clock
}

func NewClient(
	awsClient AWSClient,
	vpcName string,
	options ...OptionFunc,
) Client {
	client := &client{
		awsClient: awsClient,
		vpcName:   vpcName,
		timeout:   60 * time.Second,
		clock:     clock.NewClock(),
	}

	for _, option := range options {
		option(client)
	}

	return client
}

type OptionFunc func(*client)

func Timeout(timeout time.Duration) OptionFunc {
	return func(c *client) {
		c.timeout = timeout
	}
}

func Clock(clock clock.Clock) OptionFunc {
	return func(c *client) {
		c.clock = clock
	}
}

func (c *client) WaitForStartedVM(instanceName string) error {
	doneCh := make(chan struct{})

	go func() {
		for {
			<-c.clock.After(time.Second)

			instance, err := c.GetVMInfo(instanceName)
			if err != nil {
				continue
			}

			if *instance.State.Name == ec2.InstanceStateNameRunning {
				close(doneCh)
				return
			}
		}
	}()

	select {
	case <-doneCh:
		return nil
	case <-c.clock.After(c.timeout):
		return errwrap.New("polling for status timed out")
	}
}

func (c *client) AssignPublicIP(instance ec2.Instance, ip string) error {
	err := c.awsClient.AssociateElasticIP(*instance.InstanceId, ip)
	if err != nil {
		return errwrap.Wrap(err, "call associateElasticIP on aws client failed")
	}
	return nil
}

func (c *client) CreateVM(instance ec2.Instance, ami, instanceType, name string) (*ec2.Instance, error) {
	securityGroupID := ""
	if len(instance.SecurityGroups) > 0 {
		securityGroupID = *instance.SecurityGroups[0].GroupId
	}
	newInstance, err := c.awsClient.Create(ami, instanceType, name, *instance.KeyName, *instance.SubnetId, securityGroupID)
	if err != nil {
		return nil, errwrap.Wrap(err, "call create on aws client failed")
	}
	return newInstance, nil
}

func (c *client) DeleteVM(instance ec2.Instance) error {
	err := c.awsClient.Delete(*instance.InstanceId)
	if err != nil {
		return errwrap.Wrap(err, "call delete on aws client failed")
	}
	return nil
}

//StopVM - will try to stop the VM
func (c *client) StopVM(instance ec2.Instance) error {
	err := c.awsClient.Stop(*instance.InstanceId)
	if err != nil {
		return errwrap.Wrap(err, "call stop on aws client failed")
	}
	return nil
}

//GetVMInfo - gets the information on the first VM to match the given filter argument
// currently filter will only do a regex on teh tag||name regex fields against
// the List's result set
func (c *client) GetVMInfo(name string) (*ec2.Instance, error) {
	list, err := c.awsClient.List(name, c.vpcName)
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
