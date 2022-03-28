package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestVMListShouldNotFail(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	if _, err := client.ListVMs(); err != nil {
		t.Fatal(err)
	}
}
func TestGetVMByName(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()
	vmName := fmt.Sprintf("test-%s", helper.GenerateRandomID(5))

	assertCanCreateVM(
		t,
		helper,
		vmName,
		nil,
	)

	fetchedVM, err := client.GetVMByName(vmName)
	if err != nil {
		t.Fatal(err)
	}
	if fetchedVM == nil {
		t.Fatal("returned VM is nil")
	}

	t.Logf("fetched VM Name %s mismatches original created VM Name %s", fetchedVM.Name(), vmName)
	if fetchedVM.Name() != vmName {
		t.Fatalf("fetched VM Name %s mismatches original created VM Name %s", fetchedVM.Name(), vmName)
	}

}
func TestAfterVMCreationShouldBePresent(t *testing.T) {
	t.Parallel()
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

	params := map[string]ovirtclient.OptionalVMParameters{
		"nocpu":   ovirtclient.CreateVMParams(),
		"withcpu": ovirtclient.CreateVMParams().MustWithCPUParameters(1, 1, 1),
	}
	for name, param := range params {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			helper := getHelper(t)
			vm := assertCanCreateVM(
				t,
				helper,
				fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
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
	t.Parallel()
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

func TestVMCreationWithInit(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	vm1 := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	tpl := assertCanCreateTemplate(t, helper, vm1)
	vm2 := assertCanCreateVMFromTemplate(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		tpl.ID(),
		ovirtclient.CreateVMParams().MustWithInitializationParameters("script-test", "test-vm"),
	)

	if vm2.Initialization().CustomScript() != "script-test" {
		t.Fatalf("got Unexpected output from the CustomScript (%s) init field ", vm2.Initialization().CustomScript())
	}

	if vm2.Initialization().HostName() != "test-vm" {
		t.Fatalf("got Unexpected output from the HostName (%s) init field ", vm2.Initialization().HostName())
	}
}

// TestVMStartStop creates a micro VM with a tiny operating system, starts it and then stops it. The OS doesn't support
// ACPI, so shutdown cannot be tested.
func TestVMStartStop(t *testing.T) {
	t.Parallel()
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

func TestVMHugePages(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().MustWithHugePages(ovirtclient.VMHugePages2M),
	)
	if hugePages := vm.HugePages(); hugePages == nil || *hugePages != ovirtclient.VMHugePages2M {
		if hugePages == nil {
			t.Fatalf("Hugepages not set on VM")
		} else {
			t.Fatalf("Incorrect value for Hugepages: %d", hugePages)
		}
	}
}

func TestCanRemoveTemplateIfVMIsCloned(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	templateVM := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		nil,
	)
	template := assertCanCreateTemplate(t, helper, templateVM)
	vm := assertCanCreateVMFromTemplate(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		template.ID(),
		ovirtclient.CreateVMParams().MustWithClone(true),
	)
	if vm.TemplateID() != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("Template ID is not set correctly on cloned VM (%s vs %s)", vm.TemplateID(), template.ID())
	}
	assertCanRemoveTemplate(t, helper, template.ID())
}

func TestCannotRemoveTemplateIfVMIsNotCloned(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	templateVM := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		nil,
	)
	template := assertCanCreateTemplate(t, helper, templateVM)
	vm := assertCanCreateVMFromTemplate(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		template.ID(),
		ovirtclient.CreateVMParams().MustWithClone(false),
	)
	if vm.TemplateID() != template.ID() {
		t.Fatalf("Template ID is not set correctly on non-cloned VM (%s vs %s)", vm.TemplateID(), template.ID())
	}
	assertCannotRemoveTemplate(t, helper, template.ID())
}

func TestCannotRemoveTemplateIfVMCloneIsNotSet(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	templateVM := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		nil,
	)
	template := assertCanCreateTemplate(t, helper, templateVM)
	vm := assertCanCreateVMFromTemplate(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		template.ID(),
		nil,
	)
	if vm.TemplateID() != template.ID() {
		t.Fatalf("Template ID is not set correctly on VM if cloning is not specified (%s vs %s)", vm.TemplateID(), template.ID())
	}
	assertCannotRemoveTemplate(t, helper, template.ID())
}

func assertCanCreateVM(
	t *testing.T,
	helper ovirtclient.TestHelper,
	name string,
	params ovirtclient.OptionalVMParameters,
) ovirtclient.VM {
	return assertCanCreateVMFromTemplate(t, helper, name, helper.GetBlankTemplateID(), params)
}

func TestVMWithoutInitialization(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test_%s", helper.GenerateRandomID(5)), nil)

	vm, err := helper.GetClient().GetVM(vm.ID())
	if err != nil {
		t.Fatalf("Failed to re-fetch VM after creation (%v)", err)
	}

	if vm.Initialization() == nil {
		t.Fatalf("Initialization is nil on VM without initialization.")
	}
}

func assertCanCreateVMFromTemplate(
	t *testing.T,
	helper ovirtclient.TestHelper,
	name string,
	templateID ovirtclient.TemplateID,
	params ovirtclient.OptionalVMParameters,
) ovirtclient.VM {
	t.Logf("Creating VM %s from template %s...", name, templateID)
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
			if vm != nil {
				t.Logf("Cleaning up test VM %s...", vm.ID())
				if err := vm.Remove(); err != nil && !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
					t.Fatalf("Failed to remove test VM %s (%v)", vm.ID(), err)
				}
			}
		},
	)
	t.Logf("Created VM %s.", vm.ID())
	return vm
}

func assertCanStartVM(t *testing.T, vm ovirtclient.VM) {
	if err := vm.Start(); err != nil {
		t.Fatalf("Failed to start VM (%v)", err)
	}
	t.Cleanup(func() {
		if vm.Status() == ovirtclient.VMStatusDown {
			return
		}
		t.Logf("Stopping test VM %s...", vm.ID())
		if err := vm.Stop(true); err != nil {
			t.Fatalf("Failed to stop VM %s after test (%v)", vm.ID(), err)
		}
		if _, err := vm.WaitForStatus(ovirtclient.VMStatusDown); err != nil {
			t.Fatalf("Failed to wait for VM %s to stop (%v)", vm.ID(), err)
		}
	})
}

func assertVMWillStart(t *testing.T, vm ovirtclient.VM) ovirtclient.VM {
	vm, err := vm.WaitForStatus(ovirtclient.VMStatusUp)
	if err != nil {
		t.Fatalf("Failed to wait for VM status to reach \"up\". (%v)", err)
	}
	return vm
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

func assertCanCreateBootableVM(t *testing.T, helper ovirtclient.TestHelper) ovirtclient.VM {
	vm1 := assertCanCreateVM(
		t,
		helper,
		helper.GenerateRandomID(5),
		nil,
	)
	disk1 := assertCanCreateDisk(t, helper)
	assertCanUploadDiskImage(t, helper, disk1)
	assertCanAttachDisk(t, vm1, disk1)
	return vm1
}
