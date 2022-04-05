package ovirtclient_test

import (
	"bytes"
	_ "embed"
	"encoding/binary"
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

//go:embed testimage/full.qcow
var fullTestImage []byte

//go:generate go run scripts/get_test_image/get_test_image.go

type qcowHeader struct {
	Magic                 [4]byte
	Version               uint32
	BackingFileOffset     uint64
	BackingFileSize       uint32
	ClusterBits           uint32
	Size                  uint64
	CryptMethod           uint32
	L1Size                uint32
	L1TableOffset         uint64
	RefcountTableOffset   uint64
	RefcountTableClusters uint32
	NBSnapshots           uint32
	SnapshotsOffset       uint64
}

// getFullTestImage downloads a fully functional test image with the QEMU guest image to a temporary directory and
// offers it as a reader.
func getFullTestImage(t *testing.T) (io.ReadSeekCloser, uint64, uint64) {
	if len(fullTestImage) == 0 {
		t.Skipf("Skipping test, full test image is not available. Did you run go generate?")
	}

	header := &qcowHeader{}
	if err := binary.Read(bytes.NewReader(fullTestImage), binary.BigEndian, header); err != nil {
		panic(fmt.Errorf("cannot read QCOW header from full test image (%w)", err))
	}

	return &nopReadCloser{bytes.NewReader(fullTestImage)}, uint64(len(fullTestImage)), header.Size
}

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
