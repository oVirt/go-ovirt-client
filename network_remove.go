package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) RemoveNetwork(
	id NetworkID,
	retries ...RetryStrategy,
) (err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	err = retry(
		fmt.Sprintf("removing network %s", id),
		o.logger,
		retries,
		func() error {
			_, err := o.conn.SystemService().NetworksService().NetworkService(string(id)).Remove().Send()
			return err
		},
	)
	return err
}

func (m *mockClient) RemoveNetwork(
	id NetworkID,
	_ ...RetryStrategy,
) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.networks[id]; !ok {
		return newError(ENotFound, "network with ID %s not found", id)
	}

	for _, profile := range m.vnicProfiles {
		if profile.networkID == id {
			return newError(EConflict, "cannot remove network with ID %s because it is still in use by VNIC profile %s", id, profile.id)
		}
	}

	delete(m.networks, id)
	return nil
}
