package gcp

import (
	"fmt"
	"regexp"
	"strings"

	errwrap "github.com/pkg/errors"
	compute "google.golang.org/api/compute/v1"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
)

type GoogleComputeClient interface {
	List(project string, zone string) (*compute.InstanceList, error)
	Delete(project string, zone string, instance string) (*compute.Operation, error)
	Insert(project string, zone string, instance *compute.Instance) (*compute.Operation, error)
}

type ClientAPI interface {
	CreateVM(instance compute.Instance) error
	DeleteVM(instanceId uint64) error
	GetVMInfo(filter Filter) (*compute.Instance, error)
	StopVM(instanceName string) error
}

type GCPClientAPI struct {
	credPath     string
	projectName  string
	zoneName     string
	googleClient GoogleComputeClient
}

//NewDefaultGoogleComputeClient -- builds a gcp client which connects to your gcp using `GOOGLE_APPLICATION_CREDENTIALS`
func NewDefaultGoogleComputeClient(credpath string) (GoogleComputeClient, error) {
	ctx := context.Background()
	hc, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		return nil, errwrap.Wrap(err, "we have an DefaultClient error")
	}

	c, err := compute.New(hc)
	if err != nil {
		return nil, errwrap.Wrap(err, "we have an compute.New error")
	}
	return &googleComputeClientWrapper{instanceService: c.Instances}, nil
}

func NewGCPClientAPI(configs ...func(*GCPClientAPI) error) (*GCPClientAPI, error) {
	gcpClient := new(GCPClientAPI)

	for _, cfg := range configs {
		err := cfg(gcpClient)
		if err != nil {
			return nil, errwrap.Wrap(err, "new GCP Client config loading error")
		}
	}

	if gcpClient.googleClient == nil {
		return nil, fmt.Errorf("You have an incomplete GCPClientAPI.googleClient")
	}
	return gcpClient, nil
}

func ConfigGoogleClient(value GoogleComputeClient) func(*GCPClientAPI) error {
	return func(gcpClient *GCPClientAPI) error {
		gcpClient.googleClient = value
		return nil
	}
}

func ConfigZoneName(value string) func(*GCPClientAPI) error {
	return func(gcpClient *GCPClientAPI) error {
		gcpClient.zoneName = value
		return nil
	}
}

func ConfigProjectName(value string) func(*GCPClientAPI) error {
	return func(gcpClient *GCPClientAPI) error {
		gcpClient.projectName = value
		return nil
	}
}

func (s *GCPClientAPI) CreateVM(instanceName string, sourceImageTarballUrl string) error {
	return nil
}

func (s *GCPClientAPI) DeleteVM(instanceId uint64) error {
	return nil
}

//GetVMInfo - gets the information on the first VM to match the given filter argument
// currently filter will only do a regex on teh tag||name regex fields against
// the List's result set
func (s *GCPClientAPI) GetVMInfo(filter Filter) (*compute.Instance, error) {
	list, err := s.googleClient.List(s.projectName, s.zoneName)
	if err != nil {
		return nil, errwrap.Wrap(err, "call.Do() on google client List() failed")
	}

	for _, item := range list.Items {
		var validID = regexp.MustCompile(filter.TagRegexString)
		var validName = regexp.MustCompile(filter.NameRegexString)
		taglist := strings.Join(item.Tags.Items, " ")
		tagMatch := validID.MatchString(taglist)
		nameMatch := validName.MatchString(item.Name)

		if tagMatch == true && nameMatch == true {
			return item, nil
		}
	}
	return nil, fmt.Errorf("No instance matches found")
}

func (s *GCPClientAPI) StopVM(instanceName string) error {
	return nil
}

type googleComputeClientWrapper struct {
	instanceService *compute.InstancesService
}

func (s *googleComputeClientWrapper) List(project string, zone string) (*compute.InstanceList, error) {
	return s.instanceService.List(project, zone).Do()
}

func (s *googleComputeClientWrapper) Delete(project string, zone string, instance string) (*compute.Operation, error) {
	return s.instanceService.Delete(project, zone, instance).Do()
}

func (s *googleComputeClientWrapper) Insert(project string, zone string, instance *compute.Instance) (*compute.Operation, error) {
	return s.instanceService.Insert(project, zone, instance).Do()
}
