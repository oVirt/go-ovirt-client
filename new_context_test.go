package ovirtclient_test

import (
	"context"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

type ctxKey string

func TestContext(t *testing.T) {
	helper := getHelper(t)

	cli1 := helper.GetClient()

	ctx := context.WithValue(context.Background(), ctxKey("foo"), "bar")

	cli2 := cli1.WithContext(ctx)

	if err := cli1.Reconnect(); err != nil {
		t.Fatalf("failed to reconnect (%v)", err)
	}

	if cli1.GetContext() != nil {
		t.Fatalf("Context of client 1 is not nil.")
	}
	if cli2.GetContext() == nil {
		t.Fatalf("Context of client 2 is nil.")
	}
	if cli2.GetContext().Value(ctxKey("foo")) != "bar" {
		t.Fatalf("Incorrect context value on client 2.")
	}

	vm, err := cli1.CreateVM(
		helper.GetClusterID(),
		helper.GetBlankTemplateID(),
		helper.GenerateTestResourceName(t),
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test VM (%v)", err)
	}
	t.Cleanup(
		func() {
			if vm != nil {
				t.Logf("Cleaning up test VM %s...", vm.ID())
				if err := vm.Remove(); err != nil && !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
					t.Fatalf("Failed to remove test VM %s (%v)", vm.ID(), err)
				}
			}
		},
	)

	if _, err := cli2.GetVM(vm.ID()); err != nil {
		t.Fatalf("Failed to fetch VM from secondary connection (%v)", err)
	}
}
