package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseTraceparent(t *testing.T) {
	valid := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	tp := ParseTraceparent(valid)
	if !tp.Valid() || !tp.Sampled() {
		t.Fatalf("expected valid sampled traceparent, got %+v", tp)
	}
	if ParseTraceparent("not-a-header").Valid() {
		t.Fatal("invalid header should not parse")
	}
	if ParseTraceparent("00-" + strings.Repeat("0", 32) + "-00f067aa0ba902b7-01").Valid() {
		t.Fatal("zero trace id should be rejected")
	}
}

func TestContinueTraceparentKeepsTraceID(t *testing.T) {
	parent := ParseTraceparent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	child := ContinueTraceparent(parent)
	if child.TraceID != parent.TraceID {
		t.Fatalf("trace id changed: %s -> %s", parent.TraceID, child.TraceID)
	}
	if child.ParentID == parent.ParentID {
		t.Fatal("span id should be renewed")
	}
}

func TestPropagatorInjectsTraceparent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set(HeaderTraceparent, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	tracer := NewPropagator(Config{Enabled: true, SamplingRate: 1})
	ctx, _ := tracer.StartRequest(req.Context(), req, "http.server")

	outbound := make(http.Header)
	Inject(ctx, outbound)
	got := ParseTraceparent(outbound.Get(HeaderTraceparent))
	if !got.Valid() {
		t.Fatal("expected outbound traceparent")
	}
	if got.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace id not propagated: %s", got.TraceID)
	}
	if TraceparentValue(ctx) == "" {
		t.Fatal("expected context traceparent")
	}
}

func TestNoopDoesNotInject(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, _ := NewNoop().StartRequest(req.Context(), req, "http.server")
	header := make(http.Header)
	Inject(ctx, header)
	if header.Get(HeaderTraceparent) != "" {
		t.Fatal("noop must not write traceparent")
	}
}

func TestNormalizeClampsSamplingRate(t *testing.T) {
	cfg := Normalize(Config{Enabled: true, SamplingRate: 2, Protocol: "GRPC", Endpoint: "http://127.0.0.1:4317"})
	if cfg.SamplingRate != 1 {
		t.Fatalf("sampling rate = %v", cfg.SamplingRate)
	}
	if cfg.Protocol != "grpc" {
		t.Fatalf("protocol = %s", cfg.Protocol)
	}
	if cfg.Endpoint != "127.0.0.1:4317" {
		t.Fatalf("endpoint = %s", cfg.Endpoint)
	}
	if !cfg.Insecure {
		t.Fatal("grpc should default to insecure")
	}
}

func TestBuildSelectsImplementation(t *testing.T) {
	t.Cleanup(func() { SetGlobal(NewNoop()) })

	noop, err := Build(Config{Enabled: false, Endpoint: "127.0.0.1:4317"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, _ := noop.StartRequest(req.Context(), req, "http.server")
	header := make(http.Header)
	noop.Inject(ctx, header)
	if header.Get(HeaderTraceparent) != "" {
		t.Fatal("disabled tracing must not inject")
	}

	prop, err := Build(Config{Enabled: true, SamplingRate: 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(HeaderTraceparent, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	ctx, _ = prop.StartRequest(req.Context(), req, "http.server")
	header = make(http.Header)
	prop.Inject(ctx, header)
	if ParseTraceparent(header.Get(HeaderTraceparent)).TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatal("propagator should continue incoming trace")
	}

	if _, err := Build(Config{Enabled: true, Endpoint: "127.0.0.1:4317"}, nil); err == nil {
		t.Fatal("missing exporter must fail when endpoint is set")
	}

	called := false
	fake := func(Config) (Tracer, error) {
		called = true
		return NewNoop(), nil
	}
	if _, err := Build(Config{Enabled: true, Endpoint: "127.0.0.1:4317"}, fake); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("exporter factory must be used when endpoint is set")
	}
}

func TestOpenSetsGlobalAndCloseRestoresNoop(t *testing.T) {
	t.Cleanup(func() { SetGlobal(NewNoop()) })

	if _, err := Open(Config{Enabled: true, SamplingRate: 1}, nil); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderTraceparent, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	ctx, _ := Global().StartRequest(req.Context(), req, "http.server")
	header := make(http.Header)
	Inject(ctx, header)
	if header.Get(HeaderTraceparent) == "" {
		t.Fatal("Open should install a process-wide tracer")
	}

	if err := Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	header = make(http.Header)
	Inject(req.Context(), header)
	if header.Get(HeaderTraceparent) != "" {
		t.Fatal("Close must restore noop")
	}
}
