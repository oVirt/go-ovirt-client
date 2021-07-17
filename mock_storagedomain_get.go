// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) GetStorageDomain(id string, retries ...RetryStrategy) (StorageDomain, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if disk, ok := m.storageDomains[id]; ok {
		return disk, nil
	}
	return nil, newError(ENotFound, "storage domain with ID %s not found", id)
}
