//
// This file implements The basic test suite for the oVirt client.
//

package ovirtclient_test

import (
	"fmt"
	"os"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

func getHelper(t *testing.T) ovirtclient.TestHelper {
	helper, err := getLiveHelper(t)
	if err != nil {
		t.Logf("⚠ Warning: failed to create live helper for tests, falling back to mock backend. (%v)", err)
		return getMockHelper(t)
	}
	return helper
}

func getMockHelper(t *testing.T) ovirtclient.TestHelper {
	helper, err := ovirtclient.NewTestHelper(
		"https://localhost/ovirt-engine/api",
		"admin@internal",
		"",
		ovirtclient.TLS().Insecure(),
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
	url, tls, err := getConnectionParametersForLiveTesting()
	if err != nil {
		return nil, err
	}
	user := os.Getenv("OVIRT_USERNAME")
	if user == "" {
		return nil, fmt.Errorf("the OVIRT_USER environment variable must not be empty")
	}
	password := os.Getenv("OVIRT_PASSWORD")
	clusterID := os.Getenv("OVIRT_CLUSTER_ID")
	blankTemplateID := os.Getenv("OVIRT_BLANK_TEMPLATE_ID")
	storageDomainID := os.Getenv("OVIRT_STORAGE_DOMAIN_ID")

	helper, err := ovirtclient.NewTestHelper(
		url,
		user,
		password,
		tls,
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
