package cliaas

import (
	"errors"
)

type VMDeleter interface {
	Delete(identifier string) error
}

func NewVMDeleter(config Config) (VMDeleter, error) {
	if config.AWS.AccessKeyID != "" {
		ec2Client, err := NewEC2Client(
			config.AWS.AccessKeyID,
			config.AWS.SecretAccessKey,
			config.AWS.Region,
		)
		if err != nil {
			return nil, err
		}

		return NewAWSVMDeleter(
			NewAWSClient(ec2Client, config.AWS.VPC),
		)
	}

	return nil, errors.New("no vm deleter exists for provided config")
}
