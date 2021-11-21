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
		fmt.Sprintf("nic_test_%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	assertNICCount(t, vm, 0)
	nic := assertCanCreateNIC(t, helper, vm, "test", ovirtclient.CreateNICParams())
	assertNICCount(t, vm, 1)
	assertCanRemoveNIC(t, nic)
	assertNICCount(t, vm, 0)
}
