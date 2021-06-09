package govirt_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestImageUploadDiskCreated(t *testing.T) {
	testImageFile := "./testimage/image"
	fh, err := os.Open(testImageFile)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to open test image file %s (%w)", testImageFile, err))
	}
	defer func() {
		_ = fh.Close()
	}()

	stat, err := fh.Stat()
	if err != nil {
		t.Fatal(fmt.Errorf("failed to stat test image file %s (%w)", testImageFile, err))
	}

	helper := getHelper(t)
	client := helper.GetClient()

	imageName := fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5))

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Minute)
	defer cancel()

	disk, err := client.UploadImage(
		ctx,
		imageName,
		helper.GetStorageDomainID(),
		true,
		uint64(stat.Size()),
		fh,
	)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to upload image (%w)", err))
	}
	disk, err = client.GetDisk(disk.ID())
	if err != nil {
		t.Fatal(fmt.Errorf("failed to fetch disk after image upload (%w)", err))
	}
	if err := client.RemoveDisk(disk.ID()); err != nil {
		t.Fatal(fmt.Errorf("failed to remove disk (%w)", err))
	}
}
