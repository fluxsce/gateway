package webapp

import "testing"

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
