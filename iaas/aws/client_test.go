package aws_test

import (
	"errors"

	iaasaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/c0-ops/cliaas/iaas/aws"
	"github.com/c0-ops/cliaas/iaas/aws/awsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		client    aws.Client
		ec2Client *awsfakes.FakeEC2Client
	)

	BeforeEach(func() {
		ec2Client = new(awsfakes.FakeEC2Client)

		client = aws.NewClient(ec2Client, "some vpc")
	})

	Describe("GetVMInfo", func() {
		var (
			instances []*ec2.Instance
			vmInfo    aws.VMInfo
			err       error
			apiErr    error
		)

		JustBeforeEach(func() {
			output := &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: instances,
					},
				},
			}
			ec2Client.DescribeInstancesReturns(output, apiErr)
			vmInfo, err = client.GetVMInfo("some-identifier")
		})

		Context("when a single instance is found", func() {
			BeforeEach(func() {
				instances = []*ec2.Instance{
					&ec2.Instance{
						InstanceId:   iaasaws.String("some-instance-id"),
						InstanceType: iaasaws.String("some-instance-type"),
						KeyName:      iaasaws.String("some-key-name"),
						SubnetId:     iaasaws.String("some-subnet-id"),
						SecurityGroups: []*ec2.GroupIdentifier{
							{
								GroupId: iaasaws.String("some-group-id"),
							},
							{
								GroupId: iaasaws.String("some-other-group-id"),
							},
						},
						NetworkInterfaces: []*ec2.InstanceNetworkInterface{
							{
								Association: &ec2.InstanceNetworkInterfaceAssociation{
									PublicIp: iaasaws.String("some-public-ip"),
								},
							},
						},
					},
				}
			})

			It("returns vm info for the instance", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(vmInfo).To(Equal(aws.VMInfo{
					InstanceID:       "some-instance-id",
					InstanceType:     "some-instance-type",
					KeyName:          "some-key-name",
					SubnetID:         "some-subnet-id",
					SecurityGroupIDs: []string{"some-group-id", "some-other-group-id"},
					PublicIP:         "some-public-ip",
				}))
			})
		})

		Context("when more than one instance is found", func() {
			BeforeEach(func() {
				instances = []*ec2.Instance{
					&ec2.Instance{},
					&ec2.Instance{},
				}
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("more than one matching instance found"))
			})
		})

		Context("when no instances are found", func() {
			BeforeEach(func() {
				instances = []*ec2.Instance{}
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no matching instances found"))
			})
		})

		Context("when there is an api error", func() {
			BeforeEach(func() {
				apiErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("describe instances failed: an error"))
			})
		})
	})

	Describe("Stop", func() {
		var (
			err    error
			apiErr error
		)

		JustBeforeEach(func() {
			ec2Client.StopInstancesReturns(&ec2.StopInstancesOutput{}, apiErr)
			err = client.StopVM("foo")
		})

		It("tries to stop the instance", func() {
			Expect(ec2Client.StopInstancesCallCount()).To(Equal(1))
			input := ec2Client.StopInstancesArgsForCall(0)
			Expect(*input).To(Equal(ec2.StopInstancesInput{
				InstanceIds: []*string{
					iaasaws.String("foo"),
				},
				DryRun: iaasaws.Bool(false),
				Force:  iaasaws.Bool(true),
			}))
		})

		Context("when there is an api error", func() {
			BeforeEach(func() {
				apiErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Delete", func() {
		var (
			err    error
			apiErr error
		)

		JustBeforeEach(func() {
			ec2Client.TerminateInstancesReturns(&ec2.TerminateInstancesOutput{}, apiErr)
			err = client.DeleteVM("foo")
		})

		It("tries to delete the instance", func() {
			Expect(ec2Client.TerminateInstancesCallCount()).To(Equal(1))
			input := ec2Client.TerminateInstancesArgsForCall(0)
			Expect(*input).To(Equal(ec2.TerminateInstancesInput{
				InstanceIds: []*string{
					iaasaws.String("foo"),
				},
				DryRun: iaasaws.Bool(false),
			}))
		})

		Context("when there is an api error", func() {
			BeforeEach(func() {
				apiErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("terminate instances failed: an error"))
			})
		})
	})

	Describe("AssignPublicIP", func() {
		var (
			err    error
			apiErr error
		)

		JustBeforeEach(func() {
			ec2Client.AssociateAddressReturns(&ec2.AssociateAddressOutput{}, apiErr)
			err = client.AssignPublicIP("foo", "1.1.1.1")
		})

		It("tries to assign the public IP", func() {
			Expect(ec2Client.AssociateAddressCallCount()).To(Equal(1))
			input := ec2Client.AssociateAddressArgsForCall(0)
			Expect(*input).To(Equal(ec2.AssociateAddressInput{
				InstanceId: iaasaws.String("foo"),
				PublicIp:   iaasaws.String("1.1.1.1"),
			}))
		})

		Context("when there is an api error", func() {
			BeforeEach(func() {
				apiErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("associate address failed: an error"))
			})
		})
	})

	Describe("Create", func() {
		var (
			err    error
			apiErr error

			name            = "some-instance-name"
			ami             = "some-instance-ami"
			instanceType    = "some-instance-type"
			keyName         = "some-key-name"
			subnetID        = "some-subnet-id"
			securityGroupID = "some-security-group-id"
		)

		JustBeforeEach(func() {
			reservation := &ec2.Reservation{
				Instances: []*ec2.Instance{
					&ec2.Instance{
						InstanceId: iaasaws.String("some-instance-id"),
					},
				},
			}

			ec2Client.RunInstancesReturns(reservation, apiErr)
			_, err = client.CreateVM(
				ami,
				instanceType,
				name,
				keyName,
				subnetID,
				securityGroupID,
			)
		})

		It("tries to create the instance", func() {
			Expect(ec2Client.RunInstancesCallCount()).To(Equal(1))
			input := ec2Client.RunInstancesArgsForCall(0)
			Expect(*input).To(Equal(ec2.RunInstancesInput{
				ImageId:          iaasaws.String(ami),
				InstanceType:     iaasaws.String(instanceType),
				MinCount:         iaasaws.Int64(1),
				MaxCount:         iaasaws.Int64(1),
				KeyName:          iaasaws.String(keyName),
				SubnetId:         iaasaws.String(subnetID),
				SecurityGroupIds: iaasaws.StringSlice([]string{securityGroupID}),
			}))
		})

		Context("when no security groups are set", func() {
			BeforeEach(func() {
				securityGroupID = ""
			})

			It("should create an instance with a blank security group", func() {
				Expect(ec2Client.RunInstancesCallCount()).To(Equal(1))
				input := ec2Client.RunInstancesArgsForCall(0)
				Expect(input.SecurityGroupIds).To(BeEmpty())
			})
		})

		Context("when creating the instance fails", func() {
			BeforeEach(func() {
				apiErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("run instances failed: an error"))
			})
		})
	})
})
