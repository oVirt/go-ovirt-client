// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) 	CopyTemplateDiskToStorageDomain(
	diskId string,
	storageDomainId string,
	retries ...RetryStrategy) (result Disk, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	disk, ok := m.disks[diskId]
	if !ok {
		return nil, newError(ENotFound, "disk with ID %s not found", diskId)
	}
	if err := disk.Lock(); err != nil {
		return nil, err
	}
	update := &mockDiskCopy{
		client: m,
		disk:   disk,
		done:   make(chan struct{}),
	}
	defer update.do()
	return disk, nil
}
