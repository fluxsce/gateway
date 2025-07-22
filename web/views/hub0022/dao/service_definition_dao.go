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

// ServiceDefinitionQueryFilter 服务定义查询过滤条件
type ServiceDefinitionQueryFilter struct {
	ServiceName          string `json:"serviceName,omitempty" form:"serviceName" query:"serviceName"`          // 服务名称（模糊查询）
	ServiceType          string `json:"serviceType,omitempty" form:"serviceType" query:"serviceType"`          // 服务类型（精确匹配）
	LoadBalanceStrategy  string `json:"loadBalanceStrategy,omitempty" form:"loadBalanceStrategy" query:"loadBalanceStrategy"`  // 负载均衡策略（精确匹配）
	ActiveFlag           string `json:"activeFlag,omitempty" form:"activeFlag" query:"activeFlag"`           // 激活状态（精确匹配）
	ProxyConfigId        string `json:"proxyConfigId,omitempty" form:"proxyConfigId" query:"proxyConfigId"`        // 代理配置ID（精确匹配）
}

// ServiceDefinitionDAO 服务定义数据访问对象
type ServiceDefinitionDAO struct {
	db database.Database
}

// NewServiceDefinitionDAO 创建服务定义DAO实例
func NewServiceDefinitionDAO(db database.Database) *ServiceDefinitionDAO {
	return &ServiceDefinitionDAO{
		db: db,
	}
}

// generateServiceDefinitionId 生成服务定义ID
// 格式：SD + YYYYMMDD + HHMMSS + 4位随机数
// 示例：SD20240615143022A1B2
func (dao *ServiceDefinitionDAO) generateServiceDefinitionId() string {
	now := time.Now()
	// 生成时间部分：YYYYMMDDHHMMSS
	timeStr := now.Format("20060102150405")
	
	// 生成4位随机字符（大写字母和数字）
	randomStr := random.GenerateRandomString(4)
	
	return fmt.Sprintf("SD%s%s", timeStr, randomStr)
}

