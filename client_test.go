//
// This file implements The basic test suite for the oVirt client.
//

package ovirtclient_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/oVirt/go-ovirt-client"
)

func getHelper(t *testing.T) ovirtclient.TestHelper {
	url := os.Getenv("OVIRT_URL")
	if url == "" {
		t.Fatal(fmt.Errorf("the OVIRT_URL environment variable must not be empty"))
	}
	user := os.Getenv("OVIRT_USERNAME")
	if user == "" {
		t.Fatal(fmt.Errorf("the OVIRT_USER environment variable must not be empty"))
	}
	password := os.Getenv("OVIRT_PASSWORD")
	caFile := os.Getenv("OVIRT_CAFILE")
	caCert := os.Getenv("OVIRT_CA_CERT")
	insecure := os.Getenv("OVIRT_INSECURE") != ""
	if caFile == "" && caCert == "" && !insecure {
		t.Fatal(fmt.Errorf("one of OVIRT_CAFILE, OVIRT_CA_CERT, or OVIRT_INSECURE must be set"))
	}
	clusterID := os.Getenv("OVIRT_CLUSTER_ID")
	blankTemplateID := os.Getenv("OVIRT_BLANK_TEMPLATE_ID")
	storageDomainID := os.Getenv("OVIRT_STORAGE_DOMAIN_ID")

	helper, err := ovirtclient.NewTestHelper(
		url,
		user,
		password,
		caFile,
		[]byte(caCert),
		insecure,
		clusterID,
		blankTemplateID,
		storageDomainID,
		ovirtclient.NewGoTestLogger(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return helper
}
