package ovirtclient_test

import "testing"

func TestStoppingAlreadyStoppedVM(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(t, helper, helper.GenerateTestResourceName(t), nil)
	if err := vm.Stop(false); err != nil {
		t.Fatalf("Failed to issue stop command on already-stopped VM (%v)", err)
	}
}
