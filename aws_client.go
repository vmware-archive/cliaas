package cliaas

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/clock"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	errwrap "github.com/pkg/errors"
)

//go:generate counterfeiter . AWSClient

type AWSClient interface {
	CreateVM(ami, instanceType, name, keyName, subnetID, securityGroupID string) (string, error)
	DeleteVM(instanceID string) error
	GetVMInfo(name string) (VMInfo, error)
	StopVM(instanceID string) error
	AssignPublicIP(instance, ip string) error
	WaitForStatus(instanceID string, status string) error
}

type client struct {
	ec2Client EC2Client
	vpcName   string
	timeout   time.Duration
	clock     clock.Clock
}

func NewAWSClient(
	ec2Client EC2Client,
	vpcName string,
	options ...OptionFunc,
) AWSClient {
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
		IncludeAllInstances: aws.Bool(true),
		InstanceIds: []*string{
			aws.String(instanceID),
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
		InstanceId: aws.String(instanceID),
		PublicIp:   aws.String(ip),
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
) (string, error) {
	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String(ami),
		InstanceType: aws.String(instanceType),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      aws.String(keyName),
	}

	if subnetID != "" {
		runInput.SubnetId = aws.String(subnetID)
	}

	if securityGroupID != "" {
		runInput.SecurityGroupIds = aws.StringSlice([]string{securityGroupID})
	}

	runResult, err := c.ec2Client.RunInstances(runInput)
	if err != nil {
		return "", errwrap.Wrap(err, "run instances failed")
	}

	_, err = c.ec2Client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
		},
	})
	if err != nil {
		return "", err
	}

	return *runResult.Instances[0].InstanceId, nil
}

func (c *client) DeleteVM(instanceID string) error {
	_, err := c.ec2Client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
		DryRun: aws.Bool(false),
	})

	if err != nil {
		return errwrap.Wrap(err, "terminate instances failed")
	}

	return nil
}

func (c *client) StopVM(instanceID string) error {
	_, err := c.ec2Client.StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
		DryRun: aws.Bool(false),
		Force:  aws.Bool(true),
	})

	if err != nil {
		return errwrap.Wrap(err, "stop instances failed")
	}

	return nil
}

type VMInfo struct {
	InstanceID       string
	InstanceType     string
	KeyName          string
	SubnetID         string
	SecurityGroupIDs []string
	PublicIP         string
}

func (c *client) GetVMInfo(name string) (VMInfo, error) {
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(name),
				},
			},
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(c.vpcName),
				},
			},
		},
	}
	resp, err := c.ec2Client.DescribeInstances(params)
	if err != nil {
		return VMInfo{}, errwrap.Wrap(err, "describe instances failed")
	}

	var list []*ec2.Instance

	for idx := range resp.Reservations {
		list = append(list, resp.Reservations[idx].Instances...)
	}

	if len(list) == 0 {
		return VMInfo{}, errwrap.New("no matching instances found")
	}

	if len(list) > 1 {
		return VMInfo{}, errwrap.New("more than one matching instance found")
	}

	instance := list[0]

	var securityGroupIDs []string
	for _, sg := range instance.SecurityGroups {
		securityGroupIDs = append(securityGroupIDs, *sg.GroupId)
	}

	var publicIP string
	if len(instance.NetworkInterfaces) > 0 {
		publicIP = *instance.NetworkInterfaces[0].Association.PublicIp
	}

	vmInfo := VMInfo{
		InstanceID:       *instance.InstanceId,
		InstanceType:     *instance.InstanceType,
		KeyName:          *instance.KeyName,
		SubnetID:         *instance.SubnetId,
		SecurityGroupIDs: securityGroupIDs,
		PublicIP:         publicIP,
	}

	return vmInfo, nil
}
