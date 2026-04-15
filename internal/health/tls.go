package health

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// TLSConfig holds paths to TLS material used when dialing a gRPC target.
type TLSConfig struct {
	// CACertFile is the path to a PEM-encoded CA certificate bundle.
	// When set, the server certificate is verified against this CA.
	CACertFile string

	// CertFile and KeyFile are the client certificate/key pair for mTLS.
	CertFile string
	KeyFile  string

	// InsecureSkipVerify disables server certificate verification.
	// Intended for testing only.
	InsecureSkipVerify bool
}

// BuildTransportCredentials constructs gRPC transport credentials from cfg.
// If cfg is nil, plain-text credentials are returned (no TLS).
func BuildTransportCredentials(cfg *TLSConfig) (credentials.TransportCredentials, error) {
	if cfg == nil {
		return credentials.NewClientTLSFromCert(nil, ""), nil
	}

	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec // intentional opt-in
	}

	if cfg.CACertFile != "" {
		pem, err := os.ReadFile(cfg.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("tls: read CA cert %q: %w", cfg.CACertFile, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("tls: no valid certificates found in %q", cfg.CACertFile)
		}
		tlsCfg.RootCAs = pool
	}

	if cfg.CertFile != "" || cfg.KeyFile != "" {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, fmt.Errorf("tls: both cert-file and key-file must be provided together")
		}
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("tls: load client key pair: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return credentials.NewTLS(tlsCfg), nil
}
