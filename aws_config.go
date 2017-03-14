package cliaas

import (
	"errors"

	errwrap "github.com/pkg/errors"
)

type AWS struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
	VPCID           string `yaml:"vpc_id"`
	AMI             string `yaml:"ami"`
}

func (c AWS) IsValid() bool {
	return c.AccessKeyID != "" &&
		c.SecretAccessKey != "" &&
		c.AMI != "" &&
		c.VPCID != "" &&
		c.Region != ""
}

func (c AWS) NewDeleter() (VMDeleter, error) {
	if c.IsValid() == false {
		return nil, errors.New("aws config is not valid")
	}
	ec2Client, err := NewEC2Client(
		c.AccessKeyID,
		c.SecretAccessKey,
		c.Region,
	)
	if err != nil {
		return nil, errwrap.Wrap(err, "unable to create NewEC2Client")
	}

	return NewAWSVMDeleter(
		NewAWSClient(ec2Client, c.VPCID),
	)
}
