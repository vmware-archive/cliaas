package gcp

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	errwrap "github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

type GoogleComputeClient interface {
	List(project string, zone string) (*compute.InstanceList, error)
	Delete(project string, zone string, instanceName string) (*compute.Operation, error)
	Insert(project string, zone string, instance *compute.Instance) (*compute.Operation, error)
	ImageInsert(project string, image *compute.Image, timeout time.Duration) (*compute.Operation, error)
	Stop(project string, zone string, instanceName string) (*compute.Operation, error)
}

type ClientAPI interface {
	CreateVM(instance compute.Instance) error
	DeleteVM(instanceName string) error
	GetVMInfo(filter Filter) (*compute.Instance, error)
	StopVM(instanceName string) error
	CreateImage(tarball string) (string, error)
	SwapLb(identifier string, vmidentifiers []string) error
}

type Client struct {
	projectName  string
	zoneName     string
	googleClient GoogleComputeClient
	timeout      time.Duration
}

//NewDefaultGoogleComputeClient -- builds a gcp client which connects to your gcp using `GOOGLE_APPLICATION_CREDENTIALS`
func NewDefaultGoogleComputeClient(credpath string) (GoogleComputeClient, error) {
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credpath)
	if err != nil {
		return nil, errwrap.Wrap(err, "couldnt set credentials ENV Var")
	}

	ctx := context.Background()
	hc, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		return nil, errwrap.Wrap(err, "we have a DefaultClient error")
	}

	c, err := compute.New(hc)
	if err != nil {
		return nil, errwrap.Wrap(err, "we have a compute.New error")
	}
	return &googleComputeClientWrapper{instanceService: c.Instances, imageService: c.Images, ctx: ctx}, nil
}

func NewClient(configs ...func(*Client) error) (*Client, error) {
	gcpClient := new(Client)
	gcpClient.timeout = 60 * time.Second

	for _, cfg := range configs {
		err := cfg(gcpClient)
		if err != nil {
			return nil, errwrap.Wrap(err, "new GCP Client config loading error")
		}
	}

	if gcpClient.googleClient == nil {
		return nil, fmt.Errorf("You have an incomplete GCPClientAPI.googleClient")
	}

	if gcpClient.zoneName == "" {
		return nil, fmt.Errorf("You have an incomplete GCPClientAPI.zoneName")
	}

	if gcpClient.projectName == "" {
		return nil, fmt.Errorf("You have an incomplete GCPClientAPI.projectName")
	}
	return gcpClient, nil
}

func ConfigTimeout(value time.Duration) func(*Client) error {
	return func(gcpClient *Client) error {
		gcpClient.timeout = value * time.Second
		return nil
	}
}

func ConfigGoogleClient(value GoogleComputeClient) func(*Client) error {
	return func(gcpClient *Client) error {
		gcpClient.googleClient = value
		return nil
	}
}

func ConfigZoneName(value string) func(*Client) error {
	return func(gcpClient *Client) error {
		gcpClient.zoneName = value
		return nil
	}
}

func ConfigProjectName(value string) func(*Client) error {
	return func(gcpClient *Client) error {
		gcpClient.projectName = value
		return nil
	}
}

func (s *Client) CreateImage(tarball string) (string, error) {
	diskName := fmt.Sprintf("opsman-disk-%v", time.Now().Format("2006-01-02-15-04-05"))
	_, err := s.googleClient.ImageInsert(s.projectName, &compute.Image{
		Name: diskName,
		RawDisk: &compute.ImageRawDisk{
			Source: fmt.Sprintf("http://storage.googleapis.com/%v", tarball),
		},
	}, s.timeout)
	diskURL := fmt.Sprintf("projects/%s/global/images/%s", s.projectName, diskName)
	return diskURL, err
}

func (s *Client) CreateVM(instance compute.Instance) error {
	operation, err := s.googleClient.Insert(s.projectName, s.zoneName, &instance)
	if err != nil {
		return errwrap.Wrap(err, "call to googleclient.Insert yielded error")
	}

	if operation.Error != nil {
		return errors.New("unexpected errors from operation response from google client")
	}

	return nil
}

func (s *Client) DeleteVM(instanceName string) error {
	operation, err := s.googleClient.Delete(s.projectName, s.zoneName, instanceName)
	if err != nil {
		return errwrap.Wrap(err, "call to googleclient.Delete yielded error")
	}

	if operation.Error != nil {
		return errors.New("unexpected errors from operation response from google client")
	}

	return nil
}

