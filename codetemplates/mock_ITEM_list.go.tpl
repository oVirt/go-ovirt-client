// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (m *mockClient) List{{ .Object }}s(retries ...RetryStrategy) ([]{{ .Object }}, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	result := make([]{{ .Object }}, len(m.{{ .ID | toLower }}s))
	i := 0
	for _, item := range m.{{ .ID | toLower }}s {
		result[i] = item
		i++
	}
	return result, nil
}
