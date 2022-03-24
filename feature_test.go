package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestFeatureFlags(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	supported, err := helper.GetClient().SupportsFeature(ovirtclient.FeatureAutoPinning)
	if err != nil {
		t.Fatalf("Failed to check autopinning support (%v)", err)
	}
	if supported {
		t.Logf("Autopinning is supported.")
	} else {
		t.Logf("Autopinning is not supported.")
	}
}
