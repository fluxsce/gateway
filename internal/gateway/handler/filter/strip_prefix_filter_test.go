package filter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/internal/gateway/core"
)

func TestStripPrefixFilterFromConfigReadsNestedPrefix(t *testing.T) {
	f, err := StripPrefixFilterFromConfig(FilterConfig{
		ID:     "strip-1",
		Name:   "strip api",
		Type:   string(StripFilterType),
		Action: string(PreRouting),
		Order:  10,
		Config: map[string]interface{}{
			"stripConfig": map[string]interface{}{
				"prefix": "/api/v1",
			},
		},
	})
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	strip := f.(*StripPrefixFilter)
	if strip.Prefix != "/api/v1" {
		t.Fatalf("prefix=%q", strip.Prefix)
	}
	if strip.GetAction() != PreRouting {
		t.Fatalf("action=%s", strip.GetAction())
	}
}

func TestStripPrefixFilterFromConfigReadsFlatPrefix(t *testing.T) {
	f, err := StripPrefixFilterFromConfig(FilterConfig{
		ID:   "strip-2",
		Name: "strip",
		Config: map[string]interface{}{
			"prefix": "api/v1/",
		},
	})
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if got := f.(*StripPrefixFilter).Prefix; got != "/api/v1" {
		t.Fatalf("prefix=%q", got)
	}
}

func TestStripPrefixFilterFromConfigRejectsEmptyPrefix(t *testing.T) {
	_, err := StripPrefixFilterFromConfig(FilterConfig{
		ID:   "strip-empty",
		Name: "strip",
		Config: map[string]interface{}{
			"stripConfig": map[string]interface{}{},
		},
	})
	if err == nil {
		t.Fatal("空前缀应返回错误")
	}
}

func TestStripPrefixFilterApplyStripsOnSegmentBoundary(t *testing.T) {
	f := mustStripFilter(t, "/api")

	got := applyStrip(t, f, "/api/users?x=1")
	if got != "/users" {
		t.Fatalf("path=%q", got)
	}

	unchanged := applyStrip(t, f, "/apix/users")
	if unchanged != "/apix/users" {
		t.Fatalf("非路径段不应剥离, path=%q", unchanged)
	}

	root := applyStrip(t, f, "/api")
	if root != "/" {
		t.Fatalf("剥完应落到根路径, path=%q", root)
	}
}

func mustStripFilter(t *testing.T, prefix string) Filter {
	t.Helper()
	f, err := StripPrefixFilterFromConfig(FilterConfig{
		ID:   "strip",
		Name: "strip",
		Config: map[string]interface{}{
			"stripConfig": map[string]interface{}{"prefix": prefix},
		},
	})
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	return f
}

func applyStrip(t *testing.T, f Filter, rawURL string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, rawURL, nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	if err := f.Apply(ctx); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return req.URL.Path
}
