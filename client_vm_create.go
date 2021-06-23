package ovirtclient

import (
	"context"
	"fmt"
	ovirtsdk "github.com/ovirt/go-ovirt"
	"github.com/pkg/errors"
	"time"
)

type VMCreator interface {
	VM() *ovirtsdk.Vm
	preCreateConfigurations() error
	create() error
	postCreateConfigurations() error
	setVmInitialization() error
	handlePlacementPolicy() error
	handleVMHugepages() error
	handleNics() error
	handleAutoPinning() error
}

type vmCreator struct {
	ctx              context.Context
	ovirtClient      *oVirtClient
	vmBuilder        *ovirtsdk.VmBuilder
	name             string
	clusterID        string
	vmInitialization VMInitialization
	cpuTopo          VMCPUTopo
	memoryMB         int64
	instanceTypeId   string
	vmType           ovirtsdk.VmType
	templateID       string
	blockDevices     []VMBlockDevice
	autoPinningPolicy ovirtsdk.AutoPinningPolicy
	hugepages        int
	vNICsIDs         []string
	vm *ovirtsdk.Vm
}

func newVmCreator(
	ctx context.Context,
	o *oVirtClient,
	name string,
	clusterID string,
	vmInitialization VMInitialization,
	cpuTopo VMCPUTopo,
	memoryMB int64,
	instanceTypeId string,
	vmType ovirtsdk.VmType,
	templateID string,
	autoPinningPolicy string,
	blockDevices []VMBlockDevice) (VMCreator, error) {
	// TODO: ADD FIELD VALIDATIONS
	if name != "" {
		return nil, errors.New(
			"VM name must be specified")
	}
	if clusterID != "" {
		return nil, errors.New(
			"VM cluster ID must be specified")
	}
	if instanceTypeId != "" {
		if memoryMB != 0 || cpuTopo != nil {
			return nil, errors.New(
				"InstanceTypeID and MemoryMB OR CPU cannot be set at the same time")
		}
	}
	if templateID == "" {
		templateID = BlankTemplateID
	}
	mappedAutoPinningPolicy := mapAutoPinningPolicy(autoPinningPolicy)
	// Add new fields
	return &vmCreator{
		ctx:              ctx,
		ovirtClient:      o,
		vmBuilder:        ovirtsdk.NewVmBuilder(),
		name:             name,
		clusterID:        clusterID,
		vmInitialization: vmInitialization,
		cpuTopo:          cpuTopo,
		memoryMB:         memoryMB,
		instanceTypeId:   instanceTypeId,
		vmType:           vmType,
		templateID:       templateID,
		blockDevices:     blockDevices,
		autoPinningPolicy: mappedAutoPinningPolicy,
	}, nil
}


func (v *vmCreator) VM() *ovirtsdk.Vm {
	return v.vm
}

// name(Required) - Name of the VM
// clusterID(Required) - Cluster which contains VM
// cpuTopo - The CPU Topology of the VM, if not specified values from the template will be used.
//			 Cannot be specified with InstanceType
// templateID - The ID of the template which will be used to create the VM. Default to Blank template.
// customScript - Custom script to pass to the VM
// hostname - Hostname of the machine
func (o *oVirtClient) CreateVM(
	ctx context.Context,
	autoPinningPolicy string,
	name string,
	clusterID string,
	vmInitialization VMInitialization,
	cpuTopo VMCPUTopo,
	memoryMB int64,
	instanceTypeId string,
	vmType ovirtsdk.VmType,
	templateID string,
	blockDevices []VMBlockDevice,
) error {
	v, err := newVmCreator(
		ctx,
		o,
		name,
		clusterID,
		vmInitialization,
		cpuTopo,
		memoryMB,
		instanceTypeId,
		vmType,
		templateID,
		autoPinningPolicy,
		blockDevices,
		)
	if err != nil {
		return errors.Wrap(err, "error setting VM creation parameters")
	}
	if err := v.preCreateConfigurations(); err != nil {
		return errors.Wrap(err, "error configuring VM")
	}
	if err := v.create(); err != nil {
		return errors.Wrap(err, "error creating VM")
	}
	if err := v.postCreateConfigurations(); err != nil {
		return errors.Wrapf(err, "error configuring VM %s", v.VM().MustId())
	}
	return nil
}

// createVM creates the configured VM on the oVirt engine and waits for it to reach DOWN state
func (v *vmCreator) create() error {
	vm, err := v.vmBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to construct VM struct")
	}

	response, err := v.ovirtClient.conn.SystemService().VmsService().Add().Vm(vm).Send()
	if err != nil {
		return errors.Wrap(err, "error creating VM")
	}
	vmID := response.MustVm().MustId()

	err = v.ovirtClient.conn.WaitForVM(vmID, ovirtsdk.VMSTATUS_DOWN, time.Minute)
	if err != nil {
		return errors.Wrap(err, "timed out waiting for the VM creation to finish")
	}
	return nil
}

