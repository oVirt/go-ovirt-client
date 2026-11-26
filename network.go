package ovirtclient

import (
	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

//go:generate go run scripts/rest/rest.go -i "Network" -n "network" -T NetworkID

// NetworkID is the UUID if the network.
type NetworkID string

// NetworkClient describes the functions related to oVirt networks.
//
// See https://www.ovirt.org/documentation/administration_guide/#chap-Logical_Networks for details.
type NetworkClient interface {
	// CreateNetwork creates a new network in the specified datacenter.
	CreateNetwork(datacenterID DatacenterID, name string, params CreateNetworkParams, retries ...RetryStrategy) (Network, error)
	// GetNetwork returns a single network based on its ID.
	GetNetwork(id NetworkID, retries ...RetryStrategy) (Network, error)
	// ListNetworks returns all networks on the oVirt engine.
	ListNetworks(retries ...RetryStrategy) ([]Network, error)
	// UpdateNetwork updates a network.
	UpdateNetwork(id NetworkID, params UpdateNetworkParams, retries ...RetryStrategy) (Network, error)
	// RemoveNetwork removes a network.
	RemoveNetwork(id NetworkID, retries ...RetryStrategy) error
}

// NetworkData is the core of Network, providing only the data access functions, but not the client
// functions.
type NetworkData interface {
	// ID returns the auto-generated identifier for this network.
	ID() NetworkID
	// Name returns the user-given name for this network.
	Name() string
	// Description returns the description for this network.
	Description() *string
	// Comment returns the comment for this network.
	Comment() *string
	// VLAN returns the optional VLAN tag (nil if no VLAN tag is set).
	VLAN() *uint
	// DatacenterID is the identifier of the datacenter object.
	DatacenterID() DatacenterID
}

// Network is the interface defining the fields for networks.
type Network interface {
	NetworkData

	// Datacenter fetches the datacenter associated with this network. This is a network call and may be slow.
	Datacenter(retries ...RetryStrategy) (Datacenter, error)
	// Update updates the current network.
	Update(params UpdateNetworkParams, retries ...RetryStrategy) (Network, error)
	// Remove removes the current network.
	Remove(retries ...RetryStrategy) error
}

// CreateNetworkParams contains the optional parameters for network creation.
type CreateNetworkParams interface {
	Description() *string
	Comment() *string
	VLAN() *uint
}

// BuildableNetworkParams is a buildable version of CreateNetworkParams.
type BuildableNetworkParams interface {
	CreateNetworkParams

	WithDescription(description string) (BuildableNetworkParams, error)
	MustWithDescription(description string) BuildableNetworkParams
	WithComment(comment string) (BuildableNetworkParams, error)
	MustWithComment(comment string) BuildableNetworkParams
	WithVLAN(vlan uint) (BuildableNetworkParams, error)
	MustWithVLAN(vlan uint) BuildableNetworkParams
}

// NewCreateNetworkParams creates a buildable set of CreateNetworkParams to pass to CreateNetwork.
func NewCreateNetworkParams() BuildableNetworkParams {
	return &createNetworkParams{}
}

type createNetworkParams struct {
	description *string
	comment     *string
	vlan        *uint
}

func (c *createNetworkParams) Description() *string {
	return c.description
}

func (c *createNetworkParams) Comment() *string {
	return c.comment
}

func (c *createNetworkParams) VLAN() *uint {
	return c.vlan
}

func (c *createNetworkParams) WithDescription(description string) (BuildableNetworkParams, error) {
	c.description = &description
	return c, nil
}

func (c *createNetworkParams) MustWithDescription(description string) BuildableNetworkParams {
	b, err := c.WithDescription(description)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createNetworkParams) WithComment(comment string) (BuildableNetworkParams, error) {
	c.comment = &comment
	return c, nil
}

func (c *createNetworkParams) MustWithComment(comment string) BuildableNetworkParams {
	b, err := c.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createNetworkParams) WithVLAN(vlan uint) (BuildableNetworkParams, error) {
	if vlan < 1 || vlan > 4094 {
		return nil, newError(EBadArgument, "VLAN ID must be between 1 and 4094")
	}
	c.vlan = &vlan
	return c, nil
}

func (c *createNetworkParams) MustWithVLAN(vlan uint) BuildableNetworkParams {
	b, err := c.WithVLAN(vlan)
	if err != nil {
		panic(err)
	}
	return b
}

// UpdateNetworkParams contains the parameters for updating a network.
type UpdateNetworkParams interface {
	Name() *string
	Description() *string
	Comment() *string
	VLAN() *uint
}

// BuildableUpdateNetworkParams is a buildable version of UpdateNetworkParams.
type BuildableUpdateNetworkParams interface {
	UpdateNetworkParams

	WithName(name string) (BuildableUpdateNetworkParams, error)
	MustWithName(name string) BuildableUpdateNetworkParams
	WithDescription(description string) (BuildableUpdateNetworkParams, error)
	MustWithDescription(description string) BuildableUpdateNetworkParams
	WithComment(comment string) (BuildableUpdateNetworkParams, error)
	MustWithComment(comment string) BuildableUpdateNetworkParams
	WithVLAN(vlan uint) (BuildableUpdateNetworkParams, error)
	MustWithVLAN(vlan uint) BuildableUpdateNetworkParams
}

// NewUpdateNetworkParams creates a buildable set of UpdateNetworkParams.
func NewUpdateNetworkParams() BuildableUpdateNetworkParams {
	return &updateNetworkParams{}
}

type updateNetworkParams struct {
	name        *string
	description *string
	comment     *string
	vlan        *uint
}

func (u *updateNetworkParams) Name() *string {
	return u.name
}

func (u *updateNetworkParams) Description() *string {
	return u.description
}

func (u *updateNetworkParams) Comment() *string {
	return u.comment
}

func (u *updateNetworkParams) VLAN() *uint {
	return u.vlan
}

func (u *updateNetworkParams) WithName(name string) (BuildableUpdateNetworkParams, error) {
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty")
	}
	u.name = &name
	return u, nil
}

