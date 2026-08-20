package middleware

import (
	"gateway/pkg/database"
	"gateway/web/middleware/audit"

	"github.com/gin-gonic/gin"
)

// InitAudit 初始化审计写入器，应用启动时与权限服务一并调用。
func InitAudit(db database.Database) {
	if db == nil {
		return
	}
	audit.Init(audit.NewDAO(db))
}

// AuditMiddleware 写操作审计中间件。
// handler 返回后落库；仅当业务 SetEvent 时写入，鉴权中间件不参与。
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		audit.Flush(c)
	}
}
