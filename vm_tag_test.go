package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestVMTagAssignemnt(t *testing.T) {
	helper := getHelper(t)
	vm := assertCanCreateVM(t, helper, fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)), nil)
	tag1 := assertCanCreateTag(t, helper, fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)), "")
	tag2 := assertCanCreateTag(t, helper, fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)), "")

	assertCanAddTagToVM(t, vm, tag1)
	assertCanAddTagToVM(t, vm, tag2)

	vmTags, err := vm.ListTags()
	if err != nil {
		t.Fatalf("Failed to list VM %s tags (%v).", vm.ID(), err)
	}
	if len(vmTags) != 2 {
		t.Fatalf("Number of tags on VM %s is incorrect (got: %d, expected: %d)", vm.ID(), len(vmTags), 2)
	}
	tag1Found := false
	tag2Found := false
	for _, tag := range vmTags {
		switch tag.ID() {
		case tag1.ID():
			tag1Found = true
		case tag2.ID():
			tag2Found = true
		}
	}
	if !tag1Found {
		t.Fatalf("Tag 1 (%s) was not found on VM %s.", tag1.Name(), vm.ID())
	}
	if !tag2Found {
		t.Fatalf("Tag 2 (%s) was not found on VM %s.", tag2.Name(), vm.ID())
	}

	if err := vm.RemoveTag(tag1.ID()); err != nil {
		t.Fatalf("Failed to remove tag 1 (%s) from VM %s. (%v)", tag1.ID(), vm.ID(), err)
	}

	vmTags, err = vm.ListTags()
	if err != nil {
		t.Fatalf("Failed to list VM %s tags (%v).", vm.ID(), err)
	}
	if len(vmTags) != 1 {
		t.Fatalf("Number of tags on VM %s is incorrect (got: %d, expected: %d)", vm.ID(), len(vmTags), 1)
	}
	if vmTags[0].ID() != tag2.ID() {
		t.Fatalf("Tag mismatch on VM %s (expected: %s, got: %s)", vm.ID(), tag2.ID(), vmTags[0].ID())
	}
}

func assertCanAddTagToVM(t *testing.T, vm ovirtclient.VM, tag ovirtclient.Tag) {
	if err := vm.AddTag(tag.ID()); err != nil {
		t.Fatalf("Failed to add tag %s to VM %s.", tag.ID(), vm.ID())
	}
}
