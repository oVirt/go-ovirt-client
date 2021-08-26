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
	// Deprecated: Use StartUploadToNewDisk instead.
	StartImageUpload(
		alias string,
		storageDomainID string,
		sparse bool,
		size uint64,
		reader readSeekCloser,
		retries ...RetryStrategy,
	) (UploadImageProgress, error)

	// StartUploadToNewDisk uploads an image file into a disk. The actual upload takes place in the
	// background and can be tracked using the returned UploadImageProgress object. If the process fails a removal
	// of the created disk is attempted.
	//
	// Parameters are as follows:
	//
	// - storageDomainID: this is the UUID of the storage domain that the image should be uploaded to.
	// - format: format of the created disk. This does not necessarily have to be identical to the format of the image
	//   being uploaded as the oVirt engine converts images on upload.
	// - size: file size of the uploaded image on the disk.
	// - reader: this is the source of the image data. It is a reader that must support seek and close operations.
	// - retries: a set of optional retry options.
	//
	// You can wait for the upload to complete using the Done() method:
	//
	//     progress, err := cli.StartUploadToNewDisk(...)
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
	StartUploadToNewDisk(
		storageDomainID string,
		format ImageFormat,
		size uint64,
		params CreateDiskOptionalParameters,
		reader readSeekCloser,
		retries ...RetryStrategy,
	) (UploadImageProgress, error)

	// UploadImage is identical to StartImageUpload, but waits until the upload is complete. It returns the disk ID
	// as a result, or the error if one happened.
	//
	// Deprecated: Use UploadToNewDisk instead.
	UploadImage(
		alias string,
		storageDomainID string,
		sparse bool,
		size uint64,
		reader readSeekCloser,
		retries ...RetryStrategy,
	) (UploadImageResult, error)

	// UploadToNewDisk is identical to StartUploadToNewDisk, but waits until the upload is complete. It
	// returns the disk ID as a result, or the error if one happened.
	UploadToNewDisk(
		storageDomainID string,
		format ImageFormat,
		size uint64,
		params CreateDiskOptionalParameters,
		reader readSeekCloser,
		retries ...RetryStrategy,
	) (UploadImageResult, error)

	// StartUploadToDisk uploads a disk image to an existing disk. The actual upload takes place in the background
	// and can be tracked using the returned UploadImageProgress object. Parameters are as follows:
	//
	// - diskID: ID of the disk to upload to.
	// - reader This is the source of the image data.
	// - retries: A set of optional retry options.
	//
	// You can wait for the upload to complete using the Done() method:
	//
	//     progress, err := cli.StartUploadToDisk(...)
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
	StartUploadToDisk(
		diskID string,
		size uint64,
		reader readSeekCloser,
		retries ...RetryStrategy,
	) (UploadImageProgress, error)

	// UploadToDisk runs StartUploadDisk and then waits for the upload to complete. It returns an error if the upload
	// failed despite retries.
	//
	// Parameters are as follows:
	//
	// - diskID: ID of the disk to upload to.
	// - size: size of the file on disk.
	// - reader: this is the source of the image data. The format is automatically determined from the file being
	//   uploaded. The reader must support seeking and close.
	// - retries: a set of optional retry options.
	UploadToDisk(
		diskID string,
		size uint64,
		reader readSeekCloser,
		retries ...RetryStrategy,
	) error

	// StartImageDownload starts the download of the image file of a specific disk.
	// The caller can then wait for the initialization using the Initialized() call:
	//
	//     <-download.Initialized()
	//
	// Alternatively, the downloader can use the Read() function to wait for the download to become available
	// and then read immediately.
	//
	// The caller MUST close the returned reader, otherwise the disk will remain locked in the oVirt engine.
	//
	// Deprecated: please use StartDownloadDisk instead.
	StartImageDownload(
		diskID string,
		format ImageFormat,
		retries ...RetryStrategy,
	) (ImageDownload, error)

	// StartDownloadDisk starts the download of the image file of a specific disk.
	// The caller can then wait for the initialization using the Initialized() call:
	//
	//     <-download.Initialized()
	//
	// Alternatively, the downloader can use the Read() function to wait for the download to become available
	// and then read immediately.
	//
	// The caller MUST close the returned reader, otherwise the disk will remain locked in the oVirt engine.
	StartDownloadDisk(
		diskID string,
		format ImageFormat,
		retries ...RetryStrategy,
	) (ImageDownload, error)

	// DownloadImage runs StartDownloadDisk, then waits for the download to be ready before returning the reader.
	// The caller MUST close the ImageDownloadReader in order to properly unlock the disk in the oVirt engine.
	//
	// Deprecated: please use DownloadDisk instead.
	DownloadImage(
		diskID string,
		format ImageFormat,
		retries ...RetryStrategy,
	) (ImageDownloadReader, error)

	// DownloadDisk runs StartDownloadDisk, then waits for the download to be ready before returning the reader.
	// The caller MUST close the ImageDownloadReader in order to properly unlock the disk in the oVirt engine.
	DownloadDisk(
		diskID string,
		format ImageFormat,
		retries ...RetryStrategy,
	) (ImageDownloadReader, error)

	// StartCreateDisk starts creating an empty disk with the specified parameters and returns a DiskCreation object,
	// which can be queried for completion. Optional parameters can be created using CreateDiskParams().
	StartCreateDisk(
		storageDomainID string,
		format ImageFormat,
		size uint64,
		params CreateDiskOptionalParameters,
		retries ...RetryStrategy,
	) (DiskCreation, error)

	// CreateDisk is a shorthand for calling StartCreateDisk, and then waiting for the disk creation to complete.
	// Optional parameters can be created using CreateDiskParams().
	//
	// Caution! The CreateDisk method may return both an error and a disk that has been created, but has not reached
	// the ready state. Since the disk is almost certainly in a locked state, this may mean that there is a disk left
	// behind.
	CreateDisk(
		storageDomainID string,
		format ImageFormat,
		size uint64,
		params CreateDiskOptionalParameters,
		retries ...RetryStrategy,
	) (Disk, error)

	// ListDisks lists all disks.
	ListDisks(retries ...RetryStrategy) ([]Disk, error)
	// GetDisk fetches a disk with a specific ID from the oVirt Engine.
	GetDisk(diskID string, retries ...RetryStrategy) (Disk, error)
	// RemoveDisk removes a disk with a specific ID.
	RemoveDisk(diskID string, retries ...RetryStrategy) error
}

