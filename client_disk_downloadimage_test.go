package ovirtclient_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestImageDownload(t *testing.T) {
	testImageFile := "./testimage/image"
	testImageData, err := ioutil.ReadFile(testImageFile)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to read test image file %s (%w)", testImageFile, err))
	}
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

	uploadResult, err := client.UploadImage(
		imageName,
		helper.GetStorageDomainID(),
		true,
		uint64(stat.Size()),
		fh,
	)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to upload image (%w)", err))
	}

	defer func() {
		disk := uploadResult.Disk()
		if disk != nil {
			diskID := uploadResult.Disk().ID()
			if err := client.RemoveDisk(diskID); err != nil {
				t.Fatal(fmt.Errorf("failed to remove disk (%w)", err))
			}
		}
	}()

	data := downloadImage(t, client, uploadResult)

	// Note about this check: this will work only on RAW images. For QCOW images the blocks may
	// be reordered, resulting in different files after download.
	if !bytes.Equal(data[:len(testImageData)], testImageData) {
		t.Fatal(fmt.Errorf("the downloaded image did not match the original upload"))
	}
}

func downloadImage(
	t *testing.T,
	client ovirtclient.Client,
	uploadResult ovirtclient.UploadImageResult,
) []byte {
	imageDownload, err := client.DownloadImage(uploadResult.Disk().ID(), ovirtclient.ImageFormatRaw)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to download image (%w)", err))
	}
	defer func() {
		_ = imageDownload.Close()
	}()

	data, err := ioutil.ReadAll(imageDownload)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to download image (%w)", err))
	}
	return data
}
