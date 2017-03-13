package cliaas

import (
	"errors"
)

type VMReplacer interface {
	Replace(identifier string) error
}

func NewVMReplacer(config Config) (VMReplacer, error) {
	if config.AWS.AccessKeyID != "" {
		ec2Client, err := NewEC2Client(
			config.AWS.AccessKeyID,
			config.AWS.SecretAccessKey,
			config.AWS.Region,
		)
		if err != nil {
			return nil, err
		}

		return NewAWSVMReplacer(
			NewAWSClient(ec2Client, config.AWS.VPCID),
			config.AWS.AMI,
		), nil
	}

	return nil, errors.New("no vm replacer exists for provided config")
}
