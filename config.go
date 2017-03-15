package cliaas

import "errors"

type Config struct {
	AWS AWS `yaml:"aws"`
	GCP GCP `yaml:"gcp"`
}

func (c Config) NewVMDeleter() (VMDeleter, error) {

	switch {
	case c.AWS.IsValid() && c.GCP.IsValid():
		return nil, errors.New("You've given a config which defines more than one iaas. This is not allowed")

	case c.AWS.IsValid():
		return c.AWS.NewDeleter()

	case c.GCP.IsValid():
		return c.GCP.NewDeleter()
	}

	return nil, errors.New("no vm deleter exists for provided config")
}

func (c Config) NewVMReplacer() (VMReplacer, error) {

	switch {
	case c.AWS.IsValid() && c.GCP.IsValid():
		return nil, errors.New("You've given a config which defines more than one iaas. This is not allowed")

	case c.AWS.IsValid():
		return c.AWS.NewReplacer()

	case c.GCP.IsValid():
		return c.GCP.NewReplacer()
	}

	return nil, errors.New("no vm replacer exists for provided config")
}
