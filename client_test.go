//
// This file implements The basic test suite for the oVirt client.
//

package govirt_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/janoszen/govirt"
)

func getClient(t *testing.T) govirt.OVirtClient {
	url := os.Getenv("OVIRT_URL")
	if url == "" {
		t.Fatal(fmt.Errorf("the OVIRT_URL environment variable must not be empty"))
	}
	user := os.Getenv("OVIRT_USER")
	if user == "" {
		t.Fatal(fmt.Errorf("the OVIRT_USER environment variable must not be empty"))
	}
	password := os.Getenv("OVIRT_PASSWORD")
	caFile := os.Getenv("OVIRT_CAFILE")
	caCert := os.Getenv("OVIRT_CA_CERT")
	insecure := os.Getenv("OVIRT_INSECURE") != ""

	cli, err := govirt.New(
		url,
		user,
		password,
		caFile,
		[]byte(caCert),
		insecure,
		nil,
		govirt.NewGoTestLogger(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return cli
}
