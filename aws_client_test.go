package cliaas_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cliaas"
	"github.com/pivotal-cf/cliaas/cliaasfakes"
)

var _ = Describe("AWSClient", func() {
	var (
		client            cliaas.AWSClient
		ec2Client         *cliaasfakes.FakeEC2Client
		elbClient         *cliaasfakes.FakeElbClient
		runningState      *ec2.InstanceState
		pendingState      *ec2.InstanceState
		shuttingDownState *ec2.InstanceState
		terminatedState   *ec2.InstanceState
		stoppingState     *ec2.InstanceState
		stoppedState      *ec2.InstanceState
	)

	BeforeEach(func() {
		runningState = &ec2.InstanceState{}
		runningState.SetCode(16)
		runningState.SetName(ec2.InstanceStateNameRunning)
		pendingState = &ec2.InstanceState{}
		pendingState.SetCode(0)
		pendingState.SetName(ec2.InstanceStateNamePending)
		shuttingDownState = &ec2.InstanceState{}
		shuttingDownState.SetCode(32)
		shuttingDownState.SetName(ec2.InstanceStateNameShuttingDown)
		terminatedState = &ec2.InstanceState{}
		terminatedState.SetCode(48)
		terminatedState.SetName(ec2.InstanceStateNameTerminated)
		stoppingState = &ec2.InstanceState{}
		stoppingState.SetCode(64)
		stoppingState.SetName(ec2.InstanceStateNameStopping)
		stoppedState = &ec2.InstanceState{}
		stoppedState.SetCode(80)
		stoppedState.SetName(ec2.InstanceStateNameStopped)
		ec2Client = new(cliaasfakes.FakeEC2Client)
		elbClient = new(cliaasfakes.FakeElbClient)
		clock := fakeclock.NewFakeClock(time.Now())

		client = cliaas.NewAWSClient(ec2Client, elbClient, "some vpc", clock)
	})

	Describe("GetVMInfo", func() {
		Context("when a single `running` instance is found", func() {
			BeforeEach(func() {

				instances := []*ec2.Instance{
					createEC2Instance(pendingState),
					createEC2Instance(runningState),
					createEC2Instance(shuttingDownState),
					createEC2Instance(terminatedState),
					createEC2Instance(stoppingState),
					createEC2Instance(stoppedState),
				}

				ec2Client.DescribeInstancesReturns(&ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{
						{
							Instances: instances,
						},
					},
				}, nil)

				ec2Client.DescribeVolumesReturns(&ec2.DescribeVolumesOutput{
					Volumes: []*ec2.Volume{
						{
							Encrypted:  aws.Bool(true),
							Size:       aws.Int64(1),
							VolumeType: aws.String("some-volume-type"),
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
					BlockDeviceMappings: []cliaas.BlockDeviceMapping{
						{
							DeviceName: "/dev/sda1",
							EBS: cliaas.EBS{
								DeleteOnTermination: true,
								VolumeSize:          1,
								VolumeType:          "some-volume-type",
							},
						},
						{
							DeviceName: "/dev/sda2",
							EBS: cliaas.EBS{
								DeleteOnTermination: true,
								VolumeSize:          1,
								VolumeType:          "some-volume-type",
							},
						},
					},
					IAMInstanceProfileARN: "some-instance-profile-arn",
				}))
			})
		})

		Context("when more than one `running` instance is found", func() {
			BeforeEach(func() {
				instances := []*ec2.Instance{
					createEC2Instance(runningState),
					createEC2Instance(runningState),
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
			reservation  *ec2.Reservation
			name         = "some-instance-name"
			ami          = "some-instance-ami"
			vmInfoConfig = VMInfoConfig{
				InstanceType:    "some-instance-type",
				KeyName:         "some-key-name",
				SubnetID:        "some-subnet-id",
				SecurityGroupID: "some-security-group-id",
			}
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
			_, err := client.CreateVM(ami, name, createVMInfo("/dev/sda1", "", true, vmInfoConfig))
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2Client.RunInstancesCallCount()).To(Equal(1))
			input := ec2Client.RunInstancesArgsForCall(0)
			Expect(*input).To(Equal(ec2.RunInstancesInput{
				ImageId:      aws.String(ami),
				InstanceType: aws.String(vmInfoConfig.InstanceType),
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					{
						DeviceName: aws.String("/dev/sda1"),
						Ebs: &ec2.EbsBlockDevice{
							DeleteOnTermination: aws.Bool(true),
							Encrypted:           nil,
							SnapshotId:          nil,
							VolumeSize:          aws.Int64(1),
							VolumeType:          aws.String("some-volume-type"),
						},
					},
				},
				IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
					Arn: aws.String("some-instance-profile-arn"),
				},
				MinCount:         aws.Int64(1),
				MaxCount:         aws.Int64(1),
				KeyName:          aws.String(vmInfoConfig.KeyName),
				SubnetId:         aws.String(vmInfoConfig.SubnetID),
				SecurityGroupIds: aws.StringSlice([]string{vmInfoConfig.SecurityGroupID}),
			}))
		})

		Context("when there are mulitple blockdevices on the original VM", func() {
			var (
				deviceName1 = "/dev/sda1"
				deviceName2 = "/dev/sda2"
				vmInfo      cliaas.VMInfo
			)

			BeforeEach(func() {
				vmInfo = createVMInfo(deviceName1, "some-snapshot-id", true, vmInfoConfig)
				vmInfo.BlockDeviceMappings = append(vmInfo.BlockDeviceMappings, cliaas.BlockDeviceMapping{
					DeviceName: deviceName2,
					EBS: cliaas.EBS{
						DeleteOnTermination: true,
						VolumeSize:          1,
						VolumeType:          "some-volume-type",
					},
				})
			})

			It("creates a new VM with all blockdevices defined", func() {
				_, err := client.CreateVM(ami, name, vmInfo)
				Expect(err).NotTo(HaveOccurred())

				input := ec2Client.RunInstancesArgsForCall(0)
				Expect(input.BlockDeviceMappings).To(HaveLen(2))
				Expect(*input.BlockDeviceMappings[0].DeviceName).To(Equal(deviceName1))
				Expect(*input.BlockDeviceMappings[1].DeviceName).To(Equal(deviceName2))
			})
		})

		It("tries to create an instance with a blank security group when no security groups are set", func() {
			_, err := client.CreateVM(ami, name, cliaas.VMInfo{
				KeyName:          vmInfoConfig.KeyName,
				SubnetID:         vmInfoConfig.SubnetID,
				SecurityGroupIDs: []string{},
				InstanceType:     vmInfoConfig.InstanceType,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2Client.RunInstancesCallCount()).To(Equal(1))
			input := ec2Client.RunInstancesArgsForCall(0)
			Expect(input.SecurityGroupIds).To(BeEmpty())
		})

		Context("when creating the instance fails", func() {
			BeforeEach(func() {
				ec2Client.RunInstancesReturns(reservation, errors.New("an error"))
			})

			It("returns an error", func() {
				_, err := client.CreateVM(ami, name, cliaas.VMInfo{
					KeyName:          vmInfoConfig.KeyName,
					SubnetID:         vmInfoConfig.SubnetID,
					SecurityGroupIDs: []string{vmInfoConfig.SecurityGroupID},
					InstanceType:     vmInfoConfig.InstanceType,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("run instances failed: an error"))
			})
		})
	})

	Describe("SwapLb", func() {
		var (
			lbName       string
			newInstances []string
		)

		BeforeEach(func() {

			elbClient.DescribeErr = false
			elbClient.LoadBalancerExist = true
			elbClient.DeregisterErr = false
			elbClient.RegisterErr = false
			lbName = "MyLoadBalancer"
			newInstances = []string{"new-instance"}
			var oldInstance = "old-instance"
			elbClient.Instances = []*elb.Instance{&elb.Instance{InstanceId: &oldInstance}}
		})

		Context("when Describe Load Balancer throw err", func() {
			BeforeEach(func() {
				elbClient.DescribeErr = true
			})
			It("should error out", func() {
				err := client.SwapLb(lbName, newInstances)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when load balancer does not exists", func() {
			BeforeEach(func() {
				elbClient.LoadBalancerExist = false
			})
			It("should error out", func() {
				err := client.SwapLb(lbName, newInstances)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When Deregister Instance failed", func() {
			BeforeEach(func() {
				elbClient.DescribeErr = true
			})
			It("should error out", func() {
				err := client.SwapLb(lbName, newInstances)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When Degister Instance succeed", func() {
			It("should deregister the old-instance", func() {
				client.SwapLb(lbName, newInstances)
				oldInstance := *elbClient.DeregisterCapture[0].InstanceId
				Expect(oldInstance).To(Equal(oldInstance))
			})
		})

		Context("when register Instance failed", func() {
			BeforeEach(func() {
				elbClient.RegisterErr = true
			})
			It("should error out", func() {
				err := client.SwapLb(lbName, newInstances)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When Register Instance succeed", func() {
			It("should register the new-instance", func() {
				client.SwapLb(lbName, newInstances)
				newInstance := *elbClient.RegisterCapture[0].InstanceId
				Expect(newInstance).To(Equal("new-instance"))
			})
		})
	})
})

func createEC2Instance(state *ec2.InstanceState) *ec2.Instance {
	return &ec2.Instance{
		State:        state,
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
		RootDeviceName: aws.String("/dev/sda2"),
		BlockDeviceMappings: []*ec2.InstanceBlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &ec2.EbsInstanceBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					VolumeId:            aws.String("some-root-volume-id"),
				},
			},
			{
				DeviceName: aws.String("/dev/sda2"),
				Ebs: &ec2.EbsInstanceBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					VolumeId:            aws.String("some-volume-id"),
				},
			},
		},
		IamInstanceProfile: &ec2.IamInstanceProfile{
			Arn: aws.String("some-instance-profile-arn"),
		},
	}
}

type VMInfoConfig struct {
	InstanceType    string
	KeyName         string
	SubnetID        string
	SecurityGroupID string
}

func createVMInfo(deviceName, snapshotID string, encrypted bool, vmInfoConfig VMInfoConfig) cliaas.VMInfo {
	return cliaas.VMInfo{
		KeyName:          vmInfoConfig.KeyName,
		SubnetID:         vmInfoConfig.SubnetID,
		SecurityGroupIDs: []string{vmInfoConfig.SecurityGroupID},
		InstanceType:     vmInfoConfig.InstanceType,
		BlockDeviceMappings: []cliaas.BlockDeviceMapping{
			{
				DeviceName: deviceName,
				EBS: cliaas.EBS{
					DeleteOnTermination: true,
					VolumeSize:          1,
					VolumeType:          "some-volume-type",
				},
			},
		},
		IAMInstanceProfileARN: "some-instance-profile-arn",
	}
}
