package webapp

import "testing"

func TestUnmatchedPathReturnsJSON(t *testing.T) {
	cases := []struct {
		path    string
		apiRoot string
		want    bool
	}{
		{"/gatewayweb/api/hub0002", "/gatewayweb/api", true},
		{"/gatewayweb/api", "/gatewayweb/api", true},
		{"/gateway/hub0002/queryUsers", "/gatewayweb/api", true},
		{"/gateway/user/login", "/gatewayweb/api", true},
		{"/gateway", "/gatewayweb/api", true},
		{"/gatewayweb/login", "/gatewayweb/api", false},
		{"/gatewayweb", "/gatewayweb/api", false},
		{"/health", "/gatewayweb/api", false},
		{"/gateway/login", "/gateway/api", false},
		{"/gateway", "/gateway/api", false},
	}
	for _, tc := range cases {
		if got := unmatchedPathReturnsJSON(tc.path, tc.apiRoot); got != tc.want {
			t.Fatalf("unmatchedPathReturnsJSON(%q, %q)=%v, want %v", tc.path, tc.apiRoot, got, tc.want)
		}
	}
}

func TestNormalizeHealthPath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "/health"},
		{"/", "/health"},
		{"   ", "/health"},
		{"/health", "/health"},
		{"healthz", "/healthz"},
		{"/ready", "/ready"},
	}
	for _, tc := range cases {
		if got := normalizeHealthPath(tc.in); got != tc.want {
			t.Fatalf("normalizeHealthPath(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}
