package ovirtclient

import "github.com/google/uuid"

func (m *mockClient) CreateTag(name string, description string, retries ...RetryStrategy) (result Tag, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	tag := m.createTag(name, description)

	result = tag

	return
}

func (m *mockClient) createTag(
	name string,
	description string,
) *tag {
	id := uuid.Must(uuid.NewUUID()).String()
	tag := &tag{
		client:      m,
		id:          id,
		name:        name,
		description: description,
	}
	m.tags[id] = tag
	return tag
}
