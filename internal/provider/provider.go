package provider

// Provider is the interface implemented by all reverse-proxy data sources.
type Provider interface {
	Routers() ([]Router, error)
	Services() ([]Service, error)
	Certificates() ([]Certificate, error)
}
