package ovirtclient

func (m *mockClient) RemoveTemplate(id TemplateID, _ ...RetryStrategy) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	tpl, ok := m.templates[id]
	if !ok {
		return newError(ENotFound, "Template with ID %s was not found", id)
	}
	if tpl.status != TemplateStatusOK {
		return newError(EConflict, "Template %s is in status %s, not %s.", id, tpl.status, TemplateStatusOK)
	}

	delete(m.templates, id)
	return nil
}
