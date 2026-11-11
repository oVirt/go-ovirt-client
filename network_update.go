package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

func (o *oVirtClient) UpdateNetwork(
	id NetworkID,
	params UpdateNetworkParams,
	retries ...RetryStrategy,
) (result Network, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	if params == nil {
		params = NewUpdateNetworkParams()
	}

	err = retry(
		fmt.Sprintf("updating network %s", id),
		o.logger,
		retries,
		func() error {
			builder := ovirtsdk4.NewNetworkBuilder().Id(string(id))
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
			if vlan := params.VLAN(); vlan != nil {
				builder.Vlan(ovirtsdk4.NewVlanBuilder().Id(int64(*vlan)).MustBuild())
				hasChanges = true
			}
			if !hasChanges {
				return nil
			}

			response, e := o.conn.SystemService().NetworksService().NetworkService(string(id)).Update().Network(builder.MustBuild()).Send()
			if e != nil {
				return e
			}
			sdkNetwork, ok := response.Network()
			if !ok {
				return newFieldNotFound("response from network update", "network")
			}
			result, err = convertSDKNetwork(sdkNetwork, o)
			return err
		},
	)
	if err == nil && result == nil {
		return o.GetNetwork(id, retries...)
	}
	return result, err
}

func (m *mockClient) UpdateNetwork(
	id NetworkID,
	params UpdateNetworkParams,
	_ ...RetryStrategy,
) (result Network, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	net, ok := m.networks[id]
	if !ok {
		return nil, newError(ENotFound, "network with ID %s not found", id)
	}

	name := net.name
	if newName := params.Name(); newName != nil {
		if *newName == "" {
			return nil, newError(EBadArgument, "network name cannot be empty")
		}
		for _, existing := range m.networks {
			if existing.id != id && existing.dcID == net.dcID && existing.name == *newName {
				return nil, newError(EConflict, "network with name %s already exists in datacenter %s", *newName, net.dcID)
			}
		}
		name = *newName
	}

	description := net.description
	if newDesc := params.Description(); newDesc != nil {
		description = newDesc
	}

	comment := net.comment
	if newComment := params.Comment(); newComment != nil {
		comment = newComment
	}

	vlan := net.vlan
	if newVlan := params.VLAN(); newVlan != nil {
		if *newVlan < 1 || *newVlan > 4094 {
			return nil, newError(EBadArgument, "VLAN ID must be between 1 and 4094")
		}
		vlan = newVlan
	}

	updated := &network{
		client:      m,
		id:          id,
		name:        name,
		description: description,
		comment:     comment,
		vlan:        vlan,
		dcID:        net.dcID,
	}
	m.networks[id] = updated
	return updated, nil
}
