package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) GetStorageDomainDisk(id string, diskId string, retries ...RetryStrategy) (result Disk, err error) {
	retries = defaultRetries(retries, defaultReadTimeouts())
	err = retry(
		fmt.Sprintf("getting disk %s from storage domain %s", diskId, id),
		o.logger,
		retries,
		func() error {
			response, err := o.conn.SystemService().StorageDomainsService().
				StorageDomainService(id).DisksService().DiskService(diskId).Get().Send()
			if err != nil {
				return err
			}
			sdkObject, ok := response.Disk()
			if !ok {
				return newError(
					ENotFound,
					"disk %s not found on storage domain ID %s",
					diskId,
					id,
				)
			}
			result, err = convertSDKDisk(sdkObject, o)
			if err != nil {
				return wrap(
					err,
					EBug,
					"failed to convert disk %s",
					diskId,
				)
			}
			return nil
		})
	return
}
