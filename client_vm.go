package ovirtclient

import (
	"fmt"
	ovirtsdk "github.com/ovirt/go-ovirt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

//go:generate go run scripts/rest.go -i "Vm" -n "vm" -o "VM"

// VMClient includes the methods required to deal with virtual machines.
type VMClient interface {
	// CreateVM creates a virtual machine of the oVirt engine with the parameters provided
	// and waits for it to be created and reach Down state.
	// It doesn't start the VM and returns an error if one happened.
	CreateVM(
		clusterID string,
		templateID string,
		name string,
		optional OptionalVMParameters,
		retries ...RetryStrategy,
	) (VM, error)
	// GetVM returns a single virtual machine based on an ID.
	GetVM(id string, retries ...RetryStrategy) (VM, error)
	// UpdateVM updates the virtual machine with the given parameters.
	// Use UpdateVMParams to obtain a builder for the params.
	UpdateVM(id string, params UpdateVMParameters, retries ...RetryStrategy) (VM, error)
	// ListVMs returns a list of all virtual machines.
	ListVMs(retries ...RetryStrategy) ([]VM, error)
	// RemoveVM removes a virtual machine specified by id.
	RemoveVM(id string, retries ...RetryStrategy) error
	// WaitForStatus waits till VM reaches the desired status,
	// it returns the VM and an error if one happened
	WaitForStatus(id string, desiredStatus VMStatus, retries ...RetryStrategy) (VM, error)
	// StartVM starts an existing VM with the given ID and waits till it reaches status UP,
	// it returns an error if one happened
	StartVM(id string, retries ...RetryStrategy) (VM, error)
	// StopVM stops an existing VM with the given ID and waits till it reaches status Down,
	// it returns an error if one happened
	StopVM(id string, retries ...RetryStrategy) (VM, error)
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
	// cpu returns the VM cpu, made of (Sockets * Cores * Threads)
	CPU() CPU
	// MemoryMB returns the size of a VM's memory in MiBs.
	MemoryMB() uint64
	// VMType returns the type of the VM, values can be: desktop, server or high_performance.
	VMType() VMType
	// AutoPinningPolicy returns the auto pining policy set for the VM, values can be: none, resize_and_pin.
	AutoPinningPolicy() VMAutoPinningPolicy
	//PlacementPolicy returns the vm placement policy configuration.
	PlacementPolicy() VMPlacementPolicy
	// Hugepages returns the size of a VM's hugepages custom property in KiBs.
	Hugepages() *VMHugepages
	// GuaranteedMemoryMB returns the amount of memory, in MiBs,
	// that is guaranteed to not be drained by the balloon mechanism.
	GuaranteedMemoryMB() *uint64
	//Initialization returns the virtual machine’s initialization configuration.
	Initialization() Initialization
	//Tags defines is a list of tags which are assigned to the machine
	Tags() []string
}

// VM is the implementation of the virtual machine in oVirt.
type VM interface {
	VMData

	// Update updates the virtual machine with the given parameters. Use UpdateVMParams to
	// get a builder for the parameters.
	Update(params UpdateVMParameters, retries ...RetryStrategy) (VM, error)
	// Remove removes the current VM. This involves an API call and may be slow.
	Remove(retries ...RetryStrategy) error

	// CreateNIC creates a network interface on the current VM. This involves an API call and may be slow.
	CreateNIC(name string, vnicProfileID string, params OptionalNICParameters, retries ...RetryStrategy) (NIC, error)
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
	// AddToAffinityGroup adds the VM to an existing affinity group.
	// it returns an error if one happened
	AddToAffinityGroup(affinityGroupID string, retries ...RetryStrategy) error
	// RemoveFromAffinityGroup removes the VM to an existing affinity group.
	// it returns an error if one happened
	RemoveFromAffinityGroup(affinityGroupID string, retries ...RetryStrategy) error
	// WaitForStatus waits till VM reaches the desired status,
	// it returns an error if one happened
	WaitForStatus(desiredStatus VMStatus, retries ...RetryStrategy) error
	// Start starts the existing VM and waits till it reaches status UP,
	// it returns an error if one happened
	Start(retries ...RetryStrategy) error
	// Stop stops the existing VM and waits till it reaches status UP,
	// it returns an error if one happened
	Stop(retries ...RetryStrategy) error
}

// OptionalVMParameters are a list of parameters that can be, but must not necessarily be added on VM creation. This
// interface is expected to be extended in the future.
type OptionalVMParameters interface {
	// Comment returns the comment for the VM.
	Comment() *string
	// cpu defines the VM cpu, made of (Sockets * Cores * Threads)
	CPU() CPU
	// MemoryMB is the size of a VM's memory in MiBs.
	MemoryMB() *uint64
	// VMType represent what the virtual machine is optimized for, can be one of the following:
	// desktop - The virtual machine is intended to be used as a desktop
	// server - The virtual machine is intended to be used as a server.
	// high_performance - The virtual machine is intended to be used as a high performance virtual machine.
	VMType() *VMType
	// AutoPinningPolicy specifies if and how the auto cpu and NUMA configuration is applied, can be one of the following:
	// none - The cpu and NUMA pinning won’t be calculated.
	// resize_and_pin - The cpu and NUMA pinning will be configured by the dedicated host,
	// the VM will consume the host cpu cores(number-of-host-cores - 1), regardless of the VM cpu settings.
	AutoPinningPolicy() *VMAutoPinningPolicy
	//PlacementPolicy specifies the hosts which the VM can be schedualled on and how.
	PlacementPolicy() VMPlacementPolicy
	// Hugepages is the size of a VM's hugepages to use in KiBs.
	// Only 2048 and 1048576 supported.
	Hugepages() *VMHugepages
	// GuaranteedMemoryMB defines amount of memory, in MiBs,
	// that is guaranteed to not be drained by the balloon mechanism.
	GuaranteedMemoryMB() *uint64
	//Initialization defines the virtual machine’s initialization configuration.
	Initialization() Initialization
}

// BuildableVMParameters is a variant of OptionalVMParameters that can be changed using the supplied
// builder functions. This is placed here for future use.
type BuildableVMParameters interface {
	OptionalVMParameters

	// Copy creates a new copy of BuildableVMParameters that can be used to construct the optional VM parameters.
	Copy() BuildableVMParameters
	// WithComment adds a common to the VM.
	WithComment(comment string) (BuildableVMParameters, error)
	// MustWithComment is identical to WithComment, but panics instead of returning an error.
	MustWithComment(comment string) BuildableVMParameters

	// WithCPU sets the cpu settings of the VM.
	WithCPU(cpu CPU) (BuildableVMParameters, error)
	// MustWithCPU is identical to WithCPU, but panics instead of returning an error.
	MustWithCPU(cpu CPU) BuildableVMParameters

	// WithMemoryMB sets the memory settings of the VM.
	WithMemoryMB(memoryMB uint64) (BuildableVMParameters, error)
	// MustWithMemoryMB is identical to WithMemoryMB, but panics instead of returning an error.
	MustWithMemoryMB(memoryMB uint64) BuildableVMParameters

	// WithVMType sets the VM Type of the VM.
	WithVMType(vmType VMType) (BuildableVMParameters, error)
	// MustWithVMType is identical to WithVMType, but panics instead of returning an error.
	MustWithVMType(vmType VMType) BuildableVMParameters

	// WithAutoPinningPolicy sets the auto pinning policy of the VM.
	WithAutoPinningPolicy(autoPinningPolicy VMAutoPinningPolicy) (BuildableVMParameters, error)
	// MustWithAutoPinningPolicy is identical to WithAutoPinningPolicy, but panics instead of returning an error.
	MustWithAutoPinningPolicy(autoPinningPolicy VMAutoPinningPolicy) BuildableVMParameters

	// WithPlacementPolicy sets the placement policy of the VM.
	WithPlacementPolicy(placementPolicy VMPlacementPolicy) (BuildableVMParameters, error)
	// MustWithPlacementPolicy is identical to WithPlacementPolicy, but panics instead of returning an error.
	MustWithPlacementPolicy(placementPolicy VMPlacementPolicy) BuildableVMParameters

	// WithHugepages sets the hugepages settings of the VM.
	WithHugepages(hugepages VMHugepages) (BuildableVMParameters, error)
	// MustWithHugepages is identical to WithHugepages, but panics instead of returning an error.
	MustWithHugepages(hugepages VMHugepages) BuildableVMParameters

	// WithGuaranteedMemoryMB sets the guaranteed memory of the VM.
	WithGuaranteedMemoryMB(guaranteedMemory uint64) (BuildableVMParameters, error)
	// MustWithGuaranteedMemoryMB is identical to WithGuaranteedMemoryMB, but panics instead of returning an error.
	MustWithGuaranteedMemoryMB(guaranteedMemory uint64) BuildableVMParameters

	// WithInitialization sets the virtual machine’s initialization configuration.
	WithInitialization(initialization Initialization) (BuildableVMParameters, error)
	// MustWithInitialization is identical to WithInitialization, but panics instead of returning an error.
	MustWithInitialization(initialization Initialization) BuildableVMParameters
}

// UpdateVMParameters returns a set of parameters to change on a VM.
type UpdateVMParameters interface {
	// Name returns the name for the VM. Return nil if the name should not be changed.
	Name() *string
	// Comment returns the comment for the VM. Return nil if the name should not be changed.
	Comment() *string
}

// BuildableUpdateVMParameters is a buildable version of UpdateVMParameters.
type BuildableUpdateVMParameters interface {
	UpdateVMParameters

	// WithName adds an updated name to the request.
	WithName(name string) (BuildableUpdateVMParameters, error)

	// MustWithName is identical to WithName, but panics instead of returning an error
	MustWithName(name string) BuildableUpdateVMParameters

	// WithComment adds a comment to the request
	WithComment(comment string) (BuildableUpdateVMParameters, error)

	// MustWithComment is identical to WithComment, but panics instead of returning an error.
	MustWithComment(comment string) BuildableUpdateVMParameters
}

// UpdateVMParams returns a buildable set of update parameters.
func UpdateVMParams() BuildableUpdateVMParameters {
	return &updateVMParams{}
}

type updateVMParams struct {
	name    *string
	comment *string
}

func (u *updateVMParams) MustWithName(name string) BuildableUpdateVMParameters {
	builder, err := u.WithName(name)
	if err != nil {
		panic(err)
	}
	return builder
}

func (u *updateVMParams) MustWithComment(comment string) BuildableUpdateVMParameters {
	builder, err := u.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return builder
}

func (u *updateVMParams) Name() *string {
	return u.name
}

func (u *updateVMParams) Comment() *string {
	return u.comment
}

func (u *updateVMParams) WithName(name string) (BuildableUpdateVMParameters, error) {
	if err := validateVMName(name); err != nil {
		return nil, err
	}
	u.name = &name
	return u, nil
}

func (u *updateVMParams) WithComment(comment string) (BuildableUpdateVMParameters, error) {
	u.comment = &comment
	return u, nil
}

// CreateVMParams creates a set of BuildableVMParameters that can be used to construct the optional VM parameters.
func CreateVMParams() BuildableVMParameters {
	return &vmParams{
		lock: &sync.Mutex{},
	}
}

type vmParams struct {
	lock *sync.Mutex

	comment            *string
	cpu                CPU
	memoryMB           *uint64
	vmType             *VMType
	autoPinningPolicy  *VMAutoPinningPolicy
	placementPolicy    VMPlacementPolicy
	hugepages          *VMHugepages
	guaranteedMemoryMB *uint64
	initialization     Initialization
}

// CopyVMParams creates a new copy of params that can be used to construct the optional VM parameters.
func (v *vmParams) Copy() BuildableVMParameters {
	newparams := CreateVMParams()
	if comment := v.Comment(); comment != nil {
		newparams.WithComment(*comment)
	}
	if cpu := v.CPU(); cpu != nil {
		newparams.WithCPU(cpu)
	}
	if memoryMB := v.MemoryMB(); memoryMB != nil {
		newparams.WithMemoryMB(*memoryMB)
	}
	if vmType := v.VMType(); vmType != nil {
		newparams.WithVMType(*vmType)
	}
	if autoPinningPolicy := v.AutoPinningPolicy(); autoPinningPolicy != nil {
		newparams.WithAutoPinningPolicy(*autoPinningPolicy)
	}
	if placementPolicy := v.PlacementPolicy(); placementPolicy != nil {
		newparams.WithPlacementPolicy(placementPolicy)
	}
	if hugepages := v.Hugepages(); hugepages != nil {
		newparams.WithHugepages(*hugepages)
	}
	if guaranteedMemoryMB := v.GuaranteedMemoryMB(); guaranteedMemoryMB != nil {
		newparams.WithGuaranteedMemoryMB(*guaranteedMemoryMB)
	}
	if init := v.Initialization(); init != nil {
		newparams.WithInitialization(init)
	}
	return newparams
}

func (v *vmParams) WithComment(comment string) (BuildableVMParameters, error) {
	v.comment = &comment
	return v, nil
}

func (v *vmParams) WithCPU(cpu CPU) (BuildableVMParameters, error) {
	v.cpu = cpu
	return v, nil
}

func (v *vmParams) WithMemoryMB(memoryMB uint64) (BuildableVMParameters, error) {
	v.memoryMB = &memoryMB
	return v, nil
}

func (v *vmParams) WithVMType(vmType VMType) (BuildableVMParameters, error) {
	if err := vmType.Validate(); err != nil {
		return nil, err
	}
	v.vmType = &vmType
	return v, nil
}

func (v *vmParams) WithAutoPinningPolicy(autoPinningPolicy VMAutoPinningPolicy) (BuildableVMParameters, error) {
	if err := autoPinningPolicy.Validate(); err != nil {
		return nil, err
	}
	v.autoPinningPolicy = &autoPinningPolicy
	return v, nil
}

func (v *vmParams) WithPlacementPolicy(placementPolicy VMPlacementPolicy) (BuildableVMParameters, error) {
	if err := placementPolicy.Affinity().Validate(); err != nil {
		return nil, err
	}
	v.placementPolicy = placementPolicy
	return v, nil
}

func (v *vmParams) WithHugepages(hugepages VMHugepages) (BuildableVMParameters, error) {
	if err := hugepages.Validate(); err != nil {
		return nil, err
	}
	v.hugepages = &hugepages
	return v, nil
}

func (v *vmParams) WithGuaranteedMemoryMB(guaranteedMemory uint64) (BuildableVMParameters, error) {
	v.guaranteedMemoryMB = &guaranteedMemory
	return v, nil
}

func (v *vmParams) WithInitialization(initialization Initialization) (BuildableVMParameters, error) {
	v.initialization = initialization
	return v, nil
}

func (v *vmParams) MustWithComment(comment string) BuildableVMParameters {
	builder, err := v.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return builder
}

func (v *vmParams) MustWithCPU(cpu CPU) BuildableVMParameters {
	builder, err := v.WithCPU(cpu)
	if err != nil {
		panic(err)
	}
	return builder
}

func (v *vmParams) MustWithMemoryMB(memoryMB uint64) BuildableVMParameters {
	builder, err := v.WithMemoryMB(memoryMB)
	if err != nil {
		panic(err)
	}
	return builder
}

func (v *vmParams) MustWithVMType(vmType VMType) BuildableVMParameters {
	builder, err := v.WithVMType(vmType)
	if err != nil {
		panic(err)
	}
	return builder
}

func (v *vmParams) MustWithAutoPinningPolicy(autoPinningPolicy VMAutoPinningPolicy) BuildableVMParameters {
	builder, err := v.WithAutoPinningPolicy(autoPinningPolicy)
	if err != nil {
		panic(err)
	}
	return builder
}

func (v *vmParams) MustWithPlacementPolicy(placementPolicy VMPlacementPolicy) BuildableVMParameters {
	builder, err := v.WithPlacementPolicy(placementPolicy)
	if err != nil {
		panic(err)
	}
	return builder
}
func (v *vmParams) MustWithHugepages(hugepages VMHugepages) BuildableVMParameters {
	builder, err := v.WithHugepages(hugepages)
	if err != nil {
		panic(err)
	}
	return builder
}

func (v *vmParams) MustWithGuaranteedMemoryMB(guaranteedMemory uint64) BuildableVMParameters {
	builder, err := v.WithGuaranteedMemoryMB(guaranteedMemory)
	if err != nil {
		panic(err)
	}
	return builder
}

func (v *vmParams) MustWithInitialization(initialization Initialization) BuildableVMParameters {
	builder, err := v.WithInitialization(initialization)
	if err != nil {
		panic(err)
	}
	return builder
}

func (v vmParams) Comment() *string {
	return v.comment
}

func (v *vmParams) CPU() CPU {
	return v.cpu
}

func (v *vmParams) MemoryMB() *uint64 {
	return v.memoryMB
}

func (v *vmParams) VMType() *VMType {
	return v.vmType
}

func (v *vmParams) AutoPinningPolicy() *VMAutoPinningPolicy {
	return v.autoPinningPolicy
}

func (v *vmParams) PlacementPolicy() VMPlacementPolicy {
	return v.placementPolicy
}

func (v *vmParams) Hugepages() *VMHugepages {
	return v.hugepages
}

func (v *vmParams) GuaranteedMemoryMB() *uint64 {
	return v.guaranteedMemoryMB
}

func (v *vmParams) Initialization() Initialization {
	return v.initialization
}

type vm struct {
	client Client

	id                 string
	name               string
	comment            string
	clusterID          string
	templateID         string
	status             VMStatus
	cpu                CPU
	memoryMB           uint64
	vmType             VMType
	autoPiningPolicy   VMAutoPinningPolicy
	placementPolicy    VMPlacementPolicy
	hugepages          *VMHugepages
	guaranteedMemoryMB *uint64
	initialization     Initialization
	tags               []string
}

func convertSDKVM(sdkObject *ovirtsdk.Vm, client Client) (VM, error) {
	id, ok := sdkObject.Id()
	if !ok {
		return nil, newFieldNotFound("vm", "id")
	}
	template, ok := sdkObject.Template()
	if !ok {
		return nil, newFieldNotFound("vm", "template")
	}
	templateID, ok := template.Id()
	if !ok {
		return nil, newError(EBug, "template found with no id for VM")
	}
	cluster, ok := sdkObject.Cluster()
	if !ok {
		return nil, newFieldNotFound("vm", "cluster")
	}
	clusterID, ok := cluster.Id()
	if !ok {
		return nil, newError(EBug, "cluster found with no id for VM")
	}
	name, ok := sdkObject.Name()
	if !ok {
		return nil, newFieldNotFound("vm", "name")
	}
	comment, ok := sdkObject.Comment()
	if !ok {
		return nil, newFieldNotFound("vm", "comment")
	}
	status, ok := sdkObject.Status()
	if !ok {
		return nil, newFieldNotFound("vm", "status")
	}
	sdkCPU, ok := sdkObject.Cpu()
	if !ok {
		return nil, newFieldNotFound("vm", "cpu")
	}
	cpu, err := ConvertSDKCPU(*sdkCPU)
	if err != nil {
		return nil, err
	}
	memByte, ok := sdkObject.Memory()
	if !ok {
		return nil, newError(EFieldMissing, "memory field missing from VM object")
	}
	mem := convertByteTMib(uint64(memByte))
	vmtype, ok := sdkObject.Type()
	if !ok {
		return nil, newFieldNotFound("vm", "vmtype")
	}
	hugepages, ok := hugepagesFromVM(sdkObject)
	if ok {
		if err := hugepages.Validate(); err != nil {
			return nil, err
		}
	}
	memoryPolicy, ok := sdkObject.MemoryPolicy()
	if !ok {
		return nil, newFieldNotFound("vm", "memoryPolicy")
	}
	guaranteedMemory, ok := memoryPolicy.Guaranteed()
	if !ok {
		return nil, newFieldNotFound("vm", "guaranteedMemory")
	}
	guaranteedMemoryConverted := convertByteTMib(uint64(guaranteedMemory))
	var vmInitialization Initialization
	sdkInitialization, ok := sdkObject.Initialization()
	if ok {
		vmInitialization, err = ConvertSDKInitialization(*sdkInitialization)
		if err != nil {
			return nil, err
		}
	}
	// TODO: Extract to a seperate method
	tagsSDK, ok := sdkObject.Tags()
	if !ok {
		return nil, newFieldNotFound("vm", "tags")
	}
	var tags []string
	for _, tagSDK := range tagsSDK.Slice() {
		tagName, ok := tagSDK.Name()
		if !ok {
			return nil, newFieldNotFound("tag", "name")
		}
		tags = append(tags, tagName)
	}
	placementPolicySDK, ok := sdkObject.PlacementPolicy()
	if !ok {
		return nil, newFieldNotFound("vm", "cpu")
	}
	placementPolicy, err := ConvertSDKVmPlacementPolicy(*placementPolicySDK, client)
	if err != nil {
		return nil, err
	}
	// TODO: we always set the autopining to none, this is how the ovirt API works but not sure what should happen if the VM is auto pinned ?
	// As far as I know there is no way to extract the autopining policy from the VM object on ovirt < 4.5
	autoPiningPolicy := VMAutoPinningPolicyNone
	return &vm{
		client:             client,
		id:                 id,
		name:               name,
		comment:            comment,
		clusterID:          clusterID,
		templateID:         templateID,
		status:             VMStatus(status),
		cpu:                cpu,
		memoryMB:           mem,
		vmType:             VMType(vmtype),
		hugepages:          hugepages,
		autoPiningPolicy:   autoPiningPolicy,
		placementPolicy:    placementPolicy,
		guaranteedMemoryMB: &guaranteedMemoryConverted,
		initialization:     vmInitialization,
		tags:               tags,
	}, nil
}

func (v *vm) ID() string {
	return v.id
}

func (v *vm) Name() string {
	return v.name
}

func (v *vm) Comment() string {
	return v.comment
}

func (v *vm) ClusterID() string {
	return v.clusterID
}

func (v *vm) VMStatus() VMStatus {
	return v.status
}

func (v *vm) TemplateID() string {
	return v.templateID
}

func (v *vm) CPU() CPU {
	return v.cpu
}

func (v *vm) MemoryMB() uint64 {
	return v.memoryMB
}

func (v *vm) VMType() VMType {
	return v.vmType
}

func (v *vm) AutoPinningPolicy() VMAutoPinningPolicy {
	return v.autoPiningPolicy
}

func (v *vm) PlacementPolicy() VMPlacementPolicy {
	return v.placementPolicy
}

func (v *vm) Hugepages() *VMHugepages {
	return v.hugepages
}

func (v *vm) GuaranteedMemoryMB() *uint64 {
	return v.guaranteedMemoryMB
}

func (v *vm) Initialization() Initialization {
	return v.initialization
}

func (v *vm) Tags() []string {
	return v.tags
}

// withComment returns a copy of the VM with the new comment. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withComment(comment string) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                v.cpu,
		memoryMB:           v.memoryMB,
		vmType:             v.vmType,
		autoPiningPolicy:   v.autoPiningPolicy,
		placementPolicy:    v.placementPolicy,
		hugepages:          v.hugepages,
		guaranteedMemoryMB: v.guaranteedMemoryMB,
		initialization:     v.initialization,
		tags:               v.tags,
	}
}

