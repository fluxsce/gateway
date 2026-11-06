package routes

import (
	"testing"

	"gateway/web/utils/constants"
)

func TestModuleAPIPrefixFollowsFrontendPrefix(t *testing.T) {
	constants.SetAPIRoot("/gatewayweb")
	t.Cleanup(func() { constants.SetAPIRoot("/") })

	got := ModuleAPIPrefix("hub0007")
	if got != "/gatewayweb/api/hub0007" {
		t.Fatalf("ModuleAPIPrefix: got %q, want /gatewayweb/api/hub0007", got)
	}
}

func TestAPIRootFromFrontendPrefix(t *testing.T) {
	cases := []struct {
		prefix string
		want   string
	}{
		{"", "/api"},
		{"/", "/api"},
		{"gatewayweb", "/gatewayweb/api"},
		{"/gatewayweb", "/gatewayweb/api"},
		{"/gatewayweb/", "/gatewayweb/api"},
	}
	for _, tc := range cases {
		got := constants.APIRootFromFrontendPrefix(tc.prefix)
		if got != tc.want {
			t.Fatalf("prefix %q: got %q, want %q", tc.prefix, got, tc.want)
		}
	}
}
