package cliaas

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf/cliaas/iaas/gcp"
	errwrap "github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
)

type Client interface {
	Delete(vmIdentifier string) error
	Replace(vmIdentifier string, imageIdentifier string) error
	SwapLb(identifier string, vmidentifiers []string) error
}

func NewAWSAPIClientAdaptor(client AWSClient) Client {
	return &awsAPIClientAdaptor{
		client: client,
	}
}

type awsAPIClientAdaptor struct {
	client AWSClient
}

func (v *awsAPIClientAdaptor) Delete(identifier string) error {
	return v.client.DeleteVM(identifier)
}

func (v *awsAPIClientAdaptor) Replace(identifier string, ami string) error {
	vmInfo, err := v.client.GetVMInfo(identifier + "*")
	if err != nil {
		return err
	}

	err = v.client.StopVM(vmInfo.InstanceID)
	if err != nil {
		_ = v.client.StartVM(vmInfo.InstanceID)
		return err
	}

	err = v.client.WaitForStatus(vmInfo.InstanceID, ec2.InstanceStateNameStopped)
	if err != nil {
		_ = v.client.StartVM(vmInfo.InstanceID)
		return err
	}

	instanceID, err := v.client.CreateVM(
		ami,
		identifier,
		vmInfo,
	)
	if err != nil {
		_ = v.client.StartVM(vmInfo.InstanceID)
		return err
	}

	err = v.client.WaitForStatus(instanceID, ec2.InstanceStateNameRunning)
	if err != nil {
		_ = v.client.DeleteVM(instanceID)
		return err
	}

	if vmInfo.PublicIP != "" {
		err = v.client.AssignPublicIP(instanceID, vmInfo.PublicIP)
		if err != nil {
			_ = v.client.DeleteVM(instanceID)
			_ = v.client.AssignPublicIP(vmInfo.InstanceID, vmInfo.PublicIP)
			_ = v.client.StartVM(vmInfo.InstanceID)
			return err
		}
	}

	return nil
}

func (v *awsAPIClientAdaptor) SwapLb(identifier string, vmidentifiers []string) error {
	return v.client.SwapLb(identifier, vmidentifiers)
}

type gcpClient struct {
	client *gcp.Client
}

func (c *gcpClient) Delete(identifier string) error {
	return c.client.DeleteVM(identifier)
}

func (c *gcpClient) Replace(identifier string, sourceImageTarballURL string) error {
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

	diskName, err := c.client.CreateImage(sourceImageTarballURL)
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

func (c *gcpClient) SwapLb(identifier string, vmidentifiers []string) error {
	return c.client.SwapLb(identifier, vmidentifiers)
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
