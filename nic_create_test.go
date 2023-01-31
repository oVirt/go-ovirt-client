package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestVMNICCreation(t *testing.T) {
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
		ovirtclient.CreateNICParams())
	assertNICCount(t, vm, 1)
	assertCanRemoveNIC(t, nic)
	assertNICCount(t, vm, 0)
}

func TestDuplicateVMNICCreation(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	nicName := "test_duplicate_name"

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("nic_test_%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	assertNICCount(t, vm, 0)
	nic1 := assertCanCreateNIC(
		t,
		helper,
		vm,
		nicName,
		ovirtclient.CreateNICParams())
	assertNICCount(t, vm, 1)
	assertCannotCreateNIC(
		t,
		helper,
		vm,
		nicName,
		ovirtclient.CreateNICParams())
	assertNICCount(t, vm, 1)
	assertCanRemoveNIC(t, nic1)
	assertNICCount(t, vm, 0)
}

func TestVMNICWithMacCreation(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	mac := "a1:b2:c3:d4:e5:f6"
	invalidMac := "invalid mac address"

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("nic_test_%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	assertNICCount(t, vm, 0)
	nic := assertCanCreateNICMac(
		t,
		helper,
		vm,
		mac)
	assertNICCount(t, vm, 1)
	assertCantCreateNICMac(
		t,
		helper,
		vm,
		invalidMac)
	assertNICCount(t, vm, 1)
	assertCanRemoveNIC(t, nic)
	assertNICCount(t, vm, 0)
}
