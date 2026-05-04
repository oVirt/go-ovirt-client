package ovirtclient

import (
	"fmt"
	"strconv"

	ovirtsdk "github.com/ovirt/go-ovirt"
)

func (o *oVirtClient) UpdateVM(
	id VMID,
	params UpdateVMParameters,
	retries ...RetryStrategy,
) (result VM, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))

	vm := &ovirtsdk.Vm{}
	vm.SetId(string(id))
	if name := params.Name(); name != nil {
		if *name == "" {
			return nil, newError(EBadArgument, "name must not be empty for VM update")
		}
		vm.SetName(*name)
	}
	if comment := params.Comment(); comment != nil {
		vm.SetComment(*comment)
	}
	if description := params.Description(); description != nil {
		vm.SetDescription(*description)
	}
	if cpu := params.CPU(); cpu != nil {
		cpuUpdate := &ovirtsdk.Cpu{}
		if topo := (*cpu).Topo(); topo != nil {
			cpuUpdate.SetTopology(
				ovirtsdk.NewCpuTopologyBuilder().
					Cores(int64((*cpu).Topo().Cores())).
					Threads(int64((*cpu).Topo().Threads())).
					Sockets(int64((*cpu).Topo().Sockets())).
					MustBuild(),
			)
		}
		if mode := (*cpu).Mode(); mode != nil {
			cpuUpdate.SetMode(ovirtsdk.CpuMode(*mode))
		}
		vm.SetCpu(cpuUpdate)
	}
	if hugePages := params.HugePages(); hugePages != nil {
		var customProperties []*ovirtsdk.CustomProperty
		customProp, err := ovirtsdk.NewCustomPropertyBuilder().
			Name("hugepages").
			Value(strconv.FormatUint(uint64(*hugePages), 10)).
			Build()
		if err != nil {
			panic(newError(EBug, "Failed to build 'hugepages' custom property from value %d", hugePages))
		}
		customProperties = append(customProperties, customProp)
		customPropertiesSlice := &ovirtsdk.CustomPropertySlice{}
		customPropertiesSlice.SetSlice(customProperties)
		vm.SetCustomProperties(customPropertiesSlice)
	}
	if initialization := params.Initialization(); initialization != nil {
		init := *params.Initialization()
		initBuilder := ovirtsdk.NewInitializationBuilder()

		if init.CustomScript() != "" {
			initBuilder.CustomScript(init.CustomScript())
		}
		if init.HostName() != "" {
			initBuilder.HostName(init.HostName())
		}
		if nicConf := init.NicConfiguration(); nicConf != nil {

			nicBuilder := ovirtsdk.NewNicConfigurationBuilder()
			nicBuilder.BootProtocol(ovirtsdk.BOOTPROTOCOL_STATIC)
			nicBuilder.OnBoot(true)
			nicBuilder.Name(nicConf.Name())

			ipBuilder := ovirtsdk.NewIpBuilder().
				Address(nicConf.IP().Address).
				Gateway(nicConf.IP().Gateway).
				Netmask(nicConf.IP().Netmask).
				Version(ovirtsdk.IPVERSION_V4)
			nicBuilder.Ip(ipBuilder.MustBuild())

			if nicConf.IPV6() != nil {
				ipV6Builder := ovirtsdk.NewIpBuilder().
					Address(nicConf.IPV6().Address).
					Gateway(nicConf.IPV6().Gateway).
					Netmask(nicConf.IPV6().Netmask).
					Version(ovirtsdk.IPVERSION_V6)
				nicBuilder.Ipv6(ipV6Builder.MustBuild())

			}

			initBuilder.NicConfigurationsOfAny(nicBuilder.MustBuild())
		}
		vm.SetInitialization(initBuilder.MustBuild())
	}
	if memory := params.Memory(); memory != nil {
		vm.SetMemory(*memory)
	}
	if memoryPolicyParams := params.MemoryPolicy(); memoryPolicyParams != nil {
		memPolicy := &ovirtsdk.MemoryPolicy{}
		if guaranteedMemory := (*memoryPolicyParams).Guaranteed(); guaranteedMemory != nil {
			memPolicy.SetGuaranteed(*guaranteedMemory)
		}

		if maxMemory := (*memoryPolicyParams).Max(); maxMemory != nil {
			memPolicy.SetMax(*maxMemory)
		}

		if memBallooning := (*memoryPolicyParams).Ballooning(); memBallooning != nil {
			memPolicy.SetBallooning(*memBallooning)
		}
		vm.SetMemoryPolicy(memPolicy)
	}
	if placementPolicy := params.PlacementPolicy(); placementPolicy != nil {
		placementPolicyBuilder := ovirtsdk.NewVmPlacementPolicyBuilder()
		if affinity := (*placementPolicy).Affinity(); affinity != nil {
			placementPolicyBuilder.Affinity(ovirtsdk.VmAffinity(*affinity))
		}
		hosts := make([]ovirtsdk.HostBuilder, len((*placementPolicy).HostIDs()))
		for i, hostID := range (*placementPolicy).HostIDs() {
			hostBuilder := ovirtsdk.NewHostBuilder().Id(string(hostID))
			hosts[i] = *hostBuilder
		}
		placementPolicyBuilder.HostsBuilderOfAny(hosts...)
		vm.SetPlacementPolicy(placementPolicyBuilder.MustBuild())
	}
	if instanceType := params.InstanceTypeID(); instanceType != nil {
		vm.SetInstanceType(ovirtsdk.NewInstanceTypeBuilder().Id(string(*instanceType)).MustBuild())
	}
	if vmType := params.VMType(); vmType != nil {
		vm.SetType(ovirtsdk.VmType(*vmType))
	}
	if osParams, isSet := params.OS(); isSet {
		os := &ovirtsdk.OperatingSystem{}
		if osType := (*osParams).Type(); osType != nil {
			os.SetType(*osType)
		}
		vm.SetOs(os)
	}
	if serialConsole := params.SerialConsole(); serialConsole != nil {
		vm.SetConsole(ovirtsdk.NewConsoleBuilder().Enabled(*serialConsole).MustBuild())
	}
	if soundcardEnabled := params.SoundcardEnabled(); soundcardEnabled != nil {
		vm.SetSoundcardEnabled(*soundcardEnabled)
	}

	err = retry(
		fmt.Sprintf("updating vm %s", id),
		o.logger,
		retries,
		func() error {
			response, err := o.conn.SystemService().VmsService().VmService(string(id)).Update().Vm(vm).Send()
			if err != nil {
				return wrap(err, EUnidentified, "failed to update VM")
			}
			vm, ok := response.Vm()
			if !ok {
				return newError(EFieldMissing, "missing VM in VM update response")
			}
			result, err = convertSDKVM(vm, o)
			if err != nil {
				return wrap(
					err,
					EBug,
					"failed to convert VM",
				)
			}
			return nil
		})
	return result, err
}

