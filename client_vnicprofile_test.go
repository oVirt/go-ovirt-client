package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestVNICProfile(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	networks, err := client.ListNetworks()
	if err != nil {
		t.Fatalf("failed to list networks (%v)", err)
	}
	if len(networks) == 0 {
		t.Fatalf("no networks found")
	}

	vnicProfile, err := client.CreateVNICProfile("test", networks[0].ID(), ovirtclient.CreateVNICProfileParams())
	if err != nil {
		t.Fatalf("failed to create VNIC profile (%v)", err)
	}
	if err := vnicProfile.Remove(); err != nil {
		t.Fatalf("failed to remove VNIC profile (%v)", err)
	}
}
