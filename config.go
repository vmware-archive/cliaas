package cliaas

import (
	"errors"

	errwrap "github.com/pkg/errors"

	"github.com/pivotal-cf/cliaas/iaas/gcp"
)

type Config struct {
	AWS struct {
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
		Region          string `yaml:"region"`
		VPCID           string `yaml:"vpc_id"`
		AMI             string `yaml:"ami"`
	} `yaml:"aws"`

	GCP struct {
		CredfilePath string `yaml:"credfile"`
		Zone         string `yaml:"zone"`
		Project      string `yaml:"project"`
		DiskImageURL string `yaml:"disk_image_url"`
	} `yaml:"gcp"`
}

func (c *Config) ForAWS() bool {
	return c.AWS.AccessKeyID != ""
}

func (c *Config) ForGCP() bool {
	return c.GCP.CredfilePath != ""
}

func (c *Config) NewVMDeleter() (VMDeleter, error) {
	switch {
	case c.ForAWS():
		awsClient, err := c.newAWSClient()
		if err != nil {
			return nil, err
		}

		return NewAWSVMDeleter(awsClient)
	case c.ForGCP():
		gcpClientAPI, err := c.newGCPClient()
		if err != nil {
			return nil, errwrap.Wrap(err, "Failed to create new GCP API client")
		}

		return NewGCPVMDeleter(gcpClientAPI)
	}

	return nil, errors.New("no vm deleter exists for provided config")
}

func (c *Config) NewVMReplacer() (VMReplacer, error) {
	switch {
	case c.ForAWS():
		awsClient, err := c.newAWSClient()
		if err != nil {
			return nil, err
		}

		return NewAWSVMReplacer(awsClient, c.AWS.AMI), nil
	case c.ForGCP():
		gcpClientAPI, err := c.newGCPClient()
		if err != nil {
			return nil, errwrap.Wrap(err, "Failed to create new GCP API client")
		}

		return NewGCPVMReplacer(gcpClientAPI)
	}

	return nil, errors.New("no vm replacer exists for provided config")
}

func (c *Config) newAWSClient() (AWSClient, error) {
	ec2Client, err := NewEC2Client(
		c.AWS.AccessKeyID,
		c.AWS.SecretAccessKey,
		c.AWS.Region,
	)
	if err != nil {
		return nil, err
	}

	return NewAWSClient(ec2Client, c.AWS.VPCID), nil
}

func (c *Config) newGCPClient() (*gcp.GCPClientAPI, error) {
	gcpClient, err := gcp.NewDefaultGoogleComputeClient(c.GCP.CredfilePath)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to create gcp default client")
	}

	gcpClientAPI, err := gcp.NewGCPClientAPI(
		gcp.ConfigGoogleClient(gcpClient),
		gcp.ConfigZoneName(c.GCP.Zone),
		gcp.ConfigProjectName(c.GCP.Project),
	)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to create gcp client api")
	}

	return gcpClientAPI, nil
}
