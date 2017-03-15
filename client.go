package cliaas

import (
	"fmt"
	"time"

	"github.com/pivotal-cf/cliaas/iaas"
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

	if vmInfo.PublicIP == "" {
		return fmt.Errorf("instance %s does not have an elastic ip", vmInfo.InstanceID)
	}

	err = v.client.StopVM(vmInfo.InstanceID)
	if err != nil {
		return err
	}

	err = v.client.WaitForStatus(vmInfo.InstanceID, "stopped")
	if err != nil {
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
		return err
	}

	err = v.client.WaitForStatus(instanceID, "running")
	if err != nil {
		return err
	}

	return v.client.AssignPublicIP(instanceID, vmInfo.PublicIP)
}

type gcpClient struct {
	client                *gcp.GCPClientAPI
	sourceImageTarballURL string
}

func (c *gcpClient) Delete(identifier string) error {
	return c.client.DeleteVM(identifier)
}

func (c *gcpClient) Replace(identifier string, sourceImageTarballURL string) error {
	vmInstance, err := c.client.GetVMInfo(iaas.Filter{
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
