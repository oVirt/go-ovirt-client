package ovirtclient

import (
	"sync"
	"time"
)

func (m *mockClient) StartCreateDisk(
	storageDomainID string,
	format ImageFormat,
	size uint64,
	params CreateDiskOptionalParameters,
	_ ...RetryStrategy,
) (DiskCreation, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.storageDomains[storageDomainID]; !ok {
		return nil, newError(ENotFound, "storage domain with ID %s not found", storageDomainID)
	}

	disk := &diskWithData{
		disk: disk{
			client:          m,
			id:              m.GenerateUUID(),
			format:          format,
			provisionedSize: size,
			totalSize:       size,
			storageDomainID: storageDomainID,
			status:          DiskStatusLocked,
		},
		lock:   &sync.Mutex{},
		locked: true,
		data:   nil,
	}

	if params != nil {
		if alias := params.Alias(); alias != "" {
			disk.disk.alias = alias
		}
		if sparse := params.Sparse(); sparse != nil {
			disk.disk.sparse = *sparse
		}
	}

	m.disks[disk.id] = disk

	return &mockDiskCreation{
		client: m,
		disk:   disk,
	}, nil
}

func (m *mockClient) CreateDisk(
	storageDomainID string,
	format ImageFormat,
	size uint64,
	params CreateDiskOptionalParameters,
	retries ...RetryStrategy,
) (Disk, error) {
	result, err := m.StartCreateDisk(storageDomainID, format, size, params, retries...)
	if err != nil {
		return nil, err
	}
	return result.Wait()
}

type mockDiskCreation struct {
	client *mockClient
	disk   *diskWithData
}

func (c *mockDiskCreation) Disk() Disk {
	c.client.lock.Lock()
	defer c.client.lock.Unlock()
	return c.disk
}

func (c *mockDiskCreation) Wait(_ ...RetryStrategy) (Disk, error) {
	c.client.lock.Lock()
	if !c.disk.locked {
		disk := c.disk
		c.client.lock.Unlock()
		return disk, nil
	}
	c.client.lock.Unlock()
	time.Sleep(time.Second)
	c.client.lock.Lock()
	newDisk := *c.disk
	newDisk.status = DiskStatusOK
	newDisk.locked = false
	c.client.disks[newDisk.id] = &newDisk
	c.disk = &newDisk
	c.client.lock.Unlock()
	return c.disk, nil
}
