package ovirtclient

import "fmt"

// WaitForDiskOk waits for a disk to be in the OK status, then additionally queries the job that was in progress with
// the correlation ID. This is necessary because the disk returns OK status before the job has actually finished,
// resulting in a "disk locked" error on subsequent operations. It uses checkDiskOk as an underlying function.
func (o *oVirtClient) WaitForDiskOk(diskID string, correlationID string, retries []RetryStrategy) (err error) {
	err = retry(
		fmt.Sprintf("waiting for disk %s to become OK", diskID),
		o.logger,
		retries,
		func() error {
			_, err = o.checkDiskOk(diskID)
			return err
		},
	)
	if err != nil {
		return err
	}
	if err := o.waitForJobFinished(correlationID, retries); err != nil {
		return err
	}
	return nil
}

// checkDiskOk fetches the disk for the transfer and checks if it is in the OK status. It returns an EPending error if
// it is not.
func (o *oVirtClient) checkDiskOk(diskID string) (Disk, error) {
	disk, err := o.GetDisk(diskID)
	if err != nil {
		return nil, err
	}
	switch disk.Status() {
	case DiskStatusOK:
		return disk, nil
	case DiskStatusLocked:
		return nil, newError(EPending, "disk status is %s, not %s", disk.Status(), DiskStatusOK)
	default:
		return nil, newError(EUnexpectedDiskStatus, "disk status is %s, not %s", disk.Status(), DiskStatusOK)
	}
}
