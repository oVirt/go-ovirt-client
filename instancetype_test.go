package ovirtclient_test

import "testing"

func TestListInstanceTypes(t *testing.T) {
	helper := getHelper(t)
	_, err := helper.GetClient().ListInstanceTypes()
	if err != nil {
		t.Fatalf("failed to list instance types (%v)", err)
	}
}
