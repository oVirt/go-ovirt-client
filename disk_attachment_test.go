package ovirtclient_test

import (
	"errors"
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestDiskAttachmentCreation(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("disk_attachment_test_%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	disk := assertCanCreateDisk(t, helper)
	assertDiskAttachmentCount(t, vm, 0)
	attachment := assertCanAttachDisk(t, vm, disk)
	assertDiskAttachmentMatches(t, attachment, disk, vm)

	if attachment.Active() {
		t.Fatalf("Incorrect value for 'active' on newly created disk attachment.")
	}
	if attachment.Bootable() {
		t.Fatalf("Incorrect value for 'bootable' on newly created disk attachment.")
	}

	assertDiskAttachmentCount(t, vm, 1)
	assertCanDetachDisk(t, attachment)
}

func TestDiskAttachmentCannotBeAttachedToSecondVM(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	vm1 := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("disk_attachment_test_%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	vm2 := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("disk_attachment_test_%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	disk := assertCanCreateDisk(t, helper)
	_ = assertCanAttachDisk(t, vm1, disk)
	assertCannotAttachDisk(t, vm2, disk, ovirtclient.EConflict)
}

func assertCanCreateDisk(t *testing.T, helper ovirtclient.TestHelper) ovirtclient.Disk {
	return assertCanCreateDiskWithParameters(t, helper, ovirtclient.ImageFormatRaw, nil)
}

func assertCanCreateDiskWithParameters(
	t *testing.T,
	helper ovirtclient.TestHelper,
	format ovirtclient.ImageFormat,
	parameters ovirtclient.CreateDiskOptionalParameters,
) ovirtclient.Disk {
	t.Logf("Creating test disk...")
	client := helper.GetClient()
	disk, err := client.CreateDisk(
		helper.GetStorageDomainID(),
		format,
		1048576,
		parameters,
	)
	if disk != nil {
		t.Cleanup(
			func() {
				if err := disk.Remove(); err != nil {
					if !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
						t.Fatalf("Failed to remove test disk %s (%v)", disk.ID(), err)
					}
				}
			},
		)
	}
	if err != nil {
		t.Fatalf("Failed to create test disk (%v)", err)
	}
	return disk
}

func assertCanDetachDisk(t *testing.T, attachment ovirtclient.DiskAttachment) {
	if err := attachment.Remove(); err != nil {
		t.Fatalf("Failed to detach disk %s from VM %s (%v)", attachment.DiskID(), attachment.VMID(), err)
	}
}

func assertCanAttachDisk(t *testing.T, vm ovirtclient.VM, disk ovirtclient.Disk) ovirtclient.DiskAttachment {
	return assertCanAttachDiskWithParams(t, vm, disk, nil)
}

func assertCanAttachDiskWithParams(
	t *testing.T,
	vm ovirtclient.VM,
	disk ovirtclient.Disk,
	params ovirtclient.CreateDiskAttachmentOptionalParams,
) ovirtclient.DiskAttachment {
	attachment, err := vm.AttachDisk(disk.ID(), ovirtclient.DiskInterfaceVirtIO, params)
	if err != nil {
		t.Fatalf("Failed to create disk attachment (%v)", err)
	}
	if attachment.VMID() != vm.ID() {
		t.Fatalf("Mismatching VM ID after creation (%s != %s)", attachment.VMID(), vm.ID())
	}
	if attachment.DiskID() != disk.ID() {
		t.Fatalf("Mismatching disk ID after creation (%s != %s)", attachment.DiskID(), disk.ID())
	}
	if attachment.DiskInterface() != ovirtclient.DiskInterfaceVirtIO {
		t.Fatalf(
			"Mismatching disk interface after creation (%s != %s)",
			attachment.DiskInterface(),
			ovirtclient.DiskInterfaceVirtIO,
		)
	}
	return attachment
}

func assertCannotAttachDisk(t *testing.T, vm ovirtclient.VM, disk ovirtclient.Disk, errorCode ovirtclient.ErrorCode) {
	if _, err := vm.AttachDisk(disk.ID(), ovirtclient.DiskInterfaceVirtIO, nil, ovirtclient.MaxTries(5)); err != nil {
		if errorCode == "" {
			return
		}
		var e ovirtclient.EngineError
		if !errors.As(err, &e) {
			t.Fatalf("Failed to convert returned error to EngineError (%v)", err)
		}
		if !e.HasCode(errorCode) {
			t.Fatalf("Unexpected error code: %s, instead of: %s", e.Code(), errorCode)
		}
		return
	}
	t.Fatalf("Unexectedly attached disk %s to VM %s", disk.ID(), vm.ID())
}

func assertDiskAttachmentMatches(
	t *testing.T,
	attachment ovirtclient.DiskAttachment,
	disk ovirtclient.Disk,
	vm ovirtclient.VM,
) {
	if attachment.DiskID() != disk.ID() {
		t.Fatalf("Disk attachment disk ID does not match (is %s should be %s)", attachment.DiskID(), disk.ID())
	}
	if attachment.VMID() != vm.ID() {
		t.Fatalf("Disk attachment vm ID does not match (is %s should be %s)", attachment.DiskID(), vm.ID())
	}
}

func assertDiskAttachmentCount(t *testing.T, vm ovirtclient.VM, count uint) {
	attachments, err := vm.ListDiskAttachments()
	if err != nil {
		t.Fatalf("Failed to list disk attachments for VM %s (%v)", vm.ID(), err)
	}
	if len(attachments) != int(count) {
		t.Fatalf("Invalid number of attachments on VM %s, expected %d, is %d", vm.ID(), count, len(attachments))
	}
}
