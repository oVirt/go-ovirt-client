package ovirtclient_test

import (
	"fmt"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

// The following example demonstrates how to create a virtual machine. It is set up
// using the test helper, but can be easily modified to use the client directly.
func ExampleVMClient_create() {
	// Create the helper for testing. Alternatively, you could create a production client with ovirtclient.New()
	helper := ovirtclient.NewTestHelperFromEnv(ovirtclientlog.NewNOOPLogger())
	// Get the oVirt client
	client := helper.GetClient()

	// This is the cluster the VM will be created on.
	clusterID := helper.GetClusterID()
	// Use the blank template as a starting point.
	templateID := helper.GetBlankTemplateID()
	// Set the VM name
	name := "test-vm"
	// Create the optional parameters.
	params := ovirtclient.CreateVMParams()

	// Create the VM...
	vm, err := client.CreateVM(clusterID, templateID, name, params)
	if err != nil {
		panic(fmt.Sprintf("failed to create VM (%v)", err))
	}

	// ... and then remove it. Alternatively, you could call client.RemoveVM(vm.ID()).
	if err := vm.Remove(); err != nil {
		panic(fmt.Sprintf("failed to remove VM (%v)", err))
	}
}
