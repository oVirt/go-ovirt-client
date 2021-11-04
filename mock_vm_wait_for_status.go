package ovirtclient

func (m *mockClient) WaitForStatus(id string, desiredStatus VMStatus, _ ...RetryStrategy) (VM, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.vms[id]; !ok {
		return nil, newError(ENotFound, "VM with ID %s not found", id)
	}
	if m.vms[id].status != desiredStatus {
		return nil, newError(
			ETimeout,
			"timeout while waiting for VM to reach state %s, current state %s",
			desiredStatus,
			m.vms[id].status)
	}
	return m.vms[id], nil
}
