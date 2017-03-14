package cliaas

import (
	"errors"
	"reflect"

	errwrap "github.com/pkg/errors"
)

type Config struct {
	AWS AWS `yaml:"aws"`
	GCP GCP `yaml:"gcp"`
}

type ValidConfig interface {
	IsValidChecker
	ReplacerDeleter
}

type ReplacerDeleter interface {
	DeleterCreater
	ReplacerCreator
}

type IsValidChecker interface {
	IsValid() bool
}

type DeleterCreater interface {
	NewDeleter() (VMDeleter, error)
}

type ReplacerCreator interface {
	NewReplacer() (VMReplacer, error)
}

type ConfigParser struct {
	Config Config
}

func (s ConfigParser) NewVMDeleter() (VMDeleter, error) {
	deleter, err := s.newVMReplacerDeleter()
	if err != nil {
		return nil, errwrap.Wrap(err, "error in calling newVMReplacerDeleter")
	}
	return deleter.NewDeleter()
}

func (s ConfigParser) NewVMReplacer() (VMReplacer, error) {
	replacer, err := s.newVMReplacerDeleter()
	if err != nil {
		return nil, errwrap.Wrap(err, "error in calling newVMReplacerDeleter")
	}
	return replacer.NewReplacer()
}

func (s ConfigParser) newVMReplacerDeleter() (ReplacerDeleter, error) {
	var replacerDeleters = make([]ReplacerDeleter, 0)
	var configReflect = reflect.ValueOf(s.Config)

	for i := 0; i < configReflect.NumField(); i++ {
		if iaasElement, ok := configReflect.Field(i).Interface().(ValidConfig); ok {
			if iaasElement.IsValid() {
				replacerDeleters = append(replacerDeleters, iaasElement)
			}
		}
	}
	if len(replacerDeleters) == 0 {
		return nil, errors.New("Couldn't find any IaaS objects in your config")
	}

	if len(replacerDeleters) > 1 {
		return nil, errors.New("Found more than one IaaS object in your config")
	}
	return replacerDeleters[0], nil
}
