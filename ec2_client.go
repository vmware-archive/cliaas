package cliaas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//go:generate counterfeiter . EC2Client

type EC2Client interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	DescribeInstanceStatus(*ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error)
	AssociateAddress(*ec2.AssociateAddressInput) (*ec2.AssociateAddressOutput, error)
	TerminateInstances(*ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)
	StopInstances(*ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error)
	StartInstances(*ec2.StartInstancesInput) (*ec2.StartInstancesOutput, error)
	CreateTags(*ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error)
	RunInstances(*ec2.RunInstancesInput) (*ec2.Reservation, error)
}

func NewEC2Client(
	accessKeyID string,
	secretAccessKey string,
	region string,
) (EC2Client, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return ec2.New(sess, &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      aws.String(region),
	}), nil
}
