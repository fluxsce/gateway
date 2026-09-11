package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"
	"gateway/internal/gateway/logwrite/types"
)

func TestPrepareForwardAssistStreamsWithoutCopyWhenNoRetryNoRecord(t *testing.T) {
	payload := strings.Repeat("x", 64)
	req := httptest.NewRequest(http.MethodPost, "http://gateway/p", strings.NewReader(payload))
	fwd, err := prepareForwardAssist(req, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if fwd.full != nil {
		t.Fatal("不重试且不记日志时不应整包缓冲")
	}
	got, err := io.ReadAll(fwd.Body())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != payload {
		t.Fatalf("上游应收到完整体, got len=%d", len(got))
	}
}

func TestPrepareForwardAssistKeepsFullBodyWhenRecordingWithoutRetry(t *testing.T) {
	payload := []byte("abcdefghijklmnop")
	req := httptest.NewRequest(http.MethodPost, "http://gateway/p", bytes.NewReader(payload))
	fwd, err := prepareForwardAssist(req, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fwd.full, payload) {
		t.Fatalf("记日志应整包缓冲, got %q", fwd.full)
	}
	got, err := io.ReadAll(fwd.Body())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("上游应收到完整体, got %q", got)
	}
	ctx := core.NewContext(httptest.NewRecorder(), req)
	fwd.CommitBody(ctx)
	body, ok := ctx.Get("request_body")
	if !ok {
		t.Fatal("记日志时应写入 request_body")
	}
	if stored, _ := body.([]byte); !bytes.Equal(stored, payload) {
		t.Fatalf("request_body 应是完整转发体, got %q", stored)
	}
	if fwd.KnownSize() != len(payload) {
		t.Fatalf("requestSize 应为完整长度 %d, got %d", len(payload), fwd.KnownSize())
	}
}

func TestPrepareForwardAssistBuffersFullBodyWhenRetryEnabled(t *testing.T) {
	payload := []byte("retry-body-full")
	req := httptest.NewRequest(http.MethodPost, "http://gateway/p", bytes.NewReader(payload))
	fwd, err := prepareForwardAssist(req, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fwd.full, payload) {
		t.Fatalf("重试应整包缓冲, got %q", fwd.full)
	}
	first, _ := io.ReadAll(fwd.Body())
	second, _ := io.ReadAll(req.Body)
	if !bytes.Equal(first, payload) || !bytes.Equal(second, payload) {
		t.Fatalf("重试后 Body 应仍能读完整包 first=%q second=%q", first, second)
	}
	ctx := core.NewContext(httptest.NewRecorder(), req)
	fwd.CommitBody(ctx)
	body, _ := ctx.Get("request_body")
	if got, _ := body.([]byte); !bytes.Equal(got, payload) {
		t.Fatalf("重试记日志应写入完整转发体, got %q", got)
	}
}

func TestPrepareForwardAssistRecordsGETWithBody(t *testing.T) {
	payload := []byte(`{"q":1}`)
	req := httptest.NewRequest(http.MethodGet, "http://gateway/search", bytes.NewReader(payload))
	fwd, err := prepareForwardAssist(req, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fwd.full, payload) {
		t.Fatalf("GET 带 body 且记日志应整包缓冲, got %q", fwd.full)
	}
	ctx := core.NewContext(httptest.NewRecorder(), req)
	fwd.CommitBody(ctx)
	body, ok := ctx.Get("request_body")
	if !ok {
		t.Fatal("GET 带 body 应写入 request_body")
	}
	if stored, _ := body.([]byte); !bytes.Equal(stored, payload) {
		t.Fatalf("GET request_body 应是完整转发体, got %q", stored)
	}
}

func TestForwardAssistCloneHeaderDeepCopy(t *testing.T) {
	src := http.Header{
		"X-Trace": []string{"a", "b"},
		"Accept":  []string{"*/*"},
	}
	got := new(forwardAssist).CloneHeader(src)
	if got.Get("X-Trace") != "a" || len(got["X-Trace"]) != 2 {
		t.Fatalf("克隆结果不正确: %v", got)
	}
	src.Set("X-Trace", "mutated")
	src["Accept"][0] = "text/plain"
	if got.Get("X-Trace") != "a" || got.Get("Accept") != "*/*" {
		t.Fatal("克隆后改原 header 不应影响副本")
	}
}

