package cliaas

import (
	"os"

	"code.cloudfoundry.org/clock"

	"github.com/Azure/azure-storage-go"
	"github.com/pivotal-cf/cliaas/iaas/azure"
	"github.com/pivotal-cf/cliaas/iaas/gcp"
	errwrap "github.com/pkg/errors"
)

type Config interface {
	Image() string
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
	VHDImageURL             string `yaml:"vhd_image_url"`
	SubscriptionID          string `yaml:"subscription_id"`
	ClientID                string `yaml:"client_id"`
	ClientSecret            string `yaml:"client_secret"`
	TenantID                string `yaml:"tenant_id"`
	ResourceGroupName       string `yaml:"resource_group_name"`
	ResourceManagerEndpoint string `yaml:"resource_manager_endpoint"`
	StorageAccountName      string `yaml:"storage_account_name"`
	StorageAccountKey       string `yaml:"storage_account_key"`
	StorageContainerName    string `yaml:"storage_container_name"`
	StorageURL              string `yaml:"storage_url"`
	VMAdminPassword         string `yaml:"vm_admin_password"`
}

func (c *AzureConfig) Image() string {
	return c.VHDImageURL
}

func (c *AzureConfig) Complete() bool {
	return c.SubscriptionID != "" &&
		c.ClientID != "" &&
		c.ClientSecret != "" &&
		c.TenantID != "" &&
		c.ResourceGroupName != "" &&
		c.StorageAccountName != "" &&
		c.StorageAccountKey != "" &&
		c.VHDImageURL != "" &&
		c.StorageContainerName != ""
}

func (c *AzureConfig) NewClient() (Client, error) {
	client, err := azure.NewClient(c.SubscriptionID, c.ClientID, c.ClientSecret, c.TenantID, c.ResourceGroupName, c.ResourceManagerEndpoint)
	if err != nil {
		return nil, errwrap.Wrap(err, "azure newclient failed to create a client")
	}

	if c.StorageURL == "" {
		c.StorageURL = storage.DefaultBaseURL
	}
	client.SetStorageContainerName(c.StorageContainerName)
	client.SetStorageAccountName(c.StorageAccountName)
	client.SetStorageBaseURL(c.StorageURL)
	client.SetVMAdminPassword(c.VMAdminPassword)
	err = client.SetBlobServiceClient(c.StorageAccountName, c.StorageAccountKey, c.StorageURL)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed setting blobstore client")
	}

	return client, nil
}

type AWSConfig struct {
	AMI             string `yaml:"ami"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
	VPCID           string `yaml:"vpc"`
}

func (c *AWSConfig) Image() string {
	return c.AMI
}

func (c *AWSConfig) Complete() bool {
	return c.AccessKeyID != "" &&
		c.SecretAccessKey != "" &&
		c.VPCID != "" &&
		c.AMI != "" &&
		c.Region != ""
}

func (c *AWSConfig) NewClient() (Client, error) {
	ec2Client, err := NewEC2Client(c.AccessKeyID, c.SecretAccessKey, c.Region)
	elbClient, err := NewElbClient(c.AccessKeyID, c.SecretAccessKey, c.Region)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to make ec2 client")
	}

	return NewAWSAPIClientAdaptor(
		NewAWSClient(ec2Client, elbClient, c.VPCID, clock.NewClock())), nil
}

type GCPConfig struct {
	CredfilePath string `yaml:"credfile"`
	Zone         string `yaml:"zone"`
	Project      string `yaml:"project"`
	DiskImageURL string `yaml:"disk_image_url"`
}

func (c *GCPConfig) Image() string {
	return c.DiskImageURL
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
		gcp.ConfigTimeout(600),
	)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to create gcp client api")
	}

	return &gcpClient{
		client: gcpClientAPI,
	}, nil
}
