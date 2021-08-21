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

	vm, err := client.CreateVM("test", helper.GetClusterID(), helper.GetBlankTemplateID(), ovirtclient.VMParams())
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
}
