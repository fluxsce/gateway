package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gateway/web/globalmodels"
	"gateway/web/utils/response"

	"github.com/gin-gonic/gin"
)

// captureWriter 测试用写入器，只保留最近一条事件。
type captureWriter struct {
	event *AuditEvent
}

func (w *captureWriter) WriteAudit(_ context.Context, event *AuditEvent) {
	if event == nil {
		w.event = nil
		return
	}
	clone := *event
	w.event = &clone
}

func testUser() *globalmodels.UserContext {
	return &globalmodels.UserContext{UserId: "u1", TenantId: "t1", UserName: "alice", ClientIP: "1.1.1.1"}
}

func setupFlush(t *testing.T) (*gin.Context, *captureWriter) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/gateway/hub0021/editRouteConfig", strings.NewReader(`{"routeConfigId":"R1"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(userContextKey, testUser())
	cap := &captureWriter{}
	Init(cap)
	t.Cleanup(func() { Init(nil) })
	return c, cap
}

func TestFlushWithoutEventSkips(t *testing.T) {
	c, cap := setupFlush(t)
	Flush(c)
	if cap.event != nil {
		t.Fatalf("no SetEvent must not audit: %+v", cap.event)
	}
}

func TestFlushMergesSetEvent(t *testing.T) {
	c, cap := setupFlush(t)
	SetEvent(c, &AuditEvent{
		Action:       AuditActionCreate,
		ModuleCode:   "hub0002",
		TargetType:   "USER",
		TargetId:     "n1",
		TargetName:   "bob",
		ResourceCode: "hub0002:add",
		Detail:       `{"userName":"bob"}`,
	})
	Flush(c)
	if cap.event == nil {
		t.Fatal("expected resource audit")
	}
	if cap.event.TargetType != "USER" || cap.event.TargetId != "n1" || cap.event.TargetName != "bob" {
		t.Fatalf("expected business target, got %+v", cap.event)
	}
	if cap.event.UserId != "u1" || cap.event.UserName != "alice" {
		t.Fatalf("expected actor filled, got %+v", cap.event)
	}
	if cap.event.Result != AuditResultSuccess {
		t.Fatalf("expected success, got %s", cap.event.Result)
	}
	if cap.event.Detail != `{"userName":"bob"}` {
		t.Fatalf("expected business detail, got %s", cap.event.Detail)
	}
}

func TestFlushFailureOverridesSetEvent(t *testing.T) {
	c, cap := setupFlush(t)
	SetEvent(c, &AuditEvent{
		TargetType:   "ROUTE",
		TargetId:     "R1",
		ResourceCode: "hub0021:edit",
		Detail:       "should be replaced",
	})
	response.MarkBizError(c, "route not found")
	Flush(c)
	if cap.event == nil {
		t.Fatal("expected audit")
	}
	if cap.event.Result != AuditResultFail {
		t.Fatalf("expected fail, got %s", cap.event.Result)
	}
	if cap.event.Detail != "route not found" {
		t.Fatalf("fail detail should be error message, got %q", cap.event.Detail)
	}
	if cap.event.TargetId != "R1" {
		t.Fatalf("target should remain, got %s", cap.event.TargetId)
	}
}
