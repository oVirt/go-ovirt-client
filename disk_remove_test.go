package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

// TestDiskRemoveFromStoppedVMShouldNotResultInError tests if removing a disk from under a stopped VM does not result
// in an error.
func TestDiskRemoveFromStoppedVMShouldNotResultInError(t *testing.T) {
	helper := getHelper(t)
	disk := assertCanCreateDisk(t, helper)
	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	assertCanAttachDisk(t, vm, disk)
	if err := disk.Remove(); err != nil {
		t.Fatalf("Removing a disk from a stopped VM resulted in an error (%v).", err)
	}
}

// TestDiskRemoveFromRunningVMShouldResultInError tests if removing a disk from under a running VM results in an error.
func TestDiskRemoveFromRunningVMShouldResultInError(t *testing.T) {
	helper := getHelper(t)
	disk := assertCanCreateDisk(t, helper)
	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	assertCanAttachDisk(t, vm, disk)
	assertCanStartVM(t, helper, vm)
	if err := disk.Remove(ovirtclient.MaxTries(5)); err == nil {
		t.Fatalf("Removing a disk from a running VM did not result in an error.")
	}
}

// TestDiskRemoveShouldResultInIllegalTemplate tests that when removing a disk from under a template, the template
// should switch to the "illegal" status.
func TestDiskRemoveShouldResultInError(t *testing.T) {
	helper := getHelper(t)
	disk := assertCanCreateDisk(t, helper)
	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	assertCanAttachDisk(t, vm, disk)
	tpl := assertCanCreateTemplate(t, helper, vm)
	diskAttachments := assertCanListTemplateDiskAttachments(t, tpl)
	tplDisk := assertCanGetDiskFromTemplateAttachment(t, helper, diskAttachments[0])
	if err := tplDisk.Remove(ovirtclient.MaxTries(5)); err == nil {
		t.Fatalf("Trying to remove a disk from a template did not result in an error.")
	}
}
