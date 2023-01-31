package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestVMIPAddressReporting(t *testing.T) {
	helper := getHelper(t)

	disk := assertCanCreateDiskWithParameters(t, helper, ovirtclient.ImageFormatCow, nil)
	assertCanUploadFullyFunctionalDiskImage(t, helper, disk)
	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().MustWithMemory(512*1024*1024),
	)
	assertCanAttachDisk(t, vm, disk)
	assertCanCreateNIC(t, helper, vm, fmt.Sprintf("%s-%s", t.Name(), "eth0"), nil)
	assertCanStartVM(t, helper, vm)
	assertVMWillStart(t, vm)
	assertVMGetsIPAddress(t, vm)
}

func assertVMGetsIPAddress(t *testing.T, vm ovirtclient.VM) {
	result, err := vm.WaitForNonLocalIPAddress()
	if err != nil {
		t.Fatalf("failed to wait for non-local IP addresses on VM %s (%v)", vm.ID(), err)
	}
	if len(result) == 0 {
		t.Fatalf("no IP address returned from VM %s", vm.ID())
	}
	foundIP := false
loop:
	for interf, ips := range result {
		for _, ip := range ips {
			if ip.IsGlobalUnicast() {
				foundIP = true
				t.Logf("Found valid IP address %s on interface %s on VM %s", ip.String(), interf, vm.ID())
				break loop
			}
		}
	}
	if !foundIP {
		t.Fatalf("VM %s has no valid IP address", vm.ID())
	}
}
