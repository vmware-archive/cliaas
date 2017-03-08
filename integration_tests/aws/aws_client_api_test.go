package aws_test

import (
	"math/rand"
	"os"
	"time"

	iaasaws "github.com/aws/aws-sdk-go/aws"
	"github.com/pivotal-cf/cliaas/iaas/aws"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AwsClient", func() {
	var (
		awsClient aws.AWSClient
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

		ec2Client := ec2.New(sess, &iaasaws.Config{
			Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
			Region:      iaasaws.String(region),
		})

		awsClient = aws.NewAWSClient(ec2Client)
		Expect(awsClient).NotTo(BeNil())

		name = randSeq(10)
	})

	Context("Create", func() {
		var (
			instance  *ec2.Instance
			createErr error
		)

		JustBeforeEach(func() {
			instance, createErr = awsClient.Create(ami, vmType, name, keyPairName, subnetID, securityGroupID)
		})

		AfterEach(func() {
			if instance != nil {
				awsClient.Delete(*instance.InstanceId)
			}
		})

		It("creates a VM", func() {
			Expect(createErr).NotTo(HaveOccurred())
			Expect(instance).NotTo(BeNil())

			instances, err := awsClient.List(name, vpc)
			Expect(err).NotTo(HaveOccurred())
			Expect(instances).To(HaveLen(1))
		})
	})

	Context("AssociateElasticIP", func() {
		var instance *ec2.Instance

		BeforeEach(func() {
			var createErr error
			instance, createErr = awsClient.Create(ami, vmType, name, keyPairName, subnetID, securityGroupID)
			Expect(createErr).NotTo(HaveOccurred())

			client := aws.NewClient(awsClient, vpc)

			err := client.WaitForStatus(*instance.InstanceId, ec2.InstanceStateNameRunning)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if instance != nil {
				awsClient.Delete(*instance.InstanceId)
			}
		})

		It("it associates the elastic IP to the instance", func() {
			err := awsClient.AssociateElasticIP(*instance.InstanceId, "52.1.191.81")
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
