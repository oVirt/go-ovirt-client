//
// This file implements the image upload-related functions of the oVirt client.
//

package ovirtclient

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func (o *oVirtClient) UploadImage(
	ctx context.Context,
	alias string,
	storageDomainID string,
	sparse bool,
	size uint64,
	reader io.Reader,
) (UploadImageResult, error) {
	progress, err := o.StartImageUpload(ctx, alias, storageDomainID, sparse, size, reader)
	if err != nil {
		return nil, err
	}
	<-progress.Done()
	if err := progress.Err(); err != nil {
		return nil, err
	}
	return progress, nil
}

func (o *oVirtClient) StartImageUpload(
	ctx context.Context,
	alias string,
	storageDomainID string,
	sparse bool,
	size uint64,
	reader io.Reader,
) (UploadImageProgress, error) {
	bufReader := bufio.NewReaderSize(reader, qcowHeaderSize)

	format, qcowSize, err := extractQCOWParameters(size, bufReader)
	if err != nil {
		return nil, err
	}

	newCtx, cancel := context.WithCancel(ctx) //nolint:govet

	disk, err := o.createDiskForUpload(storageDomainID, alias, format, qcowSize, sparse, cancel)
	if err != nil {
		return nil, wrap(
			err,
			EUnidentified,
			"failed to create disk for image upload",
		)
	}

	return o.createProgress(alias, qcowSize, size, bufReader, storageDomainID, sparse, newCtx, cancel, disk)
}

func (o *oVirtClient) createDiskForUpload(
	storageDomainID string,
	alias string,
	format ImageFormat,
	qcowSize uint64,
	sparse bool,
	cancel context.CancelFunc,
) (*ovirtsdk4.Disk, error) {
	storageDomain, err := ovirtsdk4.NewStorageDomainBuilder().Id(storageDomainID).Build()
	if err != nil {
		return nil, wrap(
			err,
			EBug,
			"failed to build storage domain object from storage domain ID: %s",
			storageDomainID,
		)
	}
	diskBuilder := ovirtsdk4.NewDiskBuilder().
		Alias(alias).
		Format(ovirtsdk4.DiskFormat(format)).
		ProvisionedSize(int64(qcowSize)).
		InitialSize(int64(qcowSize)).
		StorageDomainsOfAny(storageDomain)
	diskBuilder.Sparse(sparse)
	disk, err := diskBuilder.Build()
	if err != nil {
		cancel()
		return nil, wrap(
			err,
			EBug,
			"failed to build disk with alias %s, format %s, provisioned and initial size %d",
			alias,
			format,
			qcowSize,
		)
	}
	return disk, nil
}

func (o *oVirtClient) createProgress(
	alias string,
	qcowSize uint64,
	size uint64,
	bufReader *bufio.Reader,
	storageDomainID string,
	sparse bool,
	newCtx context.Context,
	cancel context.CancelFunc,
	disk *ovirtsdk4.Disk,
) (UploadImageProgress, error) {
	progress := &uploadImageProgress{
		cli:             o,
		correlationID:   fmt.Sprintf("image_transfer_%s", alias),
		uploadedBytes:   0,
		cowSize:         qcowSize,
		size:            size,
		reader:          bufReader,
		storageDomainID: storageDomainID,
		sparse:          sparse,
		alias:           alias,
		ctx:             newCtx,
		done:            make(chan struct{}),
		lock:            &sync.Mutex{},
		cancel:          cancel,
		err:             nil,
		conn:            o.conn,
		httpClient:      o.httpClient,
		disk:            disk,
		client:          o,
	}
	go progress.upload()
	return progress, nil
}

type uploadImageProgress struct {
	cli             *oVirtClient
	uploadedBytes   uint64
	cowSize         uint64
	size            uint64
	reader          *bufio.Reader
	storageDomainID string
	sparse          bool
	alias           string

	// ctx is the context that indicates that the upload should terminate as soon as possible. The actual upload may run
	// longer in order to facilitate proper cleanup.
	ctx context.Context
	// done is the channel that is closed when the upload is completely done, either with an error, or successfully.
	done chan struct{}
	// lock is a lock that prevents race conditions during the upload process.
	lock *sync.Mutex
	// cancel is the cancel function for the context. HasCode is called to ensure that the context is properly canceled.
	cancel context.CancelFunc
	// err holds the error that happened during the upload. It can be queried using the Err() method.
	err error
	// conn is the underlying oVirt connection.
	conn *ovirtsdk4.Connection
	// httpClient is the raw HTTP client for connecting the oVirt Engine.
	httpClient http.Client
	// disk is the oVirt disk that will be provisioned during the upload.
	disk *ovirtsdk4.Disk
	// client is the Client instance that created this image upload.
	client *oVirtClient
	// correlationID is an identifier for the upload process.
	correlationID string
}

