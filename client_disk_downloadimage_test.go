package ovirtclient_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestImageDownload(t *testing.T) {
	t.Parallel()
	testImageData := getTestImageData()
	fh, size := getTestImageFile()

	helper := getHelper(t)
	client := helper.GetClient()

	imageName := fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5))

	uploadResult, err := client.UploadToNewDisk(
		helper.GetStorageDomainID(),
		ovirtclient.ImageFormatRaw,
		size,
		ovirtclient.CreateDiskParams().MustWithSparse(true).MustWithAlias(imageName),
		fh,
	)
	t.Cleanup(func() {
		disk := uploadResult.Disk()
		if disk != nil {
			diskID := uploadResult.Disk().ID()
			if err := client.RemoveDisk(diskID); err != nil {
				t.Fatal(fmt.Errorf("failed to remove disk (%w)", err))
			}
		}
	})
	if err != nil {
		t.Fatal(fmt.Errorf("failed to upload image (%w)", err))
	}

	data := downloadImage(t, client, uploadResult)

	// Note about this check: this will work only on RAW images. For QCOW images the blocks may
	// be reordered, resulting in different files after download.
	if !bytes.Equal(data[:len(testImageData)], testImageData) {
		t.Fatal(fmt.Errorf("the downloaded image did not match the original upload"))
	}
}

//go:embed testimage/image
var testImage []byte

func getTestImageFile() (io.ReadSeekCloser, uint64) {
	return &nopReadCloser{bytes.NewReader(testImage)}, uint64(len(testImage))
}

type nopReadCloser struct {
	io.ReadSeeker
}

func (n nopReadCloser) Close() error {
	return nil
}

func getTestImageData() []byte {
	return testImage
}

func downloadImage(
	t *testing.T,
	client ovirtclient.Client,
	uploadResult ovirtclient.UploadImageResult,
) []byte {
	imageDownload, err := client.DownloadDisk(uploadResult.Disk().ID(), ovirtclient.ImageFormatRaw)
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
