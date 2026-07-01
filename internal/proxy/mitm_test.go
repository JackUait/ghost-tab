package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureCerts_generatesLeafVerifiableByCA(t *testing.T) {
	dir := t.TempDir()
	certs, err := EnsureCerts(dir, "api.anthropic.com")
	if err != nil {
		t.Fatalf("EnsureCerts: %v", err)
	}

	caPEM, err := os.ReadFile(certs.CACertPath)
	if err != nil {
		t.Fatalf("CA cert not written: %v", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		t.Fatal("CA PEM not parseable")
	}

	// The leaf must chain to the CA and be valid for the upstream host.
	leaf, err := x509.ParseCertificate(certs.Leaf.Certificate[0])
	if err != nil {
		t.Fatalf("parse leaf: %v", err)
	}
	if _, err := leaf.Verify(x509.VerifyOptions{Roots: pool, DNSName: "api.anthropic.com"}); err != nil {
		t.Errorf("leaf does not verify for api.anthropic.com: %v", err)
	}
	if _, err := leaf.Verify(x509.VerifyOptions{Roots: pool, DNSName: TestHost}); err != nil {
		t.Errorf("leaf does not verify for the test host: %v", err)
	}
}

func TestEnsureCerts_reusesCachedChain(t *testing.T) {
	dir := t.TempDir()
	first, err := EnsureCerts(dir, "api.anthropic.com")
	if err != nil {
		t.Fatal(err)
	}
	// Second call must return the same persisted CA (stable NODE_EXTRA_CA_CERTS).
	second, err := EnsureCerts(dir, "api.anthropic.com")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := os.ReadFile(first.CACertPath)
	b, _ := os.ReadFile(second.CACertPath)
	if string(a) != string(b) {
		t.Error("CA cert should be stable across calls")
	}
	if _, err := os.Stat(filepath.Join(dir, "wisp-deck-ca.pem")); err != nil {
		t.Errorf("CA cert not at expected path: %v", err)
	}
}

func TestLeafCert_usableForTLSServer(t *testing.T) {
	dir := t.TempDir()
	certs, err := EnsureCerts(dir, "api.anthropic.com")
	if err != nil {
		t.Fatal(err)
	}
	// The leaf must be a usable tls.Certificate (has a parsed private key).
	cfg := &tls.Config{Certificates: []tls.Certificate{certs.Leaf}}
	if len(cfg.Certificates) != 1 || cfg.Certificates[0].PrivateKey == nil {
		t.Error("leaf tls.Certificate missing private key")
	}
}