func (u *uploadImageProgress) CorrelationID() string {
	return u.correlationID
}

func (u *uploadImageProgress) Disk() Disk {
	sdkDisk := u.disk
	id, ok := sdkDisk.Id()
	if !ok || id == "" {
		return nil
	}
	disk, err := convertSDKDisk(sdkDisk)
	if err != nil {
		panic(wrap(err, EBug, "bug: failed to convert disk"))
	}
	return disk
}

func (u *uploadImageProgress) UploadedBytes() uint64 {
	return u.uploadedBytes
}

func (u *uploadImageProgress) TotalBytes() uint64 {
	return u.size
}

func (u *uploadImageProgress) Err() error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.err != nil {
		return u.err
	}
	return nil
}

func (u *uploadImageProgress) Done() <-chan struct{} {
	return u.done
}

func (u *uploadImageProgress) Read(p []byte) (n int, err error) {
	select {
	case <-u.ctx.Done():
		return 0, newError(ETimeout, "timeout while uploading image")
	default:
	}
	n, err = u.reader.Read(p)
	u.uploadedBytes += uint64(n)
	return
}

// upload uploads the image file in the background. It is intended to be called as a goroutine. The error status can
// be obtained from Err(), while the done status can be queried using Done().
func (u *uploadImageProgress) upload() {
	defer func() {
		// Cancel context to indicate done.
		u.lock.Lock()
		u.cancel()
		close(u.done)
		u.lock.Unlock()
	}()

	if err := u.processUpload(); err != nil {
		u.err = err
	}
}

func (u *uploadImageProgress) processUpload() error {
	diskID, diskService, err := u.createDisk()
	if err != nil {
		return err
	}

	if err := u.waitForDiskOk(diskService); err != nil {
		u.removeDisk()
		return err
	}

	transfer, transferService, err := u.setupImageTransfer(diskID)
	if err != nil {
		u.removeDisk()
		return err
	}

	transferURL, err := u.findTransferURL(transfer)
	if err != nil {
		u.removeDisk()
		return err
	}

	if err := u.uploadImage(transferURL); err != nil {
		u.removeDisk()
		return err
	}

	if err := u.finalizeUpload(transferService); err != nil {
		u.removeDisk()
		return err
	}

	if err := u.waitForDiskOk(diskService); err != nil {
		u.removeDisk()
		return err
	}

	return nil
}

func (u *uploadImageProgress) removeDisk() {
	disk := u.disk
	if disk != nil {
		if id, ok := u.disk.Id(); ok {
			_ = u.client.RemoveDisk(u.ctx, id)
		}
	}
}

func (u *uploadImageProgress) finalizeUpload(
	transferService *ovirtsdk4.ImageTransferService,
) error {
	finalizeRequest := transferService.Finalize()
	finalizeRequest.Query("correlation_id", u.correlationID)
	_, err := finalizeRequest.Send()
	if err != nil {
		return wrap(err, EUnidentified, "failed to finalize image upload")
	}
	return nil
}

func (u *uploadImageProgress) uploadImage(transferURL *url.URL) error {
	putRequest, err := http.NewRequest(http.MethodPut, transferURL.String(), u)
	if err != nil {
		return wrap(err, EUnidentified, "failed to create HTTP request")
	}
	putRequest.Header.Add("content-type", "application/octet-stream")
	putRequest.ContentLength = int64(u.size)
	response, err := u.httpClient.Do(putRequest)
	if err != nil {
		return wrap(err, EUnidentified, "failed to upload image")
	}
	if err := response.Body.Close(); err != nil {
		return wrap(err, EUnidentified, "failed to close response body while uploading image")
	}
	return nil
}

