package aws

import (
	iaasaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	errwrap "github.com/pkg/errors"
)

//go:generate counterfeiter . AWSClient

type AWSClient interface {
	List(instanceNameRegex, vpcName string) ([]*ec2.Instance, error)
	Stop(instanceID string) error
	Delete(instanceID string) error
	Create(ami, vmType, name, keyPairName, subnetID, securityGroupID string) (*ec2.Instance, error)
	AssociateElasticIP(instanceID, elasticIP string) error
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	DescribeInstanceStatus(input *ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error)
}

type awsClient struct {
	ec2 *ec2.EC2
}

func NewAWSClient(ec2 *ec2.EC2) AWSClient {
	return &awsClient{
		ec2: ec2,
	}
}

func (c *awsClient) Create(ami, vmType, name, keyPairName, subnetID, securityGroupID string) (*ec2.Instance, error) {
	runInput := &ec2.RunInstancesInput{
		ImageId:      iaasaws.String(ami),
		InstanceType: iaasaws.String(vmType),
		MinCount:     iaasaws.Int64(1),
		MaxCount:     iaasaws.Int64(1),
		KeyName:      iaasaws.String(keyPairName),
	}

	if subnetID != "" {
		runInput.SubnetId = iaasaws.String(subnetID)
	}

	if securityGroupID != "" {
		runInput.SecurityGroupIds = iaasaws.StringSlice([]string{securityGroupID})
	}

	runResult, err := c.ec2.RunInstances(runInput)
	if err != nil {
		return nil, err
	}

	_, err = c.ec2.CreateTags(&ec2.CreateTagsInput{
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

func (c *awsClient) List(instanceNameRegex, vpcName string) ([]*ec2.Instance, error) {
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: iaasaws.String("tag:Name"),
				Values: []*string{
					iaasaws.String(instanceNameRegex),
				},
			},
			{
				Name: iaasaws.String("vpc-id"),
				Values: []*string{
					iaasaws.String(vpcName),
				},
			},
		},
	}
	resp, err := c.ec2.DescribeInstances(params)
	if err != nil {
		return nil, errwrap.Wrap(err, "DescribeInstances yielded error")
	}

	var instances []*ec2.Instance

	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			instances = append(instances, inst)
		}
	}
	return instances, nil
}

func (c *awsClient) Stop(instanceID string) error {
	_, err := c.ec2.StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{
			iaasaws.String(instanceID),
		},
		DryRun: iaasaws.Bool(false),
		Force:  iaasaws.Bool(true),
	})
	return err
}

func (c *awsClient) Delete(instanceID string) error {
	_, err := c.ec2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			iaasaws.String(instanceID),
		},
		DryRun: iaasaws.Bool(false),
	})
	return err
}

func (c *awsClient) AssociateElasticIP(instanceID, elasticIP string) error {
	_, err := c.ec2.AssociateAddress(&ec2.AssociateAddressInput{
		InstanceId: iaasaws.String(instanceID),
		PublicIp:   iaasaws.String(elasticIP),
	})
	return err
}

func (c *awsClient) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return c.ec2.DescribeInstances(input)
}

func (c *awsClient) DescribeInstanceStatus(input *ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error) {
	return c.ec2.DescribeInstanceStatus(input)
}
