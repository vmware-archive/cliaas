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
		awsClient *awsfakes.FakeAWSClient
	)

	BeforeEach(func() {
		awsClient = new(awsfakes.FakeAWSClient)

		client = aws.NewClient(awsClient, "some vpc")
	})

	Describe("GetVMInfo", func() {
		var (
			instances []*ec2.Instance
			instance  *ec2.Instance
			err       error
			listErr   error
		)

		JustBeforeEach(func() {
			awsClient.ListReturns(instances, listErr)
			instance, err = client.GetVMInfo("some-identifier")
		})

		Context("when a single instance is found", func() {
			BeforeEach(func() {
				instances = []*ec2.Instance{
					&ec2.Instance{},
				}
			})

			It("returns the instance", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(instance).To(Equal(instances[0]))
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
				Expect(err.Error()).To(Equal("Found more than one match"))
			})
		})

		Context("when no instances are found", func() {
			BeforeEach(func() {
				instances = []*ec2.Instance{}
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("No instance matches found"))
			})
		})

		Context("when there is an error listing instances", func() {
			BeforeEach(func() {
				listErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("call List on aws client failed: an error"))
			})
		})
	})

	Describe("Stop", func() {
		var (
			err     error
			stopErr error
		)

		JustBeforeEach(func() {
			awsClient.StopReturns(stopErr)
			err = client.StopVM(ec2.Instance{
				InstanceId: iaasaws.String("foo"),
			})
		})

		It("tries to stop the instance", func() {
			Expect(awsClient.StopCallCount()).To(Equal(1))
			instanceID := awsClient.StopArgsForCall(0)
			Expect(instanceID).To(Equal("foo"))
		})

		Context("when stop returns an error", func() {
			BeforeEach(func() {
				stopErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Delete", func() {
		var (
			err       error
			deleteErr error
		)

		JustBeforeEach(func() {
			awsClient.DeleteReturns(deleteErr)
			err = client.DeleteVM("foo")
		})

		It("tries to delete the instance", func() {
			Expect(awsClient.DeleteCallCount()).To(Equal(1))
			instanceID := awsClient.DeleteArgsForCall(0)
			Expect(instanceID).To(Equal("foo"))
		})

		Context("when delete returns an error", func() {
			BeforeEach(func() {
				deleteErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("call delete on aws client failed: an error"))
			})
		})
	})

	Describe("AssignPublicIP", func() {
		var (
			err          error
			associateErr error
		)

		JustBeforeEach(func() {
			awsClient.AssociateElasticIPReturns(associateErr)
			err = client.AssignPublicIP(ec2.Instance{
				InstanceId: iaasaws.String("foo"),
			}, "1.1.1.1")
		})

		It("tries to assign the public IP", func() {
			Expect(awsClient.AssociateElasticIPCallCount()).To(Equal(1))
			instanceID, ip := awsClient.AssociateElasticIPArgsForCall(0)
			Expect(instanceID).To(Equal("foo"))
			Expect(ip).To(Equal("1.1.1.1"))
		})

		Context("when assignment returns an error", func() {
			BeforeEach(func() {
				associateErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("call associateElasticIP on aws client failed: an error"))
			})
		})
	})

	Describe("Create", func() {
		var (
			instanceToCreate ec2.Instance
			createdInstance  *ec2.Instance
			newInstance      *ec2.Instance
			err              error
			createErr        error
		)

		BeforeEach(func() {
			instanceToCreate = ec2.Instance{
				ImageId:      iaasaws.String("foo"),
				InstanceType: iaasaws.String("my.type"),
				KeyName:      iaasaws.String("mykey"),
				SubnetId:     iaasaws.String("mysubnet"),
				Tags: []*ec2.Tag{
					&ec2.Tag{
						Key: iaasaws.String("Name"), Value: iaasaws.String("myname"),
					},
					&ec2.Tag{
						Key: iaasaws.String("OtherTag"), Value: iaasaws.String("some random"),
					},
				},
				SecurityGroups: []*ec2.GroupIdentifier{
					&ec2.GroupIdentifier{
						GroupId: iaasaws.String("mysecuritygroup"),
					},
				},
			}

			createdInstance = &ec2.Instance{}
		})

		JustBeforeEach(func() {
			awsClient.CreateReturns(createdInstance, createErr)
			newInstance, err = client.CreateVM(instanceToCreate, "foo", "my.type", "newName")
		})

		It("tries to create the instance", func() {
			Expect(awsClient.CreateCallCount()).To(Equal(1))
			ami, vmType, name, keyPairName, subnetID, securityGroupID := awsClient.CreateArgsForCall(0)
			Expect(ami).To(Equal("foo"))
			Expect(vmType).To(Equal("my.type"))
			Expect(name).To(Equal("newName"))
			Expect(keyPairName).To(Equal("mykey"))
			Expect(subnetID).To(Equal("mysubnet"))
			Expect(securityGroupID).To(Equal("mysecuritygroup"))
		})

		Context("when no security groups are set", func() {
			BeforeEach(func() {
				instanceToCreate.SecurityGroups = nil
			})

			It("should create an instance with a blank security group", func() {
				_, _, _, _, _, securityGroupID := awsClient.CreateArgsForCall(0)
				Expect(securityGroupID).To(Equal(""))
			})
		})

		Context("when creating the instance fails", func() {
			BeforeEach(func() {
				createErr = errors.New("an error")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("call create on aws client failed: an error"))
			})
		})
	})
})
