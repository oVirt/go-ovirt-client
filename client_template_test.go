package ovirtclient_test

import (
	"fmt"
	ovirtclient "github.com/ovirt/go-ovirt-client"
	"testing"
)

// TODO:
// - add test to create template from disk on SD + add it to the mock - https://github.com/oVirt/go-ovirt-client/issues/52
// create VM
// create Template

// - add multiple SDs to the mock + add helper function to get second SD https://github.com/oVirt/go-ovirt-client/issues/54

// - add storagedomain_remove_disk
// - cleanup remove disk after copy operation
// - move disk creation to Assert function

func TestTemplateCreation(t *testing.T){
	helper := getHelper(t)
	client := helper.GetClient()

	vm := assertCanCreateVM(
		t,
		helper,
		ovirtclient.CreateVMParams().MustWithName(fmt.Sprintf("template_creation_test_%s", helper.GenerateRandomID(5))),
	)
	disk := assertCanCreateDisk(t, client, helper)
	attachment := assertCanAttachDisk(t, vm, disk)
	assertDiskAttachmentMatches(t, attachment, disk, vm)

}


func TestTemplateDiskCopyToSD(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	//diskName := fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5))
	disk  := assertCanCreateDisk(t,client,helper)
	newSD := "abc480ae-498f-42ad-ad84-e8f1380eb79b"

	// create a disk
	// create the template
	// get the Disk ID from the template
	// copy template disk to different SD
	client.CopyTemplateDiskToStorageDomain(disk.ID(),newSD)


	// in case we copy disk which belong to a certain template its ID should remain the same on every SD
	_, err := client.GetStorageDomainDisk(newSD,disk.ID())

	if err != nil {
		t.Fatalf("Couldnt find disk %s on StorageDomain %s", disk.ID(),newSD)
	}

	t.Cleanup(func() {
		err := client.RemoveStorageDomainDisk(newSD,disk.ID())
		if err != nil {
			t.Fatalf("Failed to remove disk after disk creation test (%v)", err)
		}
	})



	//checkDiskAfterCreation(disk, t, diskName)

}
