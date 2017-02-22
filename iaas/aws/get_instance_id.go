package aws

import (
	"fmt"

	iaasaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	errwrap "github.com/pkg/errors"
)

var (
	ErrCouldNotFindInstance = fmt.Errorf("couldnt find instance match")
)

type Client struct {
	InstanceNameRegex string
	VPCName           string
	Describer         InstanceDescriber
}

type InstanceDescriber interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

func createErrWrap(err error, msg string) error {
	err = errwrap.Wrap(err, msg)

	if err == nil {
		err = fmt.Errorf(msg)
	}
	return err
}

func NewClient(region string, instanceNameRegex string, vpcname string) (*Client, error) {
	var err error

	if region == "" {
		var msg = "no region configured"
		err = createErrWrap(err, msg)
	}

	if instanceNameRegex == "" {
		var msg = "no instanceNameRegex configured"
		err = createErrWrap(err, msg)
	}

	if vpcname == "" {
		var msg = "no vpcname configured"
		err = createErrWrap(err, msg)
	}

	if err != nil {
		return nil, err
	}

	client := &Client{
		InstanceNameRegex: instanceNameRegex,
		VPCName:           vpcname,
	}

	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	client.Describer = ec2.New(sess, &iaasaws.Config{Region: iaasaws.String(region)})
	return client, nil
}

func (s *Client) GetInstanceID() (string, error) {
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: iaasaws.String("tag:Name"),
				Values: []*string{
					iaasaws.String(s.InstanceNameRegex),
				},
			},
			{
				Name: iaasaws.String("vpc-id"),
				Values: []*string{
					iaasaws.String(s.VPCName),
				},
			},
		},
	}
	resp, err := s.Describer.DescribeInstances(params)
	if err != nil {
		return "", errwrap.Wrap(err, "DescribeInstances yielded error")
	}

	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			return *inst.InstanceId, nil
		}
	}
	return "", ErrCouldNotFindInstance
}
