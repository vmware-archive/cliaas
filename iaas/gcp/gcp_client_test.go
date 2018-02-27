package gcp_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cliaas/iaas/gcp"
	"github.com/pivotal-cf/cliaas/iaas/gcp/gcpfakes"
	errwrap "github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
)

var _ = Describe("GCPClientAPI", func() {

	Describe("GCPClientAPI", func() {
		var client *Client
		var controlZone = "zone"
		var controlProject = "prj"
		var controlInstanceName = "blah"
		var controlInstanceTag = "hello"
		var controlDiskSizeGB = int64(10)

		Describe("given a CreateVM method and a valid instance", func() {
			var controlInstance compute.Instance
			Context("when called with a valid instance", func() {
				controlInstance = compute.Instance{}
				var fakeGoogleClient *gcpfakes.FakeGoogleComputeClient
				fakeOperation := &compute.Operation{
					Status: "DONE",
				}
				BeforeEach(func() {
					fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.InsertReturns(fakeOperation, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then the instance should be created in gcp", func() {
					err := client.CreateVM(controlInstance)
					Expect(fakeGoogleClient.InsertCallCount()).Should(Equal(1))
					project, zone, instance := fakeGoogleClient.InsertArgsForCall(0)
					Expect(project).Should(Equal(controlProject))
					Expect(zone).Should(Equal(controlZone))
					Expect(*instance).Should(Equal(controlInstance))
					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			Context("when called with an invalid instance", func() {
				BeforeEach(func() {
					controlInstance = compute.Instance{}
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeOperation := &compute.Operation{
						Error: &compute.OperationError{
							Errors: []*compute.OperationErrorErrors{
								&compute.OperationErrorErrors{
									Message: "Instance not found",
								},
							},
						},
						Status: "DONE",
					}
					fakeGoogleClient.InsertReturns(fakeOperation, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})
				It("then we should exit in error", func() {
					err := client.CreateVM(controlInstance)
					Expect(err).Should(HaveOccurred())
				})
			})

			Context("when gcp api call fails", func() {
				var controlErr = fmt.Errorf("Some GCP API Error")
				BeforeEach(func() {
					controlInstance = compute.Instance{}
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.InsertReturns(nil, controlErr)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})
				It("then we should exit in error", func() {
					err := client.CreateVM(controlInstance)
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(controlErr))
				})
			})
		})

		Describe("given a DeleteVM method and a valid instance", func() {
			Context("when called with the name of a valid instance", func() {
				var fakeGoogleClient *gcpfakes.FakeGoogleComputeClient
				fakeOperation := &compute.Operation{
					Status: "DONE",
				}
				BeforeEach(func() {
					fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.DeleteReturns(fakeOperation, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then the instance should be deleted from gcp", func() {
					err := client.DeleteVM(controlInstanceName)
					Expect(fakeGoogleClient.DeleteCallCount()).Should(Equal(1))
					project, zone, instanceName := fakeGoogleClient.DeleteArgsForCall(0)
					Expect(project).Should(Equal(controlProject))
					Expect(zone).Should(Equal(controlZone))
					Expect(instanceName).Should(Equal(controlInstanceName))
					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			Context("when called with an invalid (non-existent) instance name", func() {
				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeOperation := &compute.Operation{
						Error: &compute.OperationError{
							Errors: []*compute.OperationErrorErrors{
								&compute.OperationErrorErrors{
									Message: "Instance not found",
								},
							},
						},
						Status: "DONE",
					}
					fakeGoogleClient.DeleteReturns(fakeOperation, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})
				It("then we should exit in error", func() {
					err := client.DeleteVM(controlInstanceName)
					Expect(err).Should(HaveOccurred())
				})
			})

			Context("when gcp api call fails", func() {
				var controlErr = fmt.Errorf("Some GCP API Error")
				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.DeleteReturns(nil, controlErr)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})
				It("then we should exit in error", func() {
					err := client.DeleteVM(controlInstanceName)
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(controlErr))
				})
			})
		})

		Describe("given a StopVM method and a running instance", func() {
			Context("when called with the name of a valid running instance", func() {
				var fakeGoogleClient *gcpfakes.FakeGoogleComputeClient
				fakeOperation := &compute.Operation{
					Status: "DONE",
				}
				BeforeEach(func() {
					fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.StopReturns(fakeOperation, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})
				It("then the instance should be stopped in gcp", func() {
					err := client.StopVM(controlInstanceName)
					Expect(fakeGoogleClient.StopCallCount()).Should(Equal(1))
					project, zone, instanceName := fakeGoogleClient.StopArgsForCall(0)
					Expect(project).Should(Equal(controlProject))
					Expect(zone).Should(Equal(controlZone))
					Expect(instanceName).Should(Equal(controlInstanceName))
					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			Context("when called with a invalid (not-running) instance name", func() {
				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeOperation := &compute.Operation{
						Error: &compute.OperationError{
							Errors: []*compute.OperationErrorErrors{
								&compute.OperationErrorErrors{
									Message: "Instance not found",
								},
							},
						},
						Status: "DONE",
					}
					fakeGoogleClient.StopReturns(fakeOperation, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})
				It("then we should exit in error", func() {
					err := client.StopVM(controlInstanceName)
					Expect(err).Should(HaveOccurred())
				})
			})

			Context("when gcp api call fails", func() {
				var controlErr = fmt.Errorf("Some GCP API Error")
				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.StopReturns(nil, controlErr)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})
				It("then we should exit in error", func() {
					err := client.StopVM(controlInstanceName)
					Expect(err).Should(HaveOccurred())
					Expect(errwrap.Cause(err)).Should(Equal(controlErr))
				})
			})
		})

		Describe("given a GetVMInfo method and a filter object argument", func() {
			Context("when there is a matching instance", func() {
				controlInstanceList := createInstanceList(controlInstanceName, controlInstanceTag)
				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.ListReturns(controlInstanceList, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then it should yield the matching gcp instance", func() {
					inst, err := client.GetVMInfo(Filter{NameRegexString: controlInstanceName, TagRegexString: controlInstanceTag})
					Expect(inst).ShouldNot(BeNil())
					Expect(controlInstanceList.Items).To(HaveLen(1))
					Expect(inst).Should(Equal(controlInstanceList.Items[0]))
					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			Context("when there is no matching instance", func() {
				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.ListReturns(createInstanceList("nothing-to-match", "nothing-to-match"), nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then it should give an error", func() {
					inst, err := client.GetVMInfo(Filter{NameRegexString: "bbb", TagRegexString: "ddd"})
					Expect(inst).Should(BeNil())
					Expect(err).Should(HaveOccurred())
				})
			})

			Context("when there is empty instance set", func() {
				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.ListReturns(&compute.InstanceList{}, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then it should give an error", func() {
					inst, err := client.GetVMInfo(Filter{})
					Expect(inst).Should(BeNil())
					Expect(err).Should(HaveOccurred())
				})
			})
		})

		Describe("given a GetDisk method and a filter object argument", func() {
			Context("when there is a matching disk", func() {
				controlDisk := createDisk(controlInstanceName, controlDiskSizeGB)
				controlDiskList := &compute.DiskList{
					Items: []*compute.Disk{controlDisk},
				}

				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.DiskListReturns(controlDiskList, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then it should yield the filtered disk instance from a gcp disk list", func() {
					disk, err := client.GetDisk(Filter{NameRegexString: controlInstanceName})
					Expect(disk).ShouldNot(BeNil())
					Expect(controlDisk.SizeGb).To(BeEquivalentTo(controlDiskSizeGB))
					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			Context("when there is no matching disk", func() {
				BeforeEach(func() {
					var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
					fakeGoogleClient.DiskListReturns(&compute.DiskList{
						Items: []*compute.Disk{createDisk("nothing-to-match", 0)},
					}, nil)

					client, _ = NewClient(
						ConfigGoogleClient(fakeGoogleClient),
						ConfigZoneName(controlZone),
						ConfigProjectName(controlProject),
					)
				})

				It("then it should give an error", func() {
					disk, err := client.GetDisk(Filter{NameRegexString: "bbb"})
					Expect(err).Should(HaveOccurred())
					Expect(disk).Should(BeNil())
				})
			})
		})
	})

	Describe("given a NewGCPCLIentAPI()", func() {
		Context("when passed a incomplete/invalid set of configs", func() {
			var client *Client
			var err error
			var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
			BeforeEach(func() {
				client, err = NewClient(
					ConfigGoogleClient(fakeGoogleClient),
				)
			})
			It("then it should provide a properly initialized GCPCLientAPI object", func() {
				Expect(err).Should(HaveOccurred())
				Expect(client).Should(BeNil())
			})
		})

		Context("when passed a valid set of configs", func() {
			var client *Client
			var err error
			var controlZone = "zone"
			var controlProject = "prj"
			var fakeGoogleClient = new(gcpfakes.FakeGoogleComputeClient)
			BeforeEach(func() {
				client, err = NewClient(
					ConfigGoogleClient(fakeGoogleClient),
					ConfigZoneName(controlZone),
					ConfigProjectName(controlProject),
				)
			})
			It("then it should provide a properly initialized GCPCLientAPI object", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(client).ShouldNot(BeNil())
			})
		})
	})
})

func createInstanceList(name, tag string) *compute.InstanceList {
	return &compute.InstanceList{
		Items: []*compute.Instance{
			&compute.Instance{
				Status: InstanceRunning,
				Name:   name,
				Tags: &compute.Tags{
					Items: []string{
						tag,
					},
				},
			},
		},
	}
}

func createDisk(name string, sizeGB int64) *compute.Disk {
	return &compute.Disk{
		Status: InstanceRunning,
		Name:   name,
		SizeGb: sizeGB,
	}
}
