package ovirtclient

import (
	"io"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

//go:generate go run scripts/rest.go -i "Disk" -n "disk"

// DiskClient is the client interface part that deals with disks.
type DiskClient interface {
	// StartImageUpload uploads an image file into a disk. The actual upload takes place in the
	// background and can be tracked using the returned UploadImageProgress object.
	//
	// Parameters are as follows:
	//
	// - alias: this is the name used for the uploaded image.
	// - storageDomainID: this is the UUID of the storage domain that the image should be uploaded to.
	// - sparse: use sparse provisioning
	// - size: this is the file size of the image. This must match the bytes read.
	// - reader: this is the source of the image data.
	// - retries: a set of optional retry options.
	//
	// You can wait for the upload to complete using the Done() method:
	//
	//     progress, err := cli.StartImageUpload(...)
	//     if err != nil {
	//         //...
	//     }
	//     <-progress.Done()
	//
	// After the upload is complete you can check the Err() method if it completed successfully:
	//
	//     if err := progress.Err(); err != nil {
	//         //...
	//     }
	//
	StartImageUpload(
		alias string,
		storageDomainID string,
		sparse bool,
		size uint64,
		reader io.Reader,
		retries ...RetryStrategy,
	) (UploadImageProgress, error)

	// UploadImage is identical to StartImageUpload, but waits until the upload is complete. It returns the disk ID
	// as a result, or the error if one happened.
	UploadImage(
		alias string,
		storageDomainID string,
		sparse bool,
		size uint64,
		reader io.Reader,
		retry ...RetryStrategy,
	) (UploadImageResult, error)

	// ListDisks lists all disks.
	ListDisks(retries ...RetryStrategy) ([]Disk, error)
	// GetDisk fetches a disk with a specific ID from the oVirt Engine.
	GetDisk(diskID string, retries ...RetryStrategy) (Disk, error)
	// RemoveDisk removes a disk with a specific ID.
	RemoveDisk(diskID string, retries ...RetryStrategy) error
}

// UploadImageResult represents the completed image upload.
type UploadImageResult interface {
	// Disk returns the disk that has been created as the result of the image upload.
	Disk() Disk
	// CorrelationID returns the opaque correlation ID for the upload.
	CorrelationID() string
}

// Disk is a disk in oVirt.
type Disk interface {
	// ID is the unique ID for this disk.
	ID() string
	// Alias is the name for this disk set by the user.
	Alias() string
	// ProvisionedSize is the size visible to the virtual machine.
	ProvisionedSize() uint64
	// Format is the format of the image.
	Format() ImageFormat
	// StorageDomainID is the ID of the storage system used for this disk.
	StorageDomainID() string
	// Status returns the status the disk is in.
	Status() DiskStatus
}

// DiskStatus shows the status of a disk. Certain operations lock a disk, which is important because the disk can then
// not be changed.
type DiskStatus string

const (
	DiskStatusOK      DiskStatus = "ok"
	DiskStatusLocked  DiskStatus = "locked"
	DiskStatusIllegal DiskStatus = "illegal"
)

// UploadImageProgress is a tracker for the upload progress happening in the background.
type UploadImageProgress interface {
	// Disk returns the disk created as part of the upload process once the upload is complete. Before the upload
	// is complete it will return nil.
	Disk() Disk
	// CorrelationID returns the correlation ID for the upload.
	CorrelationID() string
	// UploadedBytes returns the number of bytes already uploaded.
	UploadedBytes() uint64
	// TotalBytes returns the total number of bytes to be uploaded.
	TotalBytes() uint64
	// Err returns the error of the upload once the upload is complete or errored.
	Err() error
	// Done returns a channel that will be closed when the upload is complete.
	Done() <-chan struct{}
}

// ImageFormat is a constant for representing the format that images can be in.
type ImageFormat string

const (
	ImageFormatCow ImageFormat = "cow"
	ImageFormatRaw ImageFormat = "raw"
)

func convertSDKDisk(sdkDisk *ovirtsdk4.Disk) (Disk, error) {
	id, ok := sdkDisk.Id()
	if !ok {
		return nil, newError(EFieldMissing, "disk does not contain an ID")
	}
	var storageDomainID string
	if sdkStorageDomain, ok := sdkDisk.StorageDomain(); ok {
		storageDomainID, _ = sdkStorageDomain.Id()
	}
	if storageDomainID == "" {
		if sdkStorageDomains, ok := sdkDisk.StorageDomains(); ok {
			if len(sdkStorageDomains.Slice()) == 1 {
				storageDomainID, _ = sdkStorageDomains.Slice()[0].Id()
			}
		}
	}
	if storageDomainID == "" {
		return nil, newError(EFieldMissing, "failed to find a valid storage domain ID for disk %s", id)
	}
	alias, ok := sdkDisk.Alias()
	if !ok {
		return nil, newError(EFieldMissing, "disk %s does not contain an alias", id)
	}
	provisionedSize, ok := sdkDisk.ProvisionedSize()
	if !ok {
		return nil, newError(EFieldMissing, "disk %s does not contain a provisioned size", id)
	}
	format, ok := sdkDisk.Format()
	if !ok {
		return nil, newError(EFieldMissing, "disk %s has no format field", id)
	}
	status, ok := sdkDisk.Status()
	if !ok {
		return nil, newError(EFieldMissing, "disk %s has no status field", id)
	}
	return &disk{
		id:              id,
		alias:           alias,
		provisionedSize: uint64(provisionedSize),
		format:          ImageFormat(format),
		storageDomainID: storageDomainID,
		status:          DiskStatus(status),
	}, nil
}

type disk struct {
	id              string
	alias           string
	provisionedSize uint64
	format          ImageFormat
	storageDomainID string
	status          DiskStatus
}

func (d disk) Status() DiskStatus {
	return d.status
}

func (d disk) ID() string {
	return d.id
}

func (d disk) Alias() string {
	return d.alias
}

func (d disk) ProvisionedSize() uint64 {
	return d.provisionedSize
}

func (d disk) Format() ImageFormat {
	return d.format
}

func (d disk) StorageDomainID() string {
	return d.storageDomainID
}
