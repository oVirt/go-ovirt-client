package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestVMRemovalShouldRemoveAttachedDisks(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(t, helper, fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)), nil)
	disk := assertCanCreateDisk(t, helper)
	assertCanAttachDisk(t, vm, disk)
	if err := vm.Remove(); err != nil {
		t.Fatalf("Cannot remove test VM %s (%v)", vm.ID(), err)
	}
	_, err := helper.GetClient().GetDisk(disk.ID())
	if err == nil {
		t.Fatalf("Disk was still present after the VM has been removed.")
	}
	if !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
		t.Fatalf("Getting disk after VM removal did not result in a non found error (%v).", err)
	}
}
