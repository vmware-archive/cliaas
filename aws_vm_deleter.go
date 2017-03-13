package cliaas

func NewAWSVMDeleter(awsClient AWSClient) (VMDeleter, error) {
	return &awsVMDeleter{
		client: awsClient,
	}, nil
}

type awsVMDeleter struct {
	client AWSClient
}

func (d *awsVMDeleter) Delete(identifier string) error {
	return d.client.DeleteVM(identifier)
}
