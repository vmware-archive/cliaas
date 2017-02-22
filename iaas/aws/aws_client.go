package aws

import (
	"errors"

	iaasaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/c0-ops/cliaas/iaas"
	errwrap "github.com/pkg/errors"
)

type AWSClient interface {
	List(instanceNameRegex, vpcName string) ([]*ec2.Instance, error)
	Stop(instanceID string) error
	Delete(instanceID string) error
}

type ClientAPI interface {
	CreateVM(instance ec2.Instance) error
	DeleteVM(instance ec2.Instance) error
	GetVMInfo(filter iaas.Filter) (*ec2.Instance, error)
	StopVM(instance ec2.Instance) error
}

type AWSClientAPI struct {
	vpcName   string
	awsClient AWSClient
}

type awsClientWrapper struct {
	ec2 *ec2.EC2
}

func NewAWSClientAPI(configs ...func(*AWSClientAPI) error) (*AWSClientAPI, error) {
	awsClient := new(AWSClientAPI)

	for _, cfg := range configs {
		err := cfg(awsClient)
		if err != nil {
			return nil, errwrap.Wrap(err, "new AWS Client config loading error")
		}
	}

	if awsClient.awsClient == nil {
		return nil, errors.New("must configure aws client")
	}
	return awsClient, nil
}

func CreateAWSClient(region, accessKey, secretKey string) (AWSClient, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return &awsClientWrapper{ec2: ec2.New(sess, &iaasaws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Region:      iaasaws.String(region),
	})}, nil
}

func ConfigAWSClient(value AWSClient) func(*AWSClientAPI) error {
	return func(awsClient *AWSClientAPI) error {
		awsClient.awsClient = value
		return nil
	}
}

func ConfigVPC(value string) func(*AWSClientAPI) error {
	return func(awsClient *AWSClientAPI) error {
		awsClient.vpcName = value
		return nil
	}
}

func (s *AWSClientAPI) CreateVM(instance ec2.Instance) error {

	return errors.New("Not implmented")
}

func (s *AWSClientAPI) DeleteVM(instance ec2.Instance) error {
	err := s.awsClient.Delete(*instance.InstanceId)
	if err != nil {
		return errwrap.Wrap(err, "call delete on aws client failed")
	}
	return nil
}

//StopVM - will try to stop the VM
func (s *AWSClientAPI) StopVM(instance ec2.Instance) error {
	err := s.awsClient.Stop(*instance.InstanceId)
	if err != nil {
		return errwrap.Wrap(err, "call stop on aws client failed")
	}
	return nil
}

//GetVMInfo - gets the information on the first VM to match the given filter argument
// currently filter will only do a regex on teh tag||name regex fields against
// the List's result set
func (s *AWSClientAPI) GetVMInfo(filter iaas.Filter) (*ec2.Instance, error) {
	list, err := s.awsClient.List(filter.NameRegexString, s.vpcName)
	if err != nil {
		return nil, errwrap.Wrap(err, "call List on aws client failed")
	}

	if len(list) == 0 {
		return nil, errors.New("No instance matches found")
	}

	if len(list) > 1 {
		return nil, errors.New("Found more than one match")
	}
	return list[0], nil
}

func (s *awsClientWrapper) List(instanceNameRegex, vpcName string) ([]*ec2.Instance, error) {
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: iaasaws.String("tag:Name"),
				Values: []*string{
					iaasaws.String(instanceNameRegex),
				},
			},
			{
				Name: iaasaws.String("vpc-id"),
				Values: []*string{
					iaasaws.String(vpcName),
				},
			},
		},
	}
	resp, err := s.ec2.DescribeInstances(params)
	if err != nil {
		return nil, errwrap.Wrap(err, "DescribeInstances yielded error")
	}

	var instances []*ec2.Instance

	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			instances = append(instances, inst)
		}
	}
	return instances, nil
}

func (s *awsClientWrapper) Stop(instanceID string) error {
	_, err := s.ec2.StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{
			iaasaws.String(instanceID),
		},
		DryRun: iaasaws.Bool(false),
		Force:  iaasaws.Bool(true),
	})
	return err
}

func (s *awsClientWrapper) Delete(instanceID string) error {
	_, err := s.ec2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			iaasaws.String(instanceID),
		},
		DryRun: iaasaws.Bool(false),
	})
	return err
}
