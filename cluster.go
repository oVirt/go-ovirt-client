package ovirtclient

import (
	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

//go:generate go run scripts/rest/rest.go -i "Cluster" -n "cluster" -T ClusterID

// ClusterClient is a part of the Client that deals with clusters in the oVirt Engine. A cluster is a logical grouping
// of hosts that share the same storage domains and have the same type of CPU (either Intel or AMD). If the hosts have
// different generations of CPU models, they use only the features present in all models.
//
// See https://www.ovirt.org/documentation/administration_guide/#chap-Clusters for details.
type ClusterClient interface {
	// CreateCluster creates a new cluster in the specified datacenter.
	CreateCluster(
		datacenterID DatacenterID,
		name string,
		cpuType string,
		params CreateClusterParams,
		retries ...RetryStrategy,
	) (Cluster, error)
	// ListClusters returns a list of all clusters in the oVirt engine.
	ListClusters(retries ...RetryStrategy) ([]Cluster, error)
	// GetCluster returns a specific cluster based on the cluster ID. An error is returned if the cluster doesn't exist.
	GetCluster(id ClusterID, retries ...RetryStrategy) (Cluster, error)
	// UpdateCluster updates an existing cluster.
	UpdateCluster(id ClusterID, params UpdateClusterParams, retries ...RetryStrategy) (Cluster, error)
	// RemoveCluster removes a cluster by its ID.
	RemoveCluster(id ClusterID, retries ...RetryStrategy) error
}

// ClusterID is an identifier for a cluster.
type ClusterID string

// ClusterData is the core of a Cluster when client functions are not required.
type ClusterData interface {
	ID() ClusterID
	Name() string
	DatacenterID() DatacenterID
	Description() *string
	Comment() *string
	CPUType() string
	CPUArchitecture() *string
	Ballooning() *bool
	MemoryOverCommitPercent() *int64
}

// Cluster represents a cluster returned from a ListClusters or GetCluster call.
type Cluster interface {
	ClusterData

	Update(params UpdateClusterParams, retries ...RetryStrategy) (Cluster, error)
	Remove(retries ...RetryStrategy) error
}

// CreateClusterParams contains optional parameters for cluster creation.
type CreateClusterParams interface {
	Description() *string
	Comment() *string
	CPUArchitecture() *string
	Ballooning() *bool
	MemoryOverCommitPercent() *int64
}

// BuildableClusterParams is a buildable version of CreateClusterParams.
type BuildableClusterParams interface {
	CreateClusterParams

	WithDescription(description string) (BuildableClusterParams, error)
	MustWithDescription(description string) BuildableClusterParams
	WithComment(comment string) (BuildableClusterParams, error)
	MustWithComment(comment string) BuildableClusterParams
	WithCPUArchitecture(cpuArch string) (BuildableClusterParams, error)
	MustWithCPUArchitecture(cpuArch string) BuildableClusterParams
	WithBallooning(ballooning bool) (BuildableClusterParams, error)
	MustWithBallooning(ballooning bool) BuildableClusterParams
	WithMemoryOverCommitPercent(percent int64) (BuildableClusterParams, error)
	MustWithMemoryOverCommitPercent(percent int64) BuildableClusterParams
}

// NewCreateClusterParams creates a buildable set of CreateClusterParams.
func NewCreateClusterParams() BuildableClusterParams {
	return &createClusterParams{}
}

type createClusterParams struct {
	description             *string
	comment                 *string
	cpuArch                 *string
	ballooning              *bool
	memoryOverCommitPercent *int64
}

func (c *createClusterParams) Description() *string {
	return c.description
}

func (c *createClusterParams) Comment() *string {
	return c.comment
}

func (c *createClusterParams) CPUArchitecture() *string {
	return c.cpuArch
}

func (c *createClusterParams) Ballooning() *bool {
	return c.ballooning
}

func (c *createClusterParams) MemoryOverCommitPercent() *int64 {
	return c.memoryOverCommitPercent
}

func (c *createClusterParams) WithDescription(description string) (BuildableClusterParams, error) {
	c.description = &description
	return c, nil
}

func (c *createClusterParams) MustWithDescription(description string) BuildableClusterParams {
	b, err := c.WithDescription(description)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createClusterParams) WithComment(comment string) (BuildableClusterParams, error) {
	c.comment = &comment
	return c, nil
}

func (c *createClusterParams) MustWithComment(comment string) BuildableClusterParams {
	b, err := c.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createClusterParams) WithCPUArchitecture(cpuArch string) (BuildableClusterParams, error) {
	c.cpuArch = &cpuArch
	return c, nil
}

func (c *createClusterParams) MustWithCPUArchitecture(cpuArch string) BuildableClusterParams {
	b, err := c.WithCPUArchitecture(cpuArch)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createClusterParams) WithBallooning(ballooning bool) (BuildableClusterParams, error) {
	c.ballooning = &ballooning
	return c, nil
}

func (c *createClusterParams) MustWithBallooning(ballooning bool) BuildableClusterParams {
	b, err := c.WithBallooning(ballooning)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createClusterParams) WithMemoryOverCommitPercent(percent int64) (BuildableClusterParams, error) {
	if percent < 0 {
		return nil, newError(EBadArgument, "memory overcommit percent must not be negative")
	}
	c.memoryOverCommitPercent = &percent
	return c, nil
}

func (c *createClusterParams) MustWithMemoryOverCommitPercent(percent int64) BuildableClusterParams {
	b, err := c.WithMemoryOverCommitPercent(percent)
	if err != nil {
		panic(err)
	}
	return b
}

// UpdateClusterParams contains parameters for updating an existing cluster.
type UpdateClusterParams interface {
	Name() *string
	Description() *string
	Comment() *string
	CPUType() *string
	CPUArchitecture() *string
	Ballooning() *bool
	MemoryOverCommitPercent() *int64
}

// BuildableUpdateClusterParams is a buildable version of UpdateClusterParams.
type BuildableUpdateClusterParams interface {
	UpdateClusterParams

	WithName(name string) (BuildableUpdateClusterParams, error)
	MustWithName(name string) BuildableUpdateClusterParams
	WithDescription(description string) (BuildableUpdateClusterParams, error)
	MustWithDescription(description string) BuildableUpdateClusterParams
	WithComment(comment string) (BuildableUpdateClusterParams, error)
	MustWithComment(comment string) BuildableUpdateClusterParams
	WithCPUType(cpuType string) (BuildableUpdateClusterParams, error)
	MustWithCPUType(cpuType string) BuildableUpdateClusterParams
	WithCPUArchitecture(cpuArch string) (BuildableUpdateClusterParams, error)
	MustWithCPUArchitecture(cpuArch string) BuildableUpdateClusterParams
	WithBallooning(ballooning bool) (BuildableUpdateClusterParams, error)
	MustWithBallooning(ballooning bool) BuildableUpdateClusterParams
	WithMemoryOverCommitPercent(percent int64) (BuildableUpdateClusterParams, error)
	MustWithMemoryOverCommitPercent(percent int64) BuildableUpdateClusterParams
}

// NewUpdateClusterParams creates a buildable set of UpdateClusterParams.
func NewUpdateClusterParams() BuildableUpdateClusterParams {
	return &updateClusterParams{}
}

type updateClusterParams struct {
	name                    *string
	description             *string
	comment                 *string
	cpuType                 *string
	cpuArch                 *string
	ballooning              *bool
	memoryOverCommitPercent *int64
}

func (u *updateClusterParams) Name() *string {
	return u.name
}

func (u *updateClusterParams) Description() *string {
	return u.description
}

func (u *updateClusterParams) Comment() *string {
	return u.comment
}

func (u *updateClusterParams) CPUType() *string {
	return u.cpuType
}

func (u *updateClusterParams) CPUArchitecture() *string {
	return u.cpuArch
}

func (u *updateClusterParams) Ballooning() *bool {
	return u.ballooning
}

func (u *updateClusterParams) MemoryOverCommitPercent() *int64 {
	return u.memoryOverCommitPercent
}

func (u *updateClusterParams) WithName(name string) (BuildableUpdateClusterParams, error) {
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty")
	}
	u.name = &name
	return u, nil
}

