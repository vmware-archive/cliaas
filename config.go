package cliaas

import (
	"os"
	"reflect"

	"github.com/pivotal-cf/cliaas/iaas/gcp"
	errwrap "github.com/pkg/errors"
)

type Config interface {
	Complete() bool
	NewClient() (Client, error)
}

type MultiConfig struct {
	AWS *AWSConfig `yaml:"aws"`
	GCP *GCPConfig `yaml:"gcp"`
}

func (c *MultiConfig) CompleteConfigs() []Config {
	typ := reflect.ValueOf(c)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	configs := []Config{}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		iface := field.Interface()

		if config, ok := iface.(Config); ok {
			if !field.IsNil() && config.Complete() {
				configs = append(configs, config)
			}
		}
	}

	return configs
}

type AWSConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
	VPCID           string `yaml:"vpc_id"`
}

func (c *AWSConfig) Complete() bool {
	return c.AccessKeyID != "" &&
		c.SecretAccessKey != "" &&
		c.VPCID != "" &&
		c.Region != ""
}

func (c *AWSConfig) NewClient() (Client, error) {
	ec2Client, err := NewEC2Client(c.AccessKeyID, c.SecretAccessKey, c.Region)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to make ec2 client")
	}

	return &awsClient{
		client: NewAWSClient(ec2Client, c.VPCID),
	}, nil
}

type GCPConfig struct {
	CredfilePath string `yaml:"credfile"`
	Zone         string `yaml:"zone"`
	Project      string `yaml:"project"`
	DiskImageURL string `yaml:"disk_image_url"`
}

func (c *GCPConfig) Complete() bool {
	_, err := os.Stat(c.CredfilePath)
	if err != nil {
		return false
	}

	return c.CredfilePath != "" &&
		c.Zone != "" &&
		c.Project != "" &&
		c.DiskImageURL != ""
}

func (c *GCPConfig) NewClient() (Client, error) {
	computeClient, err := gcp.NewDefaultGoogleComputeClient(c.CredfilePath)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to create gcp default client")
	}

	gcpClientAPI, err := gcp.NewGCPClientAPI(
		gcp.ConfigGoogleClient(computeClient),
		gcp.ConfigZoneName(c.Zone),
		gcp.ConfigProjectName(c.Project),
	)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to create gcp client api")
	}

	return &gcpClient{
		client: gcpClientAPI,
	}, nil
}
