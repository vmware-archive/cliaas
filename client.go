package cliaas

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf/cliaas/iaas/aws"
	"github.com/pivotal-cf/cliaas/iaas"
)

type Client interface {
	Delete(vmIdentifier string) error
	Replace(vmIdentifier string, imageIdentifier string, diskSizeGB int64) error
	GetDisk(vmIdentifier string) (iaas.Disk, error)
}

func NewAWSAPIClient(client aws.AWSClient) Client {
	return &awsAPIClient{
		client: client,
	}
}

type awsAPIClient struct {
	client aws.AWSClient
}

func (c *awsAPIClient) Delete(identifier string) error {
	return c.client.DeleteVM(identifier)
}

func (c *awsAPIClient) Replace(identifier string, ami string, diskSizeGB int64) error {
	vmInfo, err := c.client.GetVMInfo(identifier + "*")
	if err != nil {
		return err
	}

	err = c.client.StopVM(vmInfo.InstanceID)
	if err != nil {
		_ = c.client.StartVM(vmInfo.InstanceID)
		return err
	}

	err = c.client.WaitForStatus(vmInfo.InstanceID, ec2.InstanceStateNameStopped)
	if err != nil {
		_ = c.client.StartVM(vmInfo.InstanceID)
		return err
	}

	instanceID, err := c.client.CreateVM(
		ami,
		identifier,
		vmInfo,
	)
	if err != nil {
		_ = c.client.StartVM(vmInfo.InstanceID)
		return err
	}

	err = c.client.WaitForStatus(instanceID, ec2.InstanceStateNameRunning)
	if err != nil {
		_ = c.client.DeleteVM(instanceID)
		return err
	}

	if vmInfo.PublicIP != "" {
		err = c.client.AssignPublicIP(instanceID, vmInfo.PublicIP)
		if err != nil {
			_ = c.client.DeleteVM(instanceID)
			_ = c.client.AssignPublicIP(vmInfo.InstanceID, vmInfo.PublicIP)
			_ = c.client.StartVM(vmInfo.InstanceID)
			return err
		}
	}

	return nil
}

func (c *awsAPIClient) GetDisk(identifier string) (iaas.Disk, error) {
	return iaas.Disk{SizeGB: int64(0)}, nil
}
