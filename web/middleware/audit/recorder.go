// Package audit 提供控制面写操作审计的事件、落库与请求上下文辅助。
// Gin 中间件在 web/middleware/audit_middleware.go；鉴权中间件不参与审计。
// 业务 SetEvent 放入上下文后由全局中间件落库；登录无 session 时用 WriteDirect；踢会话当场写入。
package audit

import (
	"context"
	"net/http"

	"gateway/web/globalmodels"
	"gateway/web/utils/response"

	"github.com/gin-gonic/gin"
)

const (
	// eventKey 业务放入的资源事件，由审计中间件在请求结束时落库。
	eventKey = "audit.event"
	// flushedKey 本请求已落库，避免中间件重复 Flush。
	flushedKey = "audit.flushed"
	// userContextKey 与 middleware 包中的用户上下文键保持一致，本包不引用 middleware 以免循环依赖。
	userContextKey  = "userContext"
	defaultTenantId = "default"
)

// Writer 审计落库接口，由 DAO 实现。
type Writer interface {
	WriteAudit(ctx context.Context, event *AuditEvent)
}

var defaultWriter Writer

// Init 注入审计写入实现，应用启动时调用。
func Init(w Writer) {
	defaultWriter = w
}

func writer() Writer {
	return defaultWriter
}

func userFromGin(c *gin.Context) *globalmodels.UserContext {
	if c == nil {
		return nil
	}
	value, exists := c.Get(userContextKey)
	if !exists || value == nil {
		return nil
	}
	user, ok := value.(*globalmodels.UserContext)
	if !ok {
		return nil
	}
	return user
}

func requestPath(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if full := c.FullPath(); full != "" {
		return full
	}
	if c.Request != nil && c.Request.URL != nil {
		return c.Request.URL.Path
	}
	return ""
}

func requestMeta(c *gin.Context) (path, method string) {
	path = requestPath(c)
	if c != nil && c.Request != nil {
		method = c.Request.Method
		if c.Request.URL != nil && c.Request.URL.Path != "" {
			path = c.Request.URL.Path
		}
	}
	return path, method
}

func eventFrom(c *gin.Context) *AuditEvent {
	if c == nil {
		return nil
	}
	raw, ok := c.Get(eventKey)
	if !ok || raw == nil {
		return nil
	}
	event, _ := raw.(*AuditEvent)
	return event
}

func alreadyFlushed(c *gin.Context) bool {
	if c == nil {
		return true
	}
	raw, ok := c.Get(flushedKey)
	if !ok {
		return false
	}
	flag, _ := raw.(bool)
	return flag
}

func requestFailed(c *gin.Context) (failed bool, detail string) {
	if c == nil {
		return false, ""
	}
	failed = response.IsBizError(c) || c.Writer.Status() >= http.StatusBadRequest || c.IsAborted()
	if !failed {
		return false, ""
	}
	detail = response.BizErrorMessage(c)
	if detail == "" {
		detail = http.StatusText(c.Writer.Status())
	}
	return true, detail
}

// SetEvent 把资源事件放入上下文，由全局审计中间件在 handler 返回后落库。
// 调用方填写 Action、ModuleCode、TargetType、TargetId、TargetName、ResourceCode、Detail；
// 身份、请求信息与最终 result 由中间件补齐。未 SetEvent 的请求不记审计。
func SetEvent(c *gin.Context, event *AuditEvent) {
	if c == nil || event == nil {
		return
	}
	c.Set(eventKey, event)
}

// WriteFromGin 等同 SetEvent，保留给旧调用方。
func WriteFromGin(c *gin.Context, event *AuditEvent) {
	SetEvent(c, event)
}

// Flush 将业务放入上下文的审计事件写入账本。无 SetEvent 则跳过。同一请求只落一次。
func Flush(c *gin.Context) {
	w := writer()
	if w == nil || c == nil || alreadyFlushed(c) {
		return
	}
	event := eventFrom(c)
	if event == nil {
		return
	}
	user := userFromGin(c)
	if user == nil {
		return
	}

	fillActor(c, event, user)
	failed, failDetail := requestFailed(c)
	if failed {
		event.Result = AuditResultFail
		event.Detail = failDetail
	} else if event.Result == "" {
		event.Result = AuditResultSuccess
	}

	c.Set(flushedKey, true)
	ctx := context.Background()
	if c.Request != nil {
		ctx = c.Request.Context()
	}
	w.WriteAudit(ctx, event)
}

func fillActor(c *gin.Context, event *AuditEvent, user *globalmodels.UserContext) {
	path, method := requestMeta(c)
	if event.UserId == "" {
		event.UserId = user.UserId
	}
	if event.TenantId == "" {
		event.TenantId = user.TenantId
	}
	if event.UserName == "" {
		event.UserName = user.UserName
	}
	if event.RequestPath == "" {
		event.RequestPath = path
	}
	if event.RequestMethod == "" {
		event.RequestMethod = method
	}
	if event.ClientIP == "" {
		event.ClientIP = user.ClientIP
	}
	if event.Action == "" && event.ResourceCode != "" {
		event.Action = AuditActionFromResourceCode(event.ResourceCode)
	}
	if event.ModuleCode == "" {
		event.ModuleCode = ModuleCodeFromResourceOrPath(event.ResourceCode, path)
	}
}

func writeFromGin(c *gin.Context, event *AuditEvent) {
	w := writer()
	if w == nil || event == nil {
		return
	}
	userContext := userFromGin(c)
	if userContext == nil {
		return
	}
	fillActor(c, event, userContext)
	if event.Result == "" {
		event.Result = AuditResultSuccess
	}
	ctx := context.Background()
	if c != nil && c.Request != nil {
		ctx = c.Request.Context()
	}
	w.WriteAudit(ctx, event)
}

// WriteDirect 无 session 也可写审计（登录成功、进入冷却）。tenantId 为空时回退 default。
func WriteDirect(c *gin.Context, event *AuditEvent) {
	w := writer()
	if w == nil || event == nil {
		return
	}
	if event.TenantId == "" {
		event.TenantId = defaultTenantId
	}
	path, method := requestMeta(c)
	if event.RequestPath == "" {
		event.RequestPath = path
	}
	if event.RequestMethod == "" {
		event.RequestMethod = method
	}
	if event.Result == "" {
		event.Result = AuditResultSuccess
	}
	ctx := context.Background()
	if c != nil && c.Request != nil {
		ctx = c.Request.Context()
	}
	w.WriteAudit(ctx, event)
}
