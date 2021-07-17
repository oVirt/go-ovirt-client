// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) GetHost(id string, retries ...RetryStrategy) (Host, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if disk, ok := m.hosts[id]; ok {
		return disk, nil
	}
	return nil, newError(ENotFound, "host with ID %s not found", id)
}
