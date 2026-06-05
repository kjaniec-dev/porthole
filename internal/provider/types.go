package provider

import "time"

type Router struct {
	Name        string   `json:"name"`
	Rule        string   `json:"rule"`
	Service     string   `json:"service"`
	Status      string   `json:"status"`
	Provider    string   `json:"provider"`
	Middlewares []string `json:"middlewares"`
	TLS         *TLS     `json:"tls,omitempty"`
	EntryPoints []string `json:"entryPoints"`
}

type TLS struct {
	CertResolver string `json:"certResolver,omitempty"`
}

type Service struct {
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Status       string       `json:"status"`
	Provider     string       `json:"provider"`
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty"`
}

type LoadBalancer struct {
	Servers []Server `json:"servers,omitempty"`
}

type Server struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

type Certificate struct {
	Subject    CertSubject `json:"subject"`
	SANs       []string    `json:"sans"`
	NotAfter   time.Time   `json:"notAfter"`
	NotBefore  time.Time   `json:"notBefore"`
	Issuer     CertSubject `json:"issuer"`
	Store      string
	Domain     string
}

type CertSubject struct {
	CommonName string `json:"commonName"`
}

type Overview struct {
	HTTP      OverviewSection `json:"http"`
	TCP       OverviewSection `json:"tcp"`
	UDP       OverviewSection `json:"udp"`
}

type OverviewSection struct {
	Routers    ItemCount `json:"routers"`
	Services   ItemCount `json:"services"`
	Middlewares ItemCount `json:"middlewares"`
}

type ItemCount struct {
	Total   int `json:"total"`
	Enabled int `json:"enabled"`
	Errors  int `json:"errors"`
	Warnings int `json:"warnings"`
}
