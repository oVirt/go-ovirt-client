package ovirtclient

import (
	"fmt"
	ovirtsdk "github.com/ovirt/go-ovirt"
	"sync"
)

func (o *oVirtClient) CopyTemplateDiskToStorageDomain(
	diskId string,
	storageDomainId string,
	retries ...RetryStrategy) (result Disk, err error) {
	retries = defaultRetries(retries, defaultReadTimeouts())
	progress,err := o.StartCopyTemplateDiskToStorageDomain(diskId,storageDomainId, retries...)
	if err != nil {
		return progress.Disk(), err
	}

	return progress.Wait()
}

func (o *oVirtClient) StartCopyTemplateDiskToStorageDomain(
	diskId string,
	storageDomainId string,
	retries ...RetryStrategy)(DiskUpdate,error){
	o.logger.Infof("Starting copy template disk to different storage domain.")
	retries = defaultRetries(retries, defaultWriteTimeouts())
	correlationID := fmt.Sprintf("template_disk_copy_%s", generateRandomID(5, o.nonSecRand))
	sdkStorageDomain := ovirtsdk.NewStorageDomainBuilder().Id(storageDomainId)
	sdkDisk := ovirtsdk.NewDiskBuilder().Id(diskId)
	storageDomain,_ := o.GetStorageDomain(storageDomainId)
	disk,_ := o.GetDisk(diskId)

	err := retry(
		fmt.Sprintf("copying disk %s to storageDomain %s", diskId,storageDomainId),
		o.logger,
		retries,
		func() error {
			_, err := o.conn.
				SystemService().
				DisksService().
				DiskService(diskId).
				Copy().
				StorageDomain(sdkStorageDomain.MustBuild()).
				Disk(sdkDisk.MustBuild()).
				Query("correlation_id", correlationID).
				Send()

			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return &storageDomainDiskWait{
		client:        o,
		disk:          disk,
		storageDomain: storageDomain,
		correlationID: correlationID,
		lock:          &sync.Mutex{},
	},nil
}