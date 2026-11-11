package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestStorageDomainCRUD(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	hosts, err := client.ListHosts()
	if err != nil {
		t.Fatalf("failed to list hosts (%v)", err)
	}
	if len(hosts) == 0 {
		t.Fatalf("no hosts found")
	}
	hostID := hosts[0].ID()

	sdName := fmt.Sprintf("sd_test_%s", helper.GenerateRandomID(5))
	address := "nfs.example.com"
	path := "/data/export"
	desc := "test storage domain description"
	comment := "test storage domain comment"
	format := ovirtclient.StorageFormatV5
	wipe := true
	csab := int64(5)
	wlsi := int64(10)

	params, err := ovirtclient.NewCreateNFSStorageDomainParams().
		WithDescription(desc)
	if err != nil {
		t.Fatalf("failed to set description (%v)", err)
	}
	params, err = params.WithComment(comment)
	if err != nil {
		t.Fatalf("failed to set comment (%v)", err)
	}
	params, err = params.WithStorageFormat(format)
	if err != nil {
		t.Fatalf("failed to set storage format (%v)", err)
	}
	params, err = params.WithWipeAfterDelete(wipe)
	if err != nil {
		t.Fatalf("failed to set wipe after delete (%v)", err)
	}
	params, err = params.WithCriticalSpaceActionBlocker(csab)
	if err != nil {
		t.Fatalf("failed to set critical space action blocker (%v)", err)
	}
	params, err = params.WithWarningLowSpaceIndicator(wlsi)
	if err != nil {
		t.Fatalf("failed to set warning low space indicator (%v)", err)
	}

	sd, err := client.CreateNFSStorageDomain(sdName, hostID, address, path, params)
	if err != nil {
		t.Fatalf("failed to create storage domain (%v)", err)
	}

	t.Cleanup(func() {
		_ = client.RemoveStorageDomain(sd.ID(), nil)
	})

	if sd.ID() == "" {
		t.Fatalf("storage domain ID is empty")
	}
	if sd.Name() != sdName {
		t.Fatalf("expected name %s, got %s", sdName, sd.Name())
	}
	if sd.StorageType() != ovirtclient.StorageDomainTypeNFS {
		t.Fatalf("expected storage type %s, got %s", ovirtclient.StorageDomainTypeNFS, sd.StorageType())
	}
	if sd.Description() == nil || *sd.Description() != desc {
		t.Fatalf("expected description %s, got %v", desc, sd.Description())
	}
	if sd.Comment() == nil || *sd.Comment() != comment {
		t.Fatalf("expected comment %s, got %v", comment, sd.Comment())
	}
	if sd.StorageAddress() == nil || *sd.StorageAddress() != address {
		t.Fatalf("expected storage address %s, got %v", address, sd.StorageAddress())
	}
	if sd.StoragePath() == nil || *sd.StoragePath() != path {
		t.Fatalf("expected storage path %s, got %v", path, sd.StoragePath())
	}
	if sd.StorageFormat() == nil || *sd.StorageFormat() != format {
		t.Fatalf("expected storage format %s, got %v", format, sd.StorageFormat())
	}
	if sd.WipeAfterDelete() == nil || *sd.WipeAfterDelete() != wipe {
		t.Fatalf("expected wipe after delete %v, got %v", wipe, sd.WipeAfterDelete())
	}
	if sd.CriticalSpaceActionBlocker() == nil || *sd.CriticalSpaceActionBlocker() != csab {
		t.Fatalf("expected critical space action blocker %d, got %v", csab, sd.CriticalSpaceActionBlocker())
	}
	if sd.WarningLowSpaceIndicator() == nil || *sd.WarningLowSpaceIndicator() != wlsi {
		t.Fatalf("expected warning low space indicator %d, got %v", wlsi, sd.WarningLowSpaceIndicator())
	}

	// Fetch storage domain
	fetched, err := client.GetStorageDomain(sd.ID())
	if err != nil {
		t.Fatalf("failed to get storage domain (%v)", err)
	}
	if fetched.Name() != sdName {
		t.Fatalf("expected fetched name %s, got %s", sdName, fetched.Name())
	}

	// Update storage domain
	updatedName := fmt.Sprintf("%s_updated", sdName)
	updatedDesc := "updated description"
	updatedComment := "updated comment"
	updatedWipe := false
	updatedCSAB := int64(10)
	updatedWLSI := int64(20)

	updateParams, err := ovirtclient.NewUpdateStorageDomainParams().
		WithName(updatedName)
	if err != nil {
		t.Fatalf("failed to set update name (%v)", err)
	}
	updateParams, err = updateParams.WithDescription(updatedDesc)
	if err != nil {
		t.Fatalf("failed to set update description (%v)", err)
	}
	updateParams, err = updateParams.WithComment(updatedComment)
	if err != nil {
		t.Fatalf("failed to set update comment (%v)", err)
	}
	updateParams, err = updateParams.WithWipeAfterDelete(updatedWipe)
	if err != nil {
		t.Fatalf("failed to set update wipe after delete (%v)", err)
	}
	updateParams, err = updateParams.WithCriticalSpaceActionBlocker(updatedCSAB)
	if err != nil {
		t.Fatalf("failed to set update critical space action blocker (%v)", err)
	}
	updateParams, err = updateParams.WithWarningLowSpaceIndicator(updatedWLSI)
	if err != nil {
		t.Fatalf("failed to set update warning low space indicator (%v)", err)
	}

	updated, err := client.UpdateStorageDomain(sd.ID(), updateParams)
	if err != nil {
		t.Fatalf("failed to update storage domain (%v)", err)
	}
	if updated.Name() != updatedName {
		t.Fatalf("expected updated name %s, got %s", updatedName, updated.Name())
	}
	if updated.Description() == nil || *updated.Description() != updatedDesc {
		t.Fatalf("expected updated description %s, got %v", updatedDesc, updated.Description())
	}
	if updated.Comment() == nil || *updated.Comment() != updatedComment {
		t.Fatalf("expected updated comment %s, got %v", updatedComment, updated.Comment())
	}
	if updated.WipeAfterDelete() == nil || *updated.WipeAfterDelete() != updatedWipe {
		t.Fatalf("expected updated wipe after delete %v, got %v", updatedWipe, updated.WipeAfterDelete())
	}
	if updated.CriticalSpaceActionBlocker() == nil || *updated.CriticalSpaceActionBlocker() != updatedCSAB {
		t.Fatalf("expected updated critical space action blocker %d, got %v", updatedCSAB, updated.CriticalSpaceActionBlocker())
	}
	if updated.WarningLowSpaceIndicator() == nil || *updated.WarningLowSpaceIndicator() != updatedWLSI {
		t.Fatalf("expected updated warning low space indicator %d, got %v", updatedWLSI, updated.WarningLowSpaceIndicator())
	}

	// Remove storage domain
	if err := client.RemoveStorageDomain(sd.ID(), nil); err != nil {
		t.Fatalf("failed to remove storage domain (%v)", err)
	}

	// Verify removal
	_, err = client.GetStorageDomain(sd.ID())
	if err == nil {
		t.Fatalf("expected error getting removed storage domain, got nil")
	}
}

func TestStorageDomainDuplicateName(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	hosts, err := client.ListHosts()
	if err != nil {
		t.Fatalf("failed to list hosts (%v)", err)
	}
	hostID := hosts[0].ID()

	sdName := fmt.Sprintf("sd_dup_%s", helper.GenerateRandomID(5))

	sd1, err := client.CreateNFSStorageDomain(sdName, hostID, "nfs.local", "/exports/1", nil)
	if err != nil {
		t.Fatalf("failed to create storage domain (%v)", err)
	}
	t.Cleanup(func() {
		_ = client.RemoveStorageDomain(sd1.ID(), nil)
	})

	_, err = client.CreateNFSStorageDomain(sdName, hostID, "nfs.local", "/exports/2", nil)
	if err == nil {
		t.Fatalf("expected error creating storage domain with duplicate name, got nil")
	}
}
