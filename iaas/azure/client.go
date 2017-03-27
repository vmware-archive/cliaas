package azure

import (
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	errwrap "github.com/pkg/errors"
)

const defaultResourceManagerEndpoint = "https://management.azure.com/"

type Client struct {
	virtualMachinesClient compute.VirtualMachinesClient
}

func NewClient(
	subscriptionID string,
	clientID string,
	clientSecret string,
	tenantID string,
	resourceGroupName string,
	resourceManagerEndpoint string,
) (*Client, error) {
	c := map[string]string{
		"AZURE_CLIENT_ID":       clientID,
		"AZURE_CLIENT_SECRET":   clientSecret,
		"AZURE_SUBSCRIPTION_ID": subscriptionID,
		"AZURE_TENANT_ID":       tenantID,
	}
	if err := checkEnvVar(&c); err != nil {
		return nil, errwrap.Wrap(err, "failed on check of env vars")
	}
	if resourceManagerEndpoint == "" {
		resourceManagerEndpoint = defaultResourceManagerEndpoint
	}

	spt, err := helpers.NewServicePrincipalTokenFromCredentials(c, resourceManagerEndpoint)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed to generate new service principal token")
	}
	client := compute.NewVirtualMachinesClient(subscriptionID)
	client.Authorizer = spt
	return &Client{virtualMachinesClient: client}, nil
}

func (s *Client) Delete(vmIdentifier string) error {
	return errors.New("not yet implemented")
}
func (s *Client) Replace(vmIdentifier string, imageIdentifier string) error {
	return errors.New("not yet implemented")
}

func checkEnvVar(envVars *map[string]string) error {
	var missingVars []string
	for varName, value := range *envVars {
		if value == "" {
			missingVars = append(missingVars, varName)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("Missing environment variables %v", missingVars)
	}
	return nil
}
