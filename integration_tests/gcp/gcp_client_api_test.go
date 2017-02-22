package gcp_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"

	. "github.com/c0-ops/cliaas/iaas/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPCLientAPI", func() {
	Describe("given a gcpclientapi and a gcp client which targets a valid gcp account/creds", func() {
		var credContents = os.Getenv("GCP_CREDS")
		var project = os.Getenv("GCP_PROJECT")
		var zone = os.Getenv("GCP_ZONE")
		var credFile *os.File
		var gcpClientAPI *GCPClientAPI
		var instanceNameGUID, err = newUUID()
		BeforeEach(func() {
			var err error
			credFile, err = ioutil.TempFile("/tmp", "gcpCred")
			defer credFile.Close()
			Expect(err).ShouldNot(HaveOccurred())
			credFile.Write([]byte(credContents))
			gcpClient, err := NewDefaultGoogleComputeClient(credFile.Name())
			Expect(err).ShouldNot(HaveOccurred())
			gcpClientAPI, err = NewGCPClientAPI(
				ConfigGoogleClient(gcpClient),
				ConfigZoneName(zone),
				ConfigProjectName(project),
			)
			Expect(err).ShouldNot(HaveOccurred())
			createVM(instanceNameGUID, project, zone)
		})
		AfterEach(func() {
			if instanceExists(instanceNameGUID, project, zone) {
				deleteVM(instanceNameGUID, project, zone)
			}

			if diskExists(instanceNameGUID, project, zone) {
				deleteDisk(instanceNameGUID, project, zone)
			}

			os.Remove(credFile.Name())
		})
		Context("when calling CreateVM with valid arguments", func() {
			var instanceNameGUIDCreate string
			var computeInstance compute.Instance
			BeforeEach(func() {
				var err error
				instanceNameGUIDCreate, err = newUUID()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(instanceExists(instanceNameGUIDCreate, project, zone)).Should(BeFalse())
				computeInstance = newComputeInstance(instanceNameGUIDCreate, project, zone)
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
				err = gcpClientAPI.DeleteVM(instanceNameGUID)
				Expect(err).ShouldNot(HaveOccurred())
				waitForDelete(instanceNameGUID, project, zone, instanceExists)
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeFalse())
			})
		})

		Context("when calling GetVMInfo with valid arguments for a running instance", func() {
			var controlTag = "cliaas"
			It("then we should recieve all info about the matching instance (by name) in GCP", func() {
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue())
				Expect(instanceStopped(instanceNameGUID, project, zone)).Should(BeFalse())
				instance, err := gcpClientAPI.GetVMInfo(Filter{
					NameRegexString: instanceNameGUID,
					TagRegexString:  "",
				})
				Expect(err).ShouldNot(HaveOccurred())
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
				instance, err := gcpClientAPI.GetVMInfo(Filter{
					NameRegexString: "",
					TagRegexString:  controlTag,
				})
				Expect(err).ShouldNot(HaveOccurred())
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
				err = gcpClientAPI.StopVM(instanceNameGUID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue(), "does the instance exist?")
				Eventually(func() bool {
					return instanceStopped(instanceNameGUID, project, zone)
				}, 30, 5).Should(BeTrue(), "is the instance stopped?")
			})
		})
	})
})

func getCompute() *compute.Service {
	ctx := context.Background()
	hc, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		Expect(err).ShouldNot(HaveOccurred())
	}
	c, err := compute.New(hc)
	if err != nil {
		Expect(err).ShouldNot(HaveOccurred())
	}

	return c
}

func deleteDisk(instanceNameGUID string, project string, zone string) {
	c := getCompute()

	call := c.Disks.Delete(project, zone, instanceNameGUID)
	_, err := call.Do()
	Expect(err).ShouldNot(HaveOccurred())

	waitForDelete(instanceNameGUID, project, zone, diskExists)
}

func waitForDelete(instanceNameGUID string, project string, zone string, exists func(string, string, string) bool) {
	var waitSeconds time.Duration = 10 * time.Second

	for {
		if !exists(instanceNameGUID, project, zone) {
			break
		}

		time.Sleep(waitSeconds)
	}
}

func diskExists(instanceNameGUID string, project string, zone string) bool {
	c := getCompute()
	call := c.Disks.List(project, zone)
	list, err := call.Do()
	Expect(err).ShouldNot(HaveOccurred())

	for _, item := range list.Items {
		if item.Name == instanceNameGUID {
			return true
		}
	}
	return false
}

func instanceExists(instanceNameGUID string, project string, zone string) bool {
	c := getCompute()
	call := c.Instances.List(project, zone)
	list, err := call.Do()
	Expect(err).ShouldNot(HaveOccurred())

	for _, item := range list.Items {
		if item.Name == instanceNameGUID {
			return true
		}
	}

	return false
}

func instanceStopped(instanceNameGUID string, project string, zone string) bool {
	c := getCompute()
	call := c.Instances.List(project, zone)
	list, err := call.Do()
	Expect(err).ShouldNot(HaveOccurred())

	for _, item := range list.Items {
		if item.Name == instanceNameGUID && (item.Status == "STOPPED" || item.Status == "STOPPING") {
			return true
		}
	}

	return false
}

func newComputeInstance(instanceNameGUID string, project string, zone string) compute.Instance {
	machineType := "zones/" + zone + "/machineTypes/f1-micro"
	nic := []*compute.NetworkInterface{
		&compute.NetworkInterface{
			Kind:    "compute#networkInterface",
			Name:    "nic0",
			Network: "global/networks/default",
		},
	}
	sourceImage := "projects/debian-cloud/global/images/family/debian-8"
	disks := []*compute.AttachedDisk{
		&compute.AttachedDisk{
			Boot: true,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				SourceImage: sourceImage,
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

func createVM(instanceNameGUID string, project string, zone string) {
	instance := newComputeInstance(instanceNameGUID, project, zone)

	c := getCompute()

	call := c.Instances.Insert(project, zone, &instance)
	_, err := call.Do()
	Expect(err).ShouldNot(HaveOccurred())

}

func deleteVM(instanceNameGUID string, project string, zone string) {
	c := getCompute()

	call := c.Instances.Delete(project, zone, instanceNameGUID)
	_, err := call.Do()
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
	return fmt.Sprintf("cliaas-int-test-vm-%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
