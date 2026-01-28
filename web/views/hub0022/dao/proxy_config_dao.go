package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/empty"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0022/models"
)

// ProxyConfigDAO 代理配置数据访问对象
type ProxyConfigDAO struct {
	db database.Database
}

// NewProxyConfigDAO 创建代理配置DAO实例
func NewProxyConfigDAO(db database.Database) *ProxyConfigDAO {
	return &ProxyConfigDAO{
		db: db,
	}
}

// CreateProxyConfig 创建代理配置
func (dao *ProxyConfigDAO) CreateProxyConfig(ctx context.Context, proxyConfig *models.ProxyConfig, operatorId string) (string, error) {
	// 验证必填字段
	if proxyConfig.GatewayInstanceId == "" {
		return "", errors.New("网关实例ID不能为空")
	}
	if proxyConfig.ProxyName == "" {
		return "", errors.New("代理名称不能为空")
	}
	if proxyConfig.ProxyType == "" {
		return "", errors.New("代理类型不能为空")
	}

	// 自动生成代理配置ID（如果为空）
	if proxyConfig.ProxyConfigId == "" {
		// 使用公共方法生成32位唯一字符串，前缀为"PC"
		proxyConfig.ProxyConfigId = random.GenerateUniqueStringWithPrefix("PC", 32)
	}

	// 设置自动填充的字段
	now := time.Now()
	proxyConfig.AddTime = now
	proxyConfig.AddWho = operatorId
	proxyConfig.EditTime = now
	proxyConfig.EditWho = operatorId
	proxyConfig.OprSeqFlag = proxyConfig.ProxyConfigId
	proxyConfig.CurrentVersion = 1
	proxyConfig.ActiveFlag = "Y"

	// 设置默认值
	if proxyConfig.ConfigPriority == 0 {
		proxyConfig.ConfigPriority = 100
	}

	// 插入记录
	_, err := dao.db.Insert(ctx, "HUB_GW_PROXY_CONFIG", proxyConfig, true)
	if err != nil {
		return "", huberrors.WrapError(err, "创建代理配置失败")
	}

	return proxyConfig.ProxyConfigId, nil
}

// GetProxyConfigById 根据ID获取代理配置
func (dao *ProxyConfigDAO) GetProxyConfigById(ctx context.Context, proxyConfigId, tenantId string) (*models.ProxyConfig, error) {
	if proxyConfigId == "" {
		return nil, errors.New("proxyConfigId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_PROXY_CONFIG 
		WHERE proxyConfigId = ? AND tenantId = ?
	`

	var proxyConfig models.ProxyConfig
	err := dao.db.QueryOne(ctx, &proxyConfig, query, []interface{}{proxyConfigId, tenantId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询代理配置失败")
	}

	return &proxyConfig, nil
}

// UpdateProxyConfig 更新代理配置
func (dao *ProxyConfigDAO) UpdateProxyConfig(ctx context.Context, proxyConfig *models.ProxyConfig, operatorId string) error {
	if proxyConfig.ProxyConfigId == "" {
		return errors.New("proxyConfigId不能为空")
	}

	// 验证必填字段
	if proxyConfig.GatewayInstanceId == "" {
		return errors.New("网关实例ID不能为空")
	}
	if proxyConfig.ProxyName == "" {
		return errors.New("代理名称不能为空")
	}
	if proxyConfig.ProxyType == "" {
		return errors.New("代理类型不能为空")
	}

	// 检查配置是否存在
	existing, err := dao.GetProxyConfigById(ctx, proxyConfig.ProxyConfigId, proxyConfig.TenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取现有代理配置失败")
	}
	if existing == nil {
		return errors.New("代理配置不存在")
	}

	// 保留不可修改的字段
	proxyConfig.TenantId = existing.TenantId
	proxyConfig.ProxyConfigId = existing.ProxyConfigId
	proxyConfig.AddTime = existing.AddTime
	proxyConfig.AddWho = existing.AddWho
	proxyConfig.CurrentVersion = existing.CurrentVersion + 1

	// 更新修改信息
	proxyConfig.EditTime = time.Now()
	proxyConfig.EditWho = operatorId
	proxyConfig.OprSeqFlag = proxyConfig.ProxyConfigId

	// 构建更新条件
	where := "proxyConfigId = ? AND tenantId = ? AND currentVersion = ?"
	args := []interface{}{proxyConfig.ProxyConfigId, proxyConfig.TenantId, existing.CurrentVersion}

	// 执行更新
	affectedRows, err := dao.db.Update(ctx, "HUB_GW_PROXY_CONFIG", proxyConfig, where, args, true, true)
	if err != nil {
		return huberrors.WrapError(err, "更新代理配置失败")
	}

	if affectedRows == 0 {
		return errors.New("更新失败，可能是并发修改导致版本冲突")
	}

	return nil
}

// DeleteProxyConfig 删除代理配置
func (dao *ProxyConfigDAO) DeleteProxyConfig(ctx context.Context, proxyConfigId, tenantId, operatorId string) error {
	if proxyConfigId == "" {
		return errors.New("proxyConfigId不能为空")
	}

	// 检查配置是否存在
	existing, err := dao.GetProxyConfigById(ctx, proxyConfigId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取代理配置失败")
	}
	if existing == nil {
		return errors.New("代理配置不存在")
	}

	// 执行物理删除
	sql := `DELETE FROM HUB_GW_PROXY_CONFIG WHERE proxyConfigId = ? AND tenantId = ?`

	result, err := dao.db.Exec(ctx, sql, []interface{}{proxyConfigId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除代理配置失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("删除失败，代理配置不存在")
	}

	return nil
}

// ListProxyConfigs 分页查询代理配置列表
func (dao *ProxyConfigDAO) ListProxyConfigs(ctx context.Context, tenantId, gatewayInstanceId string, page, pageSize int) ([]*models.ProxyConfig, int, error) {
	// 构建查询条件
	whereConditions := []string{"tenantId = ?"}
	args := []interface{}{tenantId}

	if !empty.IsEmpty(gatewayInstanceId) {
		whereConditions = append(whereConditions, "gatewayInstanceId = ?")
		args = append(args, gatewayInstanceId)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 构建基础查询语句
	baseQuery := fmt.Sprintf("SELECT * FROM HUB_GW_PROXY_CONFIG WHERE %s ORDER BY addTime DESC", whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, args, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询代理配置总数失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return []*models.ProxyConfig{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(args, paginationArgs...)

	// 执行分页查询
	var proxyConfigs []*models.ProxyConfig
	err = dao.db.Query(ctx, &proxyConfigs, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询代理配置列表失败")
	}

	return proxyConfigs, countResult.Count, nil
}

// GetProxyConfigByGatewayInstance 根据网关实例ID获取代理配置（返回单条数据）
func (dao *ProxyConfigDAO) GetProxyConfigByGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId string) (*models.ProxyConfig, error) {
	if gatewayInstanceId == "" {
		return nil, errors.New("gatewayInstanceId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_PROXY_CONFIG 
		WHERE gatewayInstanceId = ? AND tenantId = ?
		ORDER BY addTime DESC
		LIMIT 1
	`

	var proxyConfig models.ProxyConfig
	err := dao.db.QueryOne(ctx, &proxyConfig, query, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, huberrors.WrapError(err, "查询代理配置失败")
	}

	return &proxyConfig, nil
}
