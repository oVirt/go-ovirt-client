package ovirtclient

import (
	"time"
)

func (m *mockClient) CreateTemplate(
	vmID string,
	name string,
	params OptionalTemplateCreateParameters,
	_ ...RetryStrategy,
) (Template, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if params == nil {
		params = &templateCreateParameters{}
	}

	_, ok := m.vms[vmID]
	if !ok {
		return nil, newError(ENotFound, "VM with ID %s not found", vmID)
	}
	description := ""
	if desc := params.Description(); desc != nil {
		description = *desc
	}
	tpl := &template{
		client:      m,
		id:          TemplateID(m.GenerateUUID()),
		name:        name,
		description: description,
		status:      TemplateStatusLocked,
	}
	m.templates[tpl.ID()] = tpl
	go func() {
		time.Sleep(2 * time.Second)
		m.lock.Lock()
		defer m.lock.Unlock()
		tpl.status = TemplateStatusOK
	}()
	return tpl, nil
}
