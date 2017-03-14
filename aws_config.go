package cliaas

import errwrap "github.com/pkg/errors"

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

func (c AWS) NewReplacer() (VMReplacer, error) {
	if c.IsValid() == false {
		return nil, InvalidConfigErr{s: "invalid aws config"}
	}

	ec2Client, err := NewEC2Client(
		c.AccessKeyID,
		c.SecretAccessKey,
		c.Region,
	)
	if err != nil {
		return nil, errwrap.Wrap(err, "NewEC2Client creation failed")
	}

	return NewAWSVMReplacer(
		NewAWSClient(ec2Client, c.VPCID),
		c.AMI,
	), nil
}

func (c AWS) NewDeleter() (VMDeleter, error) {
	if c.IsValid() == false {
		return nil, InvalidConfigErr{s: "invalid aws config"}
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

type InvalidConfigErr struct {
	s string
	error
}

func (e InvalidConfigErr) Error() string {
	return e.s
}
func (e InvalidConfigErr) IsInvalidConfig() {}
