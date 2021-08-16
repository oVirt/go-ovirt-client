// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) Get{{ .Object }}(id string, _ ...RetryStrategy) ({{ .Object }}, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if item, ok := m.{{ .ID | toLower }}s[id]; ok {
		return item, nil
	}
	return nil, newError(ENotFound, "{{ .Name }} with ID %s not found", id)
}
