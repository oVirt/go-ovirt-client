package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

// TestTemplateDiskListAttachments tests if listing the disk attachments of a template works properly.
func TestTemplateDiskListAttachments(t *testing.T) {
	t.Parallel()
	t.Run("empty", testTemplateDiskListAttachmentsEmpty)
	t.Run("single", testTemplateDiskListAttachmentsSingle)
}

func testTemplateDiskListAttachmentsSingle(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	disk := assertCanCreateDisk(t, helper)
	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	assertCanAttachDisk(t, vm, disk)
	tpl := assertCanCreateTemplate(t, helper, vm)
	diskAttachments, err := tpl.ListDiskAttachments()
	if err != nil {
		t.Fatalf("Failed to list disk attachments for template %s (%v).", tpl.ID(), err)
	}
	if len(diskAttachments) != 1 {
		t.Fatalf(
			"List of disk attachments for template %s does not contain 1 element (%d elements found).",
			tpl.ID(),
			len(diskAttachments),
		)
	}
	tplDisk := assertCanGetDiskFromTemplateAttachment(t, helper, diskAttachments[0])
	if tplDisk.TotalSize() != disk.TotalSize() {
		t.Fatalf(
			"Template disk %s has incorrect size (%d instead of %d bytes).",
			tplDisk.ID(),
			tplDisk.TotalSize(),
			disk.TotalSize(),
		)
	}
}

func assertCanListTemplateDiskAttachments(
	t *testing.T,
	tpl ovirtclient.Template,
) []ovirtclient.TemplateDiskAttachment {
	diskAttachments, err := tpl.ListDiskAttachments()
	if err != nil {
		t.Fatalf("Failed to list disk attachments of template %s (%v).", tpl.ID(), err)
	}
	return diskAttachments
}

func assertCanGetDiskFromTemplateAttachment(
	t *testing.T,
	helper ovirtclient.TestHelper,
	attachment ovirtclient.TemplateDiskAttachment,
) ovirtclient.Disk {
	disk, err := helper.GetClient().GetDisk(attachment.DiskID())
	if err != nil {
		t.Fatalf("Failed to get disk %s from attachment %s (%v).", attachment.DiskID(), disk.ID(), err)
	}
	return disk
}

func testTemplateDiskListAttachmentsEmpty(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	tpl := assertCanCreateTemplate(t, helper, vm)
	diskAttachments, err := tpl.ListDiskAttachments()
	if err != nil {
		t.Fatalf("Failed to list disk attachments for template %s (%v).", tpl.ID(), err)
	}
	if len(diskAttachments) != 0 {
		t.Fatalf(
			"List of disk attachments for template %s is not empty (%d elements fiund).",
			tpl.ID(),
			len(diskAttachments),
		)
	}
}
