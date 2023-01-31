package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestExtendDisk(t *testing.T) {
	helper := getHelper(t)

	disk := assertCanCreateDisk(t, helper)

	newDiskSize := disk.ProvisionedSize() + 4096
	t.Logf("Increasing disk %s size to %d bytes...", disk.ID(), newDiskSize)
	updatedDisk, err := disk.Update(ovirtclient.UpdateDiskParams().MustWithProvisionedSize(newDiskSize))
	if err != nil {
		t.Fatalf("Failed to extend disk %s (%v)", disk.ID(), err)
	}

	t.Logf("Waiting for disk to become OK...")
	updatedDisk, err = updatedDisk.WaitForOK()
	if err != nil {
		t.Fatalf("Failed to wait for disk %s to return to OK status. (%v)", disk.ID(), err)
	}

	t.Logf("Checking new disk size...")
	if updatedDisk.ProvisionedSize() < newDiskSize {
		t.Fatalf(
			"The updated disk had a size smaller than expected (%d bytes instead of %d bytes).",
			updatedDisk.ProvisionedSize(),
			newDiskSize,
		)
	}
	t.Logf("New disk size is OK.")
}
