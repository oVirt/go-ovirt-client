package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) RemoveStorageDomainDisk(id string, diskId string, retries ...RetryStrategy) (err error) {
	retries = defaultRetries(retries, defaultReadTimeouts())
	err = retry(
		fmt.Sprintf("removing disk %s from storage domain %s", diskId, id),
		o.logger,
		retries,
		func() error {
			_, err := o.conn.SystemService().StorageDomainsService().
				StorageDomainService(id).DisksService().DiskService(diskId).Remove().Send()
			if err != nil {
				o.logger.Infof("error removing disk..")
				return err
			}

			return nil
		})
	return
}