//StopVM - will try to stop the VM with the given name
func (s *Client) StopVM(instanceName string) error {
	operation, err := s.googleClient.Stop(s.projectName, s.zoneName, instanceName)
	if err != nil {
		return errwrap.Wrap(err, "call to googleclient.Stop yielded error")
	}

	if operation.Error != nil {
		return errors.New("unexpected errors from operation response from google client")
	}

	return nil
}

//GetVMInfo - gets the information on the first VM to match the given filter argument
// currently filter will only do a regex on teh tag||name regex fields against
// the List's result set
func (s *Client) GetVMInfo(filter Filter) (*compute.Instance, error) {
	return s.getVMInfo(filter, InstanceRunning)
}

func (s *Client) getVMInfo(filter Filter, status string) (*compute.Instance, error) {
	list, err := s.googleClient.List(s.projectName, s.zoneName)
	if err != nil {
		return nil, errwrap.Wrap(err, "call List on google client failed")
	}

	for _, item := range list.Items {
		var validID = regexp.MustCompile(filter.TagRegexString)
		var validName = regexp.MustCompile(filter.NameRegexString)
		taglist := strings.Join(item.Tags.Items, " ")
		tagMatch := validID.MatchString(taglist)
		nameMatch := validName.MatchString(item.Name)

		if tagMatch &&
			nameMatch &&
			(status == InstanceAll || item.Status == InstanceRunning) {
			return item, nil
		}
	}
	return nil, fmt.Errorf("No instance matches found")
}

func (s *Client) WaitForStatus(vmName string, desiredStatus string) error {
	errChannel := make(chan error)
	go func() {
		for {
			vmInfo, err := s.getVMInfo(Filter{NameRegexString: vmName}, InstanceAll)
			if err != nil {
				errChannel <- errwrap.Wrap(err, "GetVMInfo call failed")
				return
			}

			if vmInfo.Status == desiredStatus {
				errChannel <- nil
				return
			}
		}
	}()
	select {
	case res := <-errChannel:
		return res
	case <-time.After(s.timeout):
		return errors.New("polling for status timed out")
	}
}

func (s *Client) SwapLb(identifier string, vmidentifiers []string) error {
	return nil
}

type googleComputeClientWrapper struct {
	imageService    *compute.ImagesService
	instanceService *compute.InstancesService
	ctx             context.Context
}

func (s *googleComputeClientWrapper) List(project string, zone string) (*compute.InstanceList, error) {
	return s.instanceService.List(project, zone).Context(s.ctx).Do()
}

func (s *googleComputeClientWrapper) Delete(project string, zone string, instance string) (*compute.Operation, error) {
	return s.instanceService.Delete(project, zone, instance).Context(s.ctx).Do()
}

func (s *googleComputeClientWrapper) Stop(project string, zone string, instance string) (*compute.Operation, error) {
	vmInstance, err := s.instanceService.Get(project, zone, instance).Context(s.ctx).Do()
	if err != nil {
		return nil, errwrap.Wrap(err, "failed getting vm instance")
	}

	if len(vmInstance.NetworkInterfaces) > 0 && len(vmInstance.NetworkInterfaces[0].AccessConfigs) > 0 {
		accessConfigName := vmInstance.NetworkInterfaces[0].AccessConfigs[0].Name
		nicName := vmInstance.NetworkInterfaces[0].Name
		operation, err := s.instanceService.DeleteAccessConfig(project, zone, instance, accessConfigName, nicName).Context(s.ctx).Do()
		if err != nil {
			return operation, errwrap.Wrap(err, "could not delete access config")
		}

	}

	return s.instanceService.Stop(project, zone, instance).Context(s.ctx).Do()
}

func (s *googleComputeClientWrapper) Insert(project string, zone string, instance *compute.Instance) (*compute.Operation, error) {
	return s.instanceService.Insert(project, zone, instance).Context(s.ctx).Do()
}

func (s *googleComputeClientWrapper) ImageInsert(project string, image *compute.Image, timeout time.Duration) (*compute.Operation, error) {
	operation, err := s.imageService.Insert(project, image).Context(s.ctx).Do()
	if err != nil {
		return operation, errwrap.Wrap(err, "disk image insert failed")
	}

	errChannel := make(chan error)
	go func() {
		for {
			image, err := s.imageService.Get(project, image.Name).Context(s.ctx).Do()
			if err != nil {
				errChannel <- errwrap.Wrap(err, "image get failed")
			}

			if image.Status == ImageReady {
				errChannel <- nil
			}

			if image.Status == ImageFailed {
				errChannel <- errors.New("image creation failed")
			}
		}
	}()
	select {
	case res := <-errChannel:
		return operation, res
	case <-time.After(timeout):
		return nil, errors.New("polling for status timed out")
	}
	return nil, nil
}
