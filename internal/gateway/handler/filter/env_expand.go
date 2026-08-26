package filter

import (
	"strings"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/pkg/syssetting"
)

// expandEnvValue 将配置值中的 ${NAME} 替换为租户全局环境变量。
// 不含占位符时原样返回，避免热路径上的无谓查找。
func expandEnvValue(ctx *core.Context, raw string) string {
	if raw == "" || !strings.Contains(raw, "${") {
		return raw
	}
	tenantID := ""
	if ctx != nil {
		if id, ok := ctx.GetString(constants.ContextKeyTenantID); ok {
			tenantID = id
		}
	}
	return syssetting.ExpandEnvVars(tenantID, raw)
}
