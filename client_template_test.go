package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

// TestTemplateCreation is the simplest test for creating and using a template.
func TestTemplateCreation(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	template := assertCanCreateTemplate(t, helper, vm)
	tpl := assertCanGetTemplateOK(t, helper, template.ID())
	if tpl.ID() != template.ID() {
		t.Fatalf("IDs of the returned template don't match.")
	}
}

// TestTemplateCPU tests if the CPU settings are properly replicated when creating or using a template.
func TestTemplateCPU(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	vm1 := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().MustWithCPUParameters(1, 2, 1),
	)
	tpl := assertCanCreateTemplate(t, helper, vm1)
	if tpl.CPU() == nil {
		t.Fatalf("Template with explicit CPU options returned a nil CPU.")
	}
	if tpl.CPU().Topo() == nil {
		t.Fatalf("Template with explicit CPU options returned a nil topo.")
	}
	if cores := tpl.CPU().Topo().Cores(); cores != vm1.CPU().Topo().Cores() {
		t.Fatalf(
			"Template with explicit CPU options returned the incorrect number of cores (%d instead of %d).",
			cores,
			vm1.CPU().Topo().Cores(),
		)
	}
	if threads := tpl.CPU().Topo().Threads(); threads != vm1.CPU().Topo().Threads() {
		t.Fatalf(
			"Template with explicit CPU options returned the incorrect number of threads (%d instead of %d).",
			threads,
			vm1.CPU().Topo().Threads(),
		)
	}
	if sockets := tpl.CPU().Topo().Sockets(); sockets != vm1.CPU().Topo().Sockets() {
		t.Fatalf(
			"Template with explicit CPU options returned the incorrect number of sockets (%d instead of %d).",
			sockets,
			vm1.CPU().Topo().Sockets(),
		)
	}
	vm2 := assertCanCreateVMFromTemplate(t, helper, "test2", tpl.ID(), nil)
	if vm1.CPU().Topo().Cores() != vm2.CPU().Topo().Cores() {
		t.Fatalf(
			"VM created from template returned incorrect number of cores (%d instead of %d).",
			vm1.CPU().Topo().Cores(),
			vm2.CPU().Topo().Cores(),
		)
	}
	if vm1.CPU().Topo().Threads() != vm2.CPU().Topo().Threads() {
		t.Fatalf(
			"VM created from template returned incorrect number of threads (%d instead of %d).",
			vm1.CPU().Topo().Threads(),
			vm2.CPU().Topo().Threads(),
		)
	}
	if vm1.CPU().Topo().Sockets() != vm2.CPU().Topo().Sockets() {
		t.Fatalf(
			"VM created from template returned incorrect number of sockets (%d instead of %d).",
			vm1.CPU().Topo().Sockets(),
			vm2.CPU().Topo().Sockets(),
		)
	}
}

func TestTemplateDisk(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	disk := assertCanCreateDisk(t, helper)
	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	assertCanAttachDisk(t, vm, disk)
	template := assertCanCreateTemplate(t, helper, vm)

	vm2 := assertCanCreateVMFromTemplate(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), template.ID(),
		nil,
	)

	diskAttachments := assertCanListDiskAttachments(t, vm2)
	if len(diskAttachments) != 1 {
		t.Fatalf(
			"Incorrect number of disk attachments after using template (%d instead of %d).",
			len(diskAttachments),
			1,
		)
	}
	newDisk := assertCanGetDiskFromAttachment(t, diskAttachments[0])
	if newDisk.ProvisionedSize() != disk.ProvisionedSize() {
		t.Fatalf(
			"Incorrect disk size after template use (%d bytes instead of %d).",
			newDisk.ProvisionedSize(),
			disk.ProvisionedSize(),
		)
	}
}

// TestTemplateDiskCopy Copying template disks between storageDomains.
func TestTemplateDiskCopy(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	disk := assertCanCreateDisk(t, helper)
	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	assertCanAttachDisk(t, vm, disk)
	template := assertCanCreateTemplate(t, helper, vm)

	tpl := assertCanGetTemplateOK(t, helper, template.ID())

	diskAttachments := assertCanListTemplateDiskAttachments(t, tpl)
	if len(diskAttachments) != 1 {
		t.Fatalf(
			"Incorrect number of disk attachments after using template (%d instead of %d).",
			len(diskAttachments),
			1,
		)
	}

	templateDisk := assertCanGetDiskFromTemplateAttachment(t, helper, diskAttachments[0])
	secondarySD := helper.GetSecondaryStorageDomainID(t)

	t.Logf("Copying template disk %s to storage domain %s", templateDisk.ID(), secondarySD)

	newDisk, err := helper.GetClient().CopyTemplateDiskToStorageDomain(templateDisk.ID(), secondarySD)
	if err != nil {
		t.Fatalf("Failed to copy template disk to storage domain %s.", secondarySD)
	}

	assertCanGetDiskFromStorageDomain(t, helper, secondarySD, newDisk)

}