func (u *uploadImageProgress) findTransferURL(transfer *ovirtsdk4.ImageTransfer) (*url.URL, error) {
	var tryURLs []string
	if transferURL, ok := transfer.TransferUrl(); ok && transferURL != "" {
		tryURLs = append(tryURLs, transferURL)
	}
	if proxyURL, ok := transfer.ProxyUrl(); ok && proxyURL != "" {
		tryURLs = append(tryURLs, proxyURL)
	}

	if len(tryURLs) == 0 {
		return nil, newError(EBug, "neither a transfer URL nor a proxy URL was returned from the oVirt Engine")
	}

	var foundTransferURL *url.URL
	var lastError error
	for _, transferURL := range tryURLs {
		transferURL, err := url.Parse(transferURL)
		if err != nil {
			lastError = wrap(err, EUnidentified, "failed to parse transfer URL %s", transferURL)
			continue
		}

		hostUrl, err := url.Parse(transfer.MustTransferUrl())
		if err == nil {
			optionsReq, err := http.NewRequest(http.MethodOptions, hostUrl.String(), strings.NewReader(""))
			if err != nil {
				lastError = err
				continue
			}
			res, err := u.httpClient.Do(optionsReq)
			if err == nil {
				statusCode := res.StatusCode
				if err := res.Body.Close(); err != nil {
					lastError = wrap(err, EUnidentified, "failed to close response body in options request")
				} else {
					if statusCode == 200 {
						foundTransferURL = transferURL
						lastError = nil
						break
					} else {
						lastError = newError(EConnection, "non-200 status code returned from URL %s (%d)", hostUrl, res.StatusCode)
					}
				}
			} else {
				lastError = err
			}
		} else {
			lastError = err
		}
	}
	if foundTransferURL == nil {
		return nil, wrap(lastError, EUnidentified, "failed to find transfer URL")
	}
	return foundTransferURL, nil
}

func (u *uploadImageProgress) createDisk() (string, *ovirtsdk4.DiskService, error) {
	addDiskRequest := u.conn.SystemService().DisksService().Add().Disk(u.disk)
	addDiskRequest.Query("correlation_id", u.correlationID)
	addResp, err := addDiskRequest.Send()
	if err != nil {
		diskAlias, _ := u.disk.Alias()
		return "", nil, wrap(err, EUnidentified, "failed to create disk, alias: %s", diskAlias)
	}
	diskID := addResp.MustDisk().MustId()
	diskService := u.conn.SystemService().DisksService().DiskService(diskID)
	return diskID, diskService, nil
}

func (u *uploadImageProgress) setupImageTransfer(diskID string) (
	*ovirtsdk4.ImageTransfer,
	*ovirtsdk4.ImageTransferService,
	error,
) {
	var lastError EngineError
	imageTransfersService := u.conn.SystemService().ImageTransfersService()
	image := ovirtsdk4.NewImageBuilder().Id(diskID).MustBuild()
	transfer := ovirtsdk4.
		NewImageTransferBuilder().
		Image(image).
		MustBuild()
	transferReq := imageTransfersService.
		Add().
		ImageTransfer(transfer).
		Query("correlation_id", u.correlationID)
	transferRes, err := transferReq.Send()
	if err != nil {
		return nil, nil, wrap(err, EUnidentified, "failed to start image transfer")
	}
	transfer = transferRes.MustImageTransfer()
	transferService := imageTransfersService.ImageTransferService(transfer.MustId())

	for {
		req, err := transferService.Get().Send()
		if err == nil {
			if req.MustImageTransfer().MustPhase() == ovirtsdk4.IMAGETRANSFERPHASE_TRANSFERRING {
				break
			} else {
				lastError = newError(
					EPending,
					"image transfer is in phase %s instead of transferring",
					req.MustImageTransfer().MustPhase(),
				)
			}
		} else {
			lastError = wrap(err, EUnidentified, "failed to get image transfer for disk %s", diskID)
			if !lastError.CanAutoRetry() {
				return nil, nil, lastError
			}
		}
		select {
		case <-time.After(time.Second * 5):
		case <-u.ctx.Done():
			return nil, nil, wrap(lastError, ETimeout, "timeout while waiting for image transfer")
		}
	}
	return transfer, transferService, nil
}

func (u *uploadImageProgress) waitForDiskOk(diskService *ovirtsdk4.DiskService) error {
	var lastError error
	for {
		req, err := diskService.Get().Send()
		if err == nil {
			disk, ok := req.Disk()
			if !ok {
				return newError(EUnsupported, "the disk was removed after upload, probably not supported")
			}
			if disk.MustStatus() == ovirtsdk4.DISKSTATUS_OK {
				return u.cli.waitForJobFinished(u.ctx, u.correlationID)
			} else {
				lastError = newError(EPending, "disk status is %s, not ok", disk.MustStatus())
			}
			u.disk = disk
		} else {
			lastError = err
		}
		select {
		case <-time.After(5 * time.Second):
		case <-u.ctx.Done():
			return wrap(lastError, ETimeout, "timeout while waiting for disk to be ok after upload")
		}
	}
}
