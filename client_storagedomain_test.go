package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestStorageDomainDiskGet(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	diskName := fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5))

	disk, err := client.CreateDisk(
		helper.GetStorageDomainID(),
		ovirtclient.ImageFormatRaw,
		512,
		ovirtclient.CreateDiskParams().MustWithAlias(diskName),
	)
	if err != nil {
		t.Fatal(err)
	}

	if disk.Status() != ovirtclient.DiskStatusOK {
		t.Fatalf(
			"Disk is not in %s status after creation, instead it is %s",
			ovirtclient.DiskStatusOK,
			disk.Status(),
		)
	}
	t.Cleanup(func() {
		err := disk.Remove()
		if err != nil {
			t.Fatalf("Failed to remove disk after disk creation test (%v)", err)
		}
	})

	checkDiskAfterCreation(disk, t, diskName)

	fetchedDisk, err := client.GetDisk(disk.ID())
	if err != nil {
		t.Fatalf("failed to fetch disk after creation (%v)", err)
	}

	checkDiskAfterCreation(fetchedDisk, t, diskName)
	disk, err = client.GetStorageDomainDisk(helper.GetStorageDomainID(),disk.ID())
	if err != nil {
		t.Fatal(err)
	}

	checkDiskAfterCreation(disk, t, diskName)

}
