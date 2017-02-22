package aws_test

import (
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
			It("then it should list vms", func() {
				instances, err := awsClient.List("Test*", vpc)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(instances)).Should(BeEquivalentTo(1))
			})
		})
	})
})
