package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) RemoveStorageDomain(
	id StorageDomainID,
	params RemoveStorageDomainParams,
	retries ...RetryStrategy,
) (err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	if params == nil {
		params = NewRemoveStorageDomainParams()
	}

	err = retry(
		fmt.Sprintf("removing storage domain %s", id),
		o.logger,
		retries,
		func() error {
			req := o.conn.SystemService().StorageDomainsService().StorageDomainService(string(id)).Remove()
			if params != nil {
				if d := params.Destroy(); d != nil {
					req.Destroy(*d)
				}
				if f := params.Format(); f != nil {
					req.Format(*f)
				}
				if h := params.Host(); h != nil {
					req.Host(*h)
				}
			}
			_, err := req.Send()
			return err
		},
	)
	return err
}

func (m *mockClient) RemoveStorageDomain(
	id StorageDomainID,
	params RemoveStorageDomainParams,
	_ ...RetryStrategy,
) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.storageDomains[id]; !ok {
		return newError(ENotFound, "storage domain with ID %s not found", id)
	}

	destroy := false
	if params != nil && params.Destroy() != nil {
		destroy = *params.Destroy()
	}

	if !destroy {
		for _, disk := range m.disks {
			for _, sID := range disk.storageDomainIDs {
				if sID == id {
					return newError(EConflict, "cannot remove storage domain %s because it still contains disks", id)
				}
			}
		}
	} else {
		var toDelete []DiskID
		toUpdate := make(map[DiskID]*diskWithData)
		for diskID, d := range m.disks {
			found := false
			remaining := make([]StorageDomainID, 0, len(d.storageDomainIDs))
			for _, sID := range d.storageDomainIDs {
				if sID == id {
					found = true
				} else {
					remaining = append(remaining, sID)
				}
			}
			if !found {
				continue
			}
			if len(remaining) == 0 {
				if _, ok := m.templateDiskAttachmentsByDisk[diskID]; ok {
					return newError(EConflict, "cannot remove storage domain %s because disk %s is attached to a template", id, diskID)
				}
				toDelete = append(toDelete, diskID)
			} else {
				updatedDisk := *d
				updatedDisk.storageDomainIDs = remaining
				toUpdate[diskID] = &updatedDisk
			}
		}
		for _, diskID := range toDelete {
			m.deleteDiskLocked(diskID)
		}
		for diskID, updatedDisk := range toUpdate {
			m.disks[diskID] = updatedDisk
		}
	}

	delete(m.storageDomains, id)
	return nil
}
