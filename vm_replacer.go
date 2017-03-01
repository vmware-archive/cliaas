package cliaas

import (
	"errors"
)

type VMReplacer interface {
	Replace(identifier string) error
}

func NewVMReplacer(config Config) (VMReplacer, error) {
	if config.AWS.AccessKeyID != "" {
		return NewAWSVMReplacer(
			config.AWS.AccessKeyID,
			config.AWS.SecretAccessKey,
			config.AWS.Region,
			config.AWS.VPC,
			config.AWS.AMI,
		)
	}

	return nil, errors.New("no vm replacer exists for provided config")
}