func (u *updateNetworkParams) MustWithName(name string) BuildableUpdateNetworkParams {
	b, err := u.WithName(name)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateNetworkParams) WithDescription(description string) (BuildableUpdateNetworkParams, error) {
	u.description = &description
	return u, nil
}

func (u *updateNetworkParams) MustWithDescription(description string) BuildableUpdateNetworkParams {
	b, err := u.WithDescription(description)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateNetworkParams) WithComment(comment string) (BuildableUpdateNetworkParams, error) {
	u.comment = &comment
	return u, nil
}

func (u *updateNetworkParams) MustWithComment(comment string) BuildableUpdateNetworkParams {
	b, err := u.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateNetworkParams) WithVLAN(vlan uint) (BuildableUpdateNetworkParams, error) {
	if vlan < 1 || vlan > 4094 {
		return nil, newError(EBadArgument, "VLAN ID must be between 1 and 4094")
	}
	u.vlan = &vlan
	return u, nil
}

func (u *updateNetworkParams) MustWithVLAN(vlan uint) BuildableUpdateNetworkParams {
	b, err := u.WithVLAN(vlan)
	if err != nil {
		panic(err)
	}
	return b
}

func convertSDKNetwork(sdkObject *ovirtsdk4.Network, client Client) (Network, error) {
	id, ok := sdkObject.Id()
	if !ok {
		return nil, newFieldNotFound("network", "id")
	}
	name, ok := sdkObject.Name()
	if !ok {
		return nil, newFieldNotFound("network", "name")
	}
	dc, ok := sdkObject.DataCenter()
	if !ok {
		return nil, newFieldNotFound("network", "datacenter")
	}
	dcID, ok := dc.Id()
	if !ok {
		return nil, newFieldNotFound("datacenter on network", "ID")
	}
	var description *string
	if desc, ok := sdkObject.Description(); ok {
		description = &desc
	}
	var comment *string
	if c, ok := sdkObject.Comment(); ok {
		comment = &c
	}
	var vlanTag *uint
	if vlan, ok := sdkObject.Vlan(); ok {
		if vlanID, ok := vlan.Id(); ok && vlanID > 0 {
			u := uint(vlanID)
			vlanTag = &u
		}
	}
	return &network{
		client:      client,
		id:          NetworkID(id),
		name:        name,
		description: description,
		comment:     comment,
		vlan:        vlanTag,
		dcID:        DatacenterID(dcID),
	}, nil
}

type network struct {
	client Client

	id          NetworkID
	name        string
	description *string
	comment     *string
	vlan        *uint
	dcID        DatacenterID
}

func (n network) ID() NetworkID {
	return n.id
}

func (n network) Name() string {
	return n.name
}

func (n network) Description() *string {
	return n.description
}

func (n network) Comment() *string {
	return n.comment
}

func (n network) VLAN() *uint {
	return n.vlan
}

func (n network) DatacenterID() DatacenterID {
	return n.dcID
}

func (n network) Datacenter(retries ...RetryStrategy) (Datacenter, error) {
	return n.client.GetDatacenter(n.dcID, retries...)
}

func (n network) Update(params UpdateNetworkParams, retries ...RetryStrategy) (Network, error) {
	return n.client.UpdateNetwork(n.id, params, retries...)
}

func (n network) Remove(retries ...RetryStrategy) error {
	return n.client.RemoveNetwork(n.id, retries...)
}
