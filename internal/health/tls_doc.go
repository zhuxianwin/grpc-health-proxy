// Package health provides gRPC health-check primitives used by the proxy.
//
// # TLS support
//
// BuildTransportCredentials converts a [TLSConfig] into gRPC
// [credentials.TransportCredentials] suitable for use when dialling a
// health-check target over TLS or mutual-TLS (mTLS).
//
// Usage:
//
//	creds, err := health.BuildTransportCredentials(&health.TLSConfig{
//		CACertFile: "/etc/certs/ca.pem",
//		CertFile:   "/etc/certs/client.pem",
//		KeyFile:    "/etc/certs/client-key.pem",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
package health
