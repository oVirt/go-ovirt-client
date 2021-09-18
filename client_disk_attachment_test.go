package ovirtclient_test

import (
	"errors"
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestDiskAttachmentCreation(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	vm := assertCanCreateVM(t, client, helper)
	disk := assertCanCreateDisk(t, client, helper)
	assertDiskAttachmentCount(t, vm, 0)
	attachment := assertCanAttachDisk(t, vm, disk)
	assertDiskAttachmentMatches(t, attachment, disk, vm)
	assertDiskAttachmentCount(t, vm, 1)
	assertCanDetachDisk(t, attachment)
}

func TestDiskAttachmentCannotBeAttachedToSecondVM(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	vm1 := assertCanCreateVM(t, client, helper)
	vm2 := assertCanCreateVM(t, client, helper)
	disk := assertCanCreateDisk(t, client, helper)
	_ = assertCanAttachDisk(t, vm1, disk)
	assertCannotAttachDisk(t, vm2, disk, ovirtclient.EConflict)
}

func assertCanCreateDisk(t *testing.T, client ovirtclient.Client, helper ovirtclient.TestHelper) ovirtclient.Disk {
	disk, err := client.CreateDisk(
		helper.GetStorageDomainID(),
		ovirtclient.ImageFormatRaw,
		512,
		nil,
	)
	if disk != nil {
		t.Cleanup(
			func() {
				if err := disk.Remove(); err != nil {
					t.Fatalf("Failed to remove test disk %s (%v)", disk.ID(), err)
				}
			},
		)
	}
	if err != nil {
		t.Fatalf("Failed to create test disk (%v)", err)
	}
	return disk
}

func assertCanCreateVM(t *testing.T, client ovirtclient.Client, helper ovirtclient.TestHelper) ovirtclient.VM {
	vm, err := client.CreateVM(
		helper.GetClusterID(),
		helper.GetBlankTemplateID(),
		ovirtclient.CreateVMParams().MustWithName(fmt.Sprintf("disk_attachment_test_%s", helper.GenerateRandomID(5))),
	)
	if err != nil {
		t.Fatalf("Failed to create test VM (%v)", err)
	}
	t.Cleanup(
		func() {
			if err := vm.Remove(); err != nil {
				t.Fatalf("Failed to remove test VM %s (%v)", vm.ID(), err)
			}
		},
	)
	return vm
}

func assertCanDetachDisk(t *testing.T, attachment ovirtclient.DiskAttachment) {
	if err := attachment.Remove(); err != nil {
		t.Fatalf("Failed to detach disk %s from VM %s (%v)", attachment.DiskID(), attachment.VMID(), err)
	}
}

func assertCanAttachDisk(t *testing.T, vm ovirtclient.VM, disk ovirtclient.Disk) ovirtclient.DiskAttachment {
	attachment, err := vm.AttachDisk(disk.ID(), ovirtclient.DiskInterfaceVirtIO, nil)
	if err != nil {
		t.Fatalf("Failed to create disk attachment (%v)", err)
	}
	return attachment
}

func assertCannotAttachDisk(t *testing.T, vm ovirtclient.VM, disk ovirtclient.Disk, errorCode ovirtclient.ErrorCode) {
	if _, err := vm.AttachDisk(disk.ID(), ovirtclient.DiskInterfaceVirtIO, nil); err != nil {
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
