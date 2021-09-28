package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestVMNICUpdate(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		ovirtclient.CreateVMParams().MustWithName(fmt.Sprintf("nic_test_%s", helper.GenerateRandomID(5))),
	)
	assertNICCount(t, vm, 0)
	nic := assertCanCreateNIC(t, helper, vm, "test", ovirtclient.CreateNICParams())
	nic = assertCanUpdateNICName(t, nic, "test1")
	vnicProfile := assertCanCreateVNICProfile(t, helper)
	nic = assertCanUpdateNICVNICProfile(t, nic, vnicProfile.ID())
	// Go back to the original VNIC profile ID to make sure we don't block deleting the test VNIC profile.
	_ = assertCanUpdateNICVNICProfile(t, nic, helper.GetVNICProfileID())
}
