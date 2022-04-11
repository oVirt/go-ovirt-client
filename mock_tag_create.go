package ovirtclient

import "github.com/google/uuid"

func (m *mockClient) CreateTag(name string, params CreateTagParams, _ ...RetryStrategy) (result Tag, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	id := uuid.Must(uuid.NewUUID()).String()
	if params == nil {
		params = NewCreateTagParams()
	}
	tag := &tag{
		client:      m,
		id:          id,
		name:        name,
		description: params.Description(),
	}
	m.tags[id] = tag

	result = tag
	return
}
