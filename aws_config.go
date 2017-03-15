package cliaas

import (
	"errors"

	errwrap "github.com/pkg/errors"
)

var ErrInvalidConfig = errors.New("invalid configuration")

type AWS struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
	VPCID           string `yaml:"vpc_id"`
}

func (c *AWS) IsValid() bool {
	return c.AccessKeyID != "" &&
		c.SecretAccessKey != "" &&
		c.VPCID != "" &&
		c.Region != ""
}

func (c *AWS) NewReplacer() (VMReplacer, error) {
	ec2Client, err := c.getClient()
	if err != nil {
		return nil, errwrap.Wrap(err, "get EC2 client failed")
	}

	return &awsVM{
		client: NewAWSClient(ec2Client, c.VPCID),
	}, nil
}

func (c *AWS) NewDeleter() (VMDeleter, error) {
	ec2Client, err := c.getClient()
	if err != nil {
		return nil, errwrap.Wrap(err, "get EC2 client failed")
	}

	return &awsVM{
		client: NewAWSClient(ec2Client, c.VPCID),
	}, nil
}

func (c *AWS) getClient() (EC2Client, error) {
	if !c.IsValid() {
		return nil, ErrInvalidConfig
	}

	return NewEC2Client(
		c.AccessKeyID,
		c.SecretAccessKey,
		c.Region,
	)
}
