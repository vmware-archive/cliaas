package gcp_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"

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
			//createVM(instanceNameGUID)
		})
		AfterEach(func() {
			os.Remove(credFile.Name())
			//deleteVM(instanceNameGUID)
		})
		Context("when calling CreateVM with valid arguments", func() {
			var instanceNameGUID string
			var computeInstance compute.Instance
			BeforeEach(func() {
				var err error
				instanceNameGUID, err = newUUID()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeFalse())
				computeInstance = newComputeInstance(instanceNameGUID, zone, project)
				err = gcpClientAPI.CreateVM(computeInstance)
				Expect(err).ShouldNot(HaveOccurred())
			})
			AfterEach(func() {
				deleteVM(instanceNameGUID, project, zone)
			})
			It("then a new instance should have been created in GCP", func() {
				Expect(instanceExists(instanceNameGUID, project, zone)).Should(BeTrue())
			})
		})

		XContext("when calling DeleteVM with valid arguments for a running instance", func() {
			It("then the specified instance should not longer exist in GCP", func() {
				Expect(true).Should(BeFalse())
			})
		})

		XContext("when calling GetVMInfo with valid arguments for a running instance", func() {
			It("then we should recieve all info about the matching instance in GCP", func() {
				Expect(true).Should(BeFalse())
			})
		})

		XContext("when calling StopVM with valid arguments for a running instance", func() {
			It("then the specified instance should have been stopped in GCP", func() {
				Expect(true).Should(BeFalse())
			})
		})
	})
})

func instanceExists(instanceNameGUID string, project string, zone string) bool {
	ctx := context.Background()
	hc, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		Expect(err).ShouldNot(HaveOccurred())
	}
	c, err := compute.New(hc)
	if err != nil {
		Expect(err).ShouldNot(HaveOccurred())
	}

	call := c.Instances.List(project, zone)
	list, err := call.Do()

	for _, item := range list.Items {
		if item.Name == instanceNameGUID {
			return true
		}
	}
	return false
}

func newComputeInstance(instanceNameGUID string, zone string, project string) compute.Instance {
	machineType := "zones/" + zone + "/machineTypes/n1-standard-1"
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

	return compute.Instance{
		Name:              instanceNameGUID,
		MachineType:       machineType,
		NetworkInterfaces: nic,
		Disks:             disks,
	}

}

func createVM(instanceNameGUID string) {

}

func deleteVM(instanceNameGUID string, project string, zone string) {
	ctx := context.Background()
	hc, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		Expect(err).ShouldNot(HaveOccurred())
	}
	c, err := compute.New(hc)
	if err != nil {
		Expect(err).ShouldNot(HaveOccurred())
	}

	call := c.Instances.Delete(project, zone, instanceNameGUID)
	_, err = call.Do()
	Expect(err).ShouldNot(HaveOccurred())
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
