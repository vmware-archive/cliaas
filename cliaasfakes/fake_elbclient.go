package cliaasfakes

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/elb"
)

var EmptyLbDescription []*elb.LoadBalancerDescription = make([]*elb.LoadBalancerDescription, 0)

//FakeElbClient Mock the elb client behavior
type FakeElbClient struct {
	Instances         []*elb.Instance
	DescribeErr       bool
	DeregisterErr     bool
	DeregisterCapture []*elb.Instance
	RegisterCapture   []*elb.Instance
	RegisterErr       bool
	LoadBalancerExist bool
}

//DescribeLoadBalancers Mock the DescribeLoadBalancers behavior
func (c *FakeElbClient) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {

	if c.DescribeErr {
		return nil, errors.New("Describe API error")
	}
	if !c.LoadBalancerExist {
		return &elb.DescribeLoadBalancersOutput{LoadBalancerDescriptions: EmptyLbDescription}, nil
	}
	lbDescriptions := &elb.LoadBalancerDescription{
		Instances: c.Instances,
	}
	return &elb.DescribeLoadBalancersOutput{LoadBalancerDescriptions: []*elb.LoadBalancerDescription{lbDescriptions}}, nil
}

//DeregisterInstancesFromLoadBalancer Mock the deregisterInstance method
func (c *FakeElbClient) DeregisterInstancesFromLoadBalancer(input *elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error) {
	if c.DeregisterErr {
		return nil, errors.New("")
	}
	c.DeregisterCapture = input.Instances
	return nil, nil
}

//RegisterInstancesWithLoadBalancer Mock the registerInstance method
func (c *FakeElbClient) RegisterInstancesWithLoadBalancer(input *elb.RegisterInstancesWithLoadBalancerInput) (*elb.RegisterInstancesWithLoadBalancerOutput, error) {
	if c.RegisterErr {
		return nil, errors.New("")
	}
	c.RegisterCapture = input.Instances
	return nil, nil
}
