package cliaas

import (
	"errors"
)

type VMDeleter interface {
	Delete(identifier string) error
}

func NewVMDeleter(config Config) (VMDeleter, error) {
	if config.AWS.AccessKeyID != "" {
		return NewAWSVMDeleter(
			config.AWS.AccessKeyID,
			config.AWS.SecretAccessKey,
			config.AWS.Region,
			config.AWS.VPC,
		)
	}

	return nil, errors.New("no vm deleter exists for provided config")
}
