package dao

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"time"

	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0043/models"
)

// ConfigDAO 配置数据访问对象
type ConfigDAO struct {
	db database.Database
}

// NewConfigDAO 创建配置DAO
func NewConfigDAO(db database.Database) *ConfigDAO {
	return &ConfigDAO{
		db: db,
	}
}

// GetConfigById 根据主键获取配置信息
func (dao *ConfigDAO) GetConfigById(ctx context.Context, tenantId, namespaceId, groupName, configDataId string) (*types.ConfigData, error) {
	if namespaceId == "" || groupName == "" || configDataId == "" {
		return nil, errors.New("namespaceId、groupName和configDataId不能为空")
	}

	query := `
		SELECT * FROM HUB_SERVICE_CONFIG_DATA 
		WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?
	`

	var config types.ConfigData
	err := dao.db.QueryOne(ctx, &config, query, []interface{}{tenantId, namespaceId, groupName, configDataId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询配置失败")
	}

	return &config, nil
}

// InsertConfig 插入新配置
// 参数:
//   - ctx: 上下文对象
//   - config: 配置数据（必须包含所有必填字段）
//
// 返回:
//   - error: 可能的错误
func (dao *ConfigDAO) InsertConfig(ctx context.Context, config *types.ConfigData) error {
	if config == nil {
		return errors.New("配置数据不能为空")
	}

	// 注意：tenantId由前置校验保证非空，此处不再校验
	if config.NamespaceId == "" || config.GroupName == "" || config.ConfigDataId == "" {
		return errors.New("namespaceId、groupName和configDataId不能为空")
	}

	// 计算MD5值
	hash := md5.Sum([]byte(config.ConfigContent))
	config.Md5Value = fmt.Sprintf("%x", hash)

	// 设置默认值
	if config.Version == 0 {
		config.Version = 1
	}
	if config.AddTime.IsZero() {
		config.AddTime = time.Now()
	}
	if config.EditTime.IsZero() {
		config.EditTime = time.Now()
	}
	if config.ActiveFlag == "" {
		config.ActiveFlag = "Y"
	}

	// 插入配置
	_, err := dao.db.Insert(ctx, "HUB_SERVICE_CONFIG_DATA", config, true)
	if err != nil {
		return huberrors.WrapError(err, "插入配置失败")
	}

	return nil
}

// ListConfigs 获取配置列表（支持条件查询和分页）
// 注意：tenantId由前置校验保证非空，此处不再校验
func (dao *ConfigDAO) ListConfigs(ctx context.Context, tenantId string, query *models.ConfigQuery, page, pageSize int) ([]*types.ConfigData, int, error) {
	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建查询条件
	whereClause := "WHERE c.tenantId = ?"
	var params []interface{}
	params = append(params, tenantId)

	// 构建基础查询语句（列表查询不包含大字段 configContent，减少内存开销）
	// 只查询列表展示需要的字段，详情查询时再获取完整信息
	baseQuery := `SELECT 
		c.configDataId,
		c.tenantId,
		c.namespaceId,
		c.groupName,
		c.contentType,
		c.configDescription,
		c.encrypted,
		c.version,
		c.md5Value,
		c.addTime,
		c.addWho,
		c.editTime,
		c.editWho,
		c.oprSeqFlag,
		c.currentVersion,
		c.activeFlag,
		c.noteText,
		c.extProperty
		FROM HUB_SERVICE_CONFIG_DATA c`

	// 构建查询条件，只有当字段不为空时才添加对应条件
	if query != nil {
		if !empty.IsEmpty(query.NamespaceId) {
			whereClause += " AND c.namespaceId = ?"
			params = append(params, query.NamespaceId)
		}
		if !empty.IsEmpty(query.GroupName) {
			whereClause += " AND c.groupName = ?"
			params = append(params, query.GroupName)
		}
		if !empty.IsEmpty(query.ConfigDataId) {
			whereClause += " AND c.configDataId LIKE ?"
			params = append(params, "%"+query.ConfigDataId+"%")
		}
		if !empty.IsEmpty(query.ContentType) {
			whereClause += " AND c.contentType = ?"
			params = append(params, query.ContentType)
		}
		if !empty.IsEmpty(query.ActiveFlag) {
			whereClause += " AND c.activeFlag = ?"
			params = append(params, query.ActiveFlag)
		} else {
			// 默认只查询活动状态的配置
			whereClause += " AND c.activeFlag = 'Y'"
		}
	} else {
		// 默认只查询活动状态的配置
		whereClause += " AND c.activeFlag = 'Y'"
	}

	// 完整查询语句
	fullQuery := baseQuery + " " + whereClause + `
		ORDER BY c.namespaceId, c.groupName, c.configDataId, c.addTime DESC
	`

	// 构建计数查询
	countQuery, err := sqlutils.BuildCountQuery(fullQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建计数查询失败")
	}

	// 执行计数查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询配置总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*types.ConfigData{}, 0, nil
	}

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, fullQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数：基础查询参数 + 分页参数
	queryArgs := params
	queryArgs = append(queryArgs, paginationArgs...)

	// 执行分页查询
	var configs []*types.ConfigData
	err = dao.db.Query(ctx, &configs, paginatedQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询配置列表失败")
	}

	return configs, total, nil
}

// UpdateConfig 更新配置
// 参数:
//   - ctx: 上下文对象
//   - config: 配置数据（必须包含主键字段）
//
// 返回:
//   - error: 可能的错误
func (dao *ConfigDAO) UpdateConfig(ctx context.Context, config *types.ConfigData) error {
	if config == nil {
		return errors.New("配置数据不能为空")
	}

	// 注意：tenantId由前置校验保证非空，此处不再校验
	if config.NamespaceId == "" || config.GroupName == "" || config.ConfigDataId == "" {
		return errors.New("namespaceId、groupName和configDataId不能为空")
	}

	// 计算MD5值
	hash := md5.Sum([]byte(config.ConfigContent))
	config.Md5Value = fmt.Sprintf("%x", hash)

	// 更新时，版本号递增
	config.Version++
	// EditTime 应该由 controller 设置，如果未设置则使用当前时间
	if config.EditTime.IsZero() {
		config.EditTime = time.Now()
	}

	where := "tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?"
	args := []interface{}{config.TenantId, config.NamespaceId, config.GroupName, config.ConfigDataId}

	_, err := dao.db.Update(ctx, "HUB_SERVICE_CONFIG_DATA", config, where, args, true, true)
	if err != nil {
		return huberrors.WrapError(err, "更新配置失败")
	}

	return nil
}

// DeleteConfig 删除配置（物理删除）
// 参数:
//   - ctx: 上下文对象
//   - tenantId: 租户ID（由前置校验保证非空，此处不再校验）
//   - namespaceId: 命名空间ID
//   - groupName: 分组名称
//   - configDataId: 配置数据ID
//
// 返回:
//   - error: 可能的错误
func (dao *ConfigDAO) DeleteConfig(ctx context.Context, tenantId, namespaceId, groupName, configDataId string) error {
	// 注意：tenantId由前置校验保证非空，此处不再校验
	if namespaceId == "" || groupName == "" || configDataId == "" {
		return errors.New("namespaceId、groupName和configDataId不能为空")
	}

	query := "DELETE FROM HUB_SERVICE_CONFIG_DATA WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?"
	args := []interface{}{tenantId, namespaceId, groupName, configDataId}

	_, err := dao.db.Exec(ctx, query, args, true)
	if err != nil {
		return huberrors.WrapError(err, "删除配置失败")
	}

	return nil
}
