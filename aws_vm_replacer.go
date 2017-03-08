package cliaas

import (
	iaasaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/c0-ops/cliaas/iaas/aws"
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

	ec2Client := ec2.New(sess, &iaasaws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      iaasaws.String(region),
	})

	return &awsVMReplacer{
		client: aws.NewClient(ec2Client, vpc),
		ami:    ami,
	}, nil
}

type awsVMReplacer struct {
	client aws.Client
	ami    string
}

func (r *awsVMReplacer) Replace(identifier string) error {
	instance, err := r.client.GetVMInfo(identifier + "*")
	if err != nil {
		return err
	}

	err = r.client.StopVM(*instance)
	if err != nil {
		return err
	}

	err = r.client.WaitForStatus(*instance.InstanceId, "stopped")
	if err != nil {
		return err
	}

	var keyName string
	if instance.KeyName != nil {
		keyName = *instance.KeyName
	}

	var subnetID string
	if instance.SubnetId != nil {
		subnetID = *instance.SubnetId
	}

	var securityGroupID string
	if len(instance.SecurityGroups) > 0 {
		securityGroupID = *instance.SecurityGroups[0].GroupId
	}

	newInstance, err := r.client.CreateVM(
		r.ami,
		*instance.InstanceType,
		identifier,
		keyName,
		subnetID,
		securityGroupID,
	)
	if err != nil {
		return err
	}

	err = r.client.WaitForStatus(*newInstance.InstanceId, "running")
	if err != nil {
		return err
	}

	return r.client.AssignPublicIP(*newInstance, *instance.NetworkInterfaces[0].Association.PublicIp)
}
