package cliaas

import (
	"errors"
	"reflect"
)

type Config struct {
	AWS AWS `yaml:"aws"`
	GCP GCP `yaml:"gcp"`
}

type ValidConfig interface {
	IsValidChecker
	DeleterCreater
}

type IsValidChecker interface {
	IsValid() bool
}

type DeleterCreater interface {
	NewDeleter() (VMDeleter, error)
}

type ValidReplacer interface {
	NewReplacer() (VMReplacer, error)
}

type ConfigParser struct {
	Config Config
}

func (s ConfigParser) NewVMDeleter() (VMDeleter, error) {
	var deleterCreaters = make([]DeleterCreater, 0)
	var configReflect = reflect.ValueOf(s.Config)

	for i := 0; i < configReflect.NumField(); i++ {
		if iaasElement, ok := configReflect.Field(i).Interface().(ValidConfig); ok {
			if iaasElement.IsValid() {
				deleterCreaters = append(deleterCreaters, iaasElement)
			}
		}
	}
	if len(deleterCreaters) == 0 {
		return nil, errors.New("Couldn't find any IaaS objects in your config")
	}

	if len(deleterCreaters) > 1 {
		return nil, errors.New("Found more than one IaaS object in your config")
	}
	return deleterCreaters[0].NewDeleter()
}
