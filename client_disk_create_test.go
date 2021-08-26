package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

// ExampleDiskClient_CreateDisk is an example of creating an empty disk. This example works with the test
// helper, but can be modified in production.
func ExampleDiskClient_CreateDisk() {
	// Create a logger. This can be adapter to use your own logger.
	logger := ovirtclientlog.NewNOOPLogger()
	// Create the test helper. This will give us our test storage domain.
	helper := ovirtclient.NewTestHelperFromEnv(logger)
	// Create the client. Replace with ovirtclient.New() for production use.
	client := helper.GetClient()

	// Obtain the storage domain used for testing.
	storageDomainID := helper.GetStorageDomainID()
	// Let's create a raw disk.
	imageFormat := ovirtclient.ImageFormatRaw
	// 512 bytes are enough.
	diskSize := 512

	// Create the disk and wait for it to become available. Use StartCreateDisk to skip the wait.
	disk, err := client.CreateDisk(
		storageDomainID,
		imageFormat,
		uint64(diskSize),
		ovirtclient.CreateDiskParams().WithAlias("test_disk"),
	)
	if err != nil {
		panic(err)
	}

	// Remove the disk we just created.
	if err := disk.Remove(); err != nil {
		panic(err)
	}
	// Output:
}

func TestDiskCreation(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	diskName := fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5))

	disk, err := client.CreateDisk(
		helper.GetStorageDomainID(),
		ovirtclient.ImageFormatRaw,
		512,
		ovirtclient.CreateDiskParams().WithAlias(diskName),
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
	defer func() {
		err := disk.Remove()
		if err != nil {
			t.Fatalf("Failed to remove disk after disk creation test (%v)", err)
		}
	}()

	checkDiskAfterCreation(disk, t, diskName)

	fetchedDisk, err := client.GetDisk(disk.ID())
	if err != nil {
		t.Fatalf("failed to fetch disk after creation (%v)", err)
	}

	checkDiskAfterCreation(fetchedDisk, t, diskName)
}

func checkDiskAfterCreation(disk ovirtclient.Disk, t *testing.T, name string) {
	if disk.ProvisionedSize() < 512 {
		t.Fatalf("Incorrect provisioned disk size after creation: %d", disk.ProvisionedSize())
	}
	if disk.TotalSize() < 512 {
		t.Fatalf("Incorrect total disk size after creation: %d", disk.TotalSize())
	}
	if disk.Status() != ovirtclient.DiskStatusOK {
		t.Fatalf(
			"Disk is not in %s status after creation, instead it is %s",
			ovirtclient.DiskStatusOK,
			disk.Status(),
		)
	}
	if disk.Alias() != name {
		t.Fatalf("Incorrect disk alias after creation (%s instead of %s)", disk.Alias(), name)
	}
}
