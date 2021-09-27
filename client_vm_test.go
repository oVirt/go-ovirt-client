package ovirtclient_test

import (
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

	vm, err := client.CreateVM(
		helper.GetClusterID(),
		helper.GetBlankTemplateID(),
		ovirtclient.CreateVMParams().MustWithName("test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = client.RemoveVM(vm.ID())
		if err != nil {
			t.Fatalf("failed to remove VM after test, please remove manually (%v)", err)
		}
	}()
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

func assertCanCreateVM(
	t *testing.T,
	helper ovirtclient.TestHelper,
	params ovirtclient.OptionalVMParameters,
) ovirtclient.VM {
	client := helper.GetClient()
	vm, err := client.CreateVM(
		helper.GetClusterID(),
		helper.GetBlankTemplateID(),
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
