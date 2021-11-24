package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestTemplateCreation(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test-%s", helper.GenerateRandomID(5)), nil)
	template := assertCanCreateTemplate(t, helper, vm)
	tpl := assertCanGetTemplate(t, helper, template.ID())
	if tpl.ID() != template.ID() {
		t.Fatalf("IDs of the returned template don't match.")
	}
}

func TestTemplateCPU(t *testing.T) {
	helper := getHelper(t)

	vm1 := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().MustWithCPUParameters(1, 2, 1),
	)
	tpl := assertCanCreateTemplate(t, helper, vm1)
	if tpl.CPU() == nil {
		t.Fatalf("Template with explicit CPU options returned a nil CPU.")
	}
	if tpl.CPU().Topo() == nil {
		t.Fatalf("Template with explicit CPU options returned a nil topo.")
	}
	if cores := tpl.CPU().Topo().Cores(); cores != vm1.CPU().Topo().Cores() {
		t.Fatalf(
			"Template with explicit CPU options returned the incorrect number of cores (%d instead of %d).",
			cores,
			vm1.CPU().Topo().Cores(),
		)
	}
	if threads := tpl.CPU().Topo().Threads(); threads != vm1.CPU().Topo().Threads() {
		t.Fatalf(
			"Template with explicit CPU options returned the incorrect number of threads (%d instead of %d).",
			threads,
			vm1.CPU().Topo().Threads(),
		)
	}
	if sockets := tpl.CPU().Topo().Sockets(); sockets != vm1.CPU().Topo().Sockets() {
		t.Fatalf(
			"Template with explicit CPU options returned the incorrect number of sockets (%d instead of %d).",
			sockets,
			vm1.CPU().Topo().Sockets(),
		)
	}
	vm2 := assertCanCreateVMFromTemplate(t, helper, "test2", tpl.ID(), nil)
	if vm1.CPU().Topo().Cores() != vm2.CPU().Topo().Cores() {
		t.Fatalf(
			"VM created from template returned incorrect number of cores (%d instead of %d).",
			vm1.CPU().Topo().Cores(),
			vm2.CPU().Topo().Cores(),
		)
	}
	if vm1.CPU().Topo().Threads() != vm2.CPU().Topo().Threads() {
		t.Fatalf(
			"VM created from template returned incorrect number of threads (%d instead of %d).",
			vm1.CPU().Topo().Threads(),
			vm2.CPU().Topo().Threads(),
		)
	}
	if vm1.CPU().Topo().Sockets() != vm2.CPU().Topo().Sockets() {
		t.Fatalf(
			"VM created from template returned incorrect number of sockets (%d instead of %d).",
			vm1.CPU().Topo().Sockets(),
			vm2.CPU().Topo().Sockets(),
		)
	}
}

func assertCanGetTemplate(t *testing.T, helper ovirtclient.TestHelper, id ovirtclient.TemplateID) ovirtclient.Template {
	tpl, err := helper.GetClient().GetTemplate(id)
	if err != nil {
		t.Fatalf("Failed to get template %s. (%v)", id, err)
	}
	tpl, err = tpl.WaitForStatus(ovirtclient.TemplateStatusOK)
	if err != nil {
		t.Fatalf("Failed to wait for template %s to reach \"ok\" status. (%v)", tpl.ID(), err)
	}
	return tpl
}

func assertCanCreateTemplate(t *testing.T, helper ovirtclient.TestHelper, vm ovirtclient.VM) ovirtclient.Template {
	template, err := helper.GetClient().CreateTemplate(
		vm.ID(),
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create template from VM %s (%v)", vm.ID(), err)
	}
	t.Cleanup(func() {
		if err := template.Remove(); err != nil && !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
			t.Fatalf("Failed to clean up template %s after test. (%v)", template.ID(), err)
		}
	})
	return template
}
