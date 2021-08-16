package ovirtclient

import (
	ovirtsdk "github.com/ovirt/go-ovirt"
)

//go:generate go run scripts/rest.go -i "Vm" -n "vm" -o "VM"

// VMClient includes the methods required to deal with virtual machines.
type VMClient interface {
	// CreateVM creates a virtual machine.
	CreateVM(clusterID string, name string, templateID string, retries ...RetryStrategy) (VM, error)
	// GetVM returns a single virtual machine based on an ID.
	GetVM(id string, retries ...RetryStrategy) (VM, error)
	// ListVMs returns a list of all virtual machines.
	ListVMs(retries ...RetryStrategy) ([]VM, error)
	// RemoveVM removes a virtual machine specified by id.
	RemoveVM(id string, retries ...RetryStrategy) error
}

// VM is the implementation of the virtual machine in oVirt.
type VM interface {
	// ID returns the unique identifier (UUID) of the current virtual machine.
	ID() string
	// Name is the user-defined name of the virtual machine.
	Name() string
	// Comment is the comment added to the VM.
	Comment() string
	// ClusterID returns the cluster this machine belongs to.
	ClusterID() string
	// TemplateID returns the ID of the base template for this machine.
	TemplateID() string

	// Remove removes the current VM. This involves an API call and may be slow.
	Remove(retries ...RetryStrategy) error
}

type vm struct {
	client Client

	id         string
	name       string
	comment    string
	clusterID  string
	templateID string
}

func (v *vm) Remove(retries ...RetryStrategy) error {
	return v.client.RemoveVM(v.id, retries...)
}

func (v *vm) Comment() string {
	return v.comment
}

func (v *vm) ClusterID() string {
	return v.clusterID
}

func (v *vm) TemplateID() string {
	return v.templateID
}

func (v *vm) ID() string {
	return v.id
}

func (v *vm) Name() string {
	return v.name
}

func convertSDKVM(sdkObject *ovirtsdk.Vm, client Client) (VM, error) {
	id, ok := sdkObject.Id()
	if !ok {
		return nil, newError(EFieldMissing, "id field missing from VM object")
	}
	name, ok := sdkObject.Name()
	if !ok {
		return nil, newError(EFieldMissing, "name field missing from VM object")
	}
	comment, ok := sdkObject.Comment()
	if !ok {
		return nil, newError(EFieldMissing, "comment field missing from VM object")
	}
	cluster, ok := sdkObject.Cluster()
	if !ok {
		return nil, newError(EFieldMissing, "cluster field missing from VM object")
	}
	clusterID, ok := cluster.Id()
	if !ok {
		return nil, newError(EFieldMissing, "ID field missing from cluster in VM object")
	}
	template, ok := sdkObject.Template()
	if !ok {
		return nil, newFieldNotFound("VM", "template")
	}
	templateID, ok := template.Id()
	if !ok {
		return nil, newFieldNotFound("template in VM", "template ID")
	}

	return &vm{
		id:         id,
		name:       name,
		comment:    comment,
		clusterID:  clusterID,
		client:     client,
		templateID: templateID,
	}, nil
}
