package ovirtclient_test

import (
	"crypto/x509"
	"errors"
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

func TestInvalidCredentials(t *testing.T) {
	t.Parallel()
	helper, err := ovirtclient.NewLiveTestHelperFromEnv(ovirtclientlog.NewTestLogger(t))
	if err != nil {
		t.Skipf("🚧 Skipping test: no live credentials provided.")
		return
	}
	url := helper.GetClient().GetURL()
	tls := helper.GetTLS()
	logger := ovirtclientlog.NewTestLogger(t)
	_, err = ovirtclient.New(
		url,
		"nonexistent@internal",
		"invalid-password-for-testing-purposes",
		tls,
		logger,
		nil,
	)
	if err == nil {
		t.Fatal("no error returned from New on invalid credentials")
	}

	var e ovirtclient.EngineError
	if errors.As(err, &e) {
		if e.Code() != ovirtclient.EAccessDenied {
			t.Fatalf("the returned error was not an EAccessDenied error (%v)", err)
		}
	} else {
		t.Fatalf("the returned error was not an EngineError (%v)", err)
	}
}

func TestBadURL(t *testing.T) {
	t.Parallel()
	logger := ovirtclientlog.NewTestLogger(t)
	_, err := ovirtclient.New(
		"https://example.com",
		"nonexistent@internal",
		"invalid-password-for-testing-purposes",
		ovirtclient.TLS().Insecure(),
		logger,
		nil,
	)
	if err == nil {
		t.Fatal("no error returned from New on invalid URL")
	}

	var e ovirtclient.EngineError
	if errors.As(err, &e) {
		if e.Code() != ovirtclient.ENotAnOVirtEngine {
			t.Fatalf("the returned error was not an ENotAnOVirtEngine (%v)", err)
		}
	} else {
		t.Fatalf("the returned error was not an EngineError (%v)", err)
	}
}

func TestBadTLS(t *testing.T) {
	t.Parallel()
	// False CA is the CA we will give to the client
	_, _, falseCACertBytes, err := createCA()
	if err != nil {
		t.Fatalf("failed to create false CA (%v)", err)
	}

	// Real CA is the CA we will use in the server
	realCAPrivKey, realCACert, _, err := createCA()
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

	srv, err := newTestServer(port, serverCert, serverPrivKey, &noopHandler{})
	if err != nil {
		t.Fatal(err)
	}
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	logger := ovirtclientlog.NewTestLogger(t)
	_, err = ovirtclient.New(
		fmt.Sprintf("https://127.0.0.1:%d", port),
		"nonexistent@internal",
		"invalid-password-for-testing-purposes",
		ovirtclient.TLS().CACertsFromMemory(falseCACertBytes),
		logger,
		nil,
	)

	if err == nil {
		t.Fatal("no error returned from New on invalid URL")
	}

	var e ovirtclient.EngineError
	if errors.As(err, &e) {
		if e.Code() != ovirtclient.ETLSError {
			t.Fatalf("the returned error was not an ETLSError (%v)", err)
		}
	} else {
		t.Fatalf("the returned error was not an EngineError (%v)", err)
	}
}
