package cliaas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
)

type ElbClient interface {
	DeregisterInstancesFromLoadBalancer(input *elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error)
	RegisterInstancesWithLoadBalancer(input *elb.RegisterInstancesWithLoadBalancerInput) (*elb.RegisterInstancesWithLoadBalancerOutput, error)
	DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
}

func NewElbClient(
	accessKeyID string,
	secretAccessKey string,
	region string,
) (ElbClient, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return elb.New(sess, &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      aws.String(region),
	}), nil
}
