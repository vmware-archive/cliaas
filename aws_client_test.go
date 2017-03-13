package cliaas_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cliaas"
	"github.com/pivotal-cf/cliaas/cliaasfakes"
)

var _ = Describe("AWSClient", func() {
	var (
		client    cliaas.AWSClient
		ec2Client *cliaasfakes.FakeEC2Client
	)

	BeforeEach(func() {
		ec2Client = new(cliaasfakes.FakeEC2Client)

		client = cliaas.NewAWSClient(ec2Client, "some vpc")
	})

	Describe("GetVMInfo", func() {
		Context("when a single instance is found", func() {
			BeforeEach(func() {
				instances := []*ec2.Instance{
					&ec2.Instance{
						InstanceId:   aws.String("some-instance-id"),
						InstanceType: aws.String("some-instance-type"),
						KeyName:      aws.String("some-key-name"),
						SubnetId:     aws.String("some-subnet-id"),
						SecurityGroups: []*ec2.GroupIdentifier{
							{
								GroupId: aws.String("some-group-id"),
							},
							{
								GroupId: aws.String("some-other-group-id"),
							},
						},
						NetworkInterfaces: []*ec2.InstanceNetworkInterface{
							{
								Association: &ec2.InstanceNetworkInterfaceAssociation{
									PublicIp: aws.String("some-public-ip"),
								},
							},
						},
					},
				}

				ec2Client.DescribeInstancesReturns(&ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{
						{
							Instances: instances,
						},
					},
				}, nil)
			})

			It("returns vm info for the instance", func() {
				vmInfo, err := client.GetVMInfo("some-identifier")
				Expect(err).NotTo(HaveOccurred())
				Expect(vmInfo).To(Equal(cliaas.VMInfo{
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
				instances := []*ec2.Instance{
					&ec2.Instance{},
					&ec2.Instance{},
				}
				ec2Client.DescribeInstancesReturns(&ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{
						{
							Instances: instances,
						},
					},
				}, nil)
			})

			It("returns an error", func() {
				_, err := client.GetVMInfo("some-identifier")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("more than one matching instance found"))
			})
		})

		Context("when no instances are found", func() {
			BeforeEach(func() {
				instances := []*ec2.Instance{}

				ec2Client.DescribeInstancesReturns(&ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{
						{
							Instances: instances,
						},
					},
				}, nil)
			})

			It("returns an error", func() {
				_, err := client.GetVMInfo("some-identifier")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no matching instances found"))
			})
		})

		Context("when there is an api error", func() {
			BeforeEach(func() {
				ec2Client.DescribeInstancesReturns(nil, errors.New("an error"))
			})

			It("returns an error", func() {
				_, err := client.GetVMInfo("some-identifier")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("describe instances failed: an error"))
			})
		})
	})

	Describe("Stop", func() {
		It("tries to stop the instance", func() {
			err := client.StopVM("foo")
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2Client.StopInstancesCallCount()).To(Equal(1))
			input := ec2Client.StopInstancesArgsForCall(0)
			Expect(*input).To(Equal(ec2.StopInstancesInput{
				InstanceIds: []*string{
					aws.String("foo"),
				},
				DryRun: aws.Bool(false),
				Force:  aws.Bool(true),
			}))
		})

		Context("when there is an api error", func() {
			BeforeEach(func() {
				ec2Client.StopInstancesReturns(&ec2.StopInstancesOutput{}, errors.New("an error"))
			})

			It("returns an error", func() {
				err := client.StopVM("foo")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Delete", func() {
		It("tries to delete the instance", func() {
			err := client.DeleteVM("foo")
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2Client.TerminateInstancesCallCount()).To(Equal(1))
			input := ec2Client.TerminateInstancesArgsForCall(0)
			Expect(*input).To(Equal(ec2.TerminateInstancesInput{
				InstanceIds: []*string{
					aws.String("foo"),
				},
				DryRun: aws.Bool(false),
			}))
		})

		Context("when there is an api error", func() {
			BeforeEach(func() {
				ec2Client.TerminateInstancesReturns(&ec2.TerminateInstancesOutput{}, errors.New("an error"))
			})

			It("returns an error", func() {
				err := client.DeleteVM("foo")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("terminate instances failed: an error"))
			})
		})
	})

	Describe("AssignPublicIP", func() {

		It("tries to assign the public IP", func() {
			err := client.AssignPublicIP("foo", "1.1.1.1")
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2Client.AssociateAddressCallCount()).To(Equal(1))
			input := ec2Client.AssociateAddressArgsForCall(0)
			Expect(*input).To(Equal(ec2.AssociateAddressInput{
				InstanceId: aws.String("foo"),
				PublicIp:   aws.String("1.1.1.1"),
			}))
		})

		Context("when there is an api error", func() {
			BeforeEach(func() {
				ec2Client.AssociateAddressReturns(&ec2.AssociateAddressOutput{}, errors.New("an error"))
			})

			It("returns an error", func() {
				err := client.AssignPublicIP("foo", "1.1.1.1")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("associate address failed: an error"))
			})
		})
	})

	Describe("Create", func() {
		var (
			name            = "some-instance-name"
			ami             = "some-instance-ami"
			instanceType    = "some-instance-type"
			keyName         = "some-key-name"
			subnetID        = "some-subnet-id"
			securityGroupID = "some-security-group-id"
			reservation     *ec2.Reservation
		)

		BeforeEach(func() {
			reservation = &ec2.Reservation{
				Instances: []*ec2.Instance{
					&ec2.Instance{
						InstanceId: aws.String("some-instance-id"),
					},
				},
			}

			ec2Client.RunInstancesReturns(reservation, nil)
		})

		It("tries to create the instance", func() {
			_, err := client.CreateVM(ami, instanceType, name, keyName, subnetID, securityGroupID)
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2Client.RunInstancesCallCount()).To(Equal(1))
			input := ec2Client.RunInstancesArgsForCall(0)
			Expect(*input).To(Equal(ec2.RunInstancesInput{
				ImageId:          aws.String(ami),
				InstanceType:     aws.String(instanceType),
				MinCount:         aws.Int64(1),
				MaxCount:         aws.Int64(1),
				KeyName:          aws.String(keyName),
				SubnetId:         aws.String(subnetID),
				SecurityGroupIds: aws.StringSlice([]string{securityGroupID}),
			}))
		})

		Context("when no security groups are set", func() {
			BeforeEach(func() {
				securityGroupID = ""
			})

			It("should create an instance with a blank security group", func() {
				_, err := client.CreateVM(ami, instanceType, name, keyName, subnetID, securityGroupID)
				Expect(err).NotTo(HaveOccurred())

				Expect(ec2Client.RunInstancesCallCount()).To(Equal(1))
				input := ec2Client.RunInstancesArgsForCall(0)
				Expect(input.SecurityGroupIds).To(BeEmpty())
			})
		})

		Context("when creating the instance fails", func() {
			BeforeEach(func() {
				ec2Client.RunInstancesReturns(reservation, errors.New("an error"))
			})

			It("returns an error", func() {
				_, err := client.CreateVM(ami, instanceType, name, keyName, subnetID, securityGroupID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("run instances failed: an error"))
			})
		})
	})
})
