package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestVMNICCreation(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		ovirtclient.CreateVMParams().MustWithName(fmt.Sprintf("nic_test_%s", helper.GenerateRandomID(5))),
	)
	assertNICCount(t, vm, 0)
	nic := assertCanCreateNIC(t, helper, vm, "test", ovirtclient.CreateNICParams())
	assertNICCount(t, vm, 1)
	assertCanRemoveNIC(t, nic)
	assertNICCount(t, vm, 0)
}

func assertCanCreateNIC(
	t *testing.T,
	helper ovirtclient.TestHelper,
	vm ovirtclient.VM,
	name string,
	params ovirtclient.BuildableNICParameters,
) ovirtclient.NIC {
	nic, err := vm.CreateNIC(name, helper.GetVNICProfileID(), params)
	if err != nil {
		t.Fatalf("failed to create NIC on VM %s", vm.ID())
	}
	if nic.VMID() != vm.ID() {
		t.Fatalf("VM ID mismatch between NIC and VM (%s != %s)", nic.VMID(), vm.ID())
	}
	return nic
}

func assertCanRemoveNIC(t *testing.T, nic ovirtclient.NIC) {
	if err := nic.Remove(); err != nil {
		t.Fatalf("failed to remove NIC %s (%v)", nic.ID(), err)
	}
}

func assertNICCount(t *testing.T, vm ovirtclient.VM, n int) {
	nics, err := vm.ListNICs()
	if err != nil {
		t.Fatalf("failed to list NICs on VM %s", vm.ID())
	}
	if len(nics) != n {
		t.Fatalf("unexpected number of NICs after NIC removal on VM %s: %d instead of %d", vm.ID(), len(nics), n)
	}
}
