package cliaas

import "fmt"

type awsVM struct {
	client AWSClient
	ami    string
}

func (v *awsVM) Delete(identifier string) error {
	return v.client.DeleteVM(identifier)
}

func (v *awsVM) Replace(identifier string) error {
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
		v.ami,
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
