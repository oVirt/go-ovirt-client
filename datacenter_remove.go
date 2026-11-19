package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) RemoveDatacenter(
	id DatacenterID,
	retries ...RetryStrategy,
) (err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	err = retry(
		fmt.Sprintf("removing datacenter %s", id),
		o.logger,
		retries,
		func() error {
			_, err := o.conn.SystemService().DataCentersService().DataCenterService(string(id)).Remove().Send()
			return err
		},
	)
	return err
}

func (m *mockClient) RemoveDatacenter(
	id DatacenterID,
	_ ...RetryStrategy,
) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	dc, ok := m.dataCenters[id]
	if !ok {
		return newError(ENotFound, "datacenter with ID %s not found", id)
	}

	if len(dc.clusters) > 0 {
		return newError(EConflict, "cannot remove datacenter with ID %s because it still has clusters attached", id)
	}

	for _, n := range m.networks {
		if n.dcID == id {
			return newError(EConflict, "cannot remove datacenter with ID %s because it still has networks attached", id)
		}
	}

	delete(m.dataCenters, id)
	return nil
}
