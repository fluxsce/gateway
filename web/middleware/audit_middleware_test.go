package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/web/globalmodels"
	"gateway/web/middleware/audit"
	"gateway/web/utils/response"

	"github.com/gin-gonic/gin"
)

type captureWriter struct {
	event *audit.AuditEvent
}

func (w *captureWriter) WriteAudit(_ context.Context, event *audit.AuditEvent) {
	if event == nil {
		w.event = nil
		return
	}
	clone := *event
	w.event = &clone
}

func TestAuditMiddlewareFlushesAfterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cap := &captureWriter{}
	audit.Init(cap)
	t.Cleanup(func() { audit.Init(nil) })

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userContext", &globalmodels.UserContext{UserId: "u1", TenantId: "t1", UserName: "alice"})
		c.Next()
	})
	r.Use(AuditMiddleware())
	r.POST("/t", func(c *gin.Context) {
		audit.SetEvent(c, &audit.AuditEvent{
			Action:       audit.AuditActionCreate,
			TargetType:   "USER",
			TargetId:     "n1",
			TargetName:   "bob",
			ResourceCode: "hub0002:add",
			ModuleCode:   "hub0002",
		})
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/t", nil))
	if cap.event == nil {
		t.Fatal("middleware should flush after handler")
	}
	if cap.event.TargetId != "n1" || cap.event.Action != audit.AuditActionCreate {
		t.Fatalf("unexpected event: %+v", cap.event)
	}
}

func TestAuditMiddlewareSkipsWithoutEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cap := &captureWriter{}
	audit.Init(cap)
	t.Cleanup(func() { audit.Init(nil) })

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userContext", &globalmodels.UserContext{UserId: "u1", TenantId: "t1", UserName: "alice"})
		c.Next()
	})
	r.Use(AuditMiddleware())
	r.POST("/t", func(c *gin.Context) {
		response.ErrorJSON(c, "没有执行此操作的权限", "ED00010", http.StatusForbidden)
		c.Abort()
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/t", nil))
	if cap.event != nil {
		t.Fatalf("permission deny without SetEvent must not audit: %+v", cap.event)
	}
}
