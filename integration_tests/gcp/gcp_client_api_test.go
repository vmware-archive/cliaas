package gcp_test

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"time"

	"google.golang.org/api/compute/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cliaas/iaas/gcp"
	"fmt"
)

var gcpClient GoogleComputeClient

const diskSizeGB = int64(11)

var _ = Describe("GCPClientAPI", func() {
	Describe("given a gcpclientapi and a gcp client which targets a valid gcp account/creds", func() {
		var gcpClientAPI *Client
		var credContents = os.Getenv("GCP_CREDS")
		var project = os.Getenv("GCP_PROJECT")
		var zone = os.Getenv("GCP_ZONE")
		var credFile *os.File
		var instanceNameGUID string
		var replacedInstanceName string
		BeforeEach(func() {
			var err error
			instanceNameGUID, err = newUUID()
			Expect(err).NotTo(HaveOccurred())
			credFile, err = ioutil.TempFile("/tmp", "gcpCred")
			Expect(err).ShouldNot(HaveOccurred())
			defer func() {
				_ = credFile.Close()
			}()
			_, err = credFile.Write([]byte(credContents))
			Expect(err).NotTo(HaveOccurred())
			Expect(err).ShouldNot(HaveOccurred())

			gcpClient, err = NewDefaultGoogleComputeClient(credFile.Name())

			gcpClientAPI, err = NewClient(
				ConfigGoogleClient(gcpClient),
				ConfigZoneName(zone),
				ConfigProjectName(project),
			)
			Expect(err).ShouldNot(HaveOccurred())

			instance := newComputeInstanceFromSourceImage(instanceNameGUID, "projects/debian-cloud/global/images/family/debian-8", project, zone)
			gcpClientAPI.CreateVM(instance)
		})

		AfterEach(func() {
			if instanceExists(instanceNameGUID, project, zone) {
				deleteVM(instanceNameGUID, project, zone)
			}

			if diskExists(instanceNameGUID, project, zone) {
				deleteDisk(instanceNameGUID, project, zone)
			}

			if instanceExists(replacedInstanceName, project, zone) {
				deleteVM(replacedInstanceName, project, zone)
			}

			if diskExists(replacedInstanceName, project, zone) {
				deleteDisk(replacedInstanceName, project, zone)
			}

			_ = os.Remove(credFile.Name())
		})

		Context("when calling CreateVM with valid arguments", func() {
			var instanceNameGUIDCreate string
			var computeInstance compute.Instance
			BeforeEach(func() {
				var err error
				instanceNameGUIDCreate, err = newUUID()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(instanceExists(instanceNameGUIDCreate, project, zone)).Should(BeFalse())
				computeInstance = newComputeInstanceFromSourceImage(instanceNameGUIDCreate, "projects/debian-cloud/global/images/family/debian-8", project, zone)
				err = gcpClientAPI.CreateVM(computeInstance)
				Expect(err).ShouldNot(HaveOccurred())
			})
			AfterEach(func() {
				deleteVM(instanceNameGUIDCreate, project, zone)
			})
			It("then a new instance should have been created in GCP", func() {
				Expect(instanceExists(instanceNameGUIDCreate, project, zone)).Should(BeTrue())
			})
		})

		Context("when calling DeleteVM with valid arguments for a running instance", func() {
			It("then the specified instance should not longer exist in GCP", func() {
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue())
				err := gcpClientAPI.DeleteVM(instanceNameGUID)
				Expect(err).ShouldNot(HaveOccurred())
				waitForDelete(instanceNameGUID, project, zone, instanceExists)
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeFalse())
			})
		})

		Context("when calling GetVMInfo with valid arguments for a running instance", func() {
			var controlTag = "cliaas"
			It("then we should receive all info about the matching instance (by name) in GCP", func() {
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue())
				Expect(instanceStopped(instanceNameGUID, project, zone)).Should(BeFalse())

				var instance *compute.Instance
				Eventually(func() error {
					var err error
					instance, err = gcpClientAPI.GetVMInfo(Filter{
						NameRegexString: instanceNameGUID,
						TagRegexString:  "",
					})
					return err
				}, "2m", "10s").Should(Succeed())
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue())
				Expect(instanceStopped(instanceNameGUID, project, zone)).Should(BeFalse())
				Expect(instance.Name).Should(Equal(instanceNameGUID))
				Expect(instance.Tags).ShouldNot(BeNil())
				Expect(instance.Tags.Items).Should(HaveLen(1))
				Expect(instance.Tags.Items[0]).Should(Equal(controlTag))
			})

			It("then we should receive all info about the matching instance (by tag) in GCP", func() {
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue())
				Expect(instanceStopped(instanceNameGUID, project, zone)).Should(BeFalse())

				var instance *compute.Instance
				Eventually(func() error {
					var err error
					instance, err = gcpClientAPI.GetVMInfo(Filter{
						NameRegexString: instanceNameGUID,
						TagRegexString:  controlTag,
					})
					return err
				}, "2m", "10s").Should(Succeed())
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue())
				Expect(instanceStopped(instanceNameGUID, project, zone)).Should(BeFalse())
				Expect(instance.Name).Should(Equal(instanceNameGUID))
				Expect(instance.Tags).ShouldNot(BeNil())
				Expect(instance.Tags.Items).Should(HaveLen(1))
				Expect(instance.Tags.Items[0]).Should(Equal(controlTag))
			})
		})

		Context("when calling StopVM with valid arguments for a running instance", func() {
			It("then the specified instance should have been stopped in GCP", func() {
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue())
				Expect(instanceStopped(instanceNameGUID, project, zone)).Should(BeFalse())
				err := gcpClientAPI.StopVM(instanceNameGUID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue(), "does the instance exist?")
				Eventually(func() bool {
					return instanceStopped(instanceNameGUID, project, zone)
				}, 30, 5).Should(BeTrue(), "is the instance stopped?")
			})
		})

		Context("when calling GetDisk with valid arguments for a running instance", func() {
			It("then a valid disk with the specified size should be returned", func() {
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue())
				Expect(instanceStopped(instanceNameGUID, project, zone)).Should(BeFalse())
				disk, err := gcpClientAPI.Disk(Filter{
					NameRegexString: instanceNameGUID + "*",
				})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(disk).ShouldNot(BeNil())
				Expect(disk.SizeGb).Should(Equal(diskSizeGB))
			})
		})

		Describe("Replace", func() {
			JustBeforeEach(func() {
				err := gcpClientAPI.WaitForStatus(instanceNameGUID, InstanceRunning)
				Expect(err).ShouldNot(HaveOccurred())

				err = gcpClientAPI.Replace(instanceNameGUID, "ops-manager-us/pcf-gcp-2.0-build.255.tar.gz", diskSizeGB)
				Expect(err).ShouldNot(HaveOccurred())
			})

			Context("when called on a vm with a name matching the given regex", func() {
				It("should delete the matching VM and spin up a the new VM in its place", func() {
					Expect(instanceTerminated(instanceNameGUID, project, zone)).To(BeTrue())

					var instance *compute.Instance
					Eventually(func() error {
						var err error
						instance, err = gcpClientAPI.GetVMInfo(Filter{
							NameRegexString: instanceNameGUID + "-*",
						})
						replacedInstanceName = instance.Name
						return err
					}, "2m", "10s").Should(Succeed())

					Expect(instanceRunning(replacedInstanceName, project, zone)).To(BeTrue())

					disk, err := gcpClientAPI.Disk(Filter{
						NameRegexString: replacedInstanceName,
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(disk).ShouldNot(BeNil())
					Expect(disk.SizeGb).Should(Equal(diskSizeGB))
				})
			})
		})
	})
})

