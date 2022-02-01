package ovirtclient_test

import (
	"context"
	"fmt"
	"os"
	"time"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

// This example demonstrates the simplest way to upload an image without special timeout handling. The call still times
// out after a built-in timeout.
func ExampleDiskClient_uploadImage() {
	// Open image file to upload
	fh, err := os.Open("/path/to/test.img")
	if err != nil {
		panic(fmt.Errorf("failed to open image file (%w)", err))
	}
	defer func() {
		if err = fh.Close(); err != nil {
			panic(err)
		}
	}()

	// Get the file size
	stat, err := fh.Stat()
	if err != nil {
		panic(fmt.Errorf("failed to stat image file (%w)", err))
	}

	// Obtain oVirt client. Alternatively, you can call ovirtclient.New() to do this directly.
	helper := ovirtclient.NewTestHelperFromEnv(ovirtclientlog.NewNOOPLogger())
	client := helper.GetClient()

	imageName := fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5))

	// Upload image and wait for result.
	uploadResult, err := client.UploadToNewDisk(
		helper.GetStorageDomainID(),
		ovirtclient.ImageFormatRaw,
		uint64(stat.Size()),
		ovirtclient.CreateDiskParams().MustWithAlias(imageName).MustWithSparse(true),
		fh,
	)
	if err != nil {
		panic(fmt.Errorf("failed to upload image (%w)", err))
	}
	fmt.Printf("Uploaded as disk %s\n", uploadResult.Disk().ID())
}

// This example demonstrates how to upload a VM image into a disk while being able to cancel the process manually.
func ExampleDiskClient_uploadImageWithCancel() {
	// Open image file to upload
	fh, err := os.Open("/path/to/test.img")
	if err != nil {
		panic(fmt.Errorf("failed to open image file (%w)", err))
	}
	defer func() {
		if err = fh.Close(); err != nil {
			panic(err)
		}
	}()

	// Get the file size
	stat, err := fh.Stat()
	if err != nil {
		panic(fmt.Errorf("failed to stat image file (%w)", err))
	}

	// Obtain oVirt client. Alternatively, you can call ovirtclient.New() to do this directly.
	helper := ovirtclient.NewTestHelperFromEnv(ovirtclientlog.NewNOOPLogger())
	client := helper.GetClient()

	imageName := fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5))

	// Set up context so we can cancel the upload
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	uploadResult, err := client.StartUploadToNewDisk(
		helper.GetStorageDomainID(),
		ovirtclient.ImageFormatRaw,
		uint64(stat.Size()),
		ovirtclient.CreateDiskParams().MustWithSparse(true).MustWithAlias(imageName),
		fh,
		ovirtclient.ContextStrategy(ctx),
	)
	if err != nil {
		panic(fmt.Errorf("failed to upload image (%w)", err))
	}

	// Wait for image upload to initialize.
	select {
	case <-time.After(10 * time.Minute):
		// Cancel upload
		cancel()
		// Wait for it to actually finish.
		<-uploadResult.Done()
	case <-uploadResult.Done():
	}

	if err := uploadResult.Err(); err != nil {
		panic(fmt.Errorf("failed to upload image (%w)", err))
	}
	fmt.Printf("Uploaded as disk %s\n", uploadResult.Disk().ID())
}
