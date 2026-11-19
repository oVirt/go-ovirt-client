package ovirtclient_test

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

// This example shows how to set up TLS verification in a variety of ways.
func ExampleTLS() {
	tls := ovirtclient.TLS()

	// Add certificates from an in-memory byte slice. Certificates must be in PEM format.
	tls.CACertsFromMemory([]byte("-----BEGIN CERTIFICATE-----\n..."))

	// Add certificates from a single file. Certificates must be in PEM format.
	tls.CACertsFromFile("/path/to/file.pem")

	// Add certificates from a directory. Optionally, regular expressions can be passed that must match the file
	// names.
	tls.CACertsFromDir("/path/to/certs", regexp.MustCompile(`\.pem`))

	// Add system certificates. This does not work on Windows before Go 1.18.
	tls.CACertsFromSystem()

	// Disable certificate verification. This is a bad idea.
	tls.Insecure()

	// This will typically be called by the ovirtclient.New() function to create a TLS certificate.
	tlsConfig, err := tls.CreateTLSConfig()
	if err != nil {
		panic(fmt.Errorf("failed to create TLS config (%w)", err))
	}
	if tlsConfig.InsecureSkipVerify {
		fmt.Printf("Certificate verification is disabled.")
	} else {
		fmt.Printf("Certificate verification is enabled.")
	}
	// Output: Certificate verification is disabled.
}

// This example shows how to set up TLS verification from an existing certificate pool.
func ExampleBuildableTLSProvider_CACertsFromCertPool() {
	tls := ovirtclient.TLS()

	// Add custom certificate pool as a source of certificates.
	certPool := x509.NewCertPool()
	tls.CACertsFromCertPool(certPool)

	// This will typically be called by the ovirtclient.New() function to create a TLS certificate.
	tlsConfig, err := tls.CreateTLSConfig()
	if err != nil {
		panic(fmt.Errorf("failed to create TLS config (%w)", err))
	}
	if tlsConfig.InsecureSkipVerify {
		fmt.Printf("Certificate verification is disabled.")
	} else {
		fmt.Printf("Certificate verification is enabled.")
	}
	// Output: Certificate verification is enabled.
}

func TestCACertsFromDirWithDotfilesAndNonPEM(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Add a dotfile with a distinct valid CA cert (like .hidden-ca.pem) to verify dotfiles are excluded
	_, _, hiddenCAPEM, err := createCA()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden-ca.pem"), hiddenCAPEM, 0644); err != nil {
		t.Fatal(err)
	}

	// Add a non-PEM file (like README.md)
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("This is a cert directory"), 0644); err != nil {
		t.Fatal(err)
	}

	// Add a valid CA cert
	_, _, caPEM, err := createCA()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "ca.pem"), caPEM, 0644); err != nil {
		t.Fatal(err)
	}

	logger := ovirtclientlog.NewTestLogger(t)
	tls := ovirtclient.TLSWithLogger(logger)
	tls.CACertsFromDir(tmpDir)

	tlsConfig, err := tls.CreateTLSConfig()
	if err != nil {
		t.Fatalf("unexpected error creating TLS config with dotfiles and non-pem files: %v", err)
	}
	if tlsConfig.RootCAs == nil {
		t.Fatal("expected RootCAs to be populated")
	}
	expectedPool := x509.NewCertPool()
	if !expectedPool.AppendCertsFromPEM(caPEM) {
		t.Fatal("failed to append caPEM to expectedPool")
	}
	if !expectedPool.Equal(tlsConfig.RootCAs) {
		t.Fatal("expected RootCAs to match generated CA certificate pool")
	}
}

func TestCACertsFromDirEmptyOrNoValidCerts(t *testing.T) {
	t.Parallel()

	emptyDir := t.TempDir()
	if _, err := ovirtclient.TLS().CACertsFromDir(emptyDir).CreateTLSConfig(); err == nil {
		t.Fatal("expected error creating TLS config from empty directory")
	}

	nonPEMDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(nonPEMDir, "README.md"), []byte("not a cert"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ovirtclient.TLS().CACertsFromDir(nonPEMDir).CreateTLSConfig(); err == nil {
		t.Fatal("expected error creating TLS config when all files in directory are non-PEM")
	}
}
