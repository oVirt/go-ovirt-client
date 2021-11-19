package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestVMListShouldNotFail(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	if _, err := client.ListVMs(); err != nil {
		t.Fatal(err)
	}
}

func TestAfterVMCreationShouldBePresent(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	vm := assertCanCreateVM(t, helper, "test", nil)
	fetchedVM, err := client.GetVM(vm.ID())
	if err != nil {
		t.Fatal(err)
	}
	if fetchedVM == nil {
		t.Fatal("returned VM is nil")
	}
	if fetchedVM.ID() != vm.ID() {
		t.Fatalf("fetched VM ID %s mismatches original created VM ID %s", fetchedVM.ID(), vm.ID())
	}

	updatedVM, err := fetchedVM.Update(
		ovirtclient.UpdateVMParams().MustWithName("new_name").MustWithComment("new comment"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if updatedVM.ID() != vm.ID() {
		t.Fatalf("updated VM ID %s mismatches original created VM ID %s", updatedVM.ID(), vm.ID())
	}
	if updatedVM.Name() != "new_name" {
		t.Fatalf("updated VM name %s does not match update parameters", updatedVM.Name())
	}
	if updatedVM.Comment() != "new comment" {
		t.Fatalf("updated VM comment %s does not match update parameters", updatedVM.Comment())
	}

	fetchedVM, err = client.GetVM(vm.ID())
	if err != nil {
		t.Fatal(err)
	}
	if fetchedVM == nil {
		t.Fatal("returned VM is nil")
	}
	if fetchedVM.Name() != "new_name" {
		t.Fatalf("updated VM name %s does not match update parameters", fetchedVM.Name())
	}
	if fetchedVM.Comment() != "new comment" {
		t.Fatalf("updated VM comment %s does not match update parameters", fetchedVM.Comment())
	}
}

func TestVMCreationWithCPU(t *testing.T) {
	helper := getHelper(t)

	params := map[string]ovirtclient.OptionalVMParameters{
		"nocpu":   ovirtclient.CreateVMParams(),
		"withcpu": ovirtclient.CreateVMParams().MustWithCPUParameters(1, 1, 1),
	}
	for name, param := range params {
		t.Run(name, func(t *testing.T) {
			vm := assertCanCreateVM(
				t,
				helper,
				fmt.Sprintf("test-%s", name),
				param,
			)

			cpu := vm.CPU()
			if cpu == nil {
				t.Fatalf("Creating a VM with CPU settings did not return a VM with CPU.")
			}

			topo := cpu.Topo()
			if topo == nil {
				t.Fatalf("Creating a VM with CPU settings did not return a CPU topology.")
			}

			if cores := topo.Cores(); cores != 1 {
				t.Fatalf("Creating a VM with 1 CPU core returned a topology with %d cores.", cores)
			}

			if threads := topo.Threads(); threads != 1 {
				t.Fatalf("Creating a VM with 1 CPU thread returned a topology with %d threads.", threads)
			}

			if sockets := topo.Sockets(); sockets != 1 {
				t.Fatalf("Creating a VM with 1 CPU socket returned a topology with %d sockets.", sockets)
			}

		})
	}
}

func assertCanCreateVM(
	t *testing.T,
	helper ovirtclient.TestHelper,
	name string,
	params ovirtclient.OptionalVMParameters,
) ovirtclient.VM {
	client := helper.GetClient()
	vm, err := client.CreateVM(
		helper.GetClusterID(),
		helper.GetBlankTemplateID(),
		name,
		params,
	)
	if err != nil {
		t.Fatalf("Failed to create test VM (%v)", err)
	}
	t.Cleanup(
		func() {
			if err := vm.Remove(); err != nil {
				t.Fatalf("Failed to remove test VM %s (%v)", vm.ID(), err)
			}
		},
	)
	return vm
}
