package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestDatacenterCRUD(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	dcName := fmt.Sprintf("dc_test_%s", helper.GenerateRandomID(5))
	dcDesc := "test datacenter description"
	dcComment := "test datacenter comment"
	local := false
	quotaMode := ovirtclient.QuotaModeAudit

	params, err := ovirtclient.NewCreateDatacenterParams().
		WithDescription(dcDesc)
	if err != nil {
		t.Fatalf("failed to set description (%v)", err)
	}
	params, err = params.WithComment(dcComment)
	if err != nil {
		t.Fatalf("failed to set comment (%v)", err)
	}
	params, err = params.WithLocal(local)
	if err != nil {
		t.Fatalf("failed to set local (%v)", err)
	}
	params, err = params.WithQuotaMode(quotaMode)
	if err != nil {
		t.Fatalf("failed to set quota mode (%v)", err)
	}

	dc, err := client.CreateDatacenter(dcName, params)
	if err != nil {
		t.Fatalf("failed to create datacenter (%v)", err)
	}

	t.Cleanup(func() {
		_ = client.RemoveDatacenter(dc.ID())
	})

	if dc.ID() == "" {
		t.Fatalf("datacenter ID is empty")
	}
	if dc.Name() != dcName {
		t.Fatalf("expected datacenter name %s, got %s", dcName, dc.Name())
	}
	if dc.Description() == nil || *dc.Description() != dcDesc {
		t.Fatalf("expected description %s, got %v", dcDesc, dc.Description())
	}
	if dc.Comment() == nil || *dc.Comment() != dcComment {
		t.Fatalf("expected comment %s, got %v", dcComment, dc.Comment())
	}
	if dc.Local() != local {
		t.Fatalf("expected local %v, got %v", local, dc.Local())
	}
	if dc.QuotaMode() == nil || *dc.QuotaMode() != quotaMode {
		t.Fatalf("expected quota mode %v, got %v", quotaMode, dc.QuotaMode())
	}

	// Fetch datacenter
	fetched, err := client.GetDatacenter(dc.ID())
	if err != nil {
		t.Fatalf("failed to get datacenter (%v)", err)
	}
	if fetched.Name() != dcName {
		t.Fatalf("expected fetched datacenter name %s, got %s", dcName, fetched.Name())
	}

	// Update datacenter
	updatedName := fmt.Sprintf("%s_updated", dcName)
	updatedDesc := "updated description"
	updatedComment := "updated comment"
	updatedQuotaMode := ovirtclient.QuotaModeDisabled

	updateParams, err := ovirtclient.NewUpdateDatacenterParams().
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
	updateParams, err = updateParams.WithQuotaMode(updatedQuotaMode)
	if err != nil {
		t.Fatalf("failed to set update quota mode (%v)", err)
	}

	updated, err := client.UpdateDatacenter(dc.ID(), updateParams)
	if err != nil {
		t.Fatalf("failed to update datacenter (%v)", err)
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
	if updated.QuotaMode() == nil || *updated.QuotaMode() != updatedQuotaMode {
		t.Fatalf("expected updated quota mode %v, got %v", updatedQuotaMode, updated.QuotaMode())
	}

	// Remove datacenter
	if err := client.RemoveDatacenter(dc.ID()); err != nil {
		t.Fatalf("failed to remove datacenter (%v)", err)
	}

	// Verify removal
	_, err = client.GetDatacenter(dc.ID())
	if err == nil {
		t.Fatalf("expected error getting removed datacenter, got nil")
	}
}

func TestDatacenterDuplicateName(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	dcName := fmt.Sprintf("dc_dup_%s", helper.GenerateRandomID(5))

	dc1, err := client.CreateDatacenter(dcName, nil)
	if err != nil {
		t.Fatalf("failed to create datacenter (%v)", err)
	}
	t.Cleanup(func() {
		_ = client.RemoveDatacenter(dc1.ID())
	})

	_, err = client.CreateDatacenter(dcName, nil)
	if err == nil {
		t.Fatalf("expected error creating datacenter with duplicate name, got nil")
	}
}

func TestDatacenterCannotRemoveWithClusters(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	datacenters, err := client.ListDatacenters()
	if err != nil {
		t.Fatalf("failed to list datacenters (%v)", err)
	}
	if len(datacenters) == 0 {
		t.Fatalf("no datacenters found")
	}
	// Default datacenter has a cluster attached in mock
	err = client.RemoveDatacenter(datacenters[0].ID())
	if err == nil {
		t.Fatalf("expected error removing datacenter with attached clusters, got nil")
	}
}
