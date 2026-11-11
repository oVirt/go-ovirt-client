package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

func (o *oVirtClient) CreateNetwork(
	datacenterID DatacenterID,
	name string,
	params CreateNetworkParams,
	retries ...RetryStrategy,
) (result Network, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	if params == nil {
		params = NewCreateNetworkParams()
	}
	if err := validateNetworkCreationParameters(datacenterID, name, params); err != nil {
		return nil, err
	}

	err = retry(
		fmt.Sprintf("creating network %s in datacenter %s", name, datacenterID),
		o.logger,
		retries,
		func() error {
			builder := ovirtsdk4.NewNetworkBuilder().
				Name(name).
				DataCenter(ovirtsdk4.NewDataCenterBuilder().Id(string(datacenterID)).MustBuild())
			if desc := params.Description(); desc != nil {
				builder.Description(*desc)
			}
			if comment := params.Comment(); comment != nil {
				builder.Comment(*comment)
			}
			if vlan := params.VLAN(); vlan != nil {
				builder.Vlan(ovirtsdk4.NewVlanBuilder().Id(int64(*vlan)).MustBuild())
			}

			response, e := o.conn.SystemService().NetworksService().Add().Network(builder.MustBuild()).Send()
			if e != nil {
				return e
			}
			sdkNetwork, ok := response.Network()
			if !ok {
				return newFieldNotFound("response from network creation", "network")
			}
			result, err = convertSDKNetwork(sdkNetwork, o)
			return err
		},
	)
	return result, err
}

func (m *mockClient) CreateNetwork(
	datacenterID DatacenterID,
	name string,
	params CreateNetworkParams,
	_ ...RetryStrategy,
) (result Network, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if params == nil {
		params = NewCreateNetworkParams()
	}
	if err := validateNetworkCreationParameters(datacenterID, name, params); err != nil {
		return nil, err
	}

	if _, ok := m.dataCenters[datacenterID]; !ok {
		return nil, newError(ENotFound, "datacenter with ID %s not found", datacenterID)
	}

	for _, existing := range m.networks {
		if existing.dcID == datacenterID && existing.name == name {
			return nil, newError(EConflict, "network with name %s already exists in datacenter %s", name, datacenterID)
		}
	}

	id := NetworkID(m.GenerateUUID())
	net := &network{
		client:      m,
		id:          id,
		name:        name,
		description: params.Description(),
		comment:     params.Comment(),
		vlan:        params.VLAN(),
		dcID:        datacenterID,
	}
	m.networks[id] = net
	return net, nil
}

func validateNetworkCreationParameters(datacenterID DatacenterID, name string, params CreateNetworkParams) error {
	if datacenterID == "" {
		return newError(EBadArgument, "datacenter ID cannot be empty for network creation")
	}
	if name == "" {
		return newError(EBadArgument, "name cannot be empty for network creation")
	}
	if vlan := params.VLAN(); vlan != nil {
		if *vlan < 1 || *vlan > 4094 {
			return newError(EBadArgument, "VLAN ID must be between 1 and 4094")
		}
	}
	return nil
}