// withCPU returns a copy of the VM with the new cpu. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withCPU(cpu CPU) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            v.comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                cpu,
		memoryMB:           v.memoryMB,
		vmType:             v.vmType,
		autoPiningPolicy:   v.autoPiningPolicy,
		placementPolicy:    v.placementPolicy,
		hugepages:          v.hugepages,
		guaranteedMemoryMB: v.guaranteedMemoryMB,
		initialization:     v.initialization,
		tags:               v.tags,
	}
}

// withMemoryMB returns a copy of the VM with the new memoryMB. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withMemoryMB(memoryMB uint64) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            v.comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                v.cpu,
		memoryMB:           memoryMB,
		vmType:             v.vmType,
		autoPiningPolicy:   v.autoPiningPolicy,
		placementPolicy:    v.placementPolicy,
		hugepages:          v.hugepages,
		guaranteedMemoryMB: v.guaranteedMemoryMB,
		initialization:     v.initialization,
		tags:               v.tags,
	}
}

// withVMType returns a copy of the VM with the new VMType. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withVMType(vmType VMType) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            v.comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                v.cpu,
		memoryMB:           v.memoryMB,
		vmType:             vmType,
		autoPiningPolicy:   v.autoPiningPolicy,
		placementPolicy:    v.placementPolicy,
		hugepages:          v.hugepages,
		guaranteedMemoryMB: v.guaranteedMemoryMB,
		initialization:     v.initialization,
		tags:               v.tags,
	}
}

