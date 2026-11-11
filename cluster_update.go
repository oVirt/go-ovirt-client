package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

func (o *oVirtClient) UpdateCluster(
	id ClusterID,
	params UpdateClusterParams,
	retries ...RetryStrategy,
) (result Cluster, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	if params == nil {
		params = NewUpdateClusterParams()
	}

	err = retry(
		fmt.Sprintf("updating cluster %s", id),
		o.logger,
		retries,
		func() error {
			builder := ovirtsdk4.NewClusterBuilder().Id(string(id))
			hasChanges := false
			if name := params.Name(); name != nil {
				builder.Name(*name)
				hasChanges = true
			}
			if desc := params.Description(); desc != nil {
				builder.Description(*desc)
				hasChanges = true
			}
			if comment := params.Comment(); comment != nil {
				builder.Comment(*comment)
				hasChanges = true
			}
			cpuBuilder := ovirtsdk4.NewCpuBuilder()
			hasCPUChanges := false
			if ct := params.CPUType(); ct != nil {
				cpuBuilder.Type(*ct)
				hasCPUChanges = true
			}
			if ca := params.CPUArchitecture(); ca != nil {
				cpuBuilder.Architecture(ovirtsdk4.Architecture(*ca))
				hasCPUChanges = true
			}
			if hasCPUChanges {
				builder.Cpu(cpuBuilder.MustBuild())
				hasChanges = true
			}
			if b := params.Ballooning(); b != nil {
				builder.BallooningEnabled(*b)
				hasChanges = true
			}
			if overcommit := params.MemoryOverCommitPercent(); overcommit != nil {
				memPolicy := ovirtsdk4.NewMemoryPolicyBuilder().
					OverCommit(ovirtsdk4.NewMemoryOverCommitBuilder().Percent(*overcommit).MustBuild()).
					MustBuild()
				builder.MemoryPolicy(memPolicy)
				hasChanges = true
			}

			if !hasChanges {
				return nil
			}

			response, e := o.conn.SystemService().ClustersService().ClusterService(string(id)).Update().Cluster(builder.MustBuild()).Send()
			if e != nil {
				return e
			}
			sdkCluster, ok := response.Cluster()
			if !ok {
				return newFieldNotFound("response from cluster update", "cluster")
			}
			result, err = convertSDKCluster(sdkCluster, o)
			return err
		},
	)
	if err == nil && result == nil {
		return o.GetCluster(id, retries...)
	}
	return result, err
}

func (m *mockClient) UpdateCluster(
	id ClusterID,
	params UpdateClusterParams,
	_ ...RetryStrategy,
) (Cluster, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	c, ok := m.clusters[id]
	if !ok {
		return nil, newError(ENotFound, "cluster with ID %s not found", id)
	}

	name := c.name
	if newName := params.Name(); newName != nil {
		if *newName == "" {
			return nil, newError(EBadArgument, "name cannot be empty")
		}
		for _, existing := range m.clusters {
			if existing.id != id && existing.name == *newName {
				return nil, newError(EConflict, "cluster with name %s already exists", *newName)
			}
		}
		name = *newName
	}

	description := c.description
	if newDesc := params.Description(); newDesc != nil {
		description = newDesc
	}

	comment := c.comment
	if newComment := params.Comment(); newComment != nil {
		comment = newComment
	}

	cpuType := c.cpuType
	if newCPUType := params.CPUType(); newCPUType != nil {
		if *newCPUType == "" {
			return nil, newError(EBadArgument, "cpuType cannot be empty")
		}
		cpuType = *newCPUType
	}

	cpuArch := c.cpuArch
	if newCPUArch := params.CPUArchitecture(); newCPUArch != nil {
		cpuArch = newCPUArch
	}

	ballooning := c.ballooning
	if newBallooning := params.Ballooning(); newBallooning != nil {
		ballooning = newBallooning
	}

	memOvercommit := c.memoryOverCommitPercent
	if newMem := params.MemoryOverCommitPercent(); newMem != nil {
		memOvercommit = newMem
	}

	c.name = name
	c.description = description
	c.comment = comment
	c.cpuType = cpuType
	c.cpuArch = cpuArch
	c.ballooning = ballooning
	c.memoryOverCommitPercent = memOvercommit

	return c, nil
}
