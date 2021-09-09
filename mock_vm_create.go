package ovirtclient

import (
	"github.com/google/uuid"
)

func (m *mockClient) CreateVM(clusterID string, templateID string, params OptionalVMParameters, _ ...RetryStrategy) (VM, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := validateVMCreationParameters(clusterID, templateID, params); err != nil {
		return nil, err
	}
	if _, ok := m.clusters[clusterID]; !ok {
		return nil, newError(ENotFound, "cluster with ID %s not found", clusterID)
	}
	if _, ok := m.templates[templateID]; !ok {
		return nil, newError(ENotFound, "template with ID %s not found", templateID)
	}

	if params == nil {
		params = &vmParams{}
	}

	id := uuid.Must(uuid.NewUUID()).String()
	vm := &vm{
		client:     m,
		id:         id,
		name:       params.Name(),
		comment:    params.Comment(),
		clusterID:  clusterID,
		templateID: templateID,
		status:     VMStatusDown,
	}
	m.vms[id] = vm
	m.diskAttachmentsByVM[id] = map[string]*diskAttachment{}
	return vm, nil
}
