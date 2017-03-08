package commands

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/pivotal-cf/cliaas"

	yaml "gopkg.in/yaml.v2"
)

type ConfigFilePath string

func (c *ConfigFilePath) UnmarshalFlag(value string) error {
	bs, err := ioutil.ReadFile(value)
	if err != nil {
		return fmt.Errorf("failed to read config: %s", err)
	}

	var config cliaas.Config
	err = yaml.Unmarshal(bs, &config)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %s", err)
	}

	Cliaas.Config = config

	return nil
}

type CliaasCommand struct {
	Config cliaas.Config

	ConfigFile ConfigFilePath   `short:"c" long:"config" required:"true" description:"Path to config file"`
	ReplaceVM  ReplaceVMCommand `command:"replace-vm" description:"Create a new VM with the old VM's IP"`
}

var Cliaas CliaasCommand
