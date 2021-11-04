package ovirtclient

import (
	"fmt"
	ovirtsdk "github.com/ovirt/go-ovirt"
)

func (o *oVirtClient) CreateVM(
	clusterID string,
	templateID string,
	name string,
	params OptionalVMParameters,
	retries ...RetryStrategy,
) (result VM, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts())

	if err = validateVMCreationParameters(clusterID, templateID, name, params); err != nil {
		return nil, err
	}
	if params == nil {
		params = &vmParams{}
	}
	message := fmt.Sprintf("creating VM %s", name)

	err = retry(
		message,
		o.logger,
		retries,
		func() error {
			sdkVM, err := o.createVMSDKBuilder(clusterID, templateID, name, params, retries...)
			if err != nil {
				return wrap(err, EUnidentified, "failed to build VM")
			}
			vm, err := o.addVMToEngine(sdkVM)
			if err != nil {
				return wrap(err, EUnidentified, "failed to add VM to engine")
			}
			err = vm.WaitForStatus(VMStatusDown, retries...)
			if err != nil {
				// TODO: Is this the correct ERROR?
				return newError(ETimeout, "timed out waiting for the VM creation to finish")
			}
			if autoPiningPolicy := params.AutoPinningPolicy(); autoPiningPolicy != nil {
				err = o.handleAutoPining(vm.ID(), *autoPiningPolicy)
				if err != nil {
					return wrap(err, EUnidentified, "failed setting auto pining policy for VM %s", vm.ID())
				}
			}
			if tags := params.Tags(); tags != nil {
				err = o.handleVMTags(tags, vm.ID())
				if err != nil {
					return wrap(err, EUnidentified, "failed setting tags for VM %s", vm.ID())
				}
			}
			result, err = o.GetVM(vm.ID(), retries...)
			if err != nil {
				return newError(EUnidentified, "failed getting VM after creation")
			}
			return nil
		},
	)
	return result, err
}

// TODO: Is this an incurrect usage of oVirtClient? if so should we just make it a general function and send the
//  oVirt connection as a variable? or create a custom type

func (o *oVirtClient) handleVMTags(tags []string, vmID string) error {
	for _, tag := range tags {
		_, err := o.conn.SystemService().VmsService().VmService(vmID).
			TagsService().Add().
			Tag(ovirtsdk.NewTagBuilder().Name(tag).MustBuild()).
			Send()
		if err != nil {
			return newError(EUnidentified, "failed add VM %s to tag %s", vmID, tag)
		}
	}
	return nil
}

// TODO: Is this an incurrect usage of oVirtClient? if so should we just make it a general function and send the
//  oVirt connection as a variable? or create a custom type

func (o *oVirtClient) handleAutoPining(vmID string, autoPiningPolicy VMAutoPinningPolicy) error {
	vmService := o.conn.SystemService().VmsService().VmService(vmID)
	optimizeCPUSettings := autoPiningPolicy == VMAutoPinningPolicyResizeAndPin
	_, err := vmService.AutoPinCpuAndNumaNodes().OptimizeCpuSettings(optimizeCPUSettings).Send()
	if err != nil {
		// TODO: Is this the correct ERROR?
		return newError(EUnidentified, "failed to set the auto pinning policy on the VM")
	}
	return nil
}

// TODO: Is this an incorrect usage of oVirtClient? if so should we just make it a general function and send the
//  oVirt connection as a variable? or create a custom type

func (o *oVirtClient) addVMToEngine(sdkVM *ovirtsdk.Vm) (VM, error) {
	response, err := o.conn.SystemService().VmsService().Add().Vm(sdkVM).Send()
	if err != nil {
		return nil, wrap(err, EUnidentified, "failed to create VM")
	}
	sdkVM, ok := response.Vm()
	if !ok {
		return nil, newError(EFieldMissing, "missing VM in VM create response")
	}
	vm, err := convertSDKVM(sdkVM, o)
	if err != nil {
		return nil, wrap(err, EBug, "failed to convert VM")
	}
	return vm, err
}

//TODO: Try and split

