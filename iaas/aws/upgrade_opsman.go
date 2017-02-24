package aws

import (
	"fmt"
	"time"

	errwrap "github.com/pkg/errors"
)

type UpgradeOpsMan struct {
	client ClientAPI
}

func NewUpgradeOpsMan(configs ...func(*UpgradeOpsMan) error) (*UpgradeOpsMan, error) {
	upgradeOpsMan := new(UpgradeOpsMan)
	for _, cfg := range configs {
		err := cfg(upgradeOpsMan)
		if err != nil {
			return nil, errwrap.Wrap(err, "new upgradeOpsMan config loading error")
		}
	}

	if upgradeOpsMan.client == nil {
		return nil, errwrap.New("must configure client")
	}
	return upgradeOpsMan, nil
}

func NewClientAPI(region, accessKey, secretKey, vpc string) (ClientAPI, error) {
	awsClient, err := CreateAWSClient(region, accessKey, secretKey)
	if err != nil {
		return nil, err
	}
	return NewAWSClientAPI(ConfigAWSClient(awsClient), ConfigVPC(vpc))
}

func ConfigClient(value ClientAPI) func(*UpgradeOpsMan) error {
	return func(upgradeOpsMan *UpgradeOpsMan) error {
		upgradeOpsMan.client = value
		return nil
	}
}

func (s *UpgradeOpsMan) Upgrade(name, ami, instanceType, ip string) error {
	fmt.Println("Getting VM with name prefix", name)
	instance, err := s.client.GetVMInfo(fmt.Sprintf("%s*", name))
	if err != nil {
		return err
	}
	fmt.Println("Stopping VM", *instance.InstanceId)
	err = s.client.StopVM(*instance)
	if err != nil {
		return err
	}
	t := time.Now()
	dateString := fmt.Sprintf("%d-%02d-%02dT%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	newName := fmt.Sprintf("%s - %s", name, dateString)

	fmt.Println("Creating new VM with name", newName)
	newInstance, err := s.client.CreateVM(*instance, ami, instanceType, newName)
	if err != nil {
		return err
	}

	fmt.Println("Waiting for VM", newName, "to be in running status")
	err = s.client.WaitForStartedVM(newName)
	if err != nil {
		return err
	}

	fmt.Println("Associating IP", ip, "to", newName)
	err = s.client.AssignPublicIP(*newInstance, ip)
	if err != nil {
		return err
	}
	fmt.Println("Delete VM", *instance.InstanceId)
	return s.client.DeleteVM(*instance)
}
