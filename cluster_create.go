package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

func (o *oVirtClient) CreateCluster(
	datacenterID DatacenterID,
	name string,
	cpuType string,
	params CreateClusterParams,
	retries ...RetryStrategy,
) (result Cluster, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	if params == nil {
		params = NewCreateClusterParams()
	}
	if datacenterID == "" {
		return nil, newError(EBadArgument, "datacenter ID cannot be empty for cluster creation")
	}
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty for cluster creation")
	}
	if cpuType == "" {
		return nil, newError(EBadArgument, "cpuType cannot be empty for cluster creation")
	}

	err = retry(
		fmt.Sprintf("creating cluster %s", name),
		o.logger,
		retries,
		func() error {
			cpuBuilder := ovirtsdk4.NewCpuBuilder().Type(cpuType)
			if arch := params.CPUArchitecture(); arch != nil {
				cpuBuilder.Architecture(ovirtsdk4.Architecture(*arch))
			}

			builder := ovirtsdk4.NewClusterBuilder().
				Name(name).
				DataCenter(ovirtsdk4.NewDataCenterBuilder().Id(string(datacenterID)).MustBuild()).
				Cpu(cpuBuilder.MustBuild())

			if desc := params.Description(); desc != nil {
				builder.Description(*desc)
			}
			if comment := params.Comment(); comment != nil {
				builder.Comment(*comment)
			}
			if b := params.Ballooning(); b != nil {
				builder.BallooningEnabled(*b)
			}
			if overcommit := params.MemoryOverCommitPercent(); overcommit != nil {
				memPolicy := ovirtsdk4.NewMemoryPolicyBuilder().
					OverCommit(ovirtsdk4.NewMemoryOverCommitBuilder().Percent(*overcommit).MustBuild()).
					MustBuild()
				builder.MemoryPolicy(memPolicy)
			}

			response, e := o.conn.SystemService().ClustersService().Add().Cluster(builder.MustBuild()).Send()
			if e != nil {
				return e
			}
			sdkCluster, ok := response.Cluster()
			if !ok {
				return newFieldNotFound("response from cluster creation", "cluster")
			}
			result, err = convertSDKCluster(sdkCluster, o)
			return err
		},
	)
	return result, err
}

func (m *mockClient) CreateCluster(
	datacenterID DatacenterID,
	name string,
	cpuType string,
	params CreateClusterParams,
	_ ...RetryStrategy,
) (Cluster, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if params == nil {
		params = NewCreateClusterParams()
	}
	if datacenterID == "" {
		return nil, newError(EBadArgument, "datacenter ID cannot be empty for cluster creation")
	}
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty for cluster creation")
	}
	if cpuType == "" {
		return nil, newError(EBadArgument, "cpuType cannot be empty for cluster creation")
	}

	dc, ok := m.dataCenters[datacenterID]
	if !ok {
		return nil, newError(ENotFound, "datacenter with ID %s not found", datacenterID)
	}

	for _, existing := range m.clusters {
		if existing.name == name {
			return nil, newError(EConflict, "cluster with name %s already exists", name)
		}
	}

	id := ClusterID(m.GenerateUUID())
	c := &cluster{
		client:                  m,
		id:                      id,
		name:                    name,
		dcID:                    datacenterID,
		description:             params.Description(),
		comment:                 params.Comment(),
		cpuType:                 cpuType,
		cpuArch:                 params.CPUArchitecture(),
		ballooning:              params.Ballooning(),
		memoryOverCommitPercent: params.MemoryOverCommitPercent(),
	}
	m.clusters[id] = c
	dc.clusters = append(dc.clusters, id)
	return c, nil
}
