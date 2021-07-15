package ovirtclient

import (
	"context"
	"time"
)

func (o *oVirtClient) RemoveDisk(ctx context.Context, diskID string) error {
	var lastError EngineError
	for {
		_, err := o.conn.SystemService().DisksService().DiskService(diskID).Remove().Send()
		if err == nil {
			return err
		}
		lastError = wrap(
			err,
			EUnidentified,
			"failed to remove disk %s",
			diskID,
		)
		if !lastError.CanAutoRetry() {
			return lastError
		}
		select {
		case <-ctx.Done():
			return wrap(
				lastError,
				ETimeout,
				"timeout while removing disk",
			)
		case <-time.After(10 * time.Second):
		}
	}
}