func TestForwardAssistCloneHeaderEmpty(t *testing.T) {
	var assist *forwardAssist
	if assist.CloneHeader(nil) != nil || assist.CloneHeader(http.Header{}) != nil {
		t.Fatal("空头应返回 nil，nil 接收者也应可用")
	}
}

func TestHTTPProxyKeepsFullRecordedBodyInContext(t *testing.T) {
	payload := []byte("abcdefghijklmnop")
	var upstreamBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	manager := service.NewServiceManager()
	if err := manager.AddService(&service.ServiceConfig{
		ID:       "body-service",
		Name:     "body-service",
		Strategy: service.RoundRobin,
		Nodes: []*service.NodeConfig{{
			ID: "body-node", URL: upstream.URL, Weight: 1, Health: true, Enabled: true,
		}},
	}); err != nil {
		t.Fatal(err)
	}
	httpProxy, err := NewHTTPProxy(ProxyConfig{
		Type:    ProxyTypeHTTP,
		Enabled: true,
		Name:    "body-proxy",
	}, manager)
	if err != nil {
		t.Fatal(err)
	}
	defer httpProxy.Close()

	req := httptest.NewRequest(http.MethodPost, "http://gateway/p", bytes.NewReader(payload))
	req.ContentLength = int64(len(payload))
	ctx := core.NewContext(httptest.NewRecorder(), req)
	ctx.SetServiceIDs([]string{"body-service"})
	ctx.SetLogConfig(&types.LogConfig{
		RecordRequestBody: "Y",
		MaxBodySizeBytes:  8,
	})
	if !httpProxy.Handle(ctx) {
		t.Fatalf("代理失败: %v", ctx.GetErrors())
	}
	if !bytes.Equal(upstreamBody, payload) {
		t.Fatalf("上游应收到完整体, got %q", upstreamBody)
	}
	body, ok := ctx.Get("request_body")
	if !ok {
		t.Fatal("未写入 request_body")
	}
	if stored, _ := body.([]byte); !bytes.Equal(stored, payload) {
		t.Fatalf("request_body 应是完整转发体, 截断留给写库, got %q", stored)
	}
}

func TestHTTPProxyRetriesWithFullBufferedBody(t *testing.T) {
	payload := []byte("full-retry-payload")
	var seen [][]byte
	attempts := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		seen = append(seen, append([]byte(nil), body...))
		attempts++
		if attempts == 1 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("无法 Hijack")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatal(err)
			}
			_ = conn.Close()
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	manager := service.NewServiceManager()
	if err := manager.AddService(&service.ServiceConfig{
		ID:       "retry-service",
		Name:     "retry-service",
		Strategy: service.RoundRobin,
		Nodes: []*service.NodeConfig{{
			ID: "retry-node", URL: upstream.URL, Weight: 1, Health: true, Enabled: true,
		}},
	}); err != nil {
		t.Fatal(err)
	}
	httpProxy, err := NewHTTPProxy(ProxyConfig{
		Type:    ProxyTypeHTTP,
		Enabled: true,
		Name:    "retry-proxy",
		Config: map[string]interface{}{
			"retryCount":   1,
			"retryTimeout": "1ms",
		},
	}, manager)
	if err != nil {
		t.Fatal(err)
	}
	defer httpProxy.Close()

	req := httptest.NewRequest(http.MethodPost, "http://gateway/p", bytes.NewReader(payload))
	ctx := core.NewContext(httptest.NewRecorder(), req)
	ctx.SetServiceIDs([]string{"retry-service"})
	ctx.SetLogConfig(&types.LogConfig{RecordRequestBody: "Y", MaxBodySizeBytes: 4})
	if !httpProxy.Handle(ctx) {
		t.Fatalf("重试后应成功: %v", ctx.GetErrors())
	}
	if len(seen) < 2 {
		t.Fatalf("应重试一次, attempts=%d", len(seen))
	}
	for i, body := range seen {
		if !bytes.Equal(body, payload) {
			t.Fatalf("第 %d 次上游未收到完整体: %q", i+1, body)
		}
	}
	body, _ := ctx.Get("request_body")
	if got, _ := body.([]byte); !bytes.Equal(got, payload) {
		t.Fatalf("重试路径 request_body 应是完整转发体, got %q", got)
	}
}
