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

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		nil,
	)
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

func TestVMCreationFromTemplateChangedCPUValues(t *testing.T) {
	helper := getHelper(t)
	vm1 := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().MustWithCPUParameters(2, 2, 2),
	)
	tpl := assertCanCreateTemplate(t, helper, vm1)
	vm2 := assertCanCreateVMFromTemplate(
		t,
		helper,
		"test2",
		tpl.ID(),
		ovirtclient.CreateVMParams().MustWithCPUParameters(3, 3, 3),
	)
	if vm2.CPU().Topo().Cores() != 3 {
		t.Fatalf("Invalid number of cores: %d", vm2.CPU().Topo().Cores())
	}
	if vm2.CPU().Topo().Threads() != 3 {
		t.Fatalf("Invalid number of cores: %d", vm2.CPU().Topo().Threads())
	}
	if vm2.CPU().Topo().Sockets() != 3 {
		t.Fatalf("Invalid number of cores: %d", vm2.CPU().Topo().Sockets())
	}
}

// TestVMStartStop creates a micro VM with a tiny operating system, starts it and then stops it. The OS doesn't support
// ACPI, so shutdown cannot be tested.
func TestVMStartStop(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		nil,
	)
	disk := assertCanCreateDisk(t, helper)
	assertCanAttachDisk(t, vm, disk)
	assertCanUploadDiskImage(t, helper, disk)
	assertCanStartVM(t, vm)
	assertVMWillStart(t, vm)
	assertCanStopVM(t, vm)
	assertVMWillStop(t, vm)
}

func assertCanCreateVM(
	t *testing.T,
	helper ovirtclient.TestHelper,
	name string,
	params ovirtclient.OptionalVMParameters,
) ovirtclient.VM {
	return assertCanCreateVMFromTemplate(t, helper, name, helper.GetBlankTemplateID(), params)
}

func assertCanCreateVMFromTemplate(
	t *testing.T,
	helper ovirtclient.TestHelper,
	name string,
	templateID ovirtclient.TemplateID,
	params ovirtclient.OptionalVMParameters,
) ovirtclient.VM {
	client := helper.GetClient()
	vm, err := client.CreateVM(
		helper.GetClusterID(),
		templateID,
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

func assertCanStartVM(t *testing.T, vm ovirtclient.VM) {
	if err := vm.Start(); err != nil {
		t.Fatalf("Failed to start VM (%v)", err)
	}
	t.Cleanup(func() {
		if err := vm.Stop(true); err != nil {
			t.Fatalf("Failed to stop VM %s after test (%v)", vm.ID(), err)
		}
	})
}

func assertVMWillStart(t *testing.T, vm ovirtclient.VM) {
	if _, err := vm.WaitForStatus(ovirtclient.VMStatusUp); err != nil {
		t.Fatalf("Failed to wait for VM status to reach \"up\". (%v)", err)
	}
}

func assertCanStopVM(t *testing.T, vm ovirtclient.VM) {
	if err := vm.Stop(false); err != nil {
		t.Fatalf("Failed to stop VM (%v)", err)
	}
}

func assertVMWillStop(t *testing.T, vm ovirtclient.VM) {
	if _, err := vm.WaitForStatus(ovirtclient.VMStatusDown); err != nil {
		t.Fatalf("Failed to wait for VM status to reach \"down\". (%v)", err)
	}
}
