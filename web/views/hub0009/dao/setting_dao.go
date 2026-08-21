package dao

import (
	"context"
	"errors"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0009/models"
)

// ErrVersionConflict 保存时记录版本已变化，需重新读取后再改。
var ErrVersionConflict = errors.New("设置已被其他用户修改，请刷新后重试")

// SettingDAO 环境设置数据访问。
type SettingDAO struct {
	db database.Database
}

// NewSettingDAO 创建环境设置 DAO。
func NewSettingDAO(db database.Database) *SettingDAO {
	return &SettingDAO{db: db}
}

// Get 按租户与分组读取一行。不存在时返回 (nil, nil)。
func (dao *SettingDAO) Get(ctx context.Context, tenantId, groupCode string) (*models.SysSetting, error) {
	if empty.IsEmpty(tenantId) || empty.IsEmpty(groupCode) {
		return nil, errors.New("tenantId和groupCode不能为空")
	}
	query := `SELECT * FROM HUB_SYS_SETTING WHERE tenantId = ? AND groupCode = ? AND activeFlag = 'Y'`
	var row models.SysSetting
	err := dao.db.QueryOne(ctx, &row, query, []interface{}{tenantId, groupCode}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询环境设置失败")
	}
	return &row, nil
}

// ListByTenant 列出租户下全部有效分组。
func (dao *SettingDAO) ListByTenant(ctx context.Context, tenantId string) ([]*models.SysSetting, error) {
	if empty.IsEmpty(tenantId) {
		return nil, errors.New("tenantId不能为空")
	}
	query := `SELECT * FROM HUB_SYS_SETTING WHERE tenantId = ? AND activeFlag = 'Y'`
	var rows []*models.SysSetting
	if err := dao.db.Query(ctx, &rows, query, []interface{}{tenantId}, true); err != nil {
		if err == database.ErrRecordNotFound {
			return []*models.SysSetting{}, nil
		}
		return nil, huberrors.WrapError(err, "查询环境设置列表失败")
	}
	return rows, nil
}

// ListAll 列出全部有效分组，供启动时灌入进程缓存。
func (dao *SettingDAO) ListAll(ctx context.Context) ([]*models.SysSetting, error) {
	query := `SELECT * FROM HUB_SYS_SETTING WHERE activeFlag = 'Y'`
	var rows []*models.SysSetting
	if err := dao.db.Query(ctx, &rows, query, []interface{}{}, true); err != nil {
		if err == database.ErrRecordNotFound {
			return []*models.SysSetting{}, nil
		}
		return nil, huberrors.WrapError(err, "加载环境设置失败")
	}
	return rows, nil
}

// Upsert 按乐观锁写入分组内容。expectVersion 为 0 表示插入新行。
func (dao *SettingDAO) Upsert(ctx context.Context, tenantId, groupCode, content, operatorId string, expectVersion int) (int, error) {
	if empty.IsEmpty(tenantId) || empty.IsEmpty(groupCode) {
		return 0, errors.New("tenantId和groupCode不能为空")
	}
	now := time.Now()
	opr := random.GenerateUniqueStringWithPrefix("", 32)

	if expectVersion == 0 {
		row := &models.SysSetting{
			TenantId:       tenantId,
			GroupCode:      groupCode,
			Content:        content,
			AddTime:        now,
			AddWho:         operatorId,
			EditTime:       now,
			EditWho:        operatorId,
			OprSeqFlag:     opr,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}
		if _, err := dao.db.Insert(ctx, row.TableName(), row, true); err != nil {
			// 并发下可能已有行，改为按版本 1 再读，避免把冲突当普通错误
			exist, getErr := dao.Get(ctx, tenantId, groupCode)
			if getErr != nil {
				return 0, huberrors.WrapError(err, "保存环境设置失败")
			}
			if exist != nil {
				return 0, ErrVersionConflict
			}
			return 0, huberrors.WrapError(err, "保存环境设置失败")
		}
		return 1, nil
	}

	sql := `
		UPDATE HUB_SYS_SETTING SET
			content = ?, editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = ?
		WHERE tenantId = ? AND groupCode = ? AND currentVersion = ? AND activeFlag = 'Y'
	`
	affected, err := dao.db.Exec(ctx, sql, []interface{}{
		content, now, operatorId, opr, expectVersion + 1,
		tenantId, groupCode, expectVersion,
	}, true)
	if err != nil {
		return 0, huberrors.WrapError(err, "更新环境设置失败")
	}
	if affected == 0 {
		return 0, ErrVersionConflict
	}
	return expectVersion + 1, nil
}
