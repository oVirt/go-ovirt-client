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

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func (o *oVirtClient) UploadImage(
	alias string,
	storageDomainID string,
	sparse bool,
	size uint64,
	reader io.Reader,
	retries ...RetryStrategy,
) (UploadImageResult, error) {
	retries = defaultRetries(retries, defaultLongTimeouts())
	progress, err := o.StartImageUpload(alias, storageDomainID, sparse, size, reader, retries...)
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
	alias string,
	storageDomainID string,
	sparse bool,
	size uint64,
	reader io.Reader,
	retries ...RetryStrategy,
) (UploadImageProgress, error) {
	retries = defaultRetries(retries, defaultLongTimeouts())

	o.logger.Infof("Starting disk image upload...")
	bufReader := bufio.NewReaderSize(reader, qcowHeaderSize)

	format, qcowSize, err := extractQCOWParameters(size, bufReader)
	if err != nil {
		return nil, err
	}

	newCtx, cancel := context.WithCancel(context.Background()) //nolint:govet

	disk, err := o.createDiskForUpload(storageDomainID, alias, format, qcowSize, sparse, cancel)
	if err != nil {
		return nil, wrap(
			err,
			EUnidentified,
			"failed to create disk for image upload",
		)
	}

	return o.createProgress(alias, qcowSize, size, bufReader, storageDomainID, sparse, newCtx, cancel, disk, retries)
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
	retries []RetryStrategy,
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
		logger:          o.logger,
		retries:         retries,
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
	// logger contains the facility to write log messages
	logger Logger

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
	// retries contains the retry configuration
	retries []RetryStrategy
	// diskID contains the ID of the disk after it has been created.
	diskID string
	// diskService is the disk service related to the disk created.
	diskService *ovirtsdk4.DiskService
	// transfer is the set up image transfer
	transfer *ovirtsdk4.ImageTransfer
	// transferService is the transfer service related to the transfer
	transferService *ovirtsdk4.ImageTransferService
	transferURL     *url.URL
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

	steps := []func() error{
		u.createDisk,
		u.waitForDiskOk,
		u.createImageTransfer,
		u.waitForImageTransferReady,
		u.findTransferURL,
		u.uploadImage,
		u.finalizeUpload,
		u.waitForDiskOk,
	}

	for _, step := range steps {
		if err := step(); err != nil {
			u.abortTransfer()
			u.removeDisk()
			return err
		}
	}
	return nil
}

func (u *uploadImageProgress) abortTransfer() {
	if u.transfer != nil {
		if err := retry(
			fmt.Sprintf("canceling transfer for disk %s", u.diskID),
			u.logger,
			u.retries,
			func() error {
				_, err := u.transferService.Cancel().Send()
				return err
			},
		); err != nil {
			u.logger.Warningf("failed to cancel transfer for disk %s, may not be able to remove disk", u.diskID)
		}
	}
}

func (u *uploadImageProgress) removeDisk() {
	if u.diskID != "" {
		if err := u.client.RemoveDisk(u.diskID, u.retries...); err != nil {
			u.logger.Warningf("failed to remove disk %s after failed image upload, please remove manually (%w)", u.diskID, err)
		}
	}
}

func (u *uploadImageProgress) finalizeUpload() error {
	return retry(
		fmt.Sprintf("finalizing image for disk %s", u.Disk()),
		u.logger,
		u.retries,
		func() error {
			finalizeRequest := u.transferService.Finalize()
			finalizeRequest.Query("correlation_id", u.correlationID)
			_, err := finalizeRequest.Send()
			return err
		})
}

func (u *uploadImageProgress) uploadImage() error {
	u.logger.Debugf("Uploading image for disk %s via HTTP request to %s...", u.Disk().ID(), u.transferURL.String())
	putRequest, err := http.NewRequest(http.MethodPut, u.transferURL.String(), u)
	if err != nil {
		return wrap(err, EUnidentified, "failed to create HTTP request")
	}
	putRequest.Header.Add("content-type", "application/octet-stream")
	putRequest.ContentLength = int64(u.size)
	response, err := u.httpClient.Do(putRequest)
	if err != nil {
		u.logger.Debugf("Failed to upload image to disk %s via HTTP request to %s. (%v)", u.Disk().ID(), u.transferURL.String(), err)
		return wrap(err, EUnidentified, "failed to upload image")
	}
	if err := response.Body.Close(); err != nil {
		u.logger.Debugf("Failed to close response body for image transfer to disk %s via HTTP request to %s. (%v)", u.Disk().ID(), u.transferURL.String(), err)
		return wrap(err, EUnidentified, "failed to close response body while uploading image")
	}
	return nil
}

func (u *uploadImageProgress) findTransferURL() (err error) {
	u.logger.Debugf("Attempting to determine image transfer URL for disk %s...", u.Disk().ID())
	var tryURLs []string
	if transferURL, ok := u.transfer.TransferUrl(); ok && transferURL != "" {
		tryURLs = append(tryURLs, transferURL)
	}
	if proxyURL, ok := u.transfer.ProxyUrl(); ok && proxyURL != "" {
		tryURLs = append(tryURLs, proxyURL)
	}

	if len(tryURLs) == 0 {
		u.logger.Errorf("Bug: neither a transfer URL nor a proxy URL was returned from the oVirt Engine. (%v)", u.transfer)
		return newError(EBug, "neither a transfer URL nor a proxy URL was returned from the oVirt Engine")
	}

	var lastError error
	for _, transferURL := range tryURLs {
		lastError = u.verifyTransferURL(transferURL)
	}
	if lastError != nil {
		return wrap(lastError, EConnection, "failed to find a valid transfer URL; check your network connectivity to the oVirt Engine ImageIO port")
	}
	return nil
}

