package ovirtclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// New creates a new copy of the enhanced oVirt client.
func New(
	url string,
	username string,
	password string,
	caFile string,
	caCert []byte,
	insecure bool,
	extraHeaders map[string]string,
	logger Logger,
) (ClientWithLegacySupport, error) {
	return NewWithVerify(url, username, password, caFile, caCert, insecure, extraHeaders, logger, testConnection)
}

// NewWithVerify allows customizing the verification function for the connection. Alternatively, a nil can be passed to
// disable connection verification.
func NewWithVerify(
	url string,
	username string,
	password string,
	caFile string,
	caCert []byte,
	insecure bool,
	extraHeaders map[string]string,
	logger Logger,
	verify func(connection *ovirtsdk4.Connection) error,
) (ClientWithLegacySupport, error) {
	if err := validateURL(url); err != nil {
		return nil, wrap(err, EBadArgument, "invalid URL: %s", url)
	}
	if err := validateUsername(username); err != nil {
		return nil, wrap(err, "invalid username: %s", username)
	}
	if caFile == "" && len(caCert) == 0 && !insecure {
		return nil, newError(EBadArgument, "one of caFile, caCert, or insecure must be provided")
	}

	connBuilder := ovirtsdk4.NewConnectionBuilder().
		URL(url).
		Username(username).
		Password(password).
		CAFile(caFile).
		CACert(caCert).
		Insecure(insecure)
	if len(extraHeaders) > 0 {
		connBuilder.Headers(extraHeaders)
	}

	conn, err := connBuilder.Build()
	if err != nil {
		return nil, wrap(err, EUnidentified, "failed to create underlying oVirt connection")
	}

	tlsConfig, err := createTLSConfig(caFile, caCert, insecure)
	if err != nil {
		return nil, wrap(err, ETLSError, "failed to create TLS configuration")
	}

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	if verify != nil {
		if err := verify(conn); err != nil {
			return nil, err
		}
	}

	return &oVirtClient{
		conn:       conn,
		httpClient: httpClient,
		logger:     logger,
		url:        url,
	}, nil
}

func testConnection(conn *ovirtsdk4.Connection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		lastError := conn.SystemService().Connection().Test()
		if lastError == nil {
			break
		}
		if err := identify(lastError); err != nil {
			var realErr EngineError
			// This will always be an engine error
			_ = errors.As(err, &realErr)
			if !realErr.CanAutoRetry() {
				return err
			}
			lastError = err
		}
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			return wrap(
				lastError,
				ETimeout,
				"timeout while attempting to create connection",
			)
		}
	}
	return nil
}

func createTLSConfig(
	caFile string,
	caCert []byte,
	insecure bool,
) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		// Based on Mozilla intermediate compatibility:
		// https://wiki.mozilla.org/Security/Server_Side_TLS#Intermediate_compatibility_.28recommended.29
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		CurvePreferences: []tls.CurveID{
			tls.CurveP256, tls.CurveP384,
		},
		PreferServerCipherSuites: false,
		InsecureSkipVerify:       insecure,
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		// This is the case on Windows where the system certificate pool is not available.
		certPool = x509.NewCertPool()
	}
	if len(caCert) != 0 {
		if ok := certPool.AppendCertsFromPEM(caCert); !ok {
			return nil, newError(EBadArgument, "the provided CA certificate is not a valid certificate in PEM format")
		}
	}
	if caFile != "" {
		pemData, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, wrap(err, EFileReadFailed, "failed to read CA certificate from file %s", caFile)
		}
		if ok := certPool.AppendCertsFromPEM(pemData); !ok {
			return nil, newError(
				ETLSError,
				"the provided CA certificate is not a valid certificate in PEM format in file %s",
				caFile,
			)
		}
	}
	tlsConfig.RootCAs = certPool
	return tlsConfig, nil
}

func validateUsername(username string) error {
	usernameParts := strings.SplitN(username, "@", 2)
	//nolint:gomnd
	if len(usernameParts) != 2 {
		return newError(EBadArgument, "username must contain exactly one @ sign (format should be admin@internal)")
	}
	if len(usernameParts[0]) == 0 {
		return newError(EBadArgument, "no user supplied before @ sign in username (format should be admin@internal)")
	}
	if len(usernameParts[1]) == 0 {
		return newError(EBadArgument, "no scope supplied after @ sign in username (format should be admin@internal)")
	}
	return nil
}

func validateURL(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return newError(EBadArgument, "URL must start with http:// or https://")
	}
	return nil
}