// withAutoPinningPolicy returns a copy of the VM with the new autopining policy. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withAutoPinningPolicy(autoPiningPolicy VMAutoPinningPolicy) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            v.comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                v.cpu,
		memoryMB:           v.memoryMB,
		vmType:             v.vmType,
		autoPiningPolicy:   autoPiningPolicy,
		placementPolicy:    v.placementPolicy,
		hugepages:          v.hugepages,
		guaranteedMemoryMB: v.guaranteedMemoryMB,
		initialization:     v.initialization,
		tags:               v.tags,
	}
}

// withPlacementPolicy returns a copy of the VM with the new vm placement policy. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withPlacementPolicy(placementPolicy VMPlacementPolicy) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            v.comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                v.cpu,
		memoryMB:           v.memoryMB,
		vmType:             v.vmType,
		autoPiningPolicy:   v.autoPiningPolicy,
		placementPolicy:    placementPolicy,
		hugepages:          v.hugepages,
		guaranteedMemoryMB: v.guaranteedMemoryMB,
		initialization:     v.initialization,
		tags:               v.tags,
	}
}

// withHugepages returns a copy of the VM with the new hugepages. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withHugepages(hugepages VMHugepages) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            v.comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                v.cpu,
		memoryMB:           v.memoryMB,
		vmType:             v.vmType,
		autoPiningPolicy:   v.autoPiningPolicy,
		placementPolicy:    v.placementPolicy,
		hugepages:          &hugepages,
		guaranteedMemoryMB: v.guaranteedMemoryMB,
		initialization:     v.initialization,
		tags:               v.tags,
	}
}

