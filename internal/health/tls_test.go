package health

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempPEM(t *testing.T, dir, name string, block *pem.Block) string {
	t.Helper()
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer f.Close()
	if err := pem.Encode(f, block); err != nil {
		t.Fatalf("pem encode: %v", err)
	}
	return p
}

func generateSelfSignedCert(t *testing.T, dir string) (certPath, keyPath string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	certPath = writeTempPEM(t, dir, "cert.pem", &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPath = writeTempPEM(t, dir, "key.pem", &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	return certPath, keyPath
}

func TestBuildTransportCredentials_Nil(t *testing.T) {
	creds, err := BuildTransportCredentials(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil credentials")
	}
}

func TestBuildTransportCredentials_InsecureSkipVerify(t *testing.T) {
	creds, err := BuildTransportCredentials(&TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil credentials")
	}
}

func TestBuildTransportCredentials_WithCACert(t *testing.T) {
	dir := t.TempDir()
	certPath, _ := generateSelfSignedCert(t, dir)
	creds, err := BuildTransportCredentials(&TLSConfig{CACertFile: certPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil credentials")
	}
}

func TestBuildTransportCredentials_WithMTLS(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := generateSelfSignedCert(t, dir)
	creds, err := BuildTransportCredentials(&TLSConfig{CertFile: certPath, KeyFile: keyPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil credentials")
	}
}

func TestBuildTransportCredentials_MissingKey(t *testing.T) {
	_, err := BuildTransportCredentials(&TLSConfig{CertFile: "cert.pem"})
	if err == nil {
		t.Fatal("expected error when key-file is missing")
	}
}

func TestBuildTransportCredentials_InvalidCACert(t *testing.T) {
	dir := t.TempDir()
	badCA := filepath.Join(dir, "bad-ca.pem")
	if err := os.WriteFile(badCA, []byte("not a cert"), 0o600); err != nil {
		t.Fatalf("write bad ca: %v", err)
	}
	_, err := BuildTransportCredentials(&TLSConfig{CACertFile: badCA})
	if err == nil {
		t.Fatal("expected error for invalid CA cert")
	}
}
