package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestTemplateBlank(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	tpl, err := helper.GetClient().GetTemplate(ovirtclient.DefaultBlankTemplateID)
	if err != nil {
		if ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
			t.Skipf("Skipping test because the oVirt Engine does not have a factory-default blank template.")
		}
		t.Fatalf("Failed to retrieve factory-default blank template. (%v)", err)
	}
	blank, err := tpl.IsBlank()
	if err != nil {
		t.Fatalf("Failed to check if template is blank (%v).", err)
	}
	if blank != true {
		t.Fatalf("Factory-default blank template is not considered a blank template.")
	}
}
