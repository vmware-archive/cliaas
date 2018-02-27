package aws_test

import (
	"math/rand"
	"os"
	"time"

	"code.cloudfoundry.org/clock"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cliaasAWS "github.com/pivotal-cf/cliaas/iaas/aws"
)

var _ = Describe("AwsClient", func() {
	var (
		awsClient       cliaasAWS.AWSClient
		ec2Client       *ec2.EC2
		accessKey       = os.Getenv("AWS_ACCESS_KEY")
		secretKey       = os.Getenv("AWS_SECRET_KEY")
		region          = os.Getenv("AWS_REGION")
		vpc             = os.Getenv("AWS_VPC")
		subnetID        = os.Getenv("AWS_SUBNET")
		securityGroupID = os.Getenv("AWS_SECURITY_GROUP")
		keyPairName     = os.Getenv("KEYPAIR_NAME")
		ami             = "ami-0b33d91d"
		vmType          = "t2.micro"

		name string
	)

	BeforeEach(func() {
		sess, err := session.NewSession()
		Expect(err).NotTo(HaveOccurred())

		ec2Client = ec2.New(sess, &aws.Config{
			Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
			Region:      aws.String(region),
		})

		awsClient = cliaasAWS.NewAWSClient(ec2Client, vpc, clock.NewClock())
		Expect(awsClient).NotTo(BeNil())

		name = randSeq(10)
	})

	Context("Create", func() {
		var (
			instanceID string
			createErr  error
		)

		JustBeforeEach(func() {
			instanceID, createErr = awsClient.CreateVM(ami, name, cliaasAWS.VMInfo{
				InstanceType:     vmType,
				KeyName:          keyPairName,
				SubnetID:         subnetID,
				SecurityGroupIDs: []string{securityGroupID},
				BlockDeviceMappings: []cliaasAWS.BlockDeviceMapping{
					{
						DeviceName: "/dev/sda1",
						EBS: cliaasAWS.EBS{
							DeleteOnTermination: true,
							VolumeSize:          10,
							VolumeType:          "standard",
						},
					},
				},
			})
		})

		AfterEach(func() {
			if instanceID != "" {
				_ = awsClient.DeleteVM(instanceID)
			}
		})

		It("creates a VM", func() {
			Expect(createErr).NotTo(HaveOccurred())
			Expect(instanceID).NotTo(Equal(""))

			Eventually(func() error {
				_, err := awsClient.GetVMInfo(name)
				return err
			}, "1m", "10s").Should(Succeed())
		})
	})

	Context("AssignPublicIP", func() {
		var instanceID string

		BeforeEach(func() {
			var createErr error
			instanceID, createErr = awsClient.CreateVM(ami, name, cliaasAWS.VMInfo{
				InstanceType:     vmType,
				KeyName:          keyPairName,
				SubnetID:         subnetID,
				SecurityGroupIDs: []string{securityGroupID},
				BlockDeviceMappings: []cliaasAWS.BlockDeviceMapping{
					{
						DeviceName: "/dev/sda1",
						EBS: cliaasAWS.EBS{
							DeleteOnTermination: true,
							VolumeSize:          10,
							VolumeType:          "standard",
						},
					},
				},
			})
			Expect(createErr).NotTo(HaveOccurred())

			client := cliaasAWS.NewAWSClient(ec2Client, vpc, clock.NewClock())

			err := client.WaitForStatus(instanceID, ec2.InstanceStateNameRunning)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if instanceID != "" {
				_ = awsClient.DeleteVM(instanceID)
			}
		})

		It("associates the elastic IP to the instance", func() {
			err := awsClient.AssignPublicIP(instanceID, "34.205.163.20")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func randSeq(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghipqrstuvwxyz0123456789"
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return "aws_integration_test_" + string(result)
}
