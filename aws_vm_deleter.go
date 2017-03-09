package cliaas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func NewAWSVMDeleter(
	accessKeyID string,
	secretAccessKey string,
	region string,
	vpc string,
) (VMDeleter, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	ec2Client := ec2.New(sess, &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      aws.String(region),
	})

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
