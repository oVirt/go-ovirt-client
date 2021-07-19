package ovirtclient_test

import (
	"crypto/x509"
	"errors"
	"fmt"
	"testing"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

func TestCredentialChangeAfterSetup(t *testing.T) {
	// Real CA is the CA we will use in the server
	realCAPrivKey, realCACert, realCABytes, err := createCA()
	if err != nil {
		t.Fatalf("failed to create real CA (%v)", err)
	}

	serverPrivKey, serverCert, err := createSignedCert(
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		realCAPrivKey,
		realCACert,
	)
	if err != nil {
		t.Fatalf("failed to create server certificate (%v)", err)
	}

	port := getNextFreePort()

	srv, err := newTestServer(port, serverCert, serverPrivKey, &unauthorizedHandler{})
	if err != nil {
		t.Fatal(err)
	}
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	logger := ovirtclientlog.NewTestLogger(t)
	conn, err := ovirtclient.NewWithVerify(
		fmt.Sprintf("https://127.0.0.1:%d", port),
		"nonexistent@internal",
		"invalid-password-for-testing-purposes",
		"",
		realCABytes,
		false,
		nil,
		logger,
		func(connection *ovirtsdk4.Connection) error {
			// Disable connection check on setup to simulate a credential shift after the connection
			// has been established.
			return nil
		},
	)
	if err != nil {
		t.Fatalf("failed to set up connection (%v)", err)
	}

	_, err = conn.ListStorageDomains()
	if err == nil {
		t.Fatalf("listing storage domains did not result in an error")
	}
	var e ovirtclient.EngineError
	if errors.As(err, &e) {
		if e.Code() != ovirtclient.EAccessDenied {
			t.Fatalf("the returned error was not an EAccessDenied (%v)", err)
		}
	} else {
		t.Fatalf("the returned error was not an EngineError (%v)", err)
	}
}
