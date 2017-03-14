package cliaas

import (
	"errors"

	errwrap "github.com/pkg/errors"
)

type VMDeleter interface {
	Delete(identifier string) error
}

func NewVMDeleter(config Config) (VMDeleter, error) {
	var configParser = ConfigParser{Config: config}
	var iaasConfigs, err = configParser.GetValidDeleters()
	if err != nil {
		return nil, errwrap.Wrap(err, "attempt to get iaas related config elements failed")
	}

	for _, iaasConfig := range iaasConfigs {
		if iaasConfig.IsValid() {
			return iaasConfig.NewDeleter()
		}
	}
	return nil, errors.New("no vm deleter exists for provided config")
}
