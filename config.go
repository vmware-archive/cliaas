package cliaas

import (
	"errors"
	"os"

	"code.cloudfoundry.org/clock"

	"github.com/pivotal-cf/cliaas/iaas/gcp"
	errwrap "github.com/pkg/errors"
)

type Config interface {
	Complete() bool
	NewClient() (Client, error)
}

type MultiConfig struct {
	AWS   *AWSConfig   `yaml:"aws"`
	GCP   *GCPConfig   `yaml:"gcp"`
	Azure *AzureConfig `yaml:"azure"`
}

func (c *MultiConfig) Configs() []Config {
	var configs []Config

	if c.AWS != nil {
		configs = append(configs, c.AWS)
	}

	if c.GCP != nil {
		configs = append(configs, c.GCP)
	}

	if c.Azure != nil {
		configs = append(configs, c.Azure)
	}
	return configs

}

func (c *MultiConfig) CompleteConfigs() []Config {
	configs := c.Configs()

	var completeConfigs []Config
	for i := range configs {
		if configs[i].Complete() {
			completeConfigs = append(completeConfigs, configs[i])
		}
	}

	return completeConfigs
}

type AzureConfig struct {
	SubscriptionID    string `yaml:"subscription_id"`
	ClientID          string `yaml:"client_id"`
	ClientSecret      string `yaml:"client_secret"`
	TenantID          string `yaml:"tenant_id"`
	ResourceGroupName string `yaml:"resource_group_name"`
}

func (c *AzureConfig) Complete() bool {
	return c.SubscriptionID != "" &&
		c.ClientID != "" &&
		c.ClientSecret != "" &&
		c.TenantID != "" &&
		c.ResourceGroupName != ""
}

func (c *AzureConfig) NewClient() (Client, error) {
	return nil, errors.New("not yet implemented")
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
		client: NewAWSClient(ec2Client, c.VPCID, clock.NewClock()),
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

	gcpClientAPI, err := gcp.NewClient(
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