func (v *vmCreator) preCreateConfigurations() error {
	cluster := ovirtsdk.NewClusterBuilder().Id(v.clusterID).MustBuild()
	template := ovirtsdk.NewTemplateBuilder().Id(v.templateID).MustBuild()

	vmBuilder := ovirtsdk.NewVmBuilder().
		Name(v.name).
		Cluster(cluster).
		Template(template)

	if v.vmInitialization != nil {
		if err := v.setVmInitialization(); err != nil {
			return err
		}
	}
	if v.vmType != "" {
		vmBuilder.Type(v.vmType)
	}
	if v.instanceTypeId != "" {
		vmBuilder.InstanceTypeBuilder(
			ovirtsdk.NewInstanceTypeBuilder().
				Id(v.instanceTypeId))
	}
	if v.cpuTopo != nil {
		vmBuilder.CpuBuilder(
			ovirtsdk.NewCpuBuilder().
				TopologyBuilder(ovirtsdk.NewCpuTopologyBuilder().
					Cores(v.cpuTopo.Cores()).
					Sockets(v.cpuTopo.Sockets()).
					Threads(v.cpuTopo.Threads())))
	}
	if v.memoryMB > 0 {
		vmBuilder.Memory(mibBytes * v.memoryMB)
	}
	if err := v.handlePlacementPolicy(); err != nil {
		return err
	}
	return nil
}

func (v *vmCreator) postCreateConfigurations() error {
	if len(v.vNICsIDs) > 0 {
		if err := v.handleNics(); err != nil {
			return err
		}
	}
	if v.autoPinningPolicy != ovirtsdk.AUTOPINNINGPOLICY_DISABLED {
		if err := v.handleAutoPinning(); err != nil {
			return err
		}
	}
	return nil
}

// All those functions needs to be under a specific struct
func (v *vmCreator) setVmInitialization() error {
	init := ovirtsdk.NewInitializationBuilder()
	if v.vmInitialization.CustomScript() != "" {
		init.CustomScript(v.vmInitialization.CustomScript())
	}
	if v.vmInitialization.HostName() != "" {
		init.CustomScript(v.vmInitialization.HostName())
	}
	i, err := init.Build()
	if err != nil {
		return err
	}
	v.vmBuilder.Initialization(i)
	return nil
}

func (v *vmCreator) handlePlacementPolicy() error {
	// if we have a policy, we need to set the pinning to all the hosts in the cluster.
	if v.autoPinningPolicy != ovirtsdk.AUTOPINNINGPOLICY_DISABLED {
		hostsInCluster, err := v.ovirtClient.listHostsInCluster(v.clusterID)
		if err != nil {
			return errors.Wrapf(
				err, "error finding hosts in cluster %s", v.clusterID)
		}
		placementPolicyBuilder := ovirtsdk.NewVmPlacementPolicyBuilder()
		placementPolicy, err := placementPolicyBuilder.Hosts(hostsInCluster).
			Affinity(ovirtsdk.VMAFFINITY_MIGRATABLE).Build()
		if err != nil {
			return errors.Wrap(err, "failed to build the placement policy of the vm")
		}
		v.vmBuilder.PlacementPolicy(placementPolicy)
	}
	return nil
}

// handleAutoPinning updates the VM after creation to set the auto pinning policy configuration.
func (v *vmCreator) handleAutoPinning() error {
	optimizeCPUSettings := v.autoPinningPolicy == ovirtsdk.AUTOPINNINGPOLICY_ADJUST
	_, err := v.ovirtClient.conn.SystemService().VmsService().VmService(v.vm.MustId()).
		AutoPinCpuAndNumaNodes().
		OptimizeCpuSettings(optimizeCPUSettings).Send()
	if err != nil {
		return errors.Errorf("failed to set the auto pinning policy on the VM!, %v", err)
	}
	return nil
}

func (v *vmCreator) handleVMHugepages() error {
	customProp, err := ovirtsdk.NewCustomPropertyBuilder().
		Name("hugepages").
		Value(fmt.Sprint(v.hugepages)).
		Build()
	if err != nil {
		return errors.Wrap(err, "error setting hugepages custom property")
	}
	v.vmBuilder.CustomPropertiesOfAny(customProp)

	return nil
}

// handleNics replaces the NICs of the VM with nics from specified VNic profiles
func (v *vmCreator) handleNics() error {
	vmService := v.ovirtClient.conn.SystemService().VmsService().VmService(v.vm.MustId())
	nicList, err := vmService.NicsService().List().Send()
	if err != nil {
		return errors.Wrap(err, "failed fetching VM network interfaces")
	}

	// remove all existing nics
	for _, n := range nicList.MustNics().Slice() {
		_, err := vmService.NicsService().NicService(n.MustId()).Remove().Send()
		if err != nil {
			return errors.Wrap(err, "failed clearing all interfaces before populating new ones")
		}
	}

	// re-add nics
	for i, vNICsID := range v.vNICsIDs {
		_, err := vmService.NicsService().Add().Nic(
			ovirtsdk.NewNicBuilder().
				Name(fmt.Sprintf("nic%d", i+1)).
				VnicProfileBuilder(ovirtsdk.NewVnicProfileBuilder().Id(vNICsID)).
				MustBuild()).
			Send()
		if err != nil {
			return errors.Wrap(err, "failed to create network interface")
		}
	}
	return nil
}

func mapAutoPinningPolicy(policy string) ovirtsdk.AutoPinningPolicy {
	switch policy {
	case "none":
		return ovirtsdk.AUTOPINNINGPOLICY_DISABLED
	case "resize_and_pin":
		return ovirtsdk.AUTOPINNINGPOLICY_ADJUST
	default:
		return ovirtsdk.AUTOPINNINGPOLICY_DISABLED
	}
}