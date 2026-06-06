package provider

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kjaniec-dev/porthole/internal/config"
)

type TraefikClient struct {
	baseURL          string
	http             *http.Client
	certificateFiles []string
}

func NewTraefikClient(cfg config.TraefikConfig) *TraefikClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: transport}
	return &TraefikClient{
		baseURL:          cfg.URL,
		http:             client,
		certificateFiles: cfg.CertificateFiles,
	}
}

func (t *TraefikClient) get(path string, out any) error {
	resp, err := t.http.Get(t.baseURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("traefik API %s returned %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (t *TraefikClient) Routers() ([]Router, error) {
	var routers []Router
	if err := t.get("/api/http/routers", &routers); err != nil {
		return nil, err
	}
	return routers, nil
}

func (t *TraefikClient) Services() ([]Service, error) {
	var services []Service
	if err := t.get("/api/http/services", &services); err != nil {
		return nil, err
	}
	return services, nil
}

func (t *TraefikClient) Certificates() ([]Certificate, error) {
	// Traefik returns certs as map[store]map[domain]certInfo
	type rawCert struct {
		Subject   CertSubject `json:"subject"`
		SANs      []string    `json:"sans"`
		NotAfter  string      `json:"notAfter"`
		NotBefore string      `json:"notBefore"`
		Issuer    CertSubject `json:"issuer"`
	}
	type storeMap map[string]rawCert
	type tlsResponse map[string]storeMap

	var raw tlsResponse
	if err := t.get("/api/overview", nil); err != nil {
		// overview is just a check; continue
	}

	if err := t.get("/api/tls/certificates", &raw); err != nil {
		if len(t.certificateFiles) == 0 {
			return nil, err
		}
		return certificatesFromFiles(t.certificateFiles)
	}

	var certs []Certificate
	for store, domains := range raw {
		for domain, rc := range domains {
			notAfter, _ := time.Parse(time.RFC3339, rc.NotAfter)
			notBefore, _ := time.Parse(time.RFC3339, rc.NotBefore)
			certs = append(certs, Certificate{
				Subject:   rc.Subject,
				SANs:      rc.SANs,
				NotAfter:  notAfter,
				NotBefore: notBefore,
				Issuer:    rc.Issuer,
				Store:     store,
				Domain:    domain,
			})
		}
	}
	return certs, nil
}

func certificatesFromFiles(paths []string) ([]Certificate, error) {
	var certs []Certificate

	for _, path := range paths {
		pemBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		found := false
		for len(pemBytes) > 0 {
			var block *pem.Block
			block, pemBytes = pem.Decode(pemBytes)
			if block == nil {
				break
			}
			if block.Type != "CERTIFICATE" {
				continue
			}

			parsed, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}

			found = true
			domain := parsed.Subject.CommonName
			if domain == "" && len(parsed.DNSNames) > 0 {
				domain = parsed.DNSNames[0]
			}

			certs = append(certs, Certificate{
				Subject: CertSubject{
					CommonName: parsed.Subject.CommonName,
				},
				SANs:      parsed.DNSNames,
				NotAfter:  parsed.NotAfter,
				NotBefore: parsed.NotBefore,
				Issuer: CertSubject{
					CommonName: parsed.Issuer.CommonName,
				},
				Store:  "file",
				Domain: domain,
			})
		}

		if !found {
			return nil, fmt.Errorf("no certificate blocks found in %s", path)
		}
	}

	return certs, nil
}

func (t *TraefikClient) Overview() (*Overview, error) {
	var ov Overview
	if err := t.get("/api/overview", &ov); err != nil {
		return nil, err
	}
	return &ov, nil
}
