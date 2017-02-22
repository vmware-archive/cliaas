package aws_test

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/c0-ops/cliaas/iaas/aws"
	"github.com/c0-ops/cliaas/iaas/aws/awsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	errwrap "github.com/pkg/errors"
)

var _ = Describe("Aws Client", func() {

	Context("given a properly initialized client & a matching filter", func() {

		Context("when calling GetInstanceID", func() {
			var instanceID string
			var err error
			var controlInstanceID = "id-123"

			BeforeEach(func() {
				var c *Client
				fakeDescriber := new(awsfakes.FakeInstanceDescriber)
				fakeDescriber.DescribeInstancesReturns(createInstanceOutput([]*ec2.Instance{
					&ec2.Instance{
						InstanceId: &controlInstanceID,
					},
				}), nil)
				c, err = NewClient("us-east-1", "Ops*", "vpc-myguid")
				c.Describer = fakeDescriber
				Ω(err).ShouldNot(HaveOccurred())
				instanceID, err = c.GetInstanceID()
			})

			It("then it should return error and empty string", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(instanceID).Should(Equal(controlInstanceID))
			})
		})
	})

	Context("given a properly initialized client & no matches on filter config", func() {
		Context("when calling GetInstanceID", func() {
			var instanceID string
			var err error

			BeforeEach(func() {
				var c *Client
				fakeDescriber := new(awsfakes.FakeInstanceDescriber)
				fakeDescriber.DescribeInstancesReturns(createInstanceOutput([]*ec2.Instance{}), nil)
				c, err = NewClient("us-east-1", "Ops*", "vpc-myguid")
				c.Describer = fakeDescriber
				Ω(err).ShouldNot(HaveOccurred())
				instanceID, err = c.GetInstanceID()
			})

			It("then it should return error and empty string", func() {
				Ω(err).Should(HaveOccurred())
				Ω(errwrap.Cause(err)).Should(Equal(ErrCouldNotFindInstance))
				Ω(instanceID).Should(BeEmpty())
			})
		})
	})

	Context("when a client is improperly initialized", func() {
		It("then it should return error", func() {
			var err error
			Ω(func() {
				_, err = NewClient("", "", "")
			}).ShouldNot(Panic())
			Ω(err).Should(HaveOccurred())
		})
	})
})

func createInstanceOutput(instances []*ec2.Instance) *ec2.DescribeInstancesOutput {
	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			&ec2.Reservation{
				Instances: instances,
			},
		},
	}
}
