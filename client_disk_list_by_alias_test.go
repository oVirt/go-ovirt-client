package ovirtclient_test

import (
	"testing"
)

// TestSearchingForNonExistentDiskShouldNotResultInError tests that searching for a non-existent disk (0 results)
// should not result in an error.
func TestSearchingForNonExistentDiskShouldNotResultInError(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	result, err := client.ListDisksByAlias("this-disk-should-not-exist")
	if err != nil {
		t.Fatalf("Searching for a non-existent disk resulted in an error (%v)", err)
	}

	if len(result) != 0 {
		t.Fatalf("Searching for a non-existent disk resulted in a non-zero number of results (%d)", len(result))
	}
}
