package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v2"
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
	if params != nil {
		if params.Mac() != "" && params.Mac() != nic.Mac() {
			t.Fatalf("Failed to create NIC with custom mac address. Expected '%s', but created mac is '%s'", params.Mac(), nic.Mac())
		}
	}
	return nic
}

func assertCannotCreateNIC(
	t *testing.T,
	helper ovirtclient.TestHelper,
	vm ovirtclient.VM,
	name string,
	params ovirtclient.BuildableNICParameters,
) {
	nic, err := vm.CreateNIC(name, helper.GetVNICProfileID(), params)
	if nic != nil {
		t.Fatalf("create 2 NICs with same name %s", name)
	}
	if err != nil {
		print(err)
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

func assertCanUpdateNICVNICProfile(t *testing.T, nic ovirtclient.NIC, vnicProfileID ovirtclient.VNICProfileID) ovirtclient.NIC {
	newNIC, err := nic.Update(ovirtclient.UpdateNICParams().MustWithVNICProfileID(vnicProfileID))
	if err != nil {
		t.Fatalf("failed to update NIC with new VNIC profile ID (%v)", err)
	}
	if newNIC.VNICProfileID() != vnicProfileID {
		t.Fatalf("VNIC profile ID not changed after update")
	}
	return newNIC
}

func assertCanUpdateNICMac(t *testing.T, nic ovirtclient.NIC, mac string) ovirtclient.NIC {
	newNIC, err := nic.Update(ovirtclient.UpdateNICParams().MustWithMac(mac))
	if err != nil {
		t.Fatalf("failed to update NIC (%v)", err)
	}
	if newNIC.Mac() != mac {
		t.Fatalf("NIC MacAddress not changed after update call")
	}
	return newNIC
}

func assertCantUpdateNICMac(t *testing.T, nic ovirtclient.NIC, mac string) ovirtclient.NIC {
	newNIC, err := nic.Update(ovirtclient.UpdateNICParams().MustWithMac(mac))
	if err == nil {
		t.Fatalf("Mac address validation error. Invalid mac was accepted: %s", mac)
	}
	return newNIC
}

func assertCanCreateNICMac(
	t *testing.T,
	helper ovirtclient.TestHelper,
	vm ovirtclient.VM,
	mac string,
) ovirtclient.NIC {
	params := ovirtclient.CreateNICParams()
	params, err := params.WithMac(mac)

	if err != nil {
		t.Fatalf("Failed to set custom Mac address on NIC. Error: %s", err)
	}

	nic, err := vm.CreateNIC(fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), helper.GetVNICProfileID(), params)
	if err != nil {
		t.Fatalf("failed to create NIC on VM %s", vm.ID())
	}
	if nic.VMID() != vm.ID() {
		t.Fatalf("VM ID mismatch between NIC and VM (%s != %s)", nic.VMID(), vm.ID())
	}
	if nic.Mac() != params.Mac() {
		t.Fatalf("Failed to create NIC with custom mac address: %s", nic.Mac())
	}

	return nic
}

func assertCantCreateNICMac(
	t *testing.T,
	helper ovirtclient.TestHelper,
	vm ovirtclient.VM,
	mac string,
) ovirtclient.NIC {
	params := ovirtclient.CreateNICParams()
	params, err := params.WithMac(mac)
	if err != nil {
		t.Fatalf("Failed to set custom Mac address on NIC. Error: %s", err)
	}
	nic, err := vm.CreateNIC(fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), helper.GetVNICProfileID(), params)
	if err == nil {
		t.Fatalf("Mac address validation error. Invalid mac was accepted: %s", mac)
	}
	return nic
}
