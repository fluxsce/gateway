package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gohub/pkg/database"
	"gohub/pkg/database/sqlutils"
	"gohub/pkg/utils/huberrors"
	"gohub/pkg/utils/random"
	"gohub/web/views/hub0022/models"
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

// generateProxyConfigId 生成代理配置ID
// 格式：PC + YYYYMMDD + HHMMSS + 4位随机数
// 示例：PC20240615143022A1B2
func (dao *ProxyConfigDAO) generateProxyConfigId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("PC%s%s", timeStr, randomStr)
}

// isProxyConfigIdExists 检查代理配置ID是否已存在
func (dao *ProxyConfigDAO) isProxyConfigIdExists(ctx context.Context, proxyConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_PROXY_CONFIG WHERE proxyConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{proxyConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueProxyConfigId 生成唯一的代理配置ID
func (dao *ProxyConfigDAO) generateUniqueProxyConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		proxyConfigId := dao.generateProxyConfigId()
		
		exists, err := dao.isProxyConfigIdExists(ctx, proxyConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查代理配置ID是否存在失败")
		}
		
		if !exists {
			return proxyConfigId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一代理配置ID失败，已达到最大尝试次数")
}

// CreateProxyConfig 创建代理配置
func (dao *ProxyConfigDAO) CreateProxyConfig(ctx context.Context, proxyConfig *models.ProxyConfig, operatorId string) (string, error) {
	// 验证必填字段
	if proxyConfig.TenantId == "" {
		return "", errors.New("租户ID不能为空")
	}
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
		generatedId, err := dao.generateUniqueProxyConfigId(ctx)
		if err != nil {
			return "", huberrors.WrapError(err, "生成代理配置ID失败")
		}
		proxyConfig.ProxyConfigId = generatedId
	} else {
		// 如果提供了ID，检查是否已存在
		exists, err := dao.isProxyConfigIdExists(ctx, proxyConfig.ProxyConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查代理配置ID是否存在失败")
		}
		if exists {
			return "", errors.New("代理配置ID已存在")
		}
	}

	// 设置自动填充的字段
	now := time.Now()
	proxyConfig.AddTime = now
	proxyConfig.AddWho = operatorId
	proxyConfig.EditTime = now
	proxyConfig.EditWho = operatorId
	proxyConfig.OprSeqFlag = proxyConfig.ProxyConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
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
	if proxyConfigId == "" || tenantId == "" {
		return nil, errors.New("proxyConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_PROXY_CONFIG 
		WHERE proxyConfigId = ? AND tenantId = ?
	`

	var proxyConfig models.ProxyConfig
	err := dao.db.QueryOne(ctx, &proxyConfig, query, []interface{}{proxyConfigId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询代理配置失败")
	}

	return &proxyConfig, nil
}

// UpdateProxyConfig 更新代理配置
func (dao *ProxyConfigDAO) UpdateProxyConfig(ctx context.Context, proxyConfig *models.ProxyConfig, operatorId string) error {
	if proxyConfig.ProxyConfigId == "" || proxyConfig.TenantId == "" {
		return errors.New("proxyConfigId和tenantId不能为空")
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
	proxyConfig.OprSeqFlag = proxyConfig.ProxyConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	// 构建更新条件
	where := "proxyConfigId = ? AND tenantId = ? AND currentVersion = ?"
	args := []interface{}{proxyConfig.ProxyConfigId, proxyConfig.TenantId, existing.CurrentVersion}

	// 执行更新
	affectedRows, err := dao.db.Update(ctx, "HUB_GW_PROXY_CONFIG", proxyConfig, where, args, true)
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
	if proxyConfigId == "" || tenantId == "" {
		return errors.New("proxyConfigId和tenantId不能为空")
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
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建查询条件
	whereConditions := []string{"tenantId = ?", "activeFlag = 'Y'"}
	args := []interface{}{tenantId}

	if gatewayInstanceId != "" {
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

// GetProxyConfigsByGatewayInstance 根据网关实例ID获取代理配置
func (dao *ProxyConfigDAO) GetProxyConfigsByGatewayInstance(ctx context.Context, gatewayInstanceId, tenantId string) ([]*models.ProxyConfig, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return nil, errors.New("gatewayInstanceId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_PROXY_CONFIG 
		WHERE gatewayInstanceId = ? AND tenantId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC
	`

	var proxyConfigs []*models.ProxyConfig
	err := dao.db.Query(ctx, &proxyConfigs, query, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询代理配置失败")
	}

	return proxyConfigs, nil
} 