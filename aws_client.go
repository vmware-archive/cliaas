package cliaas

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/clock"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	errwrap "github.com/pkg/errors"
)

//go:generate counterfeiter . AWSClient

type AWSClient interface {
	CreateVM(ami, name string, vmInfo VMInfo) (string, error)
	DeleteVM(instanceID string) error
	GetVMInfo(name string) (VMInfo, error)
	StartVM(instanceID string) error
	StopVM(instanceID string) error
	AssignPublicIP(instance, ip string) error
	WaitForStatus(instanceID string, status string) error
	SwapLb(identifier string, vmidentifiers []string) error
}

type client struct {
	ec2Client EC2Client
	elbClient ElbClient
	vpcID     string
	timeout   time.Duration
	clock     clock.Clock
}

func NewAWSClient(
	ec2Client EC2Client,
	elbClient ElbClient,
	vpcID string,
	clock clock.Clock,
) AWSClient {
	client := &client{
		ec2Client: ec2Client,
		elbClient: elbClient,
		vpcID:     vpcID,
		timeout:   60 * time.Second,
		clock:     clock,
	}

	return client
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
	name string,
	vmInfo VMInfo,
) (string, error) {
	runInput := &ec2.RunInstancesInput{
		ImageId:             aws.String(ami),
		InstanceType:        aws.String(vmInfo.InstanceType),
		BlockDeviceMappings: convertBlockDeviceMappings(vmInfo.BlockDeviceMappings),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(vmInfo.IAMInstanceProfileARN),
		},
		MinCount: aws.Int64(1),
		MaxCount: aws.Int64(1),
		KeyName:  aws.String(vmInfo.KeyName),
	}

	if vmInfo.SubnetID != "" {
		runInput.SubnetId = aws.String(vmInfo.SubnetID)
	}

	if len(vmInfo.SecurityGroupIDs) > 0 {
		runInput.SecurityGroupIds = aws.StringSlice(vmInfo.SecurityGroupIDs)
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

func (c *client) StartVM(instanceID string) error {
	_, err := c.ec2Client.StartInstances(&ec2.StartInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
		DryRun: aws.Bool(false),
	})

	if err != nil {
		return errwrap.Wrap(err, "start instances failed")
	}

	return nil
}

type VMInfo struct {
	InstanceID            string
	InstanceType          string
	BlockDeviceMappings   []BlockDeviceMapping
	IAMInstanceProfileARN string
	KeyName               string
	SubnetID              string
	SecurityGroupIDs      []string
	PublicIP              string
}

type BlockDeviceMapping struct {
	DeviceName string
	EBS        EBS
}

type EBS struct {
	DeleteOnTermination bool
	VolumeSize          int64
	VolumeType          string
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
					aws.String(c.vpcID),
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
		for _, instance := range resp.Reservations[idx].Instances {
			if *instance.State.Name == ec2.InstanceStateNameRunning {
				list = append(list, instance)
			}
		}
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
		association := instance.NetworkInterfaces[0].Association
		if association != nil {
			publicIP = *association.PublicIp
		}
	}
	blockDeviceMappings, err := c.describeVolumes(instance.BlockDeviceMappings)
	if err != nil {
		return VMInfo{}, errwrap.Wrap(err, "describeVolumes failure")
	}

	iamInstanceProfileArn := ""
	if instance.IamInstanceProfile != nil {
		iamInstanceProfileArn = *instance.IamInstanceProfile.Arn
	}

	vmInfo := VMInfo{
		InstanceID:            *instance.InstanceId,
		InstanceType:          *instance.InstanceType,
		KeyName:               *instance.KeyName,
		SubnetID:              *instance.SubnetId,
		SecurityGroupIDs:      securityGroupIDs,
		PublicIP:              publicIP,
		BlockDeviceMappings:   blockDeviceMappings,
		IAMInstanceProfileARN: iamInstanceProfileArn,
	}

	return vmInfo, nil
}

func (c *client) describeVolumes(instanceBlockDeviceMappings []*ec2.InstanceBlockDeviceMapping) ([]BlockDeviceMapping, error) {
	blockDeviceMappings := []BlockDeviceMapping{}
	for _, blockDeviceMapping := range instanceBlockDeviceMappings {
		params := &ec2.DescribeVolumesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("volume-id"),
					Values: []*string{
						blockDeviceMapping.Ebs.VolumeId,
					},
				},
			},
		}
		resp, err := c.ec2Client.DescribeVolumes(params)
		if err != nil {
			return nil, errwrap.Wrap(err, "Describe Volume call failed")
		}

		for _, volume := range resp.Volumes {
			blockDeviceMappings = append(blockDeviceMappings, BlockDeviceMapping{
				DeviceName: aws.StringValue(blockDeviceMapping.DeviceName),
				EBS: EBS{
					DeleteOnTermination: aws.BoolValue(blockDeviceMapping.Ebs.DeleteOnTermination),
					VolumeSize:          aws.Int64Value(volume.Size),
					VolumeType:          aws.StringValue(volume.VolumeType),
				},
			})
		}
	}

	return blockDeviceMappings, nil
}
func convertBlockDeviceMappings(blockDeviceMappings []BlockDeviceMapping) []*ec2.BlockDeviceMapping {
	awsBlockDeviceMappings := []*ec2.BlockDeviceMapping{}
	for _, blockDeviceMapping := range blockDeviceMappings {

		awsBlockDeviceMappings = append(awsBlockDeviceMappings, &ec2.BlockDeviceMapping{
			DeviceName: aws.String(blockDeviceMapping.DeviceName),
			Ebs: &ec2.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(blockDeviceMapping.EBS.DeleteOnTermination),
				Encrypted:           nil,
				SnapshotId:          nil,
				VolumeSize:          aws.Int64(blockDeviceMapping.EBS.VolumeSize),
				VolumeType:          aws.String(blockDeviceMapping.EBS.VolumeType),
			},
		})
	}

	return awsBlockDeviceMappings
}

func (c *client) SwapLb(identifier string, vmidentifiers []string) error {
	loadBalancerNames := []*string{&identifier}
	describeLBInput := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: loadBalancerNames,
	}
	describeLbOutput, err := c.elbClient.DescribeLoadBalancers(describeLBInput)
	if err != nil {
		return err
	}
	if len(describeLbOutput.LoadBalancerDescriptions) != 1 {
		return fmt.Errorf("Can not find Load Balancer: %s", identifier)
	}
	oldInstances := describeLbOutput.LoadBalancerDescriptions[0].Instances
	if len(oldInstances) != 0 {
		deregisterInstancesInput := &elb.DeregisterInstancesFromLoadBalancerInput{
			Instances:        oldInstances,
			LoadBalancerName: &identifier,
		}
		_, err = c.elbClient.DeregisterInstancesFromLoadBalancer(deregisterInstancesInput)
		if err != nil {
			return err
		}
	}
	newInstances := make([]*elb.Instance, 0)
	for _, v := range vmidentifiers {
		id := v
		newInstances = append(newInstances, &elb.Instance{InstanceId: &id})
	}
	registerInstancesInput := &elb.RegisterInstancesWithLoadBalancerInput{
		Instances:        newInstances,
		LoadBalancerName: &identifier,
	}
	_, err = c.elbClient.RegisterInstancesWithLoadBalancer(registerInstancesInput)
	if err != nil {
		return err
	}
	return nil
}
