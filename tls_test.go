package ovirtclient_test

import (
	"fmt"
	"regexp"

	ovirtclient "github.com/ovirt/go-ovirt-client"
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

	// Add system certificates
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
