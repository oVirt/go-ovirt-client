package ovirtclient

import (
	"sync"

	ovirtsdk "github.com/ovirt/go-ovirt"
)

//go:generate go run scripts/rest.go -i "Vm" -n "vm" -o "VM"

// VMClient includes the methods required to deal with virtual machines.
type VMClient interface {
	// CreateVM creates a virtual machine.
	CreateVM(
		clusterID string,
		templateID string,
		optional OptionalVMParameters,
		retries ...RetryStrategy,
	) (VM, error)
	// GetVM returns a single virtual machine based on an ID.
	GetVM(id string, retries ...RetryStrategy) (VM, error)
	// ListVMs returns a list of all virtual machines.
	ListVMs(retries ...RetryStrategy) ([]VM, error)
	// RemoveVM removes a virtual machine specified by id.
	RemoveVM(id string, retries ...RetryStrategy) error
}

// VMData is the core of VM providing only data access functions.
type VMData interface {
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
	// Status returns the current status of the VM.
	Status() VMStatus
}

// VM is the implementation of the virtual machine in oVirt.
type VM interface {
	VMData

	// Remove removes the current VM. This involves an API call and may be slow.
	Remove(retries ...RetryStrategy) error

	// CreateNIC creates a network interface on the current VM. This involves an API call and may be slow.
	CreateNIC(name string, vnicProfileID string, retries ...RetryStrategy) (NIC, error)
	// GetNIC fetches a NIC with a specific ID on the current VM. This involves an API call and may be slow.
	GetNIC(id string, retries ...RetryStrategy) (NIC, error)
	// ListNICs fetches a list of network interfaces attached to this VM. This involves an API call and may be slow.
	ListNICs(retries ...RetryStrategy) ([]NIC, error)

	// AttachDisk attaches a disk to this VM.
	AttachDisk(
		diskID string,
		diskInterface DiskInterface,
		params CreateDiskAttachmentOptionalParams,
		retries ...RetryStrategy,
	) (DiskAttachment, error)
	// GetDiskAttachment returns a specific disk attachment for the current VM by ID.
	GetDiskAttachment(diskAttachmentID string, retries ...RetryStrategy) (DiskAttachment, error)
	// ListDiskAttachments lists all disk attachments for the current VM.
	ListDiskAttachments(retries ...RetryStrategy) ([]DiskAttachment, error)
	// DetachDisk removes a specific disk attachment by the disk attachment ID.
	DetachDisk(
		diskAttachmentID string,
		retries ...RetryStrategy,
	) error
}

// OptionalVMParameters are a list of parameters that can be, but must not necessarily be added on VM creation. This
// interface is expected to be extended in the future.
type OptionalVMParameters interface {
	// Name returns the name for the new VM.
	Name() string
	// Comment returns the comment for the VM.
	Comment() string
}

// BuildableVMParameters is a variant of OptionalVMParameters that can be changed using the supplied
// builder functions. This is placed here for future use.
type BuildableVMParameters interface {
	OptionalVMParameters

	// WithName adds a name to the VM.
	WithName(name string) BuildableVMParameters
	// WithComment adds a commen to the VM.
	WithComment(comment string) BuildableVMParameters
}

// CreateVMParams creates a set of BuildableVMParameters that can be used to construct the optional VM parameters.
func CreateVMParams() BuildableVMParameters {
	return &vmParams{
		lock: &sync.Mutex{},
	}
}

type vmParams struct {
	lock *sync.Mutex

	name    string
	comment string
}

func (v *vmParams) WithName(name string) BuildableVMParameters {
	v.name = name
	return v
}

func (v *vmParams) WithComment(comment string) BuildableVMParameters {
	v.comment = comment
	return v
}

func (v vmParams) Name() string {
	return v.name
}

func (v vmParams) Comment() string {
	return v.comment
}

type vm struct {
	client Client

	id         string
	name       string
	comment    string
	clusterID  string
	templateID string
	status     VMStatus
}

func (v *vm) Status() VMStatus {
	return v.status
}

func (v *vm) AttachDisk(
	diskID string,
	diskInterface DiskInterface,
	params CreateDiskAttachmentOptionalParams,
	retries ...RetryStrategy,
) (DiskAttachment, error) {
	return v.client.CreateDiskAttachment(v.id, diskID, diskInterface, params, retries...)
}

func (v *vm) GetDiskAttachment(diskAttachmentID string, retries ...RetryStrategy) (DiskAttachment, error) {
	return v.client.GetDiskAttachment(v.id, diskAttachmentID, retries...)
}

