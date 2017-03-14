package cliaas

import (
	"errors"
	"reflect"
)

type Config struct {
	AWS AWS `yaml:"aws"`
	GCP GCP `yaml:"gcp"`
}

type ValidDeleter interface {
	IsValid() bool
	NewDeleter() (VMDeleter, error)
}

type ValidReplacer interface {
	IsValid() bool
	NewReplacer() (VMReplacer, error)
}

type ConfigParser struct {
	Config Config
}

func (s ConfigParser) GetValidDeleters() ([]ValidDeleter, error) {
	var validDeleters = make([]ValidDeleter, 0)
	var configReflect = reflect.ValueOf(s.Config)

	for i := 0; i < configReflect.NumField(); i++ {
		if iaasElement, ok := configReflect.Field(i).Interface().(ValidDeleter); ok {
			validDeleters = append(validDeleters, iaasElement)
		}
	}
	if len(validDeleters) == 0 {
		return nil, errors.New("Couldn't find any IaaS objects in your config")
	}
	return validDeleters, nil
}
