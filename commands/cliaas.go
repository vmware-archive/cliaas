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

	var multiConfig cliaas.MultiConfig
	err = yaml.Unmarshal(bs, &multiConfig)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %s", err)
	}

	completeConfigs := multiConfig.CompleteConfigs()

	if len(completeConfigs) == 0 {
		log.Fatalf("zero iaas configurations exists in config")
	}

	if len(completeConfigs) > 1 {
		log.Fatalf("more than one iaas configuration exists in config")
	}

	Cliaas.Config = completeConfigs[0]

	return nil
}

type CliaasCommand struct {
	Config cliaas.Config

	ConfigFile ConfigFilePath `short:"c" long:"config" required:"true" description:"Path to config file"`

	ReplaceVM        ReplaceVMCommand        `command:"replace-vm" description:"Create a new VM with the old VM's IP"`
	DeleteVM         DeleteVMCommand         `command:"delete-vm" description:"Delete the VM that has the specified identifier"`
	SwapLoadBalancer SwapLoadBalancerCommand `command:"swap-lb-backend" description:"Replace backend instances behind a load balancer"`
}

var Cliaas CliaasCommand
