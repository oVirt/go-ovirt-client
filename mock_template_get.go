// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) GetTemplate(id string, retries ...RetryStrategy) (Template, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if disk, ok := m.templates[id]; ok {
		return disk, nil
	}
	return nil, newError(ENotFound, "template with ID %s not found", id)
}
