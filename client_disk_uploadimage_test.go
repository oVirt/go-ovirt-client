package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestImageUploadDiskCreated(t *testing.T) {
	fh, stat := getTestImageFile(t)
	defer func() {
		_ = fh.Close()
	}()

	helper := getHelper(t)
	client := helper.GetClient()

	imageName := fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5))

	uploadResult, err := client.UploadToNewDisk(
		helper.GetStorageDomainID(),
		ovirtclient.ImageFormatRaw,
		uint64(stat.Size()),
		ovirtclient.CreateDiskParams().WithSparse(true).WithAlias(imageName),
		fh,
	)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to upload image (%w)", err))
	}
	disk, err := client.GetDisk(uploadResult.Disk().ID())
	if err != nil {
		t.Fatal(fmt.Errorf("failed to fetch disk after image upload (%w)", err))
	}
	if err := client.RemoveDisk(disk.ID()); err != nil {
		t.Fatal(err)
	}
}

func TestImageUploadToExistingDisk(t *testing.T) {
	fh, stat := getTestImageFile(t)
	defer func() {
		_ = fh.Close()
	}()

	helper := getHelper(t)
	client := helper.GetClient()

	imageName := fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5))

	disk, err := client.CreateDisk(
		helper.GetStorageDomainID(),
		ovirtclient.ImageFormatRaw,
		uint64(stat.Size()),
		ovirtclient.CreateDiskParams().WithSparse(true).WithAlias(imageName),
	)
	if disk != nil {
		defer func() {
			_ = disk.Remove()
		}()
	}
	if err != nil {
		t.Fatal(err)
	}

	if err := client.UploadToDisk(
		disk.ID(),
		uint64(stat.Size()),
		fh,
	); err != nil {
		t.Fatal(fmt.Errorf("failed to upload image (%w)", err))
	}
}
