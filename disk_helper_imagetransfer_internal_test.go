package ovirtclient

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

// TestFindTransferURLFallsBackFromUnresponsiveHost checks that an unresponsive transfer URL, for example a host behind
// a firewall that drops packets, does not hold up the fallback to the proxy URL.
func TestFindTransferURLFallsBackFromUnresponsiveHost(t *testing.T) {
	oldTimeout := transferURLProbeTimeout
	transferURLProbeTimeout = 100 * time.Millisecond
	t.Cleanup(func() { transferURLProbeTimeout = oldTimeout })

	release := make(chan struct{})
	host := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-release
	}))
	proxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() {
		close(release)
		host.Close()
		proxy.Close()
	})

	transfer := &imageTransferImpl{
		logger:  ovirtclientlog.NewTestLogger(t),
		retries: []RetryStrategy{ExponentialBackoff(1)},
		httpClient: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		},
		transfer: ovirtsdk4.NewImageTransferBuilder().
			TransferUrl(host.URL + "/images/test").
			ProxyUrl(proxy.URL + "/images/test").
			MustBuild(),
	}

	done := make(chan error, 1)
	go func() {
		done <- transfer.findTransferURL()
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("failed to find a transfer URL (%v)", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatalf("findTransferURL did not fall back to the proxy URL within 15 seconds")
	}
	if expected := proxy.URL + "/images/test"; transfer.transferURL != expected {
		t.Fatalf("incorrect transfer URL selected: %s (expected: %s)", transfer.transferURL, expected)
	}
}