// isServiceDefinitionIdExists 检查服务定义ID是否已存在
func (dao *ServiceDefinitionDAO) isServiceDefinitionIdExists(ctx context.Context, serviceDefinitionId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_GW_SERVICE_DEFINITION WHERE serviceDefinitionId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{serviceDefinitionId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueServiceDefinitionId 生成唯一的服务定义ID
func (dao *ServiceDefinitionDAO) generateUniqueServiceDefinitionId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		serviceDefinitionId := dao.generateServiceDefinitionId()
		
		exists, err := dao.isServiceDefinitionIdExists(ctx, serviceDefinitionId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查服务定义ID是否存在失败")
		}
		
		if !exists {
			return serviceDefinitionId, nil
		}
		
		// 如果ID已存在，等待1毫秒后重试（确保时间戳不同）
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一服务定义ID失败，已达到最大尝试次数")
}

// CreateServiceDefinition 创建服务定义
func (dao *ServiceDefinitionDAO) CreateServiceDefinition(ctx context.Context, serviceDefinition *models.ServiceDefinition, operatorId string) (string, error) {
	// 验证必填字段
	if serviceDefinition.TenantId == "" {
		return "", errors.New("租户ID不能为空")
	}
	if serviceDefinition.ServiceName == "" {
		return "", errors.New("服务名称不能为空")
	}
	if serviceDefinition.ServiceType == 0 {
		serviceDefinition.ServiceType = 0 // 默认静态配置
	}

	// 自动生成服务定义ID（如果为空）
	if serviceDefinition.ServiceDefinitionId == "" {
		generatedId, err := dao.generateUniqueServiceDefinitionId(ctx)
		if err != nil {
			return "", huberrors.WrapError(err, "生成服务定义ID失败")
		}
		serviceDefinition.ServiceDefinitionId = generatedId
	} else {
		// 如果提供了ID，检查是否已存在
		exists, err := dao.isServiceDefinitionIdExists(ctx, serviceDefinition.ServiceDefinitionId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查服务定义ID是否存在失败")
		}
		if exists {
			return "", errors.New("服务定义ID已存在")
		}
	}

	// 设置自动填充的字段
	now := time.Now()
	serviceDefinition.AddTime = now
	serviceDefinition.AddWho = operatorId
	serviceDefinition.EditTime = now
	serviceDefinition.EditWho = operatorId
	serviceDefinition.OprSeqFlag = serviceDefinition.ServiceDefinitionId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	serviceDefinition.CurrentVersion = 1
	serviceDefinition.ActiveFlag = "Y"

	// 设置默认值
	if serviceDefinition.LoadBalanceStrategy == "" {
		serviceDefinition.LoadBalanceStrategy = "ROUND_ROBIN"
	}
	if serviceDefinition.HealthCheckEnabled == "" {
		serviceDefinition.HealthCheckEnabled = "Y"
	}

	// 插入记录
	_, err := dao.db.Insert(ctx, "HUB_GW_SERVICE_DEFINITION", serviceDefinition, true)
	if err != nil {
		return "", huberrors.WrapError(err, "创建服务定义失败")
	}

	return serviceDefinition.ServiceDefinitionId, nil
}

// GetServiceDefinitionById 根据ID获取服务定义
func (dao *ServiceDefinitionDAO) GetServiceDefinitionById(ctx context.Context, serviceDefinitionId, tenantId string) (*models.ServiceDefinition, error) {
	if serviceDefinitionId == "" || tenantId == "" {
		return nil, errors.New("serviceDefinitionId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GW_SERVICE_DEFINITION 
		WHERE serviceDefinitionId = ? AND tenantId = ?
	`

	var serviceDefinition models.ServiceDefinition
	err := dao.db.QueryOne(ctx, &serviceDefinition, query, []interface{}{serviceDefinitionId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询服务定义失败")
	}

	return &serviceDefinition, nil
}

// UpdateServiceDefinition 更新服务定义
func (dao *ServiceDefinitionDAO) UpdateServiceDefinition(ctx context.Context, serviceDefinition *models.ServiceDefinition, operatorId string) error {
	if serviceDefinition.ServiceDefinitionId == "" || serviceDefinition.TenantId == "" {
		return errors.New("serviceDefinitionId和tenantId不能为空")
	}

	// 验证必填字段
	if serviceDefinition.ServiceName == "" {
		return errors.New("服务名称不能为空")
	}
	// ServiceType为int类型，0为默认值

	// 检查服务定义是否存在
	existing, err := dao.GetServiceDefinitionById(ctx, serviceDefinition.ServiceDefinitionId, serviceDefinition.TenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取现有服务定义失败")
	}
	if existing == nil {
		return errors.New("服务定义不存在")
	}

	// 保留不可修改的字段
	serviceDefinition.TenantId = existing.TenantId
	serviceDefinition.ServiceDefinitionId = existing.ServiceDefinitionId
	serviceDefinition.AddTime = existing.AddTime
	serviceDefinition.AddWho = existing.AddWho
	serviceDefinition.CurrentVersion = existing.CurrentVersion + 1

	// 更新修改信息
	serviceDefinition.EditTime = time.Now()
	serviceDefinition.EditWho = operatorId
	serviceDefinition.OprSeqFlag = serviceDefinition.ServiceDefinitionId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	// 构建更新条件
	where := "serviceDefinitionId = ? AND tenantId = ? AND currentVersion = ?"
	args := []interface{}{serviceDefinition.ServiceDefinitionId, serviceDefinition.TenantId, existing.CurrentVersion}

	// 执行更新
	affectedRows, err := dao.db.Update(ctx, "HUB_GW_SERVICE_DEFINITION", serviceDefinition, where, args, true)
	if err != nil {
		return huberrors.WrapError(err, "更新服务定义失败")
	}

	if affectedRows == 0 {
		return errors.New("更新失败，可能是并发修改导致版本冲突")
	}

	return nil
}

// DeleteServiceDefinition 删除服务定义
func (dao *ServiceDefinitionDAO) DeleteServiceDefinition(ctx context.Context, serviceDefinitionId, tenantId, operatorId string) error {
	if serviceDefinitionId == "" || tenantId == "" {
		return errors.New("serviceDefinitionId和tenantId不能为空")
	}

	// 检查服务定义是否存在
	existing, err := dao.GetServiceDefinitionById(ctx, serviceDefinitionId, tenantId)
	if err != nil {
		return huberrors.WrapError(err, "获取服务定义失败")
	}
	if existing == nil {
		return errors.New("服务定义不存在")
	}

	// 执行物理删除
	sql := `DELETE FROM HUB_GW_SERVICE_DEFINITION WHERE serviceDefinitionId = ? AND tenantId = ?`
	
	result, err := dao.db.Exec(ctx, sql, []interface{}{serviceDefinitionId, tenantId}, true)
	if err != nil {
		return huberrors.WrapError(err, "删除服务定义失败")
	}

	// 检查是否有记录被删除
	if result == 0 {
		return errors.New("删除失败，服务定义不存在")
	}

	return nil
}

// ListServiceDefinitions 分页查询服务定义列表
func (dao *ServiceDefinitionDAO) ListServiceDefinitions(ctx context.Context, tenantId string, page, pageSize int, filter *ServiceDefinitionQueryFilter) ([]*models.ServiceDefinition, int64, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 动态构建WHERE条件
	var whereConditions []string
	var queryArgs []interface{}

	// 基础条件：租户ID
	whereConditions = append(whereConditions, "tenantId = ?")
	queryArgs = append(queryArgs, tenantId)

	// 如果没有传入filter，创建一个空的
	if filter == nil {
		filter = &ServiceDefinitionQueryFilter{}
	}

	// 激活状态过滤
	if filter.ActiveFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		queryArgs = append(queryArgs, filter.ActiveFlag)
	} else {
		// 默认只查询激活的记录
		whereConditions = append(whereConditions, "activeFlag = 'Y'")
	}

	// 服务名称（模糊查询）
	if filter.ServiceName != "" {
		whereConditions = append(whereConditions, "serviceName LIKE ?")
		queryArgs = append(queryArgs, "%"+filter.ServiceName+"%")
	}

	// 服务类型（精确匹配）
	if filter.ServiceType != "" {
		whereConditions = append(whereConditions, "serviceType = ?")
		queryArgs = append(queryArgs, filter.ServiceType)
	}

	// 负载均衡策略（精确匹配）
	if filter.LoadBalanceStrategy != "" {
		whereConditions = append(whereConditions, "loadBalanceStrategy = ?")
		queryArgs = append(queryArgs, filter.LoadBalanceStrategy)
	}

	// 代理配置ID（精确匹配）
	// 注意：根据实际表结构调整字段名
	if filter.ProxyConfigId != "" {
		// TODO: 根据实际表结构调整字段名
		whereConditions = append(whereConditions, "proxyConfigId = ?")
		queryArgs = append(queryArgs, filter.ProxyConfigId)
	}

	// 构建完整的WHERE子句
	whereClause := strings.Join(whereConditions, " AND ")

	// 构建基础查询语句
	baseQuery := fmt.Sprintf("SELECT * FROM HUB_GW_SERVICE_DEFINITION WHERE %s ORDER BY addTime DESC", whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, queryArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务定义总数失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return []*models.ServiceDefinition{}, 0, nil
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

	// 合并查询参数（先是WHERE条件参数，再是分页参数）
	allArgs := append(queryArgs, paginationArgs...)

	// 执行分页查询
	var serviceDefinitions []*models.ServiceDefinition
	err = dao.db.Query(ctx, &serviceDefinitions, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务定义列表失败")
	}

	return serviceDefinitions, int64(countResult.Count), nil
} 