// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) GetDisk(id string, retries ...RetryStrategy) (Disk, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if disk, ok := m.disks[id]; ok {
		return disk, nil
	}
	return nil, newError(ENotFound, "disk with ID %s not found", id)
}
