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
	nickName := fmt.Sprintf("duplicate_nic_%s", helper.GenerateRandomID(5))

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
		nickName,
		ovirtclient.CreateNICParams())
	assertNICCount(t, vm, 1)
	assertCannotCreateNICWithSameName(
		t,
		helper,
		vm,
		nickName,
		ovirtclient.CreateNICParams())
	assertNICCount(t, vm, 1)
	assertCanRemoveNIC(t, nic1)
	assertNICCount(t, vm, 0)
}

func TestDuplicateVMNICCreationWithSameNameDiffVNICProfileSameNetwork(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	nickName := fmt.Sprintf("duplicate_nic_%s", helper.GenerateRandomID(5))

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("nic_test_%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	assertNICCount(t, vm, 0)
	assertCanCreateNIC(
		t,
		helper,
		vm,
		nickName,
		ovirtclient.CreateNICParams())
	assertNICCount(t, vm, 1)
	diffVNICProfileSameNetwork := assertCanCreateVNICProfile(t, helper)
	assertCannotCreateNICWithVNICProfile(
		t,
		vm,
		nickName,
		diffVNICProfileSameNetwork.ID(),
		ovirtclient.ETimeout)
	assertNICCount(t, vm, 1)
}
