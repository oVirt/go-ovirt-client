package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestNetworkCRUD(t *testing.T) {
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

	netName := fmt.Sprintf("net_test_%s", helper.GenerateRandomID(5))
	netDesc := "test description"
	netComment := "test comment"
	vlan := uint(100)

	params, err := ovirtclient.NewCreateNetworkParams().
		WithDescription(netDesc)
	if err != nil {
		t.Fatalf("failed to set description (%v)", err)
	}
	params, err = params.WithComment(netComment)
	if err != nil {
		t.Fatalf("failed to set comment (%v)", err)
	}
	params, err = params.WithVLAN(vlan)
	if err != nil {
		t.Fatalf("failed to set VLAN (%v)", err)
	}

	net, err := client.CreateNetwork(dcID, netName, params)
	if err != nil {
		t.Fatalf("failed to create network (%v)", err)
	}

	t.Cleanup(func() {
		_ = client.RemoveNetwork(net.ID())
	})

	if net.ID() == "" {
		t.Fatalf("network ID is empty")
	}
	if net.Name() != netName {
		t.Fatalf("expected network name %s, got %s", netName, net.Name())
	}
	if net.DatacenterID() != dcID {
		t.Fatalf("expected datacenter ID %s, got %s", dcID, net.DatacenterID())
	}
	if net.Description() == nil || *net.Description() != netDesc {
		t.Fatalf("expected description %s, got %v", netDesc, net.Description())
	}
	if net.Comment() == nil || *net.Comment() != netComment {
		t.Fatalf("expected comment %s, got %v", netComment, net.Comment())
	}
	if net.VLAN() == nil || *net.VLAN() != vlan {
		t.Fatalf("expected VLAN %d, got %v", vlan, net.VLAN())
	}

	// Fetch network
	fetched, err := client.GetNetwork(net.ID())
	if err != nil {
		t.Fatalf("failed to get network (%v)", err)
	}
	if fetched.Name() != netName {
		t.Fatalf("expected fetched network name %s, got %s", netName, fetched.Name())
	}

	// Update network
	updatedName := fmt.Sprintf("%s_updated", netName)
	updatedDesc := "updated description"
	updatedVLAN := uint(200)
	updateParams, err := ovirtclient.NewUpdateNetworkParams().
		WithName(updatedName)
	if err != nil {
		t.Fatalf("failed to set update name (%v)", err)
	}
	updateParams, err = updateParams.WithDescription(updatedDesc)
	if err != nil {
		t.Fatalf("failed to set update description (%v)", err)
	}
	updateParams, err = updateParams.WithVLAN(updatedVLAN)
	if err != nil {
		t.Fatalf("failed to set update VLAN (%v)", err)
	}

	updated, err := client.UpdateNetwork(net.ID(), updateParams)
	if err != nil {
		t.Fatalf("failed to update network (%v)", err)
	}
	if updated.Name() != updatedName {
		t.Fatalf("expected updated name %s, got %s", updatedName, updated.Name())
	}
	if updated.Description() == nil || *updated.Description() != updatedDesc {
		t.Fatalf("expected updated description %s, got %v", updatedDesc, updated.Description())
	}
	if updated.VLAN() == nil || *updated.VLAN() != updatedVLAN {
		t.Fatalf("expected updated VLAN %d, got %v", updatedVLAN, updated.VLAN())
	}

	// Remove network
	if err := client.RemoveNetwork(net.ID()); err != nil {
		t.Fatalf("failed to remove network (%v)", err)
	}

	// Verify removal
	_, err = client.GetNetwork(net.ID())
	if err == nil {
		t.Fatalf("expected error getting removed network, got nil")
	}
}

func TestNetworkDuplicateName(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	datacenters, err := client.ListDatacenters()
	if err != nil {
		t.Fatalf("failed to list datacenters (%v)", err)
	}
	dcID := datacenters[0].ID()

	netName := fmt.Sprintf("net_dup_%s", helper.GenerateRandomID(5))

	net1, err := client.CreateNetwork(dcID, netName, nil)
	if err != nil {
		t.Fatalf("failed to create network (%v)", err)
	}
	t.Cleanup(func() {
		_ = client.RemoveNetwork(net1.ID())
	})

	_, err = client.CreateNetwork(dcID, netName, nil)
	if err == nil {
		t.Fatalf("expected error creating network with duplicate name, got nil")
	}
}

func TestNetworkInvalidVLAN(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	datacenters, err := client.ListDatacenters()
	if err != nil {
		t.Fatalf("failed to list datacenters (%v)", err)
	}
	dcID := datacenters[0].ID()

	_, err = ovirtclient.NewCreateNetworkParams().WithVLAN(5000)
	if err == nil {
		t.Fatalf("expected error setting invalid VLAN 5000, got nil")
	}

	_, err = ovirtclient.NewCreateNetworkParams().WithVLAN(0)
	if err == nil {
		t.Fatalf("expected error setting invalid VLAN 0, got nil")
	}

	netName := fmt.Sprintf("net_vlan_%s", helper.GenerateRandomID(5))
	params := ovirtclient.NewCreateNetworkParams().MustWithVLAN(100)
	net, err := client.CreateNetwork(dcID, netName, params)
	if err != nil {
		t.Fatalf("failed to create network with valid VLAN (%v)", err)
	}
	t.Cleanup(func() {
		_ = client.RemoveNetwork(net.ID())
	})
}
