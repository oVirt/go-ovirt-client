package ovirtclient

import (
	"github.com/google/uuid"
)

func (m *mockClient) CreateVM(clusterID string, templateID string, name string, params OptionalVMParameters, _ ...RetryStrategy) (VM, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := validateVMCreationParameters(clusterID, templateID, name, params); err != nil {
		return nil, err
	}
	if _, ok := m.clusters[clusterID]; !ok {
		return nil, newError(ENotFound, "cluster with ID %s not found", clusterID)
	}
	if _, ok := m.templates[templateID]; !ok {
		return nil, newError(ENotFound, "template with ID %s not found", templateID)
	}
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty for VM creation")
	}
	if params == nil {
		params = &vmParams{}
	}

	id := uuid.Must(uuid.NewUUID()).String()
	vm := &vm{
		client:     m,
		id:         id,
		clusterID:  clusterID,
		templateID: templateID,
		name:       name,
		status:     VMStatusDown,
	}
	if comment := params.Comment(); comment != nil {
		vm.comment = *comment
	}
	if cpu := params.CPU(); cpu != nil {
		vm.cpu = *cpu
	}
	if memoryMB := params.MemoryMB(); memoryMB != nil {
		vm.memoryMB = *memoryMB
	}
	if vmType := params.VMType(); vmType != nil {
		vm.vmType = *vmType
	}
	if autoPiningPolicy := params.AutoPinningPolicy(); autoPiningPolicy != nil {
		vm.autoPiningPolicy = *autoPiningPolicy
	}
	if placementPolicy := params.PlacementPolicy(); placementPolicy != nil {
		vm.placementPolicy = placementPolicy
	}
	if hugepages := params.Hugepages(); hugepages != nil {
		vm.hugepages = hugepages
	}
	if guaranteedMemoryMB := params.GuaranteedMemoryMB(); guaranteedMemoryMB != nil {
		if memory := params.MemoryMB(); *guaranteedMemoryMB > *memory {
			return nil, newError(EBadArgument, "guaranteedMemoryMB must be greater than or equal to memory")
		}
		vm.guaranteedMemoryMB = guaranteedMemoryMB
	}
	if initialization := params.Initialization(); initialization != nil {
		vm.initialization = initialization
	}
	if tags := params.Tags(); tags != nil {
		vm.tags = tags
	}

	m.vms[id] = vm
	m.diskAttachmentsByVM[id] = map[string]*diskAttachment{}
	return vm, nil
}
