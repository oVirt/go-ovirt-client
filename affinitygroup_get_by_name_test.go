package ovirtclient_test

import "testing"

func TestGetAGByName(t *testing.T) {
	helper := getHelper(t)

	// Create a dummy AG to test if the real AG is actually returned
	assertCanCreateAffinityGroup(t, helper, nil)
	// Create the AG we are looking for
	ag := assertCanCreateAffinityGroup(t, helper, nil)
	// Fetch the AG
	ag2, err := helper.GetClient().GetAffinityGroupByName(ag.ClusterID(), ag.Name())
	if err != nil {
		t.Fatalf("failed to fetch affinity group by name (%v)", err)
	}
	if ag.ID() != ag2.ID() {
		t.Fatalf("Affinity group ID mismatch after fetching by name (expected: %s, got: %s)", ag.ID(), ag2.ID())
	}
}
