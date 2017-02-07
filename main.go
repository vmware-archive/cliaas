package main

import (
	"fmt"
	"os"

	"github.com/c0-ops/cliaas/iaas/aws"
	errwrap "github.com/pkg/errors"
)

func main() {
	var instanceID string
	var client IaasClient
	var err error
	client, err = getAWSClient()

	if err != nil {
		fmt.Fprint(os.Stderr, errwrap.Wrap(err, "getaws client failed"))
		os.Exit(1)
	}

	instanceID, err = client.GetInstanceID()

	if err != nil {
		fmt.Fprint(os.Stderr, errwrap.Wrap(err, "getinstanceid failed"))
		os.Exit(1)
	}
	fmt.Println(instanceID)
	os.Exit(0)
}

func getAWSClient() (IaasClient, error) {
	region := os.Getenv("AWS_REGION")
	vpc := os.Getenv("VPC_ID")

	if len(os.Args) != 2 {
		return nil, fmt.Errorf("not enough args given to cli (requires one arg for instance tag regex)")
	}
	tagFilterString := os.Args[1]
	return iaasaws.NewClient(region, tagFilterString, vpc)
}

type IaasClient interface {
	GetInstanceID() (string, error)
}