// withGuaranteedMemoryMB returns a copy of the VM with the new guaranteed Memory in MiB. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withGuaranteedMemoryMB(guaranteedMemoryMB uint64) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            v.comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                v.cpu,
		memoryMB:           v.memoryMB,
		vmType:             v.vmType,
		autoPiningPolicy:   v.autoPiningPolicy,
		placementPolicy:    v.placementPolicy,
		hugepages:          v.hugepages,
		guaranteedMemoryMB: &guaranteedMemoryMB,
		initialization:     v.initialization,
		tags:               v.tags,
	}
}

// withInitialization returns a copy of the VM with the new initialization. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withInitialization(initialization Initialization) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            v.comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                v.cpu,
		memoryMB:           v.memoryMB,
		vmType:             v.vmType,
		autoPiningPolicy:   v.autoPiningPolicy,
		placementPolicy:    v.placementPolicy,
		hugepages:          v.hugepages,
		guaranteedMemoryMB: v.guaranteedMemoryMB,
		initialization:     initialization,
		tags:               v.tags,
	}
}

// withTags returns a copy of the VM with the new tags. It does not change the original copy to avoid
// shared state issues.
func (v *vm) withTags(tags []string) *vm {
	return &vm{
		client:             v.client,
		id:                 v.id,
		name:               v.name,
		comment:            v.comment,
		clusterID:          v.clusterID,
		templateID:         v.templateID,
		status:             v.status,
		cpu:                v.cpu,
		memoryMB:           v.memoryMB,
		vmType:             v.vmType,
		autoPiningPolicy:   v.autoPiningPolicy,
		placementPolicy:    v.placementPolicy,
		hugepages:          v.hugepages,
		guaranteedMemoryMB: v.guaranteedMemoryMB,
		initialization:     v.initialization,
		tags:               tags,
	}
}

