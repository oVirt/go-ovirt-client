package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestListAndRemoveGraphicsConsoles(t *testing.T) {
	helper := getHelper(t)
	vm := assertCanCreateVM(
		t,
		helper,
		helper.GenerateTestResourceName(t),
		ovirtclient.NewCreateVMParams().MustWithVMType(ovirtclient.VMTypeDesktop),
	)
	graphicsConsoles, err := vm.ListGraphicsConsoles()
	if err != nil {
		t.Fatalf("Failed to list graphics consoles on VM %s (%v)", vm.ID(), err)
	}
	if len(graphicsConsoles) == 0 {
		t.Fatalf("No graphics consoles found on desktop VM.")
	}
	for _, graphicsConsole := range graphicsConsoles {
		if err := graphicsConsole.Remove(); err != nil {
			t.Fatalf("failed to remove graphics console %s from VM %s (%v)", graphicsConsole.ID(), vm.ID(), err)
		}
	}
	newGraphicsConsoles, err := helper.GetClient().ListVMGraphicsConsoles(vm.ID())
	if err != nil {
		t.Fatalf("Failed to list graphics consoles for VM %s (%v)", vm.ID(), err)
	}
	if len(newGraphicsConsoles) != 0 {
		t.Fatalf("Still found graphics consoles after removing them.")
	}
}