func deleteDisk(instanceNameGUID string, project string, zone string) {
	_, err := gcpClient.Delete(project, zone, instanceNameGUID)
	Expect(err).ShouldNot(HaveOccurred())

	waitForDelete(instanceNameGUID, project, zone, diskExists)
}

func waitForDelete(instanceNameGUID string, project string, zone string, exists func(string, string, string) bool) {
	var waitSeconds = 10 * time.Second

	for {
		if !exists(instanceNameGUID, project, zone) {
			break
		}

		time.Sleep(waitSeconds)
	}
}

func diskExists(instanceNameGUID string, project string, zone string) bool {
	list, err := gcpClient.List(project, zone)
	Expect(err).ShouldNot(HaveOccurred())

	for _, item := range list.Items {
		if item.Name == instanceNameGUID {
			return true
		}
	}
	return false
}

func instanceExists(instanceNameGUID string, project string, zone string) bool {
	list, err := gcpClient.List(project, zone)
	Expect(err).ShouldNot(HaveOccurred())

	for _, item := range list.Items {
		if item.Name == instanceNameGUID {
			return true
		}
	}

	return false
}

func instanceStopped(instanceNameGUID string, project string, zone string) bool {
	list, err := gcpClient.List(project, zone)
	Expect(err).ShouldNot(HaveOccurred())

	for _, item := range list.Items {
		if item.Name == instanceNameGUID && (item.Status == "STOPPED" || item.Status == "STOPPING") {
			return true
		}
	}

	return false
}

func instanceRunning(instanceNameGUID string, project string, zone string) bool {
	list, err := gcpClient.List(project, zone)
	Expect(err).ShouldNot(HaveOccurred())

	for _, item := range list.Items {
		if item.Name == instanceNameGUID && (item.Status == "RUNNING") {
			return true
		}
	}

	return false
}

func instanceTerminated(instanceNameGUID string, project string, zone string) bool {
	list, err := gcpClient.List(project, zone)
	Expect(err).ShouldNot(HaveOccurred())

	for _, item := range list.Items {
		if item.Name == instanceNameGUID && (item.Status == "TERMINATED") {
			return true
		}

		if item.Name == instanceNameGUID {
			println("********", item.Status)
		}
	}


	return false
}

func newComputeInstanceFromSourceImage(instanceNameGUID string, sourceImage string, project string, zone string) compute.Instance {
	machineType := "zones/" + zone + "/machineTypes/f1-micro"
	nic := []*compute.NetworkInterface{
		&compute.NetworkInterface{
			Kind:    "compute#networkInterface",
			Name:    "nic0",
			Network: "global/networks/default",
		},
	}

	disks := []*compute.AttachedDisk{
		&compute.AttachedDisk{
			Boot: true,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				SourceImage: sourceImage,
				DiskSizeGb:  diskSizeGB,
			},
		},
	}
	tags := &compute.Tags{
		Items: []string{
			"cliaas",
		},
	}

	return compute.Instance{
		Name:              instanceNameGUID,
		MachineType:       machineType,
		NetworkInterfaces: nic,
		Disks:             disks,
		Tags:              tags,
	}
}

func deleteVM(instanceNameGUID string, project string, zone string) {
	_, err := gcpClient.Delete(project, zone, instanceNameGUID)
	Expect(err).ShouldNot(HaveOccurred())

	waitForDelete(instanceNameGUID, project, zone, instanceExists)
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("cliaas-int-test-vm-%x", uuid[0:4]), nil
}
