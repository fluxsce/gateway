package filter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/pkg/syssetting"
)

func TestHeaderFilterExpandsEnvVar(t *testing.T) {
	syssetting.PutEnvVars("tenant-env", syssetting.EnvVarsSettings{
		Items: []syssetting.EnvVar{{Name: "GW_TOKEN", Value: "from-env"}},
	})

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	ctx.Set(constants.ContextKeyTenantID, "tenant-env")

	h := NewRequestHeaderFilter("h", PostRouting, 1)
	h.ConfigureSet("X-Gateway-Internal", "${GW_TOKEN}")
	if err := h.Apply(ctx); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.Header.Get("X-Gateway-Internal"); got != "from-env" {
		t.Fatalf("header=%q", got)
	}
}

func TestPathRewriteExpandsEnvVar(t *testing.T) {
	syssetting.PutEnvVars("tenant-env", syssetting.EnvVarsSettings{
		Items: []syssetting.EnvVar{{Name: "API_PREFIX", Value: "/internal"}},
	})

	req := httptest.NewRequest(http.MethodGet, "/public/users", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	ctx.Set(constants.ContextKeyTenantID, "tenant-env")

	f := NewSimplePathRewriteFilter("rw", "/public", "${API_PREFIX}", 1)
	if err := f.Apply(ctx); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.URL.Path; got != "/internal/users" {
		t.Fatalf("path=%q", got)
	}
}
