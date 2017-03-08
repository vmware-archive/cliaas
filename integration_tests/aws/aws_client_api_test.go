package aws_test

import (
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cliaas"
)

var _ = Describe("AwsClient", func() {
	var (
		awsClient cliaas.AWSClient
		ec2Client *ec2.EC2
		accessKey = os.Getenv("AWS_ACCESS_KEY")
		secretKey = os.Getenv("AWS_SECRET_KEY")
		region    = os.Getenv("AWS_REGION")
		vpc       = os.Getenv("AWS_VPC")

		ami         = "ami-0b33d91d"
		vmType      = "t2.micro"
		keyPairName = "c0-cliaas"
		subnetID    = "subnet-52d6c61b"

		name            string
		securityGroupID string
	)

	BeforeEach(func() {
		sess, err := session.NewSession()
		Expect(err).NotTo(HaveOccurred())

		ec2Client = ec2.New(sess, &aws.Config{
			Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
			Region:      aws.String(region),
		})

		awsClient = cliaas.NewAWSClient(ec2Client, vpc)
		Expect(awsClient).NotTo(BeNil())

		name = randSeq(10)
	})

	Context("Create", func() {
		var (
			instanceID string
			createErr  error
		)

		JustBeforeEach(func() {
			instanceID, createErr = awsClient.CreateVM(ami, vmType, name, keyPairName, subnetID, securityGroupID)
		})

		AfterEach(func() {
			if instanceID != "" {
				awsClient.DeleteVM(instanceID)
			}
		})

		It("creates a VM", func() {
			Expect(createErr).NotTo(HaveOccurred())
			Expect(instanceID).NotTo(Equal(""))

			_, err := awsClient.GetVMInfo(name)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("AssociateElasticIP", func() {
		var instanceID string

		BeforeEach(func() {
			var createErr error
			instanceID, createErr = awsClient.CreateVM(ami, vmType, name, keyPairName, subnetID, securityGroupID)
			Expect(createErr).NotTo(HaveOccurred())

			client := cliaas.NewAWSClient(ec2Client, vpc)

			err := client.WaitForStatus(instanceID, ec2.InstanceStateNameRunning)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if instanceID != "" {
				awsClient.DeleteVM(instanceID)
			}
		})

		It("associates the elastic IP to the instance", func() {
			err := awsClient.AssignPublicIP(instanceID, "52.2.195.24")
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
	return string(result)
}
