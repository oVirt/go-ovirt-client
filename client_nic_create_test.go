package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
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

func TestDuplicateVMNICCreationWithSameName(t *testing.T) {
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
	assertCannotCreateNICWithSameName(
		t,
		helper,
		vm,
		nicName,
		ovirtclient.CreateNICParams())
	assertNICCount(t, vm, 1)
	assertCanRemoveNIC(t, nic1)
	assertNICCount(t, vm, 0)
}

func TestDuplicateVMNICCreationWithSameNameAndDiffVNICProfile(t *testing.T) {
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
	DiffVNICProfile, _ := assertCanFindDiffVNICProfile(helper, helper.GetVNICProfileID())
	if DiffVNICProfile == "" {
		assertCannotCreateNICWithSameName(
			t,
			helper,
			vm,
			nicName,
			ovirtclient.CreateNICParams())
	} else {
		assertCannotCreateNICWithVNICProfile(
			t,
			vm,
			nicName,
			DiffVNICProfile,
			ovirtclient.CreateNICParams())
	}
	assertNICCount(t, vm, 1)
	assertCanRemoveNIC(t, nic1)
	assertNICCount(t, vm, 0)
}