func (u *updateClusterParams) MustWithName(name string) BuildableUpdateClusterParams {
	b, err := u.WithName(name)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateClusterParams) WithDescription(description string) (BuildableUpdateClusterParams, error) {
	u.description = &description
	return u, nil
}

func (u *updateClusterParams) MustWithDescription(description string) BuildableUpdateClusterParams {
	b, err := u.WithDescription(description)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateClusterParams) WithComment(comment string) (BuildableUpdateClusterParams, error) {
	u.comment = &comment
	return u, nil
}

func (u *updateClusterParams) MustWithComment(comment string) BuildableUpdateClusterParams {
	b, err := u.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateClusterParams) WithCPUType(cpuType string) (BuildableUpdateClusterParams, error) {
	if cpuType == "" {
		return nil, newError(EBadArgument, "cpuType cannot be empty")
	}
	u.cpuType = &cpuType
	return u, nil
}

func (u *updateClusterParams) MustWithCPUType(cpuType string) BuildableUpdateClusterParams {
	b, err := u.WithCPUType(cpuType)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateClusterParams) WithCPUArchitecture(cpuArch string) (BuildableUpdateClusterParams, error) {
	u.cpuArch = &cpuArch
	return u, nil
}

func (u *updateClusterParams) MustWithCPUArchitecture(cpuArch string) BuildableUpdateClusterParams {
	b, err := u.WithCPUArchitecture(cpuArch)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateClusterParams) WithBallooning(ballooning bool) (BuildableUpdateClusterParams, error) {
	u.ballooning = &ballooning
	return u, nil
}

