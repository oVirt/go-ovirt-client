package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
)

func getHelper(t *testing.T) ovirtclient.TestHelper {
	return ovirtclient.NewTestHelperFromEnv(ovirtclientlog.NewTestLogger(t))
}
