package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//go:generate counterfeiter . EC2Client

type EC2Client interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	DescribeVolumes(*ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
	DescribeInstanceStatus(*ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error)
	AssociateAddress(*ec2.AssociateAddressInput) (*ec2.AssociateAddressOutput, error)
	TerminateInstances(*ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)
	StopInstances(*ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error)
	StartInstances(*ec2.StartInstancesInput) (*ec2.StartInstancesOutput, error)
	CreateTags(*ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error)
	RunInstances(*ec2.RunInstancesInput) (*ec2.Reservation, error)
}

func NewEC2Client(accessKeyID string, secretAccessKey string, region string) (EC2Client, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	ec2Client := ec2.New(sess, &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      aws.String(region),
	})

	//_, err = ec2Client.Config.Credentials.Get()
	//if err != nil {
	//	fmt.Println("Creds not found")
	//}
	//
	//name := "tidy-anteater-OpsMan az1"
	//
	//params := &ec2.DescribeInstancesInput{
	//	Filters: []*ec2.Filter{
	//		{
	//			Name: aws.String("tag:Name"),
	//			Values: []*string{
	//				aws.String(name),
	//			},
	//		},
	//	},
	//}
	//resp, err := ec2Client.DescribeInstances(params)
	//
	//var list []*ec2.Instance
	//
	//for idx := range resp.Reservations {
	//	for _, instance := range resp.Reservations[idx].Instances {
	//		if *instance.State.Name == ec2.InstanceStateNameRunning {
	//			list = append(list, instance)
	//		}
	//	}
	//}
	//
	//if len(list) == 0 {
	//	println("no matching instances found")
	//}
	//
	//if len(list) > 1 {
	//	println("more than one matching instance found")
	//}
	//
	//instance := list[0]
	//
	//_, err = describeVolumes(instance.BlockDeviceMappings)
	//if err != nil {
	//	println(err.Error(), "describeVolumes failure")
	//}

	return ec2Client, nil
}

//func describeVolumes(instanceBlockDeviceMappings []*ec2.InstanceBlockDeviceMapping) ([]BlockDeviceMapping, error) {
//	blockDeviceMappings := []BlockDeviceMapping{}
//	for _, blockDeviceMapping := range instanceBlockDeviceMappings {
//		params := &ec2.DescribeVolumesInput{
//			Filters: []*ec2.Filter{
//				{
//					Name: aws.String("volume-id"),
//					Values: []*string{
//						blockDeviceMapping.Ebs.VolumeId,
//					},
//				},
//			},
//		}
//		resp, err := ec2Client.DescribeVolumes(params)
//		if err != nil {
//			println(err.Error(), "Describe Volume call failed")
//		}
//
//		for _, volume := range resp.Volumes {
//			blockDeviceMappings = append(blockDeviceMappings, BlockDeviceMapping{
//				DeviceName: aws.StringValue(blockDeviceMapping.DeviceName),
//				EBS: EBS{
//					DeleteOnTermination: aws.BoolValue(blockDeviceMapping.Ebs.DeleteOnTermination),
//					VolumeSize:          aws.Int64Value(volume.Size),
//					VolumeType:          aws.StringValue(volume.VolumeType),
//				},
//			})
//		}
//	}
//	println("********************",blockDeviceMappings[0].EBS.VolumeSize)
//
//	return blockDeviceMappings, nil
//}
