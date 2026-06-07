package provider

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kjaniec-dev/porthole/internal/config"
)

type CaddyClient struct {
	baseURL          string
	http             *http.Client
	certificateFiles []string
}

func NewCaddyClient(cfg config.CaddyConfig) *CaddyClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}
	return &CaddyClient{
		baseURL:          cfg.URL,
		http:             &http.Client{Timeout: 10 * time.Second, Transport: transport},
		certificateFiles: cfg.CertificateFiles,
	}
}

// Caddy admin API JSON types.

type caddyConfig struct {
	Apps caddyApps `json:"apps"`
}

type caddyApps struct {
	HTTP *caddyHTTP `json:"http,omitempty"`
}

type caddyHTTP struct {
	Servers map[string]caddyServer `json:"servers"`
}

type caddyServer struct {
	Listen []string     `json:"listen"`
	Routes []caddyRoute `json:"routes"`
}

type caddyRoute struct {
	Match    []caddyMatch   `json:"match"`
	Handle   []caddyHandler `json:"handle"`
	Terminal bool           `json:"terminal"`
}

type caddyMatch struct {
	Host []string `json:"host"`
	Path []string `json:"path"`
}

type caddyHandler struct {
	Handler   string          `json:"handler"`
	Upstreams []caddyUpstream `json:"upstreams,omitempty"`
	Routes    []caddyRoute    `json:"routes,omitempty"` // subroute handler
}

type caddyUpstream struct {
	Dial string `json:"dial"`
}

func (c *CaddyClient) fetchConfig() (*caddyConfig, error) {
	resp, err := c.http.Get(c.baseURL + "/config/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("caddy admin API returned %d", resp.StatusCode)
	}

	// Caddy returns null when no config is loaded yet.
	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if string(raw) == "null" {
		return &caddyConfig{}, nil
	}

	var cfg caddyConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *CaddyClient) Routers() ([]Router, error) {
	cfg, err := c.fetchConfig()
	if err != nil {
		return nil, err
	}
	if cfg.Apps.HTTP == nil {
		return nil, nil
	}

	var routers []Router
	for serverName, server := range cfg.Apps.HTTP.Servers {
		hasTLS := serverListensTLS(server.Listen)
		for i, route := range server.Routes {
			rule := formatMatchRule(route.Match)
			svc := extractServiceName(route.Handle, serverName, i)

			var tlsVal *TLS
			if hasTLS {
				tlsVal = &TLS{}
			}

			routers = append(routers, Router{
				Name:     fmt.Sprintf("%s/route/%d", serverName, i),
				Rule:     rule,
				Service:  svc,
				Status:   "enabled",
				Provider: "caddy",
				TLS:      tlsVal,
			})
		}
	}
	return routers, nil
}

func (c *CaddyClient) Services() ([]Service, error) {
	cfg, err := c.fetchConfig()
	if err != nil {
		return nil, err
	}
	if cfg.Apps.HTTP == nil {
		return nil, nil
	}

	var services []Service
	for serverName, server := range cfg.Apps.HTTP.Servers {
		for i, route := range server.Routes {
			upstreams := extractUpstreams(route.Handle)
			if len(upstreams) == 0 {
				continue
			}

			var servers []Server
			for _, u := range upstreams {
				servers = append(servers, Server{URL: u, Status: "up"})
			}

			services = append(services, Service{
				Name:         fmt.Sprintf("%s/route/%d", serverName, i),
				Type:         "reverse_proxy",
				Status:       "enabled",
				Provider:     "caddy",
				LoadBalancer: &LoadBalancer{Servers: servers},
			})
		}
	}
	return services, nil
}

func (c *CaddyClient) Certificates() ([]Certificate, error) {
	if len(c.certificateFiles) == 0 {
		return nil, nil
	}
	return certificatesFromFiles(c.certificateFiles)
}

// serverListensTLS reports whether any listen address uses port 443.
func serverListensTLS(listen []string) bool {
	for _, l := range listen {
		if strings.HasSuffix(l, ":443") {
			return true
		}
	}
	return false
}

// formatMatchRule converts Caddy match directives to a human-readable rule string.
func formatMatchRule(matchers []caddyMatch) string {
	if len(matchers) == 0 {
		return "*"
	}

	var parts []string
	for _, m := range matchers {
		if len(m.Host) > 0 {
			parts = append(parts, fmt.Sprintf("Host(%s)", strings.Join(m.Host, ", ")))
		}
		if len(m.Path) > 0 {
			parts = append(parts, fmt.Sprintf("Path(%s)", strings.Join(m.Path, ", ")))
		}
	}
	if len(parts) == 0 {
		return "*"
	}
	return strings.Join(parts, " && ")
}

// extractServiceName returns the first upstream dial address found in the handler chain.
func extractServiceName(handlers []caddyHandler, serverName string, routeIdx int) string {
	for _, h := range handlers {
		switch h.Handler {
		case "reverse_proxy":
			if len(h.Upstreams) > 0 {
				return h.Upstreams[0].Dial
			}
		case "subroute":
			for _, sub := range h.Routes {
				if name := extractServiceName(sub.Handle, serverName, routeIdx); name != "" {
					return name
				}
			}
		}
	}
	return fmt.Sprintf("%s/route/%d", serverName, routeIdx)
}

// extractUpstreams recursively collects all upstream dial addresses from a handler chain.
func extractUpstreams(handlers []caddyHandler) []string {
	var out []string
	for _, h := range handlers {
		switch h.Handler {
		case "reverse_proxy":
			for _, u := range h.Upstreams {
				out = append(out, u.Dial)
			}
		case "subroute":
			for _, sub := range h.Routes {
				out = append(out, extractUpstreams(sub.Handle)...)
			}
		}
	}
	return out
}