func (u *updateClusterParams) MustWithBallooning(ballooning bool) BuildableUpdateClusterParams {
	b, err := u.WithBallooning(ballooning)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateClusterParams) WithMemoryOverCommitPercent(percent int64) (BuildableUpdateClusterParams, error) {
	if percent < 0 {
		return nil, newError(EBadArgument, "memory overcommit percent must not be negative")
	}
	u.memoryOverCommitPercent = &percent
	return u, nil
}

func (u *updateClusterParams) MustWithMemoryOverCommitPercent(percent int64) BuildableUpdateClusterParams {
	b, err := u.WithMemoryOverCommitPercent(percent)
	if err != nil {
		panic(err)
	}
	return b
}

func convertSDKCluster(sdkCluster *ovirtsdk4.Cluster, client Client) (Cluster, error) {
	id, ok := sdkCluster.Id()
	if !ok {
		return nil, newError(EFieldMissing, "failed to fetch ID for cluster")
	}

	name, ok := sdkCluster.Name()
	if !ok {
		return nil, newError(EFieldMissing, "failed to fetch name for cluster %s", id)
	}

	var dcID DatacenterID
	if dc, ok := sdkCluster.DataCenter(); ok {
		if rawID, ok := dc.Id(); ok {
			dcID = DatacenterID(rawID)
		}
	}

	var description *string
	if d, ok := sdkCluster.Description(); ok {
		description = &d
	}

	var comment *string
	if c, ok := sdkCluster.Comment(); ok {
		comment = &c
	}

	var cpuType string
	var cpuArch *string
	if cpu, ok := sdkCluster.Cpu(); ok {
		if t, ok := cpu.Type(); ok {
			cpuType = t
		}
		if arch, ok := cpu.Architecture(); ok {
			a := string(arch)
			cpuArch = &a
		}
	}

	var ballooning *bool
	if b, ok := sdkCluster.BallooningEnabled(); ok {
		ballooning = &b
	}

	var memOverCommit *int64
	if mp, ok := sdkCluster.MemoryPolicy(); ok {
		if oc, ok := mp.OverCommit(); ok {
			if pct, ok := oc.Percent(); ok {
				memOverCommit = &pct
			}
		}
	}

	return &cluster{
		client:                  client,
		id:                      ClusterID(id),
		name:                    name,
		dcID:                    dcID,
		description:             description,
		comment:                 comment,
		cpuType:                 cpuType,
		cpuArch:                 cpuArch,
		ballooning:              ballooning,
		memoryOverCommitPercent: memOverCommit,
	}, nil
}

type cluster struct {
	client Client

	id                      ClusterID
	name                    string
	dcID                    DatacenterID
	description             *string
	comment                 *string
	cpuType                 string
	cpuArch                 *string
	ballooning              *bool
	memoryOverCommitPercent *int64
}

func (c cluster) ID() ClusterID {
	return c.id
}

func (c cluster) Name() string {
	return c.name
}

func (c cluster) DatacenterID() DatacenterID {
	return c.dcID
}

func (c cluster) Description() *string {
	return c.description
}

func (c cluster) Comment() *string {
	return c.comment
}

func (c cluster) CPUType() string {
	return c.cpuType
}

func (c cluster) CPUArchitecture() *string {
	return c.cpuArch
}

func (c cluster) Ballooning() *bool {
	return c.ballooning
}

func (c cluster) MemoryOverCommitPercent() *int64 {
	return c.memoryOverCommitPercent
}

func (c cluster) Update(params UpdateClusterParams, retries ...RetryStrategy) (Cluster, error) {
	return c.client.UpdateCluster(c.id, params, retries...)
}

func (c cluster) Remove(retries ...RetryStrategy) error {
	return c.client.RemoveCluster(c.id, retries...)
}
