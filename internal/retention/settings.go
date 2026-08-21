package retention

import (
	"context"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/syssetting"
)

type settingRow struct {
	TenantId  string `db:"tenantId"`
	GroupCode string `db:"groupCode"`
	Content   string `db:"content"`
}

// ReloadSettings 从 HUB_SYS_SETTING 灌入进程缓存。库中没有的分组仍由 Get 返回默认值。
func ReloadSettings(ctx context.Context, db database.Database) {
	if db == nil {
		return
	}
	var rows []*settingRow
	err := db.Query(ctx, &rows, `
		SELECT tenantId, groupCode, content FROM HUB_SYS_SETTING WHERE activeFlag = 'Y'
	`, []interface{}{}, true)
	if err != nil {
		if err != database.ErrRecordNotFound {
			logger.Warn("加载环境设置失败，本轮使用内存中的值", "error", err.Error())
		}
		return
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		syssetting.ApplyGroup(row.TenantId, row.GroupCode, row.Content)
	}
}

// ReloadGroup 从库读取指定租户分组并写入进程缓存。记录不存在时保持当前内存值。
func ReloadGroup(ctx context.Context, db database.Database, tenantId, groupCode string) error {
	if db == nil || tenantId == "" || groupCode == "" {
		return nil
	}
	var row settingRow
	err := db.QueryOne(ctx, &row, `
		SELECT tenantId, groupCode, content FROM HUB_SYS_SETTING
		WHERE tenantId = ? AND groupCode = ? AND activeFlag = 'Y'
	`, []interface{}{tenantId, groupCode}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil
		}
		return err
	}
	syssetting.ApplyGroup(row.TenantId, row.GroupCode, row.Content)
	return nil
}
