package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestVMNICUpdate(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("nic_test_%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	assertNICCount(t, vm, 0)
	nic := assertCanCreateNIC(
		t,
		helper,
		vm,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateNICParams(),
	)
	nic = assertCanUpdateNICName(t, nic, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)))
	vnicProfile := assertCanCreateVNICProfile(t, helper)
	nic = assertCanUpdateNICVNICProfile(t, nic, vnicProfile.ID())
	nic = assertCanUpdateNICMac(t, nic, "a1:b2:c3:d4:e5:f6")
	_ = assertCantUpdateNICMac(t, nic, "invalid mac address")
	// Go back to the original VNIC profile ID to make sure we don't block deleting the test VNIC profile.
	_ = assertCanUpdateNICVNICProfile(t, nic, helper.GetVNICProfileID())
}
