package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/types"
)

type holdResponseWriter struct {
	http.ResponseWriter
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (w *holdResponseWriter) Write(p []byte) (int, error) {
	w.once.Do(func() { close(w.started) })
	<-w.release
	return w.ResponseWriter.Write(p)
}

func TestRecordedResponseBodyDurationStopsBeforeClientWrite(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	writer := &holdResponseWriter{
		ResponseWriter: httptest.NewRecorder(),
		started:        started,
		release:        release,
	}
	request := httptest.NewRequest(http.MethodGet, "http://gateway/sku", nil)
	ctx := core.NewContext(writer, request)
	ctx.SetLogConfig(&types.LogConfig{RecordResponseBody: "Y"})
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("sku-body")),
	}

	done := make(chan struct{})
	var backendDone time.Time
	var err error
	go func() {
		backendDone, err = (&HTTPProxy{}).handleRegularResponse(ctx, response)
		close(done)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("写给调用方没有开始")
	}
	time.Sleep(20 * time.Millisecond)
	beforeRelease := time.Now()
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("响应处理没有结束")
	}
	if err != nil {
		t.Fatal(err)
	}
	if backendDone.IsZero() || !backendDone.Before(beforeRelease) {
		t.Fatalf("后端耗时应停在写给调用方之前: done=%v release=%v", backendDone, beforeRelease)
	}
	saved, ok := ctx.Get("response_body")
	if !ok || string(saved.([]byte)) != "sku-body" {
		t.Fatalf("响应体未记录: %v", saved)
	}
}