func (v *vm) ListDiskAttachments(retries ...RetryStrategy) ([]DiskAttachment, error) {
	return v.client.ListDiskAttachments(v.id, retries...)
}

func (v *vm) DetachDisk(diskAttachmentID string, retries ...RetryStrategy) error {
	return v.client.RemoveDiskAttachment(v.id, diskAttachmentID, retries...)
}

func (v *vm) Remove(retries ...RetryStrategy) error {
	return v.client.RemoveVM(v.id, retries...)
}

func (v *vm) CreateNIC(name string, vnicProfileID string, retries ...RetryStrategy) (NIC, error) {
	return v.client.CreateNIC(v.id, name, vnicProfileID, retries...)
}

func (v *vm) GetNIC(id string, retries ...RetryStrategy) (NIC, error) {
	return v.client.GetNIC(v.id, id, retries...)
}

func (v *vm) ListNICs(retries ...RetryStrategy) ([]NIC, error) {
	return v.client.ListNICs(v.id, retries...)
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
	status, ok := sdkObject.Status()
	if !ok {
		return nil, newFieldNotFound("vm", "status")
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
		status:     VMStatus(status),
	}, nil
}

// VMStatus represents the status of a VM.
type VMStatus string

const (
	// VMStatusDown indicates that the VM is not running.
	VMStatusDown VMStatus = "down"
	// VMStatusImageLocked indicates that the virtual machine process is not running and there is some operation on the
	// disks of the virtual machine that prevents it from being started.
	VMStatusImageLocked VMStatus = "image_locked"
	// VMStatusMigrating indicates that the virtual machine process is running and the virtual machine is being migrated
	// from one host to another.
	VMStatusMigrating VMStatus = "migrating"
	// VMStatusNotResponding indicates that the hypervisor detected that the virtual machine is not responding.
	VMStatusNotResponding VMStatus = "not_responding"
	// VMStatusPaused indicates that the virtual machine process is running and the virtual machine is paused.
	// This may happen in two cases: when running a virtual machine is paused mode and when the virtual machine is being
	// automatically paused due to an error.
	VMStatusPaused VMStatus = "paused"
	// VMStatusPoweringDown indicates that the virtual machine process is running and it is about to stop running.
	VMStatusPoweringDown VMStatus = "powering_down"
	// VMStatusPoweringUp  indicates that the virtual machine process is running and the guest operating system is being
	// loaded. Note that if no guest-agent is installed, this status is set for a predefined period of time, that is by
	// default 60 seconds, when running a virtual machine.
	VMStatusPoweringUp VMStatus = "powering_up"
	// VMStatusRebooting indicates that the virtual machine process is running and the guest operating system is being
	// rebooted.
	VMStatusRebooting VMStatus = "reboot_in_progress"
	// VMStatusRestoringState indicates that the virtual machine process is about to run and the virtual machine is
	// going to awake from hibernation. In this status, the running state of the virtual machine is being restored.
	VMStatusRestoringState VMStatus = "restoring_state"
	// VMStatusSavingState indicates that the virtual machine process is running and the virtual machine is being
	// hibernated. In this status, the running state of the virtual machine is being saved. Note that this status does
	// not mean that the guest operating system is being hibernated.
	VMStatusSavingState VMStatus = "saving_state"
	// VMStatusSuspended indicates that the virtual machine process is not running and a running state of the virtual
	// machine was saved. This status is similar to Down, but when the VM is started in this status its saved running
	// state is restored instead of being booted using the normal procedure.
	VMStatusSuspended VMStatus = "suspended"
	// VMStatusUnassigned means an invalid status was received.
	VMStatusUnassigned VMStatus = "unassigned"
	// VMStatusUnknown indicates that the system failed to determine the status of the virtual machine.
	// The virtual machine process may be running or not running in this status.
	// For instance, when host becomes non-responsive the virtual machines that ran on it are set with this status.
	VMStatusUnknown VMStatus = "unknown"
	// VMStatusUp indicates that the virtual machine process is running and the guest operating system is loaded.
	// Note that if no guest-agent is installed, this status is set after a predefined period of time, that is by
	// default 60 seconds, when running a virtual machine.
	VMStatusUp VMStatus = "up"
	// VMStatusWaitForLaunch indicates that the virtual machine process is about to run.
	// This status is set when a request to run a virtual machine arrives to the host.
	// It is possible that the virtual machine process will fail to run.
	VMStatusWaitForLaunch VMStatus = "wait_for_launch"
)
