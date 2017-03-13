package cliaas

func NewAWSVMDeleter(
	accessKeyID string,
	secretAccessKey string,
	region string,
	vpc string,
) (VMDeleter, error) {
	ec2Client, err := NewEC2Client(accessKeyID, secretAccessKey, region)
	if err != nil {
		return nil, err
	}

	return &awsVMDeleter{
		client: NewAWSClient(ec2Client, vpc),
	}, nil
}

type awsVMDeleter struct {
	client AWSClient
}

func (d *awsVMDeleter) Delete(identifier string) error {
	return d.client.DeleteVM(identifier)
}
