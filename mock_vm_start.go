package ovirtclient

import (
	"time"
)

func (m *mockClient) StartVM(id string, _ ...RetryStrategy) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	item, ok := m.vms[id]
	if !ok {
		return newError(ENotFound, "vm with ID %s not found", id)
	}

	if item.Status() == VMStatusUp {
		return nil
	}

	hostID, err := m.findSuitableHost(id)
	if err != nil {
		return err
	}
	item.hostID = &hostID
	item.status = VMStatusWaitForLaunch
	go func() {
		time.Sleep(2 * time.Second)
		m.lock.Lock()
		if item.status != VMStatusWaitForLaunch {
			m.lock.Unlock()
			return
		}
		item.status = VMStatusPoweringUp
		m.lock.Unlock()
		time.Sleep(2 * time.Second)
		m.lock.Lock()
		defer m.lock.Unlock()
		if item.status != VMStatusPoweringUp {
			return
		}
		item.status = VMStatusUp
	}()
	return nil
}

func (m *mockClient) findSuitableHost(vmID string) (string, error) {
	var affectedAffinityGroups []*affinityGroup
	for _, clusterAffinityGroups := range m.affinityGroups {
		for _, affinityGroup := range clusterAffinityGroups {
			if affinityGroup.hasVM(vmID) && (affinityGroup.Enforcing() || affinityGroup.vmsRule.Enforcing()) {
				affectedAffinityGroups = append(affectedAffinityGroups, affinityGroup)
			}
		}
	}
	// Try to find a host that is suitable.
	var foundHost *host
	for _, host := range m.hosts {
		hostSuitable := true
	loop:
		for _, vm := range m.vms {
			if vm.hostID != nil && *vm.hostID == host.id {
				// If the VM resides on the current host
				for _, ag := range affectedAffinityGroups {
					// Check if the VM is a member of the AGs we care about
					if ag.hasVM(vm.id) {
						hostSuitable = false
						break loop
					}
				}
			}
		}
		if hostSuitable {
			foundHost = host
			break
		}
	}
	if foundHost == nil {
		return "", newError(EConflict, "no suitable host found matching affinity group rules")
	}
	hostID := foundHost.ID()
	return hostID, nil
}
