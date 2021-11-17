package ovirtclient

import (
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

//go:generate go run scripts/rest.go -i "AffinityGroup" -n "affinity group"

// AffinityGroupClient describes the functions related to oVirt affinity groups.
type AffinityGroupClient interface {
	// GetAffinityGroup returns a single affinity group based on its ID.
	GetAffinityGroup(id string, retries ...RetryStrategy) (AffinityGroup, error)
	// ListAffinityGroups returns all affinity groups on the oVirt engine.
	ListAffinityGroups(retries ...RetryStrategy) ([]AffinityGroup, error)
}

// AffinityGroupData is the core of affinity group, providing only the data access functions, but not the client
// functions.
type AffinityGroupData interface {
	// ID returns the auto-generated identifier for this affinity group.
	ID() string
	// Name returns the user-give name for this affinity group.
	Name() string
	// Enforcing specifies whether the affinity group uses hard or soft enforcement of the affinity
	//applied to virtual machines that are members of that affinity group.
	Enforcing() bool
	// Positive returns whether the affinity group applies positive affinity or negative affinity
	// to virtual machines that are members of that affinity group
	Positive() bool
	// Priority returns the priority of this affinity group.
	Priority() uint64
	// VmsRule returns the affinity rule applied to virtual machines that are members
	// of this affinity group.
	VmsRule() AffinityRule
	// HostsRule returns the affinity rule applied between virtual machines and hosts
	// that are members of this affinity group.
	HostsRule() AffinityRule
}

// AffinityGroup is the interface defining the fields for an affinity group.
type AffinityGroup interface {
	AffinityGroupData
}

type affinityGroup struct {
	client Client

	id          string
	name        string
	description string
	enforcing   bool
	positive    bool
	priority    uint64
	vmsRule     AffinityRule
	hostsRule   AffinityRule
}

func convertSDKAffinityGroup(sdkObject *ovirtsdk4.AffinityGroup, client *oVirtClient) (AffinityGroup, error) {
	id, ok := sdkObject.Id()
	if !ok {
		return nil, newFieldNotFound("tag", "id")
	}
	name, ok := sdkObject.Name()
	if !ok {
		return nil, newFieldNotFound("affinityGroup", "name")
	}
	description, ok := sdkObject.Description()
	if !ok {
		return nil, newFieldNotFound("affinityGroup", "description")
	}
	enforcing, ok := sdkObject.Enforcing()
	if !ok {
		return nil, newFieldNotFound("affinityGroup", "enforcing")
	}
	positive, ok := sdkObject.Positive()
	if !ok {
		return nil, newFieldNotFound("affinityGroup", "positive")
	}
	priority, ok := sdkObject.Priority()
	if !ok {
		return nil, newFieldNotFound("affinityGroup", "priority")
	}
	vmsRuleSDK, ok := sdkObject.VmsRule()
	if !ok {
		return nil, newFieldNotFound("affinityGroup", "vmsRule")
	}
	hostsRuleSDK, ok := sdkObject.HostsRule()
	if !ok {
		return nil, newFieldNotFound("affinityGroup", "hostsRule")
	}
	return &affinityGroup{
		client:      client,
		id:          id,
		name:        name,
		description: description,
		enforcing:   enforcing,
		positive:    positive,
		priority:    priority,
		vmsRule:     vmsRule,
		hostsRule:   hostsRule,
	}, nil
}

func (a affinityGroup) ID() string {
	return a.id
}

func (a affinityGroup) Name() string {
	return a.name
}

func (a affinityGroup) Description() string {
	return a.description
}

func (a affinityGroup) Enforcing() bool {
	return a.enforcing
}

func (a affinityGroup) Positive() bool {
	return a.positive
}

func (a affinityGroup) Priority() uint64 {
	return a.priority
}

func (a affinityGroup) VmsRule() AffinityRule {
	return a.vmsRule
}

func (a affinityGroup) HostsRule() AffinityRule {
	return a.hostsRule
}

type AffinityRule interface {
	// Enabled returns whether the affinity group uses this rule or not
	Enabled() bool
	// Enforcing specifies whether the affinity group uses hard or soft enforcement of the affinity
	// applied to the resources that are controlled by this rule.
	Enforcing() bool
	// Positive returns whether the affinity group applies positive affinity or negative affinity
	// to the resources that are controlled by this rule
	Positive() bool
}

type affinityRule struct {
	enabled   bool
	enforcing bool
	positive  bool
}

func NewAffinityRule() AffinityRule {
	return &affinityRule{}
}

func (a affinityRule) Enabled() bool {
	return a.enabled
}

func (a affinityRule) Enforcing() bool {
	return a.enforcing
}

func (a affinityRule) Positive() bool {
	return a.positive
}
