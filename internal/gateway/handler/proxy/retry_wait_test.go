package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gateway/internal/gateway/core"
)

func TestWaitRetryIntervalRecordsActualWait(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	gwCtx := core.NewContext(httptest.NewRecorder(), req)
	ok := waitRetryInterval(gwCtx, 20*time.Millisecond)
	if !ok {
		t.Fatal("wait should complete")
	}
	waited := gwCtx.GetRetryWait()
	if waited < 15*time.Millisecond || waited > 200*time.Millisecond {
		t.Fatalf("retry wait = %v, want around 20ms", waited)
	}
}

func TestWaitRetryIntervalCancelled(t *testing.T) {
	reqCtx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(reqCtx)
	gwCtx := core.NewContext(httptest.NewRecorder(), req)
	if waitRetryInterval(gwCtx, time.Second) {
		t.Fatal("cancelled request should stop retry wait")
	}
}
