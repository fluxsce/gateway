package dao

import (
	"context"
	"errors"
	"fmt"

	alerttypes "gateway/internal/alert/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0081/models"
)

// AlertTemplateDAO 预警模板DAO，对应表 HUB_ALERT_TEMPLATE
type AlertTemplateDAO struct {
	db database.Database
}

func NewAlertTemplateDAO(db database.Database) *AlertTemplateDAO {
	return &AlertTemplateDAO{db: db}
}

func (dao *AlertTemplateDAO) GetAlertTemplate(ctx context.Context, tenantId, templateName string) (*alerttypes.AlertTemplate, error) {
	if templateName == "" {
		return nil, errors.New("templateName不能为空")
	}

	query := `SELECT * FROM HUB_ALERT_TEMPLATE WHERE tenantId = ? AND templateName = ?`
	args := []interface{}{tenantId, templateName}

	var tpl alerttypes.AlertTemplate
	err := dao.db.QueryOne(ctx, &tpl, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询预警模板失败")
	}
	return &tpl, nil
}

// QueryAlertTemplates 分页查询预警模板
func (dao *AlertTemplateDAO) QueryAlertTemplates(ctx context.Context, tenantId string, q *models.AlertTemplateQueryRequest, page, pageSize int) ([]*alerttypes.AlertTemplate, int, error) {
	pagination := sqlutils.NewPaginationInfo(page, pageSize)
	dbType := sqlutils.GetDatabaseType(dao.db)

	whereClause := "WHERE tenantId = ?"
	params := []interface{}{tenantId}

	if q != nil {
		if !empty.IsEmpty(q.TemplateName) {
			whereClause += " AND templateName = ?"
			params = append(params, q.TemplateName)
		}
		if !empty.IsEmpty(q.ChannelType) {
			whereClause += " AND channelType = ?"
			params = append(params, q.ChannelType)
		}
		if !empty.IsEmpty(q.DisplayFormat) {
			whereClause += " AND displayFormat = ?"
			params = append(params, q.DisplayFormat)
		}
		if !empty.IsEmpty(q.ActiveFlag) {
			whereClause += " AND activeFlag = ?"
			params = append(params, q.ActiveFlag)
		}
	}

	baseQuery := fmt.Sprintf(`
		SELECT * FROM HUB_ALERT_TEMPLATE
		%s
		ORDER BY editTime DESC
	`, whereClause)

	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	if err := dao.db.QueryOne(ctx, &countResult, countQuery, params, true); err != nil {
		return nil, 0, huberrors.WrapError(err, "查询预警模板总数失败")
	}
	if countResult.Count == 0 {
		return []*alerttypes.AlertTemplate{}, 0, nil
	}

	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	allArgs := append(params, paginationArgs...)
	var rows []*alerttypes.AlertTemplate
	if err := dao.db.Query(ctx, &rows, paginatedQuery, allArgs, true); err != nil {
		return nil, 0, huberrors.WrapError(err, "查询预警模板失败")
	}
	return rows, countResult.Count, nil
}

func (dao *AlertTemplateDAO) CreateAlertTemplate(ctx context.Context, tpl *alerttypes.AlertTemplate) error {
	if tpl == nil {
		return errors.New("tpl不能为空")
	}
	_, err := dao.db.Insert(ctx, "HUB_ALERT_TEMPLATE", tpl, true)
	if err != nil {
		return huberrors.WrapError(err, "创建预警模板失败")
	}
	return nil
}

func (dao *AlertTemplateDAO) UpdateAlertTemplate(ctx context.Context, tpl *alerttypes.AlertTemplate) error {
	if tpl == nil {
		return errors.New("tpl不能为空")
	}
	where := "tenantId = ? AND templateName = ?"
	args := []interface{}{tpl.TenantId, tpl.TemplateName}
	_, err := dao.db.Update(ctx, "HUB_ALERT_TEMPLATE", tpl, where, args, true)
	if err != nil {
		return huberrors.WrapError(err, "更新预警模板失败")
	}
	return nil
}

func (dao *AlertTemplateDAO) DeleteAlertTemplate(ctx context.Context, tenantId, templateName string) error {
	if templateName == "" {
		return errors.New("templateName不能为空")
	}
	_, err := dao.db.Exec(ctx, "DELETE FROM HUB_ALERT_TEMPLATE WHERE tenantId = ? AND templateName = ?", []interface{}{tenantId, templateName}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除预警模板失败")
	}
	return nil
}
