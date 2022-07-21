package ovirtclient_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

var (
	nextFreePortLock = &sync.Mutex{} // nolint:gochecknoglobals
)

func getNextFreePort(t *testing.T) int {
	t.Helper()

	nextFreePortLock.Lock()
	defer nextFreePortLock.Unlock()

	listenConfig := net.ListenConfig{}
	ctx := context.Background()

	if deadline, ok := t.Deadline(); ok {
		var cancel func()
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	l, err := listenConfig.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to allocate port for test %s (%v)", t.Name(), err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err = l.Close(); err != nil {
		t.Fatalf(
			"Failed to close temporary listen socket on port %d for test %s (%v)",
			port,
			t.Name(),
			err,
		)
	}

	t.Logf("Allocating port %d for test %s", port, t.Name())
	return port
}

func createCA() (*rsa.PrivateKey, *x509.Certificate, []byte, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ACME, Inc"},
			Country:      []string{"US"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create private key (%w)", err)
	}
	caCert, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create CA certificate (%w)", err)
	}
	caPEM := new(bytes.Buffer)
	if err := pem.Encode(
		caPEM,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caCert,
		},
	); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encode CA cert (%w)", err)
	}
	return caPrivateKey, ca, caPEM.Bytes(), nil
}

func createSignedCert(usage []x509.ExtKeyUsage, caPrivateKey *rsa.PrivateKey, caCertificate *x509.Certificate) (
	[]byte,
	[]byte,
	error,
) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"ACME, Inc"},
			Country:      []string{"US"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 0, 1),
		SubjectKeyId: []byte{1},
		ExtKeyUsage:  usage,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		cert,
		caCertificate,
		&certPrivKey.PublicKey,
		caPrivateKey,
	)
	if err != nil {
		return nil, nil, err
	}
	certPrivKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	}); err != nil {
		return nil, nil, err
	}
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM,
		&pem.Block{Type: "CERTIFICATE", Bytes: certBytes},
	); err != nil {
		return nil, nil, err
	}
	return certPrivKeyPEM.Bytes(), certPEM.Bytes(), nil
}

func newTestServer(t *testing.T, port int, serverCert []byte, serverPrivKey []byte, handler http.Handler) (
	*testServer,
	error,
) {
	cert, err := tls.X509KeyPair(serverCert, serverPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create key pair (%w)", err)
	}
	//nolint:gosec
	srv := &http.Server{
		Addr:     fmt.Sprintf("127.0.0.1:%d", port),
		Handler:  handler,
		ErrorLog: testLogger(t),
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			Certificates: []tls.Certificate{
				cert,
			},
			MinVersion: tls.VersionTLS12,
		},
	}
	return &testServer{
		srv:  srv,
		port: port,
	}, nil
}

type testServer struct {
	srv  *http.Server
	port int
}

func (t *testServer) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", t.port))
	if err != nil {
		return fmt.Errorf("failed to start test server (%w)", err)
	}

	go func() {
		_ = t.srv.ServeTLS(ln, "", "")
	}()
	return nil
}

func (t *testServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_ = t.srv.Shutdown(ctx)
}

type noopHandler struct{}

func (t *noopHandler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(200)
}

type unauthorizedHandler struct{}

func (u *unauthorizedHandler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(401)
	_, _ = writer.Write([]byte("<html><head><title>Error</title></head><body>Unauthorized</body></html>"))
}
