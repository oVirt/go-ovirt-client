package ovirtclient_test

import (
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

func TestInvalidCredentials(t *testing.T) {
	url, tls, err := getConnectionParametersForLiveTesting()
	if err != nil {
		t.Skipf("⚠ Skipping test: no live credentials provided.")
		return
	}
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

func getConnectionParametersForLiveTesting() (string, ovirtclient.TLSProvider, error) {
	url := os.Getenv("OVIRT_URL")
	if url == "" {
		return "", nil, fmt.Errorf("the OVIRT_URL environment variable must not be empty")
	}
	tls := ovirtclient.TLS()
	configured := false
	if caFile := os.Getenv("OVIRT_CAFILE"); caFile != "" {
		configured = true
		tls.CACertsFromFile(caFile)
	}
	if caCert := os.Getenv("OVIRT_CA_CERT"); caCert != "" {
		configured = true
		tls.CACertsFromMemory([]byte(caCert))
	}
	if os.Getenv("OVIRT_INSECURE") != "" {
		configured = true
		tls.Insecure()
	}
	if !configured {
		return "", nil, fmt.Errorf("one of OVIRT_CAFILE, OVIRT_CA_CERT, or OVIRT_INSECURE must be set")
	}
	return url, tls, nil
}
