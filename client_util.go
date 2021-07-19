package ovirtclient

import (
	"context"
	"fmt"
	"time"

	ovirtsdk "github.com/ovirt/go-ovirt"
)

// waitForJobFinished waits for a job to truly finish. This is especially important when disks are involved as their
// status changes to OK prematurely.
//
// correlationID is a query parameter assigned to a job before it is sent to the ovirt engine, it must be unique and
// under 30 chars. To set a correlationID add `Query("correlation_id", correlationID)` to the engine API call, for
// example:
//
//     correlationID := fmt.Sprintf("image_transfer_%s", utilrand.String(5))
//     conn.
//         SystemService().
//         DisksService().
//         DiskService(diskId).
//         Update().
//         Query("correlation_id", correlationID).
//         Send()
func (o *oVirtClient) waitForJobFinished(ctx context.Context, correlationID string) error {
	o.logger.Debugf("Waiting for job with correlation ID %s to finish...", correlationID)
	var lastError EngineError
	for {
		jobResp, err := o.conn.SystemService().JobsService().List().Search(fmt.Sprintf("correlation_id=%s", correlationID)).Send()
		if err == nil {
			if jobSlice, ok := jobResp.Jobs(); ok {
				if len(jobSlice.Slice()) == 0 {
					return nil
				}
				for _, job := range jobSlice.Slice() {
					if status, _ := job.Status(); status != ovirtsdk.JOBSTATUS_STARTED {
						return nil
					}
				}
			}
			o.logger.Debugf("Job with correlation ID %s is still pending, retrying in 5 seconds...", correlationID)
			lastError = newError(EPending, "job for correlation ID %s still pending", correlationID)
		} else {
			realErr := wrap(err, EUnidentified, "failed to list jobs for correlation ID %s", correlationID)
			if !realErr.CanAutoRetry() {
				o.logger.Debugf("Failed to fetch job list with correlation ID %s, giving up. (%v)", correlationID, err)
				return realErr
			}
			o.logger.Debugf("Failed to fetch job list with correlation ID %s, retrying in 5 seconds... (%v)", correlationID, err)
			lastError = realErr
		}
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			o.logger.Debugf("Timeout while waiting for job with correlation ID %s to finish. (last error: %v)", correlationID, lastError)
			return wrap(
				lastError,
				ETimeout,
				"timeout while waiting for job with correlation_id %s to finish", correlationID)
		}
	}
}
