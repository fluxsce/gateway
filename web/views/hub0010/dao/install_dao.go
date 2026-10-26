package dao

import (
	"context"
	"errors"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0010/models"
)

// InstallDAO 读写用户工具安装记录。
type InstallDAO struct {
	db database.Database
}

// NewInstallDAO 创建安装记录 DAO。
func NewInstallDAO(db database.Database) *InstallDAO {
	return &InstallDAO{db: db}
}

// ListActive 列出用户当前已安装的工具。
func (dao *InstallDAO) ListActive(ctx context.Context, tenantId, userId string) ([]*models.ToolInstall, error) {
	if empty.IsEmpty(tenantId) || empty.IsEmpty(userId) {
		return nil, errors.New("tenantId和userId不能为空")
	}
	query := `SELECT * FROM HUB_TOOL_INSTALL WHERE tenantId = ? AND userId = ? AND activeFlag = 'Y'`
	var rows []*models.ToolInstall
	if err := dao.db.Query(ctx, &rows, query, []interface{}{tenantId, userId}, true); err != nil {
		if err == database.ErrRecordNotFound {
			return []*models.ToolInstall{}, nil
		}
		return nil, huberrors.WrapError(err, "查询已安装工具失败")
	}
	return rows, nil
}

// Get 读取一行，包含已卸载记录。不存在时返回 (nil, nil)。
func (dao *InstallDAO) Get(ctx context.Context, tenantId, userId, toolId string) (*models.ToolInstall, error) {
	if empty.IsEmpty(tenantId) || empty.IsEmpty(userId) || empty.IsEmpty(toolId) {
		return nil, errors.New("tenantId、userId和toolId不能为空")
	}
	query := `SELECT * FROM HUB_TOOL_INSTALL WHERE tenantId = ? AND userId = ? AND toolId = ?`
	var row models.ToolInstall
	err := dao.db.QueryOne(ctx, &row, query, []interface{}{tenantId, userId, toolId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询工具安装记录失败")
	}
	return &row, nil
}

// Save 安装或更新配置。active 为 Y 表示已安装，N 表示卸载。
func (dao *InstallDAO) Save(ctx context.Context, tenantId, userId, toolId, configJson, operatorId, active string) error {
	if active != "Y" && active != "N" {
		return errors.New("activeFlag只能是Y或N")
	}
	now := time.Now()
	exist, err := dao.Get(ctx, tenantId, userId, toolId)
	if err != nil {
		return err
	}
	opr := random.GenerateUniqueStringWithPrefix("", 32)
	if exist == nil {
		row := &models.ToolInstall{
			TenantId:       tenantId,
			UserId:         userId,
			ToolId:         toolId,
			ConfigJson:     configJson,
			InstallTime:    now,
			AddTime:        now,
			AddWho:         operatorId,
			EditTime:       now,
			EditWho:        operatorId,
			OprSeqFlag:     opr,
			CurrentVersion: 1,
			ActiveFlag:     active,
		}
		if _, err := dao.db.Insert(ctx, row.TableName(), row, true); err != nil {
			return huberrors.WrapError(err, "保存工具安装记录失败")
		}
		return nil
	}
	installTime := exist.InstallTime
	if active == "Y" && exist.ActiveFlag != "Y" {
		installTime = now
	}
	sql := `
		UPDATE HUB_TOOL_INSTALL SET
			configJson = ?, installTime = ?, editTime = ?, editWho = ?, oprSeqFlag = ?,
			currentVersion = ?, activeFlag = ?
		WHERE tenantId = ? AND userId = ? AND toolId = ?
	`
	if _, err := dao.db.Exec(ctx, sql, []interface{}{
		configJson, installTime, now, operatorId, opr, exist.CurrentVersion + 1, active,
		tenantId, userId, toolId,
	}, true); err != nil {
		return huberrors.WrapError(err, "更新工具安装记录失败")
	}
	return nil
}
