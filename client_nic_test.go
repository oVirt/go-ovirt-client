package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestVMNICCreation(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	vm, err := client.CreateVM(
		helper.GetClusterID(),
		helper.GetBlankTemplateID(),
		ovirtclient.CreateVMParams().WithName("test"),
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

	nics, err := vm.ListNICs()
	if err != nil {
		t.Fatalf("failed to list NICs on VM %s", vm.ID())
	}
	if len(nics) != 0 {
		t.Fatalf("unexpected number of NICs before NIC creation on VM %s: %d", vm.ID(), len(nics))
	}

	nic, err := vm.CreateNIC("test", helper.GetVNICProfileID())
	if err != nil {
		t.Fatalf("failed to create NIC on VM %s", vm.ID())
	}
	if nic.VMID() != vm.ID() {
		t.Fatalf("VM ID mismatch between NIC and VM (%s != %s)", nic.VMID(), vm.ID())
	}

	nics, err = vm.ListNICs()
	if err != nil {
		t.Fatalf("failed to list NICs on VM %s", vm.ID())
	}
	if len(nics) != 1 {
		t.Fatalf("unexpected number of NICs after NIC creation on VM %s: %d", vm.ID(), len(nics))
	}

	if err := nic.Remove(); err != nil {
		t.Fatalf("failed to remove NIC %s (%v)", nic.ID(), err)
	}

	nics, err = vm.ListNICs()
	if err != nil {
		t.Fatalf("failed to list NICs on VM %s", vm.ID())
	}
	if len(nics) != 0 {
		t.Fatalf("unexpected number of NICs after NIC removal on VM %s: %d", vm.ID(), len(nics))
	}
}
