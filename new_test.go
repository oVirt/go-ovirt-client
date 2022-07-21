package ovirtclient_test

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
)

func TestBadOVirtURL(t *testing.T) {
	helper, err := ovirtclient.NewLiveTestHelperFromEnv(ovirtclientlog.NewTestLogger(t))
	if err != nil {
		t.Skipf("🚧 Skipping test: no live credentials provided.")
		return
	}
	url := helper.GetClient().GetURL()
	tls := helper.GetTLS()
	username := helper.GetUsername()
	password := helper.GetPassword()

	logger := ovirtclientlog.NewTestLogger(t)
	_, err = ovirtclient.New(
		strings.TrimSuffix(strings.TrimSuffix(url, "/api"), "/api/"),
		username,
		password,
		tls,
		logger,
		nil,
	)
	if err == nil {
		t.Fatalf("Creating a connection to an endpoint not ending in /api did not result in an error.")
	}
	if !ovirtclient.HasErrorCode(err, ovirtclient.ENotAnOVirtEngine) {
		t.Fatalf("Creating a connection to an endpoint not ending in /api has not correctly resulted in an ENotAnOVirtEngine (%v)", err)
	}
}

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

	port := getNextFreePort(t)

	srv, err := newTestServer(t, port, serverCert, serverPrivKey, &noopHandler{})
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

func TestCredentialChangeAfterSetup(t *testing.T) {
	t.Parallel()
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

	port := getNextFreePort(t)

	srv, err := newTestServer(t, port, serverCert, serverPrivKey, &unauthorizedHandler{})
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
		ovirtclient.TLS().CACertsFromMemory(realCABytes),
		logger,
		nil,
		func(connection ovirtclient.Client) error {
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

func TestProxy(t *testing.T) {
	counter := 0
	proxy := startProxyServer(t, &counter)
	t.Parallel()
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

	port := getNextFreePort(t)

	srv, err := newTestServer(t, port, serverCert, serverPrivKey, &unauthorizedHandler{})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(srv.Stop)

	logger := ovirtclientlog.NewTestLogger(t)
	_, err = ovirtclient.New(
		fmt.Sprintf("https://127.0.0.1:%d", port),
		"admin@internal",
		"asdf",
		ovirtclient.TLS().CACertsFromMemory(realCABytes),
		logger,
		ovirtclient.NewExtraSettings().WithProxy(fmt.Sprintf("http://%s", proxy)),
	)
	if err == nil {
		t.Fatalf("No error on establishing a connection (%v)", err)
	}
	if !ovirtclient.HasErrorCode(err, ovirtclient.EAccessDenied) {
		t.Fatalf("incorrect error code returned: %v", err)
	}
	if counter == 0 {
		t.Fatalf("Connection did not go through the proxy.")
	}
}

func startProxyServer(t *testing.T, counter *int) string {
	wg := sync.WaitGroup{}
	wg.Add(1)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to open listen socket for the proxy (%v)", err)
	}
	var serverError error
	//nolint:gosec
	srv := http.Server{
		Addr: ln.Addr().String(),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	srv.Handler = &proxyHandler{counter: counter}
	go func() {
		defer wg.Done()
		if err := srv.Serve(ln); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				serverError = err
			}
		}
	}()
	t.Cleanup(
		func() {
			if err := srv.Close(); err != nil {
				t.Fatalf("failed to close listen socket (%v)", err)
			}
			wg.Wait()
			if serverError != nil {
				t.Fatalf("proxy server stopped unexpectedly (%v)", err)
			}
		})
	return ln.Addr().String()
}

type proxyHandler struct {
	counter *int
}

func (p *proxyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	*p.counter++
	if request.Method != "CONNECT" {
		writer.WriteHeader(400)
		return
	}
	target := request.URL.Host

	hijacker, ok := writer.(http.Hijacker)
	if !ok {
		writer.WriteHeader(500)
		return
	}

	backendConn, err := net.Dial("tcp", target)
	if err != nil {
		writer.WriteHeader(500)
		return
	}

	writer.WriteHeader(200)
	conn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		return
	}
	if clientBuf != nil {
		bytes := clientBuf.Reader.Buffered()
		buf := make([]byte, bytes)
		n, err := clientBuf.Read(buf)
		if err != nil {
			return
		}
		if n < bytes {
			// This is a bug
			return
		}
		if len(buf) > 0 {
			_, _ = backendConn.Write(buf)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(backendConn, conn)
		_ = backendConn.Close()
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(conn, backendConn)
		_ = conn.Close()
	}()
	wg.Wait()
}