func (u *uploadImageProgress) verifyTransferURL(transferURL string) error {
	parsedTransferURL, err := url.Parse(transferURL)
	if err != nil {
		return wrap(err, EUnidentified, "failed to parse transfer URL %s", transferURL)
	}

	return retry(
		fmt.Sprintf("sending OPTIONS request to %s", transferURL),
		u.logger,
		append(u.retries, MaxTries(3)),
		func() error {
			optionsReq, e := http.NewRequest(http.MethodOptions, parsedTransferURL.String(), strings.NewReader(""))
			if e != nil {
				return wrap(e, EBug, "failed to create OPTIONS request to %s", parsedTransferURL.String())
			}
			res, e := u.httpClient.Do(optionsReq)
			if e != nil {
				return wrap(e, EConnection, "HTTP request to %s failed", parsedTransferURL.String())
			}
			defer func() {
				_ = res.Body.Close()
			}()
			statusCode := res.StatusCode
			if statusCode < 199 {
				return newError(
					EConnection,
					"HTTP connection error while calling %s",
					parsedTransferURL.String(),
				)
			} else if statusCode < 399 {
				u.transferURL = parsedTransferURL
				return nil
			} else if statusCode < 499 {
				return newError(
					EPermanentHTTPError,
					"HTTP 4xx status code returned from URL %s (%d)",
					parsedTransferURL.String(),
					res.StatusCode,
				)
			} else {
				return newError(
					EConnection,
					"non-200 status code returned from URL %s (%d)",
					parsedTransferURL.String(),
					res.StatusCode,
				)
			}
		},
	)
}

func (u *uploadImageProgress) createDisk() (err error) {
	// This will never fail because we set up the disk in the previous setp
	diskAlias := u.disk.MustAlias()
	err = retry(
		fmt.Sprintf("creating disk with alias %s for image upload", diskAlias),
		u.logger,
		u.retries,
		func() error {
			addDiskRequest := u.conn.SystemService().DisksService().Add().Disk(u.disk)
			addDiskRequest.Query("correlation_id", u.correlationID)
			addResp, err := addDiskRequest.Send()
			if err != nil {
				return err
			}
			u.diskID = addResp.MustDisk().MustId()
			u.diskService = u.conn.SystemService().DisksService().DiskService(u.diskID)
			return nil
		})
	return
}

func (u *uploadImageProgress) waitForImageTransferReady() (err error) {
	return retry(
		fmt.Sprintf("waiting for image transfer to become ready for disk ID %s", u.diskID),
		u.logger,
		u.retries,
		func() error {
			req, e := u.transferService.Get().Send()
			if e != nil {
				return e
			}
			transfer, ok := req.ImageTransfer()
			if !ok {
				return newError(EFieldMissing, "fetching image transfer did not return an image transfer")
			}
			phase, ok := transfer.Phase()
			if !ok {
				return newError(EFieldMissing, "fetching image transfer did not contain a phase")
			}
			switch phase {
			case ovirtsdk4.IMAGETRANSFERPHASE_INITIALIZING:
				return newError(
					EPending,
					"image transfer is in phase %s instead of transferring",
					phase,
				)
			case ovirtsdk4.IMAGETRANSFERPHASE_TRANSFERRING:
				return nil
			default:
				return newError(EUnexpectedImageTransferPhase, "image transfer is in phase %s instead of transferring", phase)
			}
		})
}

func (u *uploadImageProgress) createImageTransfer() (err error) {
	u.logger.Debugf("Creating image transfer for image upload disk ID %s...", u.diskID)
	imageTransfersService := u.conn.SystemService().ImageTransfersService()
	image := ovirtsdk4.NewImageBuilder().Id(u.diskID).MustBuild()
	transfer := ovirtsdk4.
		NewImageTransferBuilder().
		Image(image).
		MustBuild()
	transferReq := imageTransfersService.
		Add().
		ImageTransfer(transfer).
		Query("correlation_id", u.correlationID)

	return retry(
		fmt.Sprintf("starting image transfer for disk %s", u.diskID),
		u.logger,
		u.retries,
		func() error {
			transferRes, e := transferReq.Send()
			if e != nil {
				return e
			}
			var ok bool
			u.transfer, ok = transferRes.ImageTransfer()
			if !ok {
				return newError(EFieldMissing, "missing image transfer as a response to image transfer create request")
			}
			transferID, ok := transfer.Id()
			if !ok {
				return newError(EFieldMissing, "missing image transfer ID in response to image transfer create request")
			}
			u.transferService = imageTransfersService.ImageTransferService(transferID)
			return nil
		},
	)
}

func (u *uploadImageProgress) waitForDiskOk() (err error) {
	err = retry(
		fmt.Sprintf("waiting for disk  %s to become OK", u.diskID),
		u.logger,
		u.retries,
		func() error {
			disk, e := u.cli.GetDisk(u.diskID)
			if e != nil {
				return err
			}
			if disk.Status() != DiskStatusOK {
				return newError(EPending, "disk status is %s, not ok", disk.Status())
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	return u.cli.waitForJobFinished(u.correlationID, u.retries)
}
