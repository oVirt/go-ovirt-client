package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestVMSearch(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	name1 := helper.GenerateRandomID(5)
	name2 := helper.GenerateRandomID(5)
	vm1 := assertCanCreateVM(t, helper, name1, nil)
	_ = assertCanCreateVM(t, helper, name2, nil)
	vms, err := client.SearchVMs(ovirtclient.VMSearchParams().WithName(name1))
	if err != nil {
		t.Fatalf("Failed to search for VM (%v)", err)
	}
	if len(vms) != 1 {
		t.Fatalf("Incorrect number of VMs returned (%d)", len(vms))
	}
	if vms[0].ID() != vm1.ID() {
		t.Fatalf("Incorrect VM returned: %s", vms[0].ID())
	}
}
