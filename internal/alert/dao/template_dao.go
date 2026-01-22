package dao

import (
	"context"
	"fmt"

	"gateway/internal/alert/types"
	"gateway/pkg/database"
)

// TemplateDAO 告警模板数据访问对象
type TemplateDAO struct {
	db database.Database
}

// NewTemplateDAO 创建模板DAO
func NewTemplateDAO(db database.Database) *TemplateDAO {
	return &TemplateDAO{db: db}
}

// GetTemplate 获取模板
func (d *TemplateDAO) GetTemplate(ctx context.Context, tenantId, templateName string) (*types.AlertTemplate, error) {
	query := "SELECT * FROM HUB_ALERT_TEMPLATE WHERE tenantId = ? AND templateName = ?"
	args := []interface{}{tenantId, templateName}

	var template types.AlertTemplate
	err := d.db.QueryOne(ctx, &template, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询告警模板失败: %w", err)
	}

	return &template, nil
}
