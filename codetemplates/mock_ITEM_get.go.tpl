// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) Get{{ .ID }}(id string, retries ...RetryStrategy) ({{ .ID }}, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if disk, ok := m.{{ .ID | toLower }}s[id]; ok {
		return disk, nil
	}
	return nil, newError(ENotFound, "{{ .Name }} with ID %s not found", id)
}
