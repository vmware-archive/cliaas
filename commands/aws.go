package commands

import (
	iaasaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/c0-ops/cliaas/iaas/aws"
)

type AWSCommand struct {
	AccessKey    string `long:"accesskey" env:"AWS_ACCESSKEY" description:"aws access key" required:"true"`
	SecretKey    string `long:"secretkey" env:"AWS_SECRETKEY" description:"aws secret access key" required:"true"`
	Region       string `long:"region" default:"us-east-1" env:"AWS_REGION" description:"aws region" required:"true"`
	VPC          string `long:"vpc" env:"AWS_VPC" description:"aws VPC id" required:"true"`
	Name         string `long:"name" env:"AWS_NAME" description:"aws name tag for vm" required:"true"`
	AMI          string `long:"ami" env:"AWS_AMI" description:"aws ami to provision" required:"true"`
	InstanceType string `long:"instanceType" env:"AWS_INSTANCE_TYPE" description:"aws instance type to provision" required:"true"`
	ElasticIP    string `long:"elastic-ip" env:"AWS_ELASTIC_IP" description:"aws elastic ip to associate to provisioned VM" required:"true"`
}

func (c *AWSCommand) Execute([]string) error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	ec2Client := ec2.New(sess, &iaasaws.Config{
		Credentials: credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, ""),
		Region:      iaasaws.String(c.Region),
	})

	client := aws.NewClient(
		aws.NewAWSClient(ec2Client),
		c.VPC,
	)

	opsman, err := aws.NewUpgradeOpsMan(aws.ConfigClient(client))
	if err != nil {
		return err
	}

	return opsman.Upgrade(c.Name, c.AMI, c.InstanceType, c.ElasticIP)
}
