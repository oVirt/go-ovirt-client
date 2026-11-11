package ovirtclient

import (
	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

//go:generate go run scripts/rest/rest.go -i "DataCenter" -n "datacenter" -o "Datacenter" -T DatacenterID

// DatacenterID is the UUID of a datacenter.
type DatacenterID string

// QuotaMode represents the quota enforcement mode for a datacenter.
type QuotaMode string

const (
	// QuotaModeDisabled disables quota enforcement.
	QuotaModeDisabled QuotaMode = "disabled"
	// QuotaModeAudit audits quota usage without enforcing limits.
	QuotaModeAudit QuotaMode = "audit"
	// QuotaModeEnabled enforces quota limits.
	QuotaModeEnabled QuotaMode = "enabled"
)

// QuotaModeValues returns all valid QuotaMode values.
func QuotaModeValues() []QuotaMode {
	return []QuotaMode{
		QuotaModeDisabled,
		QuotaModeAudit,
		QuotaModeEnabled,
	}
}

// DatacenterClient contains the functions related to handling datacenter objects in oVirt. Datacenters bind together
// resources of an environment (clusters, storage domains).
// See https://www.ovirt.org/documentation/administration_guide/#chap-Data_Centers for details.
type DatacenterClient interface {
	// CreateDatacenter creates a new datacenter.
	CreateDatacenter(name string, params CreateDatacenterParams, retries ...RetryStrategy) (Datacenter, error)
	// GetDatacenter returns a single datacenter by its ID.
	GetDatacenter(id DatacenterID, retries ...RetryStrategy) (Datacenter, error)
	// ListDatacenters lists all datacenters in the oVirt engine.
	ListDatacenters(retries ...RetryStrategy) ([]Datacenter, error)
	// ListDatacenterClusters lists all clusters in the specified datacenter.
	ListDatacenterClusters(id DatacenterID, retries ...RetryStrategy) ([]Cluster, error)
	// UpdateDatacenter updates an existing datacenter.
	UpdateDatacenter(id DatacenterID, params UpdateDatacenterParams, retries ...RetryStrategy) (Datacenter, error)
	// RemoveDatacenter removes a datacenter.
	RemoveDatacenter(id DatacenterID, retries ...RetryStrategy) error
}

// DatacenterData is the core of a Datacenter when client functions are not required.
type DatacenterData interface {
	ID() DatacenterID
	Name() string
	Description() *string
	Comment() *string
	Local() bool
	QuotaMode() *QuotaMode
}

// Datacenter is a logical entity that defines the set of resources used in a specific environment.
// See https://www.ovirt.org/documentation/administration_guide/#chap-Data_Centers for details.
type Datacenter interface {
	DatacenterData

	// Clusters lists the clusters for this datacenter. This is a network call and may be slow.
	Clusters(retries ...RetryStrategy) ([]Cluster, error)
	// HasCluster returns true if the cluster is in the datacenter. This is a network call and may be slow.
	HasCluster(clusterID ClusterID, retries ...RetryStrategy) (bool, error)
	// Update updates the current datacenter.
	Update(params UpdateDatacenterParams, retries ...RetryStrategy) (Datacenter, error)
	// Remove removes the current datacenter.
	Remove(retries ...RetryStrategy) error
}

// CreateDatacenterParams contains the optional parameters for datacenter creation.
type CreateDatacenterParams interface {
	Description() *string
	Comment() *string
	Local() *bool
	QuotaMode() *QuotaMode
}

// BuildableDatacenterParams is a buildable version of CreateDatacenterParams.
type BuildableDatacenterParams interface {
	CreateDatacenterParams

	WithDescription(description string) (BuildableDatacenterParams, error)
	MustWithDescription(description string) BuildableDatacenterParams
	WithComment(comment string) (BuildableDatacenterParams, error)
	MustWithComment(comment string) BuildableDatacenterParams
	WithLocal(local bool) (BuildableDatacenterParams, error)
	MustWithLocal(local bool) BuildableDatacenterParams
	WithQuotaMode(quotaMode QuotaMode) (BuildableDatacenterParams, error)
	MustWithQuotaMode(quotaMode QuotaMode) BuildableDatacenterParams
}

// NewCreateDatacenterParams creates a buildable set of CreateDatacenterParams.
func NewCreateDatacenterParams() BuildableDatacenterParams {
	return &createDatacenterParams{}
}

type createDatacenterParams struct {
	description *string
	comment     *string
	local       *bool
	quotaMode   *QuotaMode
}

func (c *createDatacenterParams) Description() *string {
	return c.description
}

func (c *createDatacenterParams) Comment() *string {
	return c.comment
}

func (c *createDatacenterParams) Local() *bool {
	return c.local
}

func (c *createDatacenterParams) QuotaMode() *QuotaMode {
	return c.quotaMode
}

func (c *createDatacenterParams) WithDescription(description string) (BuildableDatacenterParams, error) {
	c.description = &description
	return c, nil
}

func (c *createDatacenterParams) MustWithDescription(description string) BuildableDatacenterParams {
	b, err := c.WithDescription(description)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createDatacenterParams) WithComment(comment string) (BuildableDatacenterParams, error) {
	c.comment = &comment
	return c, nil
}

func (c *createDatacenterParams) MustWithComment(comment string) BuildableDatacenterParams {
	b, err := c.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createDatacenterParams) WithLocal(local bool) (BuildableDatacenterParams, error) {
	c.local = &local
	return c, nil
}

func (c *createDatacenterParams) MustWithLocal(local bool) BuildableDatacenterParams {
	b, err := c.WithLocal(local)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createDatacenterParams) WithQuotaMode(quotaMode QuotaMode) (BuildableDatacenterParams, error) {
	c.quotaMode = &quotaMode
	return c, nil
}

func (c *createDatacenterParams) MustWithQuotaMode(quotaMode QuotaMode) BuildableDatacenterParams {
	b, err := c.WithQuotaMode(quotaMode)
	if err != nil {
		panic(err)
	}
	return b
}

// UpdateDatacenterParams contains parameters for updating an existing datacenter.
type UpdateDatacenterParams interface {
	Name() *string
	Description() *string
	Comment() *string
	QuotaMode() *QuotaMode
}

// BuildableUpdateDatacenterParams is a buildable version of UpdateDatacenterParams.
type BuildableUpdateDatacenterParams interface {
	UpdateDatacenterParams

	WithName(name string) (BuildableUpdateDatacenterParams, error)
	MustWithName(name string) BuildableUpdateDatacenterParams
	WithDescription(description string) (BuildableUpdateDatacenterParams, error)
	MustWithDescription(description string) BuildableUpdateDatacenterParams
	WithComment(comment string) (BuildableUpdateDatacenterParams, error)
	MustWithComment(comment string) BuildableUpdateDatacenterParams
	WithQuotaMode(quotaMode QuotaMode) (BuildableUpdateDatacenterParams, error)
	MustWithQuotaMode(quotaMode QuotaMode) BuildableUpdateDatacenterParams
}

// NewUpdateDatacenterParams creates a buildable set of UpdateDatacenterParams.
func NewUpdateDatacenterParams() BuildableUpdateDatacenterParams {
	return &updateDatacenterParams{}
}

type updateDatacenterParams struct {
	name        *string
	description *string
	comment     *string
	quotaMode   *QuotaMode
}

func (u *updateDatacenterParams) Name() *string {
	return u.name
}

func (u *updateDatacenterParams) Description() *string {
	return u.description
}

func (u *updateDatacenterParams) Comment() *string {
	return u.comment
}

func (u *updateDatacenterParams) QuotaMode() *QuotaMode {
	return u.quotaMode
}

func (u *updateDatacenterParams) WithName(name string) (BuildableUpdateDatacenterParams, error) {
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty")
	}
	u.name = &name
	return u, nil
}

