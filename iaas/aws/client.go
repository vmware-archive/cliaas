package aws

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/clock"

	"github.com/aws/aws-sdk-go/service/ec2"
	errwrap "github.com/pkg/errors"

	iaasaws "github.com/aws/aws-sdk-go/aws"
)

//go:generate counterfeiter . Client

type Client interface {
	CreateVM(ami, instanceType, name, keyName, subnetID, securityGroupID string) (*ec2.Instance, error)
	DeleteVM(instanceID string) error
	GetVMInfo(name string) (*ec2.Instance, error)
	StopVM(instance ec2.Instance) error
	AssignPublicIP(instance ec2.Instance, ip string) error
	WaitForStatus(instanceID string, status string) error
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

func (c *client) WaitForStatus(instanceID string, status string) error {
	doneCh := make(chan struct{})

	input := &ec2.DescribeInstanceStatusInput{
		IncludeAllInstances: iaasaws.Bool(true),
		InstanceIds: []*string{
			iaasaws.String(instanceID),
		},
	}

	var lastStatus string

	go func() {
		for {
			<-c.clock.After(time.Second)

			output, err := c.awsClient.DescribeInstanceStatus(input)
			if err != nil {
				continue
			}

			if len(output.InstanceStatuses) != 1 {
				continue
			}

			lastStatus = *output.InstanceStatuses[0].InstanceState.Name

			if lastStatus == status {
				close(doneCh)
				return
			}
		}
	}()

	select {
	case <-doneCh:
		return nil
	case <-c.clock.After(c.timeout):
		return errwrap.New(fmt.Sprintf("timed out waiting for instance to become %s (last status was %s)", status, lastStatus))
	}
}

func (c *client) AssignPublicIP(instance ec2.Instance, ip string) error {
	err := c.awsClient.AssociateElasticIP(*instance.InstanceId, ip)
	if err != nil {
		return errwrap.Wrap(err, "call associateElasticIP on aws client failed")
	}
	return nil
}

func (c *client) CreateVM(
	ami string,
	instanceType string,
	name string,
	keyName string,
	subnetID string,
	securityGroupID string,
) (*ec2.Instance, error) {
	newInstance, err := c.awsClient.Create(
		ami,
		instanceType,
		name,
		keyName,
		subnetID,
		securityGroupID,
	)
	if err != nil {
		return nil, errwrap.Wrap(err, "call create on aws client failed")
	}

	return newInstance, nil
}

func (c *client) DeleteVM(instanceID string) error {
	err := c.awsClient.Delete(instanceID)
	if err != nil {
		return errwrap.Wrap(err, "call delete on aws client failed")
	}
	return nil
}

func (c *client) StopVM(instance ec2.Instance) error {
	err := c.awsClient.Stop(*instance.InstanceId)
	if err != nil {
		return errwrap.Wrap(err, "call stop on aws client failed")
	}
	return nil
}

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