func TestGetTemplateByName(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	template := assertCanCreateTemplate(t, helper, vm)

	tpl, err := helper.GetClient().GetTemplateByName(template.Name())
	if err != nil {
		t.Fatalf("Failed to get template by name %s. (%v)", template.Name(), err)
	}

	if tpl.Name() != template.Name() {
		t.Fatalf("fetched Template Name %s mismatches original created Template Name %s", tpl.Name(), template.Name())
	}

}

func assertCanGetDiskFromStorageDomain(t *testing.T, helper ovirtclient.TestHelper, storageDomainID string, disk ovirtclient.Disk) ovirtclient.Disk {
	newDisk, err := helper.GetClient().GetDiskFromStorageDomain(storageDomainID, disk.ID())

	if err != nil {
		t.Fatalf("failed to get Disk %s from storage domain %s", disk.ID(), storageDomainID)
	}

	for _, diskStorageDomain := range newDisk.StorageDomainIDs() {
		if diskStorageDomain == storageDomainID {
			return newDisk
		}

	}
	t.Fatalf("failed to get Disk %s from storage domain %s", disk.ID(), storageDomainID)
	return nil
}

func assertCanGetDiskFromAttachment(t *testing.T, diskAttachment ovirtclient.DiskAttachment) ovirtclient.Disk {
	newDisk, err := diskAttachment.Disk()
	if err != nil {
		t.Fatalf("Failed to get disk %s.", diskAttachment.DiskID())
	}
	return newDisk
}

func assertCanListDiskAttachments(t *testing.T, vm ovirtclient.VM) []ovirtclient.DiskAttachment {
	diskAttachments, err := vm.ListDiskAttachments()
	if err != nil {
		t.Fatalf("Failed to list disk attachments for VM %s. (%v)", vm.ID(), err)
	}
	return diskAttachments
}

func assertCanGetTemplateOK(t *testing.T, helper ovirtclient.TestHelper, id ovirtclient.TemplateID) ovirtclient.Template {
	tpl, err := helper.GetClient().GetTemplate(id)
	if err != nil {
		t.Fatalf("Failed to get template %s. (%v)", id, err)
	}
	t.Logf("Waiting for template %s to become \"ok\"...", tpl.ID())
	tpl, err = tpl.WaitForStatus(ovirtclient.TemplateStatusOK)
	if err != nil {
		t.Fatalf("Failed to wait for template %s to reach \"ok\" status. (%v)", tpl.ID(), err)
	}
	return tpl
}

func assertCanCreateTemplate(t *testing.T, helper ovirtclient.TestHelper, vm ovirtclient.VM) ovirtclient.Template {
	t.Logf("Creating test template from VM %s...", vm.Name())
	template, err := helper.GetClient().CreateTemplate(
		vm.ID(),
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create template from VM %s (%v)", vm.ID(), err)
	}
	t.Cleanup(func() {
		t.Logf("Cleaning up template %s...", template.ID())
		if err := template.Remove(); err != nil && !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
			t.Fatalf("Failed to clean up template %s after test. (%v)", template.ID(), err)
		}
	})
	return template
}

func assertCanRemoveTemplate(t *testing.T, helper ovirtclient.TestHelper, templateID ovirtclient.TemplateID) {
	t.Logf("Removing template %s...", templateID)
	if err := helper.GetClient().RemoveTemplate(templateID); err != nil {
		t.Fatalf("Failed to remove template %s (%v)", templateID, err)
	}
}

func assertCannotRemoveTemplate(t *testing.T, helper ovirtclient.TestHelper, templateID ovirtclient.TemplateID) {
	t.Logf("Removing template %s...", templateID)
	if err := helper.GetClient().RemoveTemplate(templateID); err == nil {
		t.Fatalf("Successfully removed template %s despite assumption.", templateID)
	}
}
