// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) GetCluster(id string, retries ...RetryStrategy) (Cluster, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if disk, ok := m.clusters[id]; ok {
		return disk, nil
	}
	return nil, newError(ENotFound, "cluster with ID %s not found", id)
}
