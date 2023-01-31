package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestVNICProfile(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	networks, err := client.ListNetworks()
	if err != nil {
		t.Fatalf("failed to list networks (%v)", err)
	}
	if len(networks) == 0 {
		t.Fatalf("no networks found")
	}

	vnicProfile, err := client.CreateVNICProfile(
		fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5)),
		networks[0].ID(),
		ovirtclient.CreateVNICProfileParams(),
	)
	if err != nil {
		t.Fatalf("failed to create VNIC profile (%v)", err)
	}
	if err := vnicProfile.Remove(); err != nil {
		t.Fatalf("failed to remove VNIC profile (%v)", err)
	}
}

func assertCanCreateVNICProfile(t *testing.T, helper ovirtclient.TestHelper) ovirtclient.VNICProfile {
	client := helper.GetClient()
	vnicProfile, err := client.GetVNICProfile(helper.GetVNICProfileID())
	if err != nil {
		t.Fatalf("failed to fetch test VNIC profile (%v)", err)
	}
	newVNICProfile, err := client.CreateVNICProfile(
		fmt.Sprintf("client_test_%s", helper.GenerateRandomID(5)),
		vnicProfile.NetworkID(),
		ovirtclient.CreateVNICProfileParams(),
	)
	if err != nil {
		t.Fatalf("failed to create test VNIC profile (%v)", err)
	}
	t.Cleanup(
		func() {
			if err := newVNICProfile.Remove(); err != nil {
				t.Fatalf("failed to clean up test VNIC profile ID %s (%v)", newVNICProfile.ID(), err)
			}
		})
	return newVNICProfile
}
