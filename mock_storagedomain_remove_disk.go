// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) RemoveStorageDomainDisk(id string, diskId string, _ ...RetryStrategy) ( error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.disks[diskId]; ok {
		return  nil
	}
	return newError(ENotFound, "storage domain with ID %s not found", id)
}
