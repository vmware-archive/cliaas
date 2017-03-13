package cliaas

func NewAWSVMReplacer(awsClient AWSClient, ami string) VMReplacer {
	return &awsVMReplacer{
		client: awsClient,
		ami:    ami,
	}
}

type awsVMReplacer struct {
	client AWSClient
	ami    string
}

func (r *awsVMReplacer) Replace(identifier string) error {
	vmInfo, err := r.client.GetVMInfo(identifier + "*")
	if err != nil {
		return err
	}

	err = r.client.StopVM(vmInfo.InstanceID)
	if err != nil {
		return err
	}

	err = r.client.WaitForStatus(vmInfo.InstanceID, "stopped")
	if err != nil {
		return err
	}

	instanceID, err := r.client.CreateVM(
		r.ami,
		vmInfo.InstanceType,
		identifier,
		vmInfo.KeyName,
		vmInfo.SubnetID,
		vmInfo.SecurityGroupIDs[0],
	)
	if err != nil {
		return err
	}

	err = r.client.WaitForStatus(instanceID, "running")
	if err != nil {
		return err
	}

	return r.client.AssignPublicIP(instanceID, vmInfo.PublicIP)
}