// CreateDiskOptionalParameters is a structure that serves to hold the optional parameters for DiskClient.CreateDisk.
type CreateDiskOptionalParameters interface {
	// Alias is a secondary name for the disk.
	Alias() string
	// Sparse returns true if sparse provisioning should be used for disks.
	Sparse() *bool
}

// BuildableCreateDiskParameters is a buildable version of CreateDiskOptionalParameters.
type BuildableCreateDiskParameters interface {
	CreateDiskOptionalParameters

	// WithAlias sets the alias of the disk.
	WithAlias(alias string) BuildableCreateDiskParameters
	// WithSparse sets the sparse flag on the disk.
	WithSparse(sparse bool) BuildableCreateDiskParameters
}

// CreateDiskParams creates a buildable set of CreateDiskOptionalParameters for use with
// Client.CreateDisk.
func CreateDiskParams() BuildableCreateDiskParameters {
	return &createDiskParams{}
}

type createDiskParams struct {
	alias  string
	sparse *bool
}

func (c *createDiskParams) Alias() string {
	return c.alias
}

func (c *createDiskParams) WithAlias(alias string) BuildableCreateDiskParameters {
	c.alias = alias
	return c
}

func (c *createDiskParams) Sparse() *bool {
	return c.sparse
}

func (c *createDiskParams) WithSparse(sparse bool) BuildableCreateDiskParameters {
	c.sparse = &sparse
	return c
}

// DiskCreation is a process object that lets you query the status of the disk creation.
type DiskCreation interface {
	// Disk returns the disk that has been created, even if it is not yet ready.
	Disk() Disk
	// Wait waits until the disk creation is complete and returns when it is done. It returns the created disk and
	// an error if one happened.
	Wait(retries ...RetryStrategy) (Disk, error)
}

// ImageDownloadReader is a special reader for reading image downloads. On the first Read call
// it waits until the image download is ready and then returns the desired bytes. It also
// tracks how many bytes are read for an async display of a progress bar.
type ImageDownloadReader interface {
	io.Reader
	// Read reads the specified bytes from the disk image. This call will block if
	// the image is not yet ready for download.
	Read(p []byte) (n int, err error)
	// Close closes the image download and unlocks the disk.
	Close() error
	// BytesRead returns the number of bytes read so far. This can be used to
	// provide a progress bar.
	BytesRead() uint64
	// Size returns the size of the disk image in bytes. This is ONLY available after the initialization is complete and
	// MAY return 0 before.
	Size() uint64
}

// ImageDownload represents an image download in progress. The caller MUST
// close the image download when it is finished otherwise the disk will not be unlocked.
type ImageDownload interface {
	ImageDownloadReader

	// Err returns the error that happened during initializing the download, or the last error reading from the
	// image server.
	Err() error
	// Initialized returns a channel that will be closed when the initialization is complete. This can be either
	// in an errored state (check Err()) or when the image is ready.
	Initialized() <-chan struct{}
}

// UploadImageResult represents the completed image upload.
type UploadImageResult interface {
	// Disk returns the disk that has been created as the result of the image upload.
	Disk() Disk
}

