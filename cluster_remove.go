package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) RemoveCluster(
	id ClusterID,
	retries ...RetryStrategy,
) (err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	err = retry(
		fmt.Sprintf("removing cluster %s", id),
		o.logger,
		retries,
		func() error {
			_, err := o.conn.SystemService().ClustersService().ClusterService(string(id)).Remove().Send()
			return err
		},
	)
	return err
}

func (m *mockClient) RemoveCluster(
	id ClusterID,
	_ ...RetryStrategy,
) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	c, ok := m.clusters[id]
	if !ok {
		return newError(ENotFound, "cluster with ID %s not found", id)
	}

	// Check if cluster has hosts attached
	for _, h := range m.hosts {
		if h.clusterID == id {
			return newError(EConflict, "cannot remove cluster with ID %s because it still has hosts attached", id)
		}
	}

	// Check if cluster has VMs attached
	for _, vm := range m.vms {
		if vm.clusterID == id {
			return newError(EConflict, "cannot remove cluster with ID %s because it still has VMs attached", id)
		}
	}

	// Check if cluster has affinity groups
	if ags, ok := m.affinityGroups[id]; ok && len(ags) > 0 {
		return newError(EConflict, "cannot remove cluster with ID %s because it still has affinity groups attached", id)
	}

	// Remove from datacenter cluster list
	if dc, ok := m.dataCenters[c.dcID]; ok {
		var newClusters []ClusterID
		for _, clID := range dc.clusters {
			if clID != id {
				newClusters = append(newClusters, clID)
			}
		}
		dc.clusters = newClusters
	}

	delete(m.clusters, id)
	return nil
}
