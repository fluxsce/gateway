package bootstrap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeListenHealthPath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"/", ""},
		{"off", ""},
		{"OFF", ""},
		{"-", ""},
		{"   ", ""},
		{"/_gw/health", "/_gw/health"},
		{"/health", "/health"},
		{"healthz", "/healthz"},
	}
	for _, tc := range cases {
		if got := normalizeListenHealthPath(tc.in); got != tc.want {
			t.Fatalf("normalizeListenHealthPath(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestWrapListenHealthDisabledUsesNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
	h := wrapListenHealth("", next)
	if _, ok := h.(http.HandlerFunc); !ok {
		t.Fatal("empty path must not wrap")
	}
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
	if !called {
		t.Fatal("business handler must run")
	}
}

func TestListenHealthHandlerExactGet(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
	h := wrapListenHealth("/_gw/health", next)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/_gw/health", nil))
	if called {
		t.Fatal("probe must not enter engine")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("status=%q", body.Status)
	}
}

func TestListenHealthHandlerMissesBusinessPaths(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
	h := wrapListenHealth("/_gw/health", next)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
	if !called {
		t.Fatal("/health must go to routes")
	}

	called = false
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/_gw/health/extra", nil))
	if !called {
		t.Fatal("prefix must not match")
	}

	called = false
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/_gw/health", nil))
	if !called {
		t.Fatal("POST must go to routes")
	}
}

func TestListenHealthHandlerHead(t *testing.T) {
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("HEAD probe must not enter engine")
	})
	h := wrapListenHealth("/_gw/health", next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, "/_gw/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatal("HEAD must not write body")
	}
}
