// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) GetStorageDomain(id string, _ ...RetryStrategy) (StorageDomain, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if item, ok := m.storageDomains[id]; ok {
		return item, nil
	}
	return nil, newError(ENotFound, "storage domain with ID %s not found", id)
}
