package aws

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/clock"

	iaasaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	errwrap "github.com/pkg/errors"
)

//go:generate counterfeiter . EC2Client

type EC2Client interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	DescribeInstanceStatus(*ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error)
	AssociateAddress(*ec2.AssociateAddressInput) (*ec2.AssociateAddressOutput, error)
	TerminateInstances(*ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)
	StopInstances(*ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error)
	CreateTags(*ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error)
	RunInstances(*ec2.RunInstancesInput) (*ec2.Reservation, error)
}

//go:generate counterfeiter . Client

type Client interface {
	CreateVM(ami, instanceType, name, keyName, subnetID, securityGroupID string) (*ec2.Instance, error)
	DeleteVM(instanceID string) error
	GetVMInfo(name string) (*ec2.Instance, error)
	StopVM(instance ec2.Instance) error
	AssignPublicIP(instance, ip string) error
	WaitForStatus(instanceID string, status string) error
}

type client struct {
	ec2Client EC2Client
	vpcName   string
	timeout   time.Duration
	clock     clock.Clock
}

func NewClient(
	ec2Client EC2Client,
	vpcName string,
	options ...OptionFunc,
) Client {
	client := &client{
		ec2Client: ec2Client,
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

			output, err := c.ec2Client.DescribeInstanceStatus(input)
			if err != nil {
				continue
			}

			if len(output.InstanceStatuses) != 1 {
				continue
			}

			instanceStatus := *output.InstanceStatuses[0].InstanceState.Name

			if instanceStatus == status {
				close(doneCh)
				return
			}

			lastStatus = instanceStatus
		}
	}()

	select {
	case <-doneCh:
		return nil
	case <-c.clock.After(c.timeout):
		return errwrap.New(fmt.Sprintf("timed out waiting for instance to become %s (last status was %s)", status, lastStatus))
	}
}

func (c *client) AssignPublicIP(instanceID, ip string) error {
	_, err := c.ec2Client.AssociateAddress(&ec2.AssociateAddressInput{
		InstanceId: iaasaws.String(instanceID),
		PublicIp:   iaasaws.String(ip),
	})

	if err != nil {
		return errwrap.Wrap(err, "associate address failed")
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
	runInput := &ec2.RunInstancesInput{
		ImageId:      iaasaws.String(ami),
		InstanceType: iaasaws.String(instanceType),
		MinCount:     iaasaws.Int64(1),
		MaxCount:     iaasaws.Int64(1),
		KeyName:      iaasaws.String(keyName),
	}

	if subnetID != "" {
		runInput.SubnetId = iaasaws.String(subnetID)
	}

	if securityGroupID != "" {
		runInput.SecurityGroupIds = iaasaws.StringSlice([]string{securityGroupID})
	}

	runResult, err := c.ec2Client.RunInstances(runInput)
	if err != nil {
		return nil, errwrap.Wrap(err, "run instances failed")
	}

	_, err = c.ec2Client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   iaasaws.String("Name"),
				Value: iaasaws.String(name),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return runResult.Instances[0], nil
}

func (c *client) DeleteVM(instanceID string) error {
	_, err := c.ec2Client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			iaasaws.String(instanceID),
		},
		DryRun: iaasaws.Bool(false),
	})

	if err != nil {
		return errwrap.Wrap(err, "terminate instances failed")
	}

	return nil
}

func (c *client) StopVM(instance ec2.Instance) error {
	_, err := c.ec2Client.StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{
			iaasaws.String(*instance.InstanceId),
		},
		DryRun: iaasaws.Bool(false),
		Force:  iaasaws.Bool(true),
	})

	if err != nil {
		return errwrap.Wrap(err, "stop instances failed")
	}

	return nil
}

func (c *client) GetVMInfo(name string) (*ec2.Instance, error) {
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: iaasaws.String("tag:Name"),
				Values: []*string{
					iaasaws.String(name),
				},
			},
			{
				Name: iaasaws.String("vpc-id"),
				Values: []*string{
					iaasaws.String(c.vpcName),
				},
			},
		},
	}
	resp, err := c.ec2Client.DescribeInstances(params)
	if err != nil {
		return nil, errwrap.Wrap(err, "describe instances failed")
	}

	var list []*ec2.Instance

	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			list = append(list, inst)
		}
	}

	if len(list) == 0 {
		return nil, errwrap.New("no matching instances found")
	}

	if len(list) > 1 {
		return nil, errwrap.New("more than one matching instance found")
	}

	return list[0], nil
}
