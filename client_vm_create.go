package ovirtclient

import (
	"fmt"
	"strconv"

	ovirtsdk "github.com/ovirt/go-ovirt"
)

type vmBuilderComponent func(params OptionalVMParameters, builder *ovirtsdk.VmBuilder)

func vmBuilderComment(params OptionalVMParameters, builder *ovirtsdk.VmBuilder) {
	if comment := params.Comment(); comment != "" {
		builder.Comment(comment)
	}
}

func vmBuilderCPU(params OptionalVMParameters, builder *ovirtsdk.VmBuilder) {
	if cpu := params.CPU(); cpu != nil {
		builder.CpuBuilder(
			ovirtsdk.NewCpuBuilder().TopologyBuilder(
				ovirtsdk.
					NewCpuTopologyBuilder().
					Cores(int64(cpu.Cores())).
					Threads(int64(cpu.Threads())).
					Sockets(int64(cpu.Sockets())),
			))
	}
}

func vmBuilderHugePages(params OptionalVMParameters, builder *ovirtsdk.VmBuilder) {
	var customProperties []*ovirtsdk.CustomProperty
	if hugePages := params.HugePages(); hugePages != nil {
		customProp, err := ovirtsdk.NewCustomPropertyBuilder().
			Name("hugepages").
			Value(strconv.FormatUint(uint64(*hugePages), 10)).
			Build()
		if err != nil {
			panic(newError(EBug, "Failed to build 'hugepages' custom property from value %d", hugePages))
		}
		customProperties = append(customProperties, customProp)
	}
	if len(customProperties) > 0 {
		builder.CustomPropertiesOfAny(customProperties...)
	}
}

func vmBuilderInitialization(params OptionalVMParameters, builder *ovirtsdk.VmBuilder) {
	if init := params.Initialization(); init != nil {
		initBuilder := ovirtsdk.NewInitializationBuilder()

		if init.CustomScript() != "" {
			initBuilder.CustomScript(init.CustomScript())
		}
		if init.HostName() != "" {
			initBuilder.HostName(init.HostName())
		}
		builder.InitializationBuilder(initBuilder)
	}
}

func (o *oVirtClient) CreateVM(clusterID ClusterID, templateID TemplateID, name string, params OptionalVMParameters, retries ...RetryStrategy) (result VM, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts())

	if err := validateVMCreationParameters(clusterID, templateID, name, params); err != nil {
		return nil, err
	}

	if params == nil {
		params = &vmParams{}
	}

	message := fmt.Sprintf("creating VM %s", name)
	vm, err := createSDKVM(clusterID, templateID, name, params)
	if err != nil {
		return nil, err
	}

	err = retry(
		message,
		o.logger,
		retries,
		func() error {
			response, err := o.conn.SystemService().VmsService().Add().Vm(vm).Send()
			if err != nil {
				return err
			}
			vm, ok := response.Vm()
			if !ok {
				return newError(EFieldMissing, "missing VM in VM create response")
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
		},
	)
	return result, err
}

func createSDKVM(
	clusterID ClusterID,
	templateID TemplateID,
	name string,
	params OptionalVMParameters,
) (*ovirtsdk.Vm, error) {
	builder := ovirtsdk.NewVmBuilder()
	builder.Cluster(ovirtsdk.NewClusterBuilder().Id(string(clusterID)).MustBuild())
	builder.Template(ovirtsdk.NewTemplateBuilder().Id(string(templateID)).MustBuild())
	builder.Name(name)
	parts := []vmBuilderComponent{
		vmBuilderComment,
		vmBuilderCPU,
		vmBuilderHugePages,
		vmBuilderInitialization,
	}

	for _, part := range parts {
		part(params, builder)
	}

	vm, err := builder.Build()
	if err != nil {
		return nil, wrap(err, EBug, "failed to build VM")
	}
	return vm, nil
}

func validateVMCreationParameters(clusterID ClusterID, templateID TemplateID, name string, _ OptionalVMParameters) error {
	if name == "" {
		return newError(EBadArgument, "name cannot be empty for VM creation")
	}
	if clusterID == "" {
		return newError(EBadArgument, "cluster ID cannot be empty for VM creation")
	}
	if templateID == "" {
		return newError(EBadArgument, "template ID cannot be empty for VM creation")
	}
	return nil
}
