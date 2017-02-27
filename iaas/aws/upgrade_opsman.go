package aws

import (
	"fmt"
	"time"
)

type UpgradeOpsMan struct {
	client Client
}

func NewUpgradeOpsMan(client Client) *UpgradeOpsMan {
	return &UpgradeOpsMan{
		client: client,
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
	return s.client.DeleteVM(*instance.InstanceId)
}
