package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestTemplateCreation(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(t, helper, "test", nil)
	template := assertCanCreateTemplate(t, helper, vm)
	tpl := assertCanGetTemplate(t, helper, template.ID())
	if tpl.ID() != template.ID() {
		t.Fatalf("IDs of the returned template don't match.")
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
	template, err := helper.GetClient().CreateTemplate(vm.ID(), "test", nil)
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