func (m *mockClient) UpdateVM(id VMID, params UpdateVMParameters, _ ...RetryStrategy) (VM, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.vms[id]; !ok {
		return nil, newError(ENotFound, "VM with ID %s not found", id)
	}

	vmOrig := *m.vms[id]
	vm := &vmOrig
	if name := params.Name(); name != nil {
		for _, otherVM := range m.vms {
			if otherVM.name == *name && otherVM.ID() != vm.ID() {
				return nil, newError(EConflict, "A VM with the name \"%s\" already exists.", *name)
			}
		}
		vm.name = *name
	}
	if comment := params.Comment(); comment != nil {
		vm.comment = *comment
	}
	if description := params.Description(); description != nil {
		vm.description = *description
	}
	if cpuParam := params.CPU(); cpuParam != nil {
		cpu := &vmCPU{}
		topo := (*cpuParam).Topo()
		cpu.topo = &vmCPUTopo{
			cores:   topo.Cores(),
			sockets: topo.Sockets(),
			threads: topo.Threads(),
		}
		if mode := (*cpuParam).Mode(); mode != nil {
			cpu.mode = mode
		}
		vm.cpu = cpu
	}
	if hugePages := params.HugePages(); hugePages != nil {
		vm.hugePages = hugePages
	}
	if initialization := params.Initialization(); initialization != nil {
		vm.initialization = *initialization
	}
	if memory := params.Memory(); memory != nil {
		vm.memory = *memory
	}
	if memoryPolicyParams := params.MemoryPolicy(); memoryPolicyParams != nil {
		memPolicy := &memoryPolicy{
			ballooning: true,
		}
		if guaranteedMemory := (*memoryPolicyParams).Guaranteed(); guaranteedMemory != nil {
			memPolicy.guaranteed = guaranteedMemory
		}

		if maxMemory := (*memoryPolicyParams).Max(); maxMemory != nil {
			memPolicy.max = maxMemory
		}

		if memBallooning := (*memoryPolicyParams).Ballooning(); memBallooning != nil {
			memPolicy.ballooning = *memBallooning
		}
		vm.memoryPolicy = memPolicy
	}
	if placementPolicy := params.PlacementPolicy(); placementPolicy != nil {
		pp := &vmPlacementPolicy{
			(*placementPolicy).Affinity(),
			(*placementPolicy).HostIDs(),
		}
		vm.placementPolicy = pp
	}
	if instanceType := params.InstanceTypeID(); instanceType != nil {
		vm.instanceTypeID = instanceType
	}
	if vmType := params.VMType(); vmType != nil {
		vm.vmType = *vmType
	}
	if osParams, isSet := params.OS(); isSet {
		os := &vmOS{
			t: "other",
		}
		if osType := (*osParams).Type(); osType != nil {
			os.t = *osType
		}
		vm.os = os
	}
	if serialConsole := params.SerialConsole(); serialConsole != nil {
		vm.serialConsole = *serialConsole
	}
	if soundcardEnabled := params.SoundcardEnabled(); soundcardEnabled != nil {
		vm.soundcardEnabled = *soundcardEnabled
	}
	m.vms[id] = vm

	return vm, nil
}
