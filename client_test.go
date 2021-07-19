//
// This file implements The basic test suite for the oVirt client.
//

package ovirtclient_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

func getHelper(t *testing.T) ovirtclient.TestHelper {
	helper, err := getLiveHelper(t)
	if err != nil {
		t.Logf("⚠ Warning: failed to create live helper for tests, falling back to mock backend.")
		return getMockHelper(t)
	}
	return helper
}

func getMockHelper(t *testing.T) ovirtclient.TestHelper {
	helper, err := ovirtclient.NewTestHelper(
		"https://localhost/ovirt-engine/api",
		"admin@internal",
		"",
		"",
		nil,
		true,
		"",
		"",
		"",
		true,
		ovirtclientlog.NewTestLogger(t),
	)
	if err != nil {
		panic(err)
	}
	return helper
}

func getLiveHelper(t *testing.T) (ovirtclient.TestHelper, error) {
	url := os.Getenv("OVIRT_URL")
	if url == "" {
		return nil, fmt.Errorf("the OVIRT_URL environment variable must not be empty")
	}
	user := os.Getenv("OVIRT_USERNAME")
	if user == "" {
		return nil, fmt.Errorf("the OVIRT_USER environment variable must not be empty")
	}
	password := os.Getenv("OVIRT_PASSWORD")
	caFile := os.Getenv("OVIRT_CAFILE")
	caCert := os.Getenv("OVIRT_CA_CERT")
	insecure := os.Getenv("OVIRT_INSECURE") != ""
	if caFile == "" && caCert == "" && !insecure {
		return nil, fmt.Errorf("one of OVIRT_CAFILE, OVIRT_CA_CERT, or OVIRT_INSECURE must be set")
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
		false,
		ovirtclientlog.NewTestLogger(t),
	)
	if err != nil {
		return nil, err
	}
	return helper, nil
}
