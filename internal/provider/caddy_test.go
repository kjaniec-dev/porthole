package provider

import (
	"testing"
)

func TestFormatMatchRule(t *testing.T) {
	tests := []struct {
		name     string
		matchers []caddyMatch
		want     string
	}{
		{
			name:     "empty matchers",
			matchers: nil,
			want:     "*",
		},
		{
			name:     "single host",
			matchers: []caddyMatch{{Host: []string{"example.com"}}},
			want:     "Host(example.com)",
		},
		{
			name:     "multiple hosts",
			matchers: []caddyMatch{{Host: []string{"a.com", "b.com"}}},
			want:     "Host(a.com, b.com)",
		},
		{
			name:     "host and path",
			matchers: []caddyMatch{{Host: []string{"api.example.com"}, Path: []string{"/v1/*"}}},
			want:     "Host(api.example.com) && Path(/v1/*)",
		},
		{
			name:     "path only",
			matchers: []caddyMatch{{Path: []string{"/health"}}},
			want:     "Path(/health)",
		},
		{
			name:     "empty matcher fields",
			matchers: []caddyMatch{{}},
			want:     "*",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatMatchRule(tc.matchers)
			if got != tc.want {
				t.Errorf("formatMatchRule() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestServerListensTLS(t *testing.T) {
	tests := []struct {
		listen []string
		want   bool
	}{
		{[]string{":443"}, true},
		{[]string{"0.0.0.0:443"}, true},
		{[]string{"127.0.0.1:443"}, true},
		{[]string{":80"}, false},
		{[]string{":80", ":443"}, true},
		{[]string{}, false},
		{nil, false},
	}

	for _, tc := range tests {
		got := serverListensTLS(tc.listen)
		if got != tc.want {
			t.Errorf("serverListensTLS(%v) = %v, want %v", tc.listen, got, tc.want)
		}
	}
}

func TestExtractUpstreams(t *testing.T) {
	tests := []struct {
		name     string
		handlers []caddyHandler
		want     []string
	}{
		{
			name:     "no handlers",
			handlers: nil,
			want:     nil,
		},
		{
			name: "direct reverse_proxy",
			handlers: []caddyHandler{
				{
					Handler:   "reverse_proxy",
					Upstreams: []caddyUpstream{{Dial: "backend:8080"}},
				},
			},
			want: []string{"backend:8080"},
		},
		{
			name: "nested in subroute",
			handlers: []caddyHandler{
				{
					Handler: "subroute",
					Routes: []caddyRoute{
						{
							Handle: []caddyHandler{
								{
									Handler:   "reverse_proxy",
									Upstreams: []caddyUpstream{{Dial: "app:3000"}},
								},
							},
						},
					},
				},
			},
			want: []string{"app:3000"},
		},
		{
			name: "multiple upstreams",
			handlers: []caddyHandler{
				{
					Handler: "reverse_proxy",
					Upstreams: []caddyUpstream{
						{Dial: "node1:8080"},
						{Dial: "node2:8080"},
					},
				},
			},
			want: []string{"node1:8080", "node2:8080"},
		},
		{
			name: "non-proxy handler ignored",
			handlers: []caddyHandler{
				{Handler: "static_response"},
			},
			want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractUpstreams(tc.handlers)
			if len(got) != len(tc.want) {
				t.Fatalf("extractUpstreams() = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("extractUpstreams()[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestExtractServiceName(t *testing.T) {
	tests := []struct {
		name     string
		handlers []caddyHandler
		want     string
	}{
		{
			name:     "no handlers falls back to route name",
			handlers: nil,
			want:     "srv0/route/0",
		},
		{
			name: "reverse_proxy returns first upstream",
			handlers: []caddyHandler{
				{
					Handler:   "reverse_proxy",
					Upstreams: []caddyUpstream{{Dial: "api:8080"}},
				},
			},
			want: "api:8080",
		},
		{
			name: "subroute unwrapped",
			handlers: []caddyHandler{
				{
					Handler: "subroute",
					Routes: []caddyRoute{
						{
							Handle: []caddyHandler{
								{
									Handler:   "reverse_proxy",
									Upstreams: []caddyUpstream{{Dial: "svc:9000"}},
								},
							},
						},
					},
				},
			},
			want: "svc:9000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractServiceName(tc.handlers, "srv0", 0)
			if got != tc.want {
				t.Errorf("extractServiceName() = %q, want %q", got, tc.want)
			}
		})
	}
}
