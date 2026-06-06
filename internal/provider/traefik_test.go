package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCertificatesFromFilesParsesPEMCertificate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	certPath := filepath.Join(dir, "example.crt")
	notBefore := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	notAfter := notBefore.Add(48 * time.Hour)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "frontend.localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		DNSNames:              []string{"frontend.localhost", "api.localhost"},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	file, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("create cert file: %v", err)
	}

	if err := pem.Encode(file, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		_ = file.Close()
		t.Fatalf("encode pem: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close cert file: %v", err)
	}

	certs, err := certificatesFromFiles([]string{certPath})
	if err != nil {
		t.Fatalf("parse certs: %v", err)
	}
	if len(certs) != 1 {
		t.Fatalf("expected 1 cert, got %d", len(certs))
	}

	cert := certs[0]
	if cert.Domain != "frontend.localhost" {
		t.Fatalf("expected domain frontend.localhost, got %q", cert.Domain)
	}
	if cert.Subject.CommonName != "frontend.localhost" {
		t.Fatalf("expected common name frontend.localhost, got %q", cert.Subject.CommonName)
	}
	if cert.Issuer.CommonName != "frontend.localhost" {
		t.Fatalf("expected issuer frontend.localhost, got %q", cert.Issuer.CommonName)
	}
	if len(cert.SANs) != 2 {
		t.Fatalf("expected 2 SANs, got %d", len(cert.SANs))
	}
	if !cert.NotAfter.Equal(notAfter) {
		t.Fatalf("expected notAfter %v, got %v", notAfter, cert.NotAfter)
	}
	if !cert.NotBefore.Equal(notBefore) {
		t.Fatalf("expected notBefore %v, got %v", notBefore, cert.NotBefore)
	}
	if cert.Store != "file" {
		t.Fatalf("expected store file, got %q", cert.Store)
	}
}
