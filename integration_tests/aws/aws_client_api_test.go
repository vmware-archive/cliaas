package aws_test

import (
	"math/rand"
	"os"

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

			//createVM(instanceNameGUID)
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
	})
})

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
