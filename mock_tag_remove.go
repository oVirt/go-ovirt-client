package ovirtclient

func (m *mockClient) RemoveTag(id string, _ ...RetryStrategy) (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.tags[id]; !ok {
		return newError(ENotFound, "Tag with ID %s not found", id)
	}

	delete(m.tags, id)

	return nil
}