// Disk is a disk in oVirt.
type Disk interface {
	// ID is the unique ID for this disk.
	ID() string
	// Alias is the name for this disk set by the user.
	Alias() string
	// ProvisionedSize is the size visible to the virtual machine.
	ProvisionedSize() uint64
	// TotalSize is the size of the image file.
	// This value can be zero in some cases, for example when the disk upload wasn't properly finalized.
	TotalSize() uint64
	// Format is the format of the image.
	Format() ImageFormat
	// StorageDomainID is the ID of the storage system used for this disk.
	StorageDomainID() string
	// Status returns the status the disk is in.
	Status() DiskStatus

	// StartDownload starts the download of the image file the current disk.
	// The caller can then wait for the initialization using the Initialized() call:
	//
	//     <-download.Initialized()
	//
	// Alternatively, the downloader can use the Read() function to wait for the download to become available
	// and then read immediately.
	//
	// The caller MUST close the returned reader, otherwise the disk will remain locked in the oVirt engine.
	// The passed context is observed only for the initialization phase.
	StartDownload(
		format ImageFormat,
		retries ...RetryStrategy,
	) (ImageDownload, error)

	// Download runs StartDownload, then waits for the download to be ready before returning the reader.
	// The caller MUST close the ImageDownloadReader in order to properly unlock the disk in the oVirt engine.
	Download(
		format ImageFormat,
		retries ...RetryStrategy,
	) (ImageDownloadReader, error)

	// Remove removes the current disk in the oVirt engine.
	Remove(retries ...RetryStrategy) error
}

// DiskStatus shows the status of a disk. Certain operations lock a disk, which is important because the disk can then
// not be changed.
type DiskStatus string

const (
	// DiskStatusOK represents a disk status that operations can be performed on.
	DiskStatusOK DiskStatus = "ok"
	// DiskStatusLocked represents a disk status where no operations can be performed on the disk.
	DiskStatusLocked DiskStatus = "locked"
	// DiskStatusIllegal indicates that the disk cannot be accessed by the virtual machine, and the user needs
	// to take action to resolve the issue.
	DiskStatusIllegal DiskStatus = "illegal"
)

// UploadImageProgress is a tracker for the upload progress happening in the background.
type UploadImageProgress interface {
	// Disk returns the disk created as part of the upload process once the upload is complete. Before the upload
	// is complete it will return nil.
	Disk() Disk
	// UploadedBytes returns the number of bytes already uploaded.
	//
	// Caution! This number may decrease or reset to 0 if the upload has to be retried.
	UploadedBytes() uint64
	// TotalBytes returns the total number of bytes to be uploaded.
	TotalBytes() uint64
	// Err returns the error of the upload once the upload is complete or errored.
	Err() error
	// Done returns a channel that will be closed when the upload is complete.
	Done() <-chan struct{}
}

// ImageFormat is a constant for representing the format that images can be in. This is relevant
// for both image uploads and image downloads, as the oVirt engine has the capability of converting
// between these formats.
//
// Note: the mocking facility cannot convert between the formats due to the complexity of the QCOW2
// format. It is recommended to write tests only using the raw format as comparing QCOW2 files
// is complex.
type ImageFormat string

// Validate returns an error if the image format doesn't have a valid value.
func (f ImageFormat) Validate() error {
	switch f {
	case ImageFormatRaw:
		return nil
	case ImageFormatCow:
		return nil
	default:
		return newError(
			EBadArgument,
			"invalid image format: %s must be one of: %s, %s",
			f,
			ImageFormatRaw,
			ImageFormatCow,
		)
	}
}

const (
	// ImageFormatCow is an image conforming to the QCOW2 image format. This image format can use
	// compression, supports snapshots, and other features.
	// See https://github.com/qemu/qemu/blob/master/docs/interop/qcow2.txt for details.
	ImageFormatCow ImageFormat = "cow"
	// ImageFormatRaw is not actually a format, it only contains the raw bytes on the block device.
	ImageFormatRaw ImageFormat = "raw"
)

func convertSDKDisk(sdkDisk *ovirtsdk4.Disk, client Client) (Disk, error) {
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
	totalSize, ok := sdkDisk.TotalSize()
	if !ok {
		return nil, newError(EFieldMissing, "disk %s does not contain a total size", id)
	}
	format, ok := sdkDisk.Format()
	if !ok {
		return nil, newError(EFieldMissing, "disk %s has no format field", id)
	}
	status, ok := sdkDisk.Status()
	if !ok {
		return nil, newError(EFieldMissing, "disk %s has no status field", id)
	}
	sparse, ok := sdkDisk.Sparse()
	if !ok {
		return nil, newError(EFieldMissing, "disk %s has no sparse field", id)
	}
	return &disk{
		client: client,

		id:              id,
		alias:           alias,
		provisionedSize: uint64(provisionedSize),
		totalSize:       uint64(totalSize),
		format:          ImageFormat(format),
		storageDomainID: storageDomainID,
		status:          DiskStatus(status),
		sparse:          sparse,
	}, nil
}

type disk struct {
	client Client

	id              string
	alias           string
	provisionedSize uint64
	format          ImageFormat
	storageDomainID string
	status          DiskStatus
	totalSize       uint64
	sparse          bool
}

func (d disk) Remove(retries ...RetryStrategy) error {
	return d.client.RemoveDisk(d.id, retries...)
}

func (d disk) TotalSize() uint64 {
	return d.totalSize
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

func (d disk) StartDownload(format ImageFormat, retries ...RetryStrategy) (ImageDownload, error) {
	return d.client.StartDownloadDisk(d.id, format, retries...)
}

func (d disk) Download(format ImageFormat, retries ...RetryStrategy) (ImageDownloadReader, error) {
	return d.client.DownloadDisk(d.id, format, retries...)
}