// createVMSDKBuilder returns an ovirt SDK VM object which was built but not added to the engine
func (o *oVirtClient) createVMSDKBuilder(
	clusterID string,
	templateID string,
	name string,
	params OptionalVMParameters,
	retries ...RetryStrategy,
) (*ovirtsdk.Vm, error) {
	builder := ovirtsdk.NewVmBuilder()
	builder.Cluster(ovirtsdk.NewClusterBuilder().Id(clusterID).MustBuild())
	builder.Template(ovirtsdk.NewTemplateBuilder().Id(templateID).MustBuild())
	builder.Name(name)

	if comment := params.Comment(); comment != nil {
		builder.Comment(*comment)
	}
	if cpu := params.CPU(); cpu != nil {
		cpuSDK, err := cpu.ConvertToSDK()
		if err != nil {
			return nil, err
		}
		builder.Cpu(cpuSDK)
	}
	if memory := params.MemoryMB(); memory != nil {
		builder.Memory(int64(convertMibToByte(*memory)))
	}
	if guaranteedMemoryMB := params.GuaranteedMemoryMB(); guaranteedMemoryMB != nil {
		memoryPolicyBuilder := ovirtsdk.NewMemoryPolicyBuilder()
		memoryPolicyBuilder.Guaranteed(int64(convertMibToByte(*guaranteedMemoryMB)))
		memoryPolicy, err := memoryPolicyBuilder.Build()
		if err != nil {
			return nil, newError(
				EBug, "failed building memory policy with guaranteedMemoryMB %v", *guaranteedMemoryMB)
		}
		builder.MemoryPolicy(memoryPolicy)
	}
	if vmType := params.VMType(); vmType != nil {
		builder.Type(ovirtsdk.VmType(*vmType))
	}
	if hugepages := params.Hugepages(); hugepages != nil {
		customProp, err := hugepages.ConvertToCustomProp()
		if err != nil {
			return nil, err
		}
		builder.CustomPropertiesOfAny(customProp)
	}
	if initialization := params.Initialization(); initialization != nil {
		initializationSDK, err := initialization.ConvertToSDK()
		if err != nil {
			return nil, err
		}
		builder.Initialization(initializationSDK)
	}
	if placementPolicy := params.PlacementPolicy(); placementPolicy != nil {
		placementPolicySDK, err := placementPolicy.ConvertToSDK()
		if err != nil {
			return nil, err
		}
		builder.PlacementPolicy(placementPolicySDK)
	}
	vm, err := builder.Build()
	if err != nil {
		return nil, wrap(err, EBug, "failed to build VM")
	}
	return vm, nil
}

func validateVMCreationParameters(clusterID string, templateID string, name string, params OptionalVMParameters) error {
	if clusterID == "" {
		return newError(EBadArgument, "cluster ID cannot be empty for VM creation")
	}
	if templateID == "" {
		return newError(EBadArgument, "template ID cannot be empty for VM creation")
	}
	if name == "" {
		return newError(EBadArgument, "name cannot be empty for VM creation")
	}
	if vmType := params.VMType(); vmType != nil {
		if err := vmType.Validate(); err != nil {
			return wrap(err, EBadArgument, "invalid VM type")
		}
	}
	if autoPinningPolicy := params.AutoPinningPolicy(); autoPinningPolicy != nil {
		if err := autoPinningPolicy.Validate(); err != nil {
			return wrap(err, EBadArgument, "invalid VM Auto Pinning Policy")
		}
		if params.PlacementPolicy() == nil && *autoPinningPolicy != VMAutoPinningPolicyNone {
			return newError(EBadArgument,
				"if auto pining policy is specified then you have to specify vm placement policy")
		}
	}
	if hugepages := params.Hugepages(); hugepages != nil {
		if err := hugepages.Validate(); err != nil {
			return wrap(err, EBadArgument, "invalid VM hugepages settings")
		}
	}
	if placementPolicy := params.PlacementPolicy(); placementPolicy != nil {
		if err := placementPolicy.affinity.Validate(); err != nil {
			return wrap(err, EBadArgument, "invalid VM placementPolicy settings")
		}
	}
	if guaranteedMemoryMB := params.GuaranteedMemoryMB(); guaranteedMemoryMB != nil {
		if memory := params.MemoryMB(); memory != nil {
			if *guaranteedMemoryMB > *memory {
				return newError(EBadArgument, "invalid VM guaranteedMemoryMB settings, guaranteedMemoryMB cannot exceed memory size")
			}
		}
	}
	return nil
}
