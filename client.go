package cliaas

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf/cliaas/iaas/aws"
	"github.com/pivotal-cf/cliaas/iaas/gcp"
	"github.com/pivotal-cf/cliaas/iaas"
	errwrap "github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
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

func NewGCPAPIClient(client gcp.ClientAPI) Client {
	return &gcpClient{
		client: client,
	}
}

type gcpClient struct {
	client gcp.ClientAPI
}

func (c *gcpClient) Delete(identifier string) error {
	return c.client.DeleteVM(identifier)
}

func (c *gcpClient) GetDisk(identifier string) (iaas.Disk, error) {
	disk, err := c.client.GetDisk(gcp.Filter{
		NameRegexString: identifier + "*",
	})
	if err != nil {
		return iaas.Disk{}, errwrap.Wrap(err, "getvminfo failed")
	}
	return iaas.Disk{SizeGB: disk.SizeGb}, nil
}

func (c *gcpClient) Replace(identifier string, sourceImageTarballURL string, diskSizeGB int64) error {
	vmInstance, err := c.client.GetVMInfo(gcp.Filter{
		NameRegexString: identifier + "*",
	})
	if err != nil {
		return errwrap.Wrap(err, "getvminfo failed")
	}

	err = c.client.StopVM(vmInstance.Name)
	if err != nil {
		return errwrap.Wrap(err, "stopvm failed")
	}

	err = c.client.WaitForStatus(vmInstance.Name, gcp.InstanceTerminated)
	if err != nil {
		return errwrap.Wrap(err, "waitforstatus after stopvm failed")
	}

	diskName, err := c.client.CreateImage(sourceImageTarballURL, diskSizeGB)
	if err != nil {
		return errwrap.Wrap(err, "could not create new disk image")
	}

	newInstance := createGCPInstanceFromExisting(vmInstance, diskName, fmt.Sprintf("%s-%s", identifier, time.Now().Format("2006-01-02-15-04-05")))
	err = c.client.CreateVM(*newInstance)
	if err != nil {
		return errwrap.Wrap(err, "CreateVM call failed")
	}

	return c.client.WaitForStatus(newInstance.Name, gcp.InstanceRunning)
}

func createGCPInstanceFromExisting(vmInstance *compute.Instance, diskName string, name string) *compute.Instance {
	newInstance := &compute.Instance{
		NetworkInterfaces: vmInstance.NetworkInterfaces,
		MachineType:       vmInstance.MachineType,
		Name:              name,
		Tags: &compute.Tags{
			Items: vmInstance.Tags.Items,
		},
		Disks: []*compute.AttachedDisk{
			&compute.AttachedDisk{
				Boot: true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: diskName,
				},
			},
		},
	}
	newInstance.NetworkInterfaces[0].NetworkIP = ""
	return newInstance
}