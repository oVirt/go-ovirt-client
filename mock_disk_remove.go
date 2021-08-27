package ovirtclient

func (m *mockClient) RemoveDisk(diskID string, _ ...RetryStrategy) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.disks[diskID]; !ok {
		return newError(ENotFound, "disk with ID %s not found", diskID)
	}

	delete(m.diskAttachmentsByDisk, diskID)
	delete(m.disks, diskID)

	return nil
}
