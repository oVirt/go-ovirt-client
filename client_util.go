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
			lastError = newError(EPending, "job for correlation ID %s still pending", correlationID)
		} else {
			realErr := wrap(err, EUnidentified, "failed to list jobs for correlation ID %s", correlationID)
			if !realErr.CanAutoRetry() {
				return realErr
			}
			lastError = realErr
		}
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return wrap(
				lastError,
				ETimeout,
				"timeout while waiting for job with correlation_id %s to finish", correlationID)
		}
	}
}
