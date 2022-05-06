package ovirtclient_test

import (
	"fmt"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
)

// The following example demonstrates how to list virtual machines.
func ExampleVMClient_list() {
	// Create the helper for testing. Alternatively, you could create a production client with ovirtclient.New()
	helper, err := ovirtclient.NewLiveTestHelperFromEnv(ovirtclientlog.NewNOOPLogger())
	if err != nil {
		panic(fmt.Errorf("failed to create live test helper (%w)", err))
	}
	// Get the oVirt client
	client := helper.GetClient()

	vms, err := client.ListVMs()
	if err != nil {
		panic(err)
	}
	for _, vm := range vms {
		fmt.Printf("Found VM %s\n", vm.ID())
	}
}
