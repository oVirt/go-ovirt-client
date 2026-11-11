package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestClusterCRUD(t *testing.T) {
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
	dcID := datacenters[0].ID()

	clusterName := fmt.Sprintf("cluster_test_%s", helper.GenerateRandomID(5))
	cpuType := "Intel SandyBridge Family"
	cpuArch := "x86_64"
	clusterDesc := "test cluster description"
	clusterComment := "test cluster comment"
	ballooning := true
	overcommit := int64(150)

	params, err := ovirtclient.NewCreateClusterParams().
		WithDescription(clusterDesc)
	if err != nil {
		t.Fatalf("failed to set description (%v)", err)
	}
	params, err = params.WithComment(clusterComment)
	if err != nil {
		t.Fatalf("failed to set comment (%v)", err)
	}
	params, err = params.WithCPUArchitecture(cpuArch)
	if err != nil {
		t.Fatalf("failed to set CPU architecture (%v)", err)
	}
	params, err = params.WithBallooning(ballooning)
	if err != nil {
		t.Fatalf("failed to set ballooning (%v)", err)
	}
	params, err = params.WithMemoryOverCommitPercent(overcommit)
	if err != nil {
		t.Fatalf("failed to set memory overcommit percent (%v)", err)
	}

	c, err := client.CreateCluster(dcID, clusterName, cpuType, params)
	if err != nil {
		t.Fatalf("failed to create cluster (%v)", err)
	}

	t.Cleanup(func() {
		_ = client.RemoveCluster(c.ID())
	})

	if c.ID() == "" {
		t.Fatalf("cluster ID is empty")
	}
	if c.Name() != clusterName {
		t.Fatalf("expected cluster name %s, got %s", clusterName, c.Name())
	}
	if c.DatacenterID() != dcID {
		t.Fatalf("expected datacenter ID %s, got %s", dcID, c.DatacenterID())
	}
	if c.CPUType() != cpuType {
		t.Fatalf("expected CPU type %s, got %s", cpuType, c.CPUType())
	}
	if c.CPUArchitecture() == nil || *c.CPUArchitecture() != cpuArch {
		t.Fatalf("expected CPU architecture %s, got %v", cpuArch, c.CPUArchitecture())
	}
	if c.Description() == nil || *c.Description() != clusterDesc {
		t.Fatalf("expected description %s, got %v", clusterDesc, c.Description())
	}
	if c.Comment() == nil || *c.Comment() != clusterComment {
		t.Fatalf("expected comment %s, got %v", clusterComment, c.Comment())
	}
	if c.Ballooning() == nil || *c.Ballooning() != ballooning {
		t.Fatalf("expected ballooning %v, got %v", ballooning, c.Ballooning())
	}
	if c.MemoryOverCommitPercent() == nil || *c.MemoryOverCommitPercent() != overcommit {
		t.Fatalf("expected memory overcommit percent %d, got %v", overcommit, c.MemoryOverCommitPercent())
	}

	// Fetch cluster
	fetched, err := client.GetCluster(c.ID())
	if err != nil {
		t.Fatalf("failed to get cluster (%v)", err)
	}
	if fetched.Name() != clusterName {
		t.Fatalf("expected fetched cluster name %s, got %s", clusterName, fetched.Name())
	}

	// Update cluster
	updatedName := fmt.Sprintf("%s_updated", clusterName)
	updatedCPUType := "Intel Haswell-noTSX Family"
	updatedDesc := "updated description"
	updatedComment := "updated comment"
	updatedBallooning := false
	updatedOvercommit := int64(200)

	updateParams, err := ovirtclient.NewUpdateClusterParams().
		WithName(updatedName)
	if err != nil {
		t.Fatalf("failed to set update name (%v)", err)
	}
	updateParams, err = updateParams.WithCPUType(updatedCPUType)
	if err != nil {
		t.Fatalf("failed to set update CPU type (%v)", err)
	}
	updateParams, err = updateParams.WithDescription(updatedDesc)
	if err != nil {
		t.Fatalf("failed to set update description (%v)", err)
	}
	updateParams, err = updateParams.WithComment(updatedComment)
	if err != nil {
		t.Fatalf("failed to set update comment (%v)", err)
	}
	updateParams, err = updateParams.WithBallooning(updatedBallooning)
	if err != nil {
		t.Fatalf("failed to set update ballooning (%v)", err)
	}
	updateParams, err = updateParams.WithMemoryOverCommitPercent(updatedOvercommit)
	if err != nil {
		t.Fatalf("failed to set update memory overcommit percent (%v)", err)
	}

	updated, err := client.UpdateCluster(c.ID(), updateParams)
	if err != nil {
		t.Fatalf("failed to update cluster (%v)", err)
	}
	if updated.Name() != updatedName {
		t.Fatalf("expected updated name %s, got %s", updatedName, updated.Name())
	}
	if updated.CPUType() != updatedCPUType {
		t.Fatalf("expected updated CPU type %s, got %s", updatedCPUType, updated.CPUType())
	}
	if updated.Description() == nil || *updated.Description() != updatedDesc {
		t.Fatalf("expected updated description %s, got %v", updatedDesc, updated.Description())
	}
	if updated.Comment() == nil || *updated.Comment() != updatedComment {
		t.Fatalf("expected updated comment %s, got %v", updatedComment, updated.Comment())
	}
	if updated.Ballooning() == nil || *updated.Ballooning() != updatedBallooning {
		t.Fatalf("expected updated ballooning %v, got %v", updatedBallooning, updated.Ballooning())
	}
	if updated.MemoryOverCommitPercent() == nil || *updated.MemoryOverCommitPercent() != updatedOvercommit {
		t.Fatalf("expected updated memory overcommit percent %d, got %v", updatedOvercommit, updated.MemoryOverCommitPercent())
	}

	// Remove cluster
	if err := client.RemoveCluster(c.ID()); err != nil {
		t.Fatalf("failed to remove cluster (%v)", err)
	}

	// Verify removal
	_, err = client.GetCluster(c.ID())
	if err == nil {
		t.Fatalf("expected error getting removed cluster, got nil")
	}
}

func TestClusterDuplicateName(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	datacenters, err := client.ListDatacenters()
	if err != nil {
		t.Fatalf("failed to list datacenters (%v)", err)
	}
	dcID := datacenters[0].ID()

	clusterName := fmt.Sprintf("cluster_dup_%s", helper.GenerateRandomID(5))

	c1, err := client.CreateCluster(dcID, clusterName, "Intel SandyBridge Family", nil)
	if err != nil {
		t.Fatalf("failed to create cluster (%v)", err)
	}
	t.Cleanup(func() {
		_ = client.RemoveCluster(c1.ID())
	})

	_, err = client.CreateCluster(dcID, clusterName, "Intel SandyBridge Family", nil)
	if err == nil {
		t.Fatalf("expected error creating cluster with duplicate name, got nil")
	}
}

func TestClusterCannotRemoveWithHosts(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	clusters, err := client.ListClusters()
	if err != nil {
		t.Fatalf("failed to list clusters (%v)", err)
	}
	if len(clusters) == 0 {
		t.Fatalf("no clusters found")
	}

	// Default cluster has a host attached in mock
	err = client.RemoveCluster(clusters[0].ID())
	if err == nil {
		t.Fatalf("expected error removing cluster with attached hosts, got nil")
	}
}
