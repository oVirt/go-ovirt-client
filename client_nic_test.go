package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func assertCanUpdateNICName(t *testing.T, nic ovirtclient.NIC, name string) ovirtclient.NIC {
	newNIC, err := nic.Update(ovirtclient.UpdateNICParams().MustWithName(name))
	if err != nil {
		t.Fatalf("failed to update NIC (%v)", err)
	}
	if newNIC.Name() != name {
		t.Fatalf("NIC name not changed after update call")
	}
	return newNIC
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

func assertCannotCreateNICWithSameName(
	t *testing.T,
	helper ovirtclient.TestHelper,
	vm ovirtclient.VM,
	name string,
	params ovirtclient.BuildableNICParameters,
) {

	_, err := vm.CreateNIC(name, helper.GetVNICProfileID(), params)
	if err == nil {
		t.Fatalf("create 2 NICs with same name %s", name)
	}
}

func assertCannotCreateNICWithVNICProfile(
	t *testing.T,
	vm ovirtclient.VM,
	name string,
	diffVNICProfile string,
	params ovirtclient.BuildableNICParameters,
) {
	_, err := vm.CreateNIC(name, diffVNICProfile, params)
	if err == nil {
		t.Fatalf("create 2 NICs with same name %s and different VNICProfile (%v)", name, err)
	}
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

func assertCanUpdateNICVNICProfile(t *testing.T, nic ovirtclient.NIC, vnicProfileID string) ovirtclient.NIC {
	newNIC, err := nic.Update(ovirtclient.UpdateNICParams().MustWithVNICProfileID(vnicProfileID))
	if err != nil {
		t.Fatalf("failed to update NIC with new VNIC profile ID (%v)", err)
	}
	if newNIC.VNICProfileID() != vnicProfileID {
		t.Fatalf("VNIC profile ID not changed after update")
	}
	return newNIC
}