func (u *updateDatacenterParams) MustWithName(name string) BuildableUpdateDatacenterParams {
	b, err := u.WithName(name)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateDatacenterParams) WithDescription(description string) (BuildableUpdateDatacenterParams, error) {
	u.description = &description
	return u, nil
}

func (u *updateDatacenterParams) MustWithDescription(description string) BuildableUpdateDatacenterParams {
	b, err := u.WithDescription(description)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateDatacenterParams) WithComment(comment string) (BuildableUpdateDatacenterParams, error) {
	u.comment = &comment
	return u, nil
}

func (u *updateDatacenterParams) MustWithComment(comment string) BuildableUpdateDatacenterParams {
	b, err := u.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateDatacenterParams) WithQuotaMode(quotaMode QuotaMode) (BuildableUpdateDatacenterParams, error) {
	u.quotaMode = &quotaMode
	return u, nil
}

func (u *updateDatacenterParams) MustWithQuotaMode(quotaMode QuotaMode) BuildableUpdateDatacenterParams {
	b, err := u.WithQuotaMode(quotaMode)
	if err != nil {
		panic(err)
	}
	return b
}

func convertSDKDatacenter(sdkObject *ovirtsdk4.DataCenter, client Client) (Datacenter, error) {
	id, ok := sdkObject.Id()
	if !ok {
		return nil, newFieldNotFound("datacenter", "id")
	}
	name, ok := sdkObject.Name()
	if !ok {
		return nil, newFieldNotFound("datacenter", "name")
	}
	var description *string
	if desc, ok := sdkObject.Description(); ok {
		description = &desc
	}
	var comment *string
	if c, ok := sdkObject.Comment(); ok {
		comment = &c
	}
	local := false
	if l, ok := sdkObject.Local(); ok {
		local = l
	}
	var quotaMode *QuotaMode
	if qm, ok := sdkObject.QuotaMode(); ok {
		q := QuotaMode(qm)
		quotaMode = &q
	}

	return &datacenter{
		client:      client,
		id:          DatacenterID(id),
		name:        name,
		description: description,
		comment:     comment,
		local:       local,
		quotaMode:   quotaMode,
	}, nil
}

type datacenter struct {
	client Client

	id          DatacenterID
	name        string
	description *string
	comment     *string
	local       bool
	quotaMode   *QuotaMode
}

func (d datacenter) Clusters(retries ...RetryStrategy) ([]Cluster, error) {
	return d.client.ListDatacenterClusters(d.id, retries...)
}

func (d datacenter) HasCluster(clusterID ClusterID, retries ...RetryStrategy) (bool, error) {
	clusters, err := d.client.ListDatacenterClusters(d.id, retries...)
	if err != nil {
		return false, err
	}
	for _, cluster := range clusters {
		if cluster.ID() == clusterID {
			return true, nil
		}
	}
	return false, nil
}

func (d datacenter) ID() DatacenterID {
	return d.id
}

func (d datacenter) Name() string {
	return d.name
}

func (d datacenter) Description() *string {
	return d.description
}

func (d datacenter) Comment() *string {
	return d.comment
}

func (d datacenter) Local() bool {
	return d.local
}

func (d datacenter) QuotaMode() *QuotaMode {
	return d.quotaMode
}

func (d datacenter) Update(params UpdateDatacenterParams, retries ...RetryStrategy) (Datacenter, error) {
	return d.client.UpdateDatacenter(d.id, params, retries...)
}

func (d datacenter) Remove(retries ...RetryStrategy) error {
	return d.client.RemoveDatacenter(d.id, retries...)
}
