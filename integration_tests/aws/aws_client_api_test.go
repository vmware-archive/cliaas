package aws_test

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	. "github.com/c0-ops/cliaas/iaas/aws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AwsClient", func() {
	Describe("given a awsclientapi and a aws client which targets a valid aws account/creds", func() {

		//var vpc = os.Getenv("AWS_VPC")
		var awsClient AWSClient
		var accessKey = os.Getenv("AWS_ACCESS_KEY")
		var secretKey = os.Getenv("AWS_SECRET_KEY")
		var region = os.Getenv("AWS_REGION")
		var vpc = os.Getenv("AWS_VPC")
		BeforeEach(func() {
			var err error
			awsClient, _ = CreateAWSClient(region, accessKey, secretKey)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(awsClient).ShouldNot(BeNil())
		})
		Context("GetVMInfo", func() {
			It("then it should create a vm, list vm and destroy vm", func() {
				ami := "ami-0b33d91d"
				vmType := "t2.micro"
				name := randSeq(10)
				keyPairName := "c0-cliaas"
				subnetID := "subnet-52d6c61b"
				securityGroupID := ""
				instance, err := awsClient.Create(ami, vmType, name, keyPairName, subnetID, securityGroupID)
				Expect(err).NotTo(HaveOccurred())
				Expect(instance).ShouldNot(BeNil())
				instances, err := awsClient.List(name, vpc)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(instances)).Should(BeEquivalentTo(1))
				err = awsClient.Delete(*instance.InstanceId)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("CreateVM", func() {
			It("then it should create and delete the VM", func() {
				ami := "ami-0b33d91d"
				vmType := "t2.micro"
				name := randSeq(10)
				keyPairName := "c0-cliaas"
				subnetID := "subnet-52d6c61b"
				securityGroupID := ""
				instance, err := awsClient.Create(ami, vmType, name, keyPairName, subnetID, securityGroupID)
				Expect(err).NotTo(HaveOccurred())
				Expect(instance).ShouldNot(BeNil())
				err = awsClient.Delete(*instance.InstanceId)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("Associate EIP to created VM", func() {
			It("then it should create vm and associate to eip", func() {
				ami := "ami-0b33d91d"
				vmType := "t2.micro"
				name := randSeq(10)
				keyPairName := "c0-cliaas"
				subnetID := "subnet-52d6c61b"
				securityGroupID := ""
				fmt.Println("Name", name)
				instance, err := awsClient.Create(ami, vmType, name, keyPairName, subnetID, securityGroupID)
				Expect(err).NotTo(HaveOccurred())
				Expect(instance).ShouldNot(BeNil())
				clientAPI, err := NewAWSClientAPI(
					ConfigAWSClient(awsClient),
					ConfigVPC(vpc),
				)
				Expect(err).NotTo(HaveOccurred())
				err = clientAPI.WaitForStartedVM(name)
				Expect(err).NotTo(HaveOccurred())
				err = awsClient.AssociateElasticIP(*instance.InstanceId, "52.1.191.81")
				Expect(err).NotTo(HaveOccurred())
				err = awsClient.Delete(*instance.InstanceId)
				Expect(err).NotTo(HaveOccurred())
			})
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
