package cliaas

import (
	"errors"
	"fmt"
	"time"

	"github.com/pivotal-cf/cliaas/iaas/gcp"
	errwrap "github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"
)

type Client interface {
	Delete(vmIdentifier string) error
	Replace(vmIdentifier string, imageIdentifier string) error
}

type awsClient struct {
	client AWSClient
}

func (v *awsClient) Delete(identifier string) error {
	return v.client.DeleteVM(identifier)
}

func (v *awsClient) Replace(identifier string, ami string) error {
	vmInfo, err := v.client.GetVMInfo(identifier + "*")
	if err != nil {
		return err
	}

	err = v.client.StopVM(vmInfo.InstanceID)
	if err != nil {
		_ = v.client.StartVM(vmInfo.InstanceID)
		return err
	}

	err = v.client.WaitForStatus(vmInfo.InstanceID, "stopped")
	if err != nil {
		_ = v.client.StartVM(vmInfo.InstanceID)
		return err
	}

	instanceID, err := v.client.CreateVM(
		ami,
		vmInfo.InstanceType,
		identifier,
		vmInfo.KeyName,
		vmInfo.SubnetID,
		vmInfo.SecurityGroupIDs[0],
	)
	if err != nil {
		_ = v.client.StartVM(vmInfo.InstanceID)
		return err
	}

	err = v.client.WaitForStatus(instanceID, "running")
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

type azureClient struct{}

func (c *azureClient) Delete(identifier string) error {
	return errors.New("not yet implemented")
}

func (c *azureClient) Replace(identifier string, sourceImageTarballURL string) error {
	return errors.New("not yet implemented")
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

	err = c.client.WaitForStatus(vmInstance.Name, "stopped")
	if err != nil {
		return errwrap.Wrap(err, "waitforstatus after stopvm failed")
	}

	vmInstance.Name = fmt.Sprintf("%s-%s", identifier, time.Now().Format("2006-01-02_15-04-05"))
	vmInstance.Disks = []*compute.AttachedDisk{
		&compute.AttachedDisk{
			Source: sourceImageTarballURL,
		},
	}
	err = c.client.CreateVM(*vmInstance)
	if err != nil {
		return errwrap.Wrap(err, "CreateVM call failed")
	}

	return c.client.WaitForStatus(vmInstance.Name, "running")
}
