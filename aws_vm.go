package cliaas

func NewAWSVMReplacer(awsClient AWSClient, ami string) VMReplacer {
	return &awsVM{
		client: awsClient,
		ami:    ami,
	}
}

func NewAWSVMDeleter(awsClient AWSClient) (VMDeleter, error) {
	return &awsVM{
		client: awsClient,
	}, nil
}

type awsVM struct {
	client AWSClient
	ami    string
}

func (d *awsVM) Delete(identifier string) error {
	return d.client.DeleteVM(identifier)
}

func (r *awsVM) Replace(identifier string) error {
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