func (v *vm) Update(params UpdateVMParameters, retries ...RetryStrategy) (VM, error) {
	return v.client.UpdateVM(v.id, params, retries...)
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

func (v *vm) CreateNIC(name string, vnicProfileID string, params OptionalNICParameters, retries ...RetryStrategy) (NIC, error) {
	return v.client.CreateNIC(v.id, vnicProfileID, name, params, retries...)
}

func (v *vm) GetNIC(id string, retries ...RetryStrategy) (NIC, error) {
	return v.client.GetNIC(v.id, id, retries...)
}

func (v *vm) ListNICs(retries ...RetryStrategy) ([]NIC, error) {
	return v.client.ListNICs(v.id, retries...)
}

//TODO: implement
func (v *vm) AddToAffinityGroup(affinityGroupID string, retries ...RetryStrategy) error {
	panic("implement me")
}

//TODO: implement
func (v *vm) RemoveFromAffinityGroup(affinityGroupID string, retries ...RetryStrategy) error {
	panic("implement me")
}

func (v *vm) WaitForStatus(desiredStatus VMStatus, retries ...RetryStrategy) error {
	_, err := v.client.WaitForStatus(v.id, desiredStatus, retries...)
	if err != nil {
		return wrap(err, EUnidentified, "failed")
	}
	return nil
}

func (v *vm) Start(retries ...RetryStrategy) error {
	_, err := v.client.StartVM(v.id, retries...)
	if err != nil {
		return wrap(err, EUnidentified, "failed to start VM")
	}
	return nil
}

func (v *vm) Stop(retries ...RetryStrategy) error {
	_, err := v.client.StopVM(v.id, retries...)
	if err != nil {
		return wrap(err, EUnidentified, "failed to stop VM")
	}
	return nil
}

var vmNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_\-.]*$`)

func validateVMName(name string) error {
	if !vmNameRegexp.MatchString(name) {
		return newError(EBadArgument, "invalid VM name: %s", name)
	}
	return nil
}

// VMStatus represents the status of a VM.
type VMStatus string

// Validate returns an error if the image format doesn't have a valid value.
func (s VMStatus) Validate() error {
	for _, status := range VMStatusValues() {
		if status == s {
			return nil
		}
	}
	return newError(
		EBadArgument,
		"invalid vm status: %s must be one of: %s",
		s,
		strings.Join(VMStatusValues().Strings(), ", "),
	)
}

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

// VMStatusList is a list of VMStatus.
type VMStatusList []VMStatus

// Strings creates a string list of the values.
func (l VMStatusList) Strings() []string {
	result := make([]string, len(l))
	for i, status := range l {
		result[i] = string(status)
	}
	return result
}

// VMStatusValues returns all possible VMStatus values.
func VMStatusValues() VMStatusList {
	return []VMStatus{
		VMStatusDown,
		VMStatusImageLocked,
		VMStatusMigrating,
		VMStatusNotResponding,
		VMStatusPaused,
		VMStatusPoweringDown,
		VMStatusPoweringUp,
		VMStatusRebooting,
		VMStatusRestoringState,
		VMStatusSavingState,
		VMStatusSuspended,
		VMStatusUnassigned,
		VMStatusUnknown,
		VMStatusUp,
		VMStatusWaitForLaunch,
	}
}

// VMType represent what the virtual machine is optimized for.
type VMType string

// Validate returns an error if the VM Type doesn't have a valid value.
func (t VMType) Validate() error {
	for _, vmType := range VMTypeValues() {
		if vmType == t {
			return nil
		}
	}
	return newError(
		EBadArgument,
		"invalid vm type: %s must be one of: %s",
		t,
		strings.Join(VMTypeValues().Strings(), ", "),
	)
}

const (
	// VMTypeDesktop indicates that the virtual machine is intended to be used as a desktop.
	// Currently, its implication is that a sound device will automatically be added to the virtual machine.
	VMTypeDesktop VMType = "desktop"
	// VMTypeServer indicates that the virtual machine is intended to be used as a server.
	// Currently, its implication is that a sound device will not automatically be added to the virtual machine.
	VMTypeServer VMType = "server"
	// VMTypeHighPerformance indicates that the virtual machine is intended to be used as a
	// high performance virtual machine.
	// The virtual machine configuration will automatically be set for running with the highest  possible performance,
	// and with performance metrics as close to bare metal as possible.
	// The following configuration changes are set automatically:
	//	- Enable headless mode.
	//	- Enable serial console.
	//	- Enable pass-through host cpu.
	//	- Enable I/O threads.
	//	- Enable I/O threads pinning and set the pinning topology.
	//	- Enable the paravirtualized random number generator PCI (virtio-rng) device.
	//	- Disable all USB devices.
	//	- Disable the soundcard device.
	//	- Disable the smartcard device.
	//	- Disable the memory balloon device.
	//	- Disable the watchdog device.
	//	- Disable migration.
	//	- Disable high availability.
	VMTypeHighPerformance VMType = "high_performance"
)

// VMTypeList is a list of VMType.
type VMTypeList []VMType

// Strings creates a string list of the values.
func (l VMTypeList) Strings() []string {
	result := make([]string, len(l))
	for i, vmType := range l {
		result[i] = string(vmType)
	}
	return result
}

// VMTypeValues returns all possible VMType values.
func VMTypeValues() VMTypeList {
	return []VMType{
		VMTypeDesktop,
		VMTypeServer,
		VMTypeHighPerformance,
	}
}

// VMAffinity represent if and how the vm will migrate between hosts.
type VMAffinity string

// Validate returns an error if the VM affinity policy doesn't have a valid value.
func (a VMAffinity) Validate() error {
	for _, affinity := range VMAffinityValues() {
		if affinity == a {
			return nil
		}
	}
	return newError(
		EBadArgument,
		"invalid vm affinity policy: %s must be one of: %s",
		a,
		strings.Join(VMAffinityValues().Strings(), ", "),
	)
}

const (
	// VMAffinityMigratable indicates that the VM can be migrated by the oVirt engine or the user between the allowed hosts.
	VMAffinityMigratable VMAffinity = "migratable"
	// VMAffinityUserMigratable indicates that the VM can only be migrated manually by the user between the allowed hosts.
	VMAffinityUserMigratable VMAffinity = "user_migratable"
	// VMAffinityPinned indicates that the VM can't be migrated between the allowed hosts.
	VMAffinityPinned VMAffinity = "pinned"
)

// VMAffinityList is a list of VMAffinity.
type VMAffinityList []VMAffinity

// Strings creates a string list of the values.
func (l VMAffinityList) Strings() []string {
	result := make([]string, len(l))
	for i, policy := range l {
		result[i] = string(policy)
	}
	return result
}

// VMAffinityValues returns all possible VMAffinity values.
func VMAffinityValues() VMAffinityList {
	return []VMAffinity{
		VMAffinityUserMigratable,
		VMAffinityMigratable,
		VMAffinityPinned,
	}
}

// VMAutoPinningPolicy represent if and how the auto cpu and NUMA configuration is applied.
type VMAutoPinningPolicy string

// Validate returns an error if the VM auto pinning policy doesn't have a valid value.
func (p VMAutoPinningPolicy) Validate() error {
	for _, policy := range VMAutoPinningPolicyValues() {
		if policy == p {
			return nil
		}
	}
	return newError(
		EBadArgument,
		"invalid vm auto pinning policy: %s must be one of: %s",
		p,
		strings.Join(VMAutoPinningPolicyValues().Strings(), ", "),
	)
}

const (
	// VMAutoPinningPolicyNone indicates that the cpu and NUMA pinning won’t be calculated.
	VMAutoPinningPolicyNone VMAutoPinningPolicy = "none"
	// VMAutoPinningPolicyResizeAndPin indicates that the cpu and NUMA pinning will be configured by the dedicated host.
	VMAutoPinningPolicyResizeAndPin VMAutoPinningPolicy = "resize_and_pin"
)

// VMAutoPinningPolicyList is a list of VMAutoPinningPolicy.
type VMAutoPinningPolicyList []VMAutoPinningPolicy

// Strings creates a string list of the values.
func (l VMAutoPinningPolicyList) Strings() []string {
	result := make([]string, len(l))
	for i, policy := range l {
		result[i] = string(policy)
	}
	return result
}

// VMAutoPinningPolicyValues returns all possible VMAutoPinningPolicy values.
func VMAutoPinningPolicyValues() VMAutoPinningPolicyList {
	return []VMAutoPinningPolicy{
		VMAutoPinningPolicyNone,
		VMAutoPinningPolicyResizeAndPin,
	}
}

// VMHugepages represent the size of a VM's hugepages custom property in KiBs
type VMHugepages uint64

// Validate returns an error if the VM hugepages doesn't have a valid value.
func (h VMHugepages) Validate() error {
	for _, hugepages := range VMHugepagesValues() {
		if hugepages == h {
			return nil
		}
	}
	return newError(
		EBadArgument,
		"invalid vm hugepages: %s must be one of: %s",
		h,
		strings.Join(VMHugepagesValues().Strings(), ", "),
	)
}

// ConvertToCustomProp returns an ovirt SDK custom property which contains the hugepages settings.
func (h VMHugepages) ConvertToCustomProp() (*ovirtsdk.CustomProperty, error) {
	customProp, err := ovirtsdk.NewCustomPropertyBuilder().
		Name("hugepages").
		Value(fmt.Sprint(h)).
		Build()
	if err != nil {
		return nil, newError(EBug, "failed building custom property hugepages")
	}
	return customProp, nil
}

const (
	// VMHugePages2M represents the small value of supported hugepages setting which is 2048 Kib.
	VMHugePages2M VMHugepages = 2048
	// VMHugePages1G represents the small value of supported hugepages setting which is 1048576 Kib.
	VMHugePages1G VMHugepages = 1048576
)

// VMHugepagesList is a list of VMHugepages.
type VMHugepagesList []VMHugepages

// Strings creates a string list of the values.
func (l VMHugepagesList) Strings() []string {
	result := make([]string, len(l))
	for i, hugepage := range l {
		result[i] = fmt.Sprint(hugepage)
	}
	return result
}

// VMHugepagesValues returns all possible VMHugepages values.
func VMHugepagesValues() VMHugepagesList {
	return []VMHugepages{
		VMHugePages2M,
		VMHugePages1G,
	}
}

func hugepagesFromVM(vm *ovirtsdk.Vm) (*VMHugepages, bool) {
	var hugepagesVal string
	customProperties, ok := vm.CustomProperties()
	if !ok {
		return nil, false
	}
	for _, c := range customProperties.Slice() {
		customPropertieName, ok := c.Name()
		if !ok {
			return nil, false
		}
		if customPropertieName == "hugepages" {
			hugepagesVal, ok = c.Value()
			if !ok {
				return nil, false
			}
			break
		}
	}
	hugepagesUint, err := strconv.ParseUint(hugepagesVal, 10, 64)
	if err != nil {
		return nil, false
	}
	hugepages := VMHugepages(hugepagesUint)
	return &hugepages, true
}

type VMPlacementPolicy interface {
	// WithHosts sets the selected hosts which the vm can be scheduled on.
	WithHosts(hosts []Host) VMPlacementPolicy
	// WithVmAffinity sets the VM affinity of the VM placement policy.
	WithVmAffinity(affinity VMAffinity) VMPlacementPolicy
	//Hosts returns the Host of the VM placement policy.
	Hosts() []Host
	//Affinity returns the VM affinity of the VM placement policy.
	Affinity() VMAffinity
	//ConvertToSDK converts VM placement policy to the oVirt SDK object of a VmPlacementPolicy.
	ConvertToSDK() (*ovirtsdk.VmPlacementPolicy, error)
}

type vmPlacementPolicy struct {
	hosts    []Host
	affinity *VMAffinity
}

func NewVmPlacementPolicy() VMPlacementPolicy {
	return &vmPlacementPolicy{}
}

func (v *vmPlacementPolicy) WithHosts(hosts []Host) VMPlacementPolicy {
	v.hosts = hosts
	return v
}

func (v *vmPlacementPolicy) WithVmAffinity(affinity VMAffinity) VMPlacementPolicy {
	v.affinity = &affinity
	return v
}

func (v *vmPlacementPolicy) Hosts() []Host {
	return v.hosts
}

func (v *vmPlacementPolicy) Affinity() VMAffinity {
	return *v.affinity
}

func (v *vmPlacementPolicy) ConvertToSDK() (*ovirtsdk.VmPlacementPolicy, error) {
	vmPlacementPolicyBuilder := ovirtsdk.NewVmPlacementPolicyBuilder()
	if len(v.hosts) > 0 {
		sdkHosts := make([]*ovirtsdk.Host, len(v.hosts))
		for _, h := range v.hosts {
			clientHost, ok := h.(host)
			if !ok {
				return nil, newError(EBug, "error converting Host interface to host type")
			}
			sdkHost, err := clientHost.convertToSDK()
			if err != nil {
				return nil, wrap(err, EUnidentified, "error converting host to SDK host")
			}
			sdkHosts = append(sdkHosts, sdkHost)
		}
		hostSlice := ovirtsdk.HostSlice{}
		hostSlice.SetSlice(sdkHosts)
		vmPlacementPolicyBuilder.Hosts(&hostSlice)
	}
	if v.affinity != nil {
		vmPlacementPolicyBuilder.Affinity(ovirtsdk.VmAffinity(*v.affinity))
	}
	placementPolicy, err := vmPlacementPolicyBuilder.Build()
	if err != nil {
		return nil, newError(EBug, "failed to build vm placement policy")
	}
	return placementPolicy, nil
}

func ConvertSDKVmPlacementPolicy(vmPlacementPolicy ovirtsdk.VmPlacementPolicy, client Client) (VMPlacementPolicy, error) {
	placementPolicy := NewVmPlacementPolicy()
	sdkHosts, ok := vmPlacementPolicy.Hosts()
	if ok {
		hosts := make([]Host, len(sdkHosts.Slice()))
		for _, h := range sdkHosts.Slice() {
			host, err := convertSDKHost(h, client)
			if err != nil {
				return nil, wrap(err, EUnidentified, "failed convertign host to SDK host")
			}
			hosts = append(hosts, host)
		}
		placementPolicy.WithHosts(hosts)
	}
	affinity, ok := vmPlacementPolicy.Affinity()
	if ok {
		placementPolicy.WithVmAffinity(VMAffinity(affinity))
	}
	return placementPolicy, nil
}

type CPU interface {
	//WithSockets sets the amount of sockets for the CPU
	WithSockets(sockets uint64) CPU
	//WithCores sets the amount of cores for the CPU
	WithCores(cores uint64) CPU
	//WithThreads sets the amount of threads for the CPU
	WithThreads(threads uint64) CPU
	//ConvertToSDK converts CPU to the oVirt SDK object of a CPU.
	ConvertToSDK() (*ovirtsdk.Cpu, error)
}

// cpu defines the VM cpu, made of (Sockets * Cores * Threads)
type cpu struct {
	// Sockets is the number of sockets for a VM.
	sockets uint64
	// Cores is the number of cores per socket.
	cores uint64
	// Thread is the number of thread per core.
	threads uint64
}

func NewCPU() CPU {
	return &cpu{
		sockets: 1,
		cores:   1,
		threads: 1,
	}
}

func (c *cpu) WithSockets(sockets uint64) CPU {
	c.sockets = sockets
	return c
}

func (c *cpu) WithCores(cores uint64) CPU {
	c.cores = cores
	return c
}

func (c *cpu) WithThreads(threads uint64) CPU {
	c.threads = threads
	return c
}

func (c *cpu) ConvertToSDK() (*ovirtsdk.Cpu, error) {
	cpuBuilder := ovirtsdk.NewCpuBuilder()
	cpuTopologyBuilder := ovirtsdk.NewCpuTopologyBuilder()
	cpuTopologyBuilder.Cores(int64(c.cores))
	cpuTopologyBuilder.Sockets(int64(c.sockets))
	cpuTopologyBuilder.Threads(int64(c.threads))
	cpuTopology, err := cpuTopologyBuilder.Build()
	if err != nil {
		return nil, newError(EBug, "failed to build cpu topology")
	}
	cpuBuilder.Topology(cpuTopology)
	cpu, err := cpuBuilder.Build()
	if err != nil {
		return nil, newError(EBug, "failed to build cpu")
	}
	return cpu, nil
}

func ConvertSDKCPU(cpu ovirtsdk.Cpu) (CPU, error) {
	c := NewCPU()
	topology, ok := cpu.Topology()
	if !ok {
		return nil, newError(EBug, "cpu with not topology set")
	}
	sockets, ok := topology.Sockets()
	if !ok {
		return nil, newError(EBug, "topology with not cores set")
	}
	c.WithSockets(uint64(sockets))
	cores, ok := topology.Cores()
	if !ok {
		return nil, newError(EBug, "topology with not cores set")
	}
	c.WithCores(uint64(cores))
	threads, ok := topology.Threads()
	if !ok {
		return nil, newError(EBug, "topology with not cores set")
	}
	c.WithThreads(uint64(threads))
	return c, nil
}

type Initialization interface {
	//WithCustomScript sets the customScript to run when the VM initializes
	WithCustomScript(customScript string) Initialization
	//WithHostname sets the hostname of the VM
	WithHostname(hostname string) Initialization
	//ConvertToSDK converts CPU to the oVirt SDK object of a CPU.
	ConvertToSDK() (*ovirtsdk.Initialization, error)
}

// initialization defines to the virtual machine’s initialization configuration.
// customScript - Cloud-init script which will be executed on Virtual Machine when deployed.
// hostname - Hostname to be set to Virtual Machine when deployed.
type initialization struct {
	customScript *string
	hostname     *string
}

func NewInitialization() Initialization {
	return &initialization{}
}

func (i *initialization) WithCustomScript(customScript string) Initialization {
	i.customScript = &customScript
	return i
}

func (i *initialization) WithHostname(hostname string) Initialization {
	i.hostname = &hostname
	return i
}

func (i *initialization) ConvertToSDK() (*ovirtsdk.Initialization, error) {
	initBuilder := ovirtsdk.NewInitializationBuilder()
	if i.customScript != nil {
		initBuilder.CustomScript(*i.customScript)
	}
	if i.hostname != nil {
		initBuilder.HostName(*i.hostname)
	}
	init, err := initBuilder.Build()
	if err != nil {
		return nil, newError(EBug, "failed to build vm initialization")
	}
	return init, nil
}

func ConvertSDKInitialization(initialization ovirtsdk.Initialization) (Initialization, error) {
	init := NewInitialization()
	customScript, ok := initialization.CustomScript()
	if ok {
		init.WithCustomScript(customScript)
	}
	hostname, ok := initialization.HostName()
	if ok {
		init.WithHostname(hostname)
	}
	return init, nil
}

func convertMibToByte(mibValue uint64) uint64 {
	return uint64(1048576) * mibValue
}

func convertByteTMib(byteValue uint64) uint64 {
	return byteValue / uint64(1048576)
}
