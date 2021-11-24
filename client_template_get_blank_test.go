package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestTemplateBlank(t *testing.T) {
	helper := getHelper(t)

	tpl, err := helper.GetClient().GetTemplate(ovirtclient.DefaultBlankTemplateID)
	if err != nil {
		if ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
			t.Skipf("Skipping test because the oVirt Engine does not have a factory-default blank template.")
		}
		t.Fatalf("Failed to retrieve factory-default blank template. (%v)", err)
	}
	if tpl.IsBlank() != true {
		t.Fatalf("Factory-default blank template is not considered a blank template.")
	}
}
