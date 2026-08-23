package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/internal/gateway/core"
)

func TestMatchPathPrefixSegmentBoundary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		routePath   string
		requestPath string
		want        bool
	}{
		{routePath: "/app", requestPath: "/app", want: true},
		{routePath: "/app", requestPath: "/app/", want: true},
		{routePath: "/app", requestPath: "/app/index.html", want: true},
		{routePath: "/app/", requestPath: "/app/index.html", want: true},
		{routePath: "/app", requestPath: "/application", want: false},
		{routePath: "/app", requestPath: "/app.js", want: false},
		{routePath: "/api/users", requestPath: "/api/users/123", want: true},
		{routePath: "/api/users", requestPath: "/api/orders", want: false},
	}
	for _, tt := range tests {
		if got := matchPathPrefix(tt.requestPath, tt.routePath); got != tt.want {
			t.Fatalf("matchPathPrefix(%q, %q) = %v, want %v", tt.requestPath, tt.routePath, got, tt.want)
		}
	}
}

func TestRoutePrefixDoesNotStealNeighborPaths(t *testing.T) {
	t.Parallel()
	route, err := NewRoute(RouteConfig{
		ID:        "app",
		Name:      "app",
		Path:      "/app",
		MatchType: MatchTypePrefix,
		ServiceID: "svc",
		Enabled:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		path string
		want bool
	}{
		{path: "/app", want: true},
		{path: "/app/page", want: true},
		{path: "/application", want: false},
		{path: "/app.js", want: false},
	} {
		req := httptest.NewRequest(http.MethodGet, item.path, nil)
		ctx := core.NewContext(httptest.NewRecorder(), req)
		got, matchErr := route.Match(ctx)
		if matchErr != nil {
			t.Fatal(matchErr)
		}
		if got != item.want {
			t.Fatalf("path %s: match=%v, want %v", item.path, got, item.want)
		}
	}
}

func TestRouteExactPathWithDoubleStarMatchesChildren(t *testing.T) {
	t.Parallel()
	route, err := NewRoute(RouteConfig{
		ID:        "users",
		Name:      "users",
		Path:      "/api/v1/users/**",
		MatchType: MatchTypeExact,
		ServiceID: "svc",
		Enabled:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/users", "/api/v1/users/123", "/api/v1/users/123/profile"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		ctx := core.NewContext(httptest.NewRecorder(), req)
		got, matchErr := route.Match(ctx)
		if matchErr != nil {
			t.Fatal(matchErr)
		}
		if !got {
			t.Fatalf("%s 应匹配 /api/v1/users/**", path)
		}
	}
}

func TestRouteWildcardSlashKeepsSegmentBoundary(t *testing.T) {
	t.Parallel()
	route, err := NewRoute(RouteConfig{
		ID:        "app-star",
		Name:      "app-star",
		Path:      "/app/*",
		MatchType: MatchTypePrefix,
		ServiceID: "svc",
		Enabled:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/application", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	got, err := route.Match(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Fatal("/app/* 不应匹配 /application")
	}
}
