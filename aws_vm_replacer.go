package cliaas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func NewAWSVMReplacer(
	accessKeyID string,
	secretAccessKey string,
	region string,
	vpc string,
	ami string,
) (VMReplacer, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	ec2Client := ec2.New(sess, &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      aws.String(region),
	})

	return &awsVMReplacer{
		client: NewAWSClient(ec2Client, vpc),
		ami:    ami,
	}, nil
}

type awsVMReplacer struct {
	client AWSClient
	ami    string
}

func (r *awsVMReplacer) Replace(identifier string) error {
	vmInfo, err := r.client.GetVMInfo(identifier + "*")
	if err != nil {
		return err
	}

	err = r.client.StopVM(vmInfo.InstanceID)
	if err != nil {
		return err
	}

	err = r.client.WaitForStatus(vmInfo.InstanceID, "stopped")
	if err != nil {
		return err
	}

	instanceID, err := r.client.CreateVM(
		r.ami,
		vmInfo.InstanceType,
		identifier,
		vmInfo.KeyName,
		vmInfo.SubnetID,
		vmInfo.SecurityGroupIDs[0],
	)
	if err != nil {
		return err
	}

	err = r.client.WaitForStatus(instanceID, "running")
	if err != nil {
		return err
	}

	return r.client.AssignPublicIP(instanceID, vmInfo.PublicIP)
}
