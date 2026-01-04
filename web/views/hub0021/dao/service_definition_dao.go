package dao

import (
	"context"
	"errors"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0021/models"
	"strings"
)

// ServiceDefinitionDAO 服务定义数据访问对象
type ServiceDefinitionDAO struct {
	db database.Database
}

// NewServiceDefinitionDAO 创建服务定义DAO
func NewServiceDefinitionDAO(db database.Database) *ServiceDefinitionDAO {
	return &ServiceDefinitionDAO{
		db: db,
	}
}

// GetServiceDefinitionsByInstance 根据网关实例ID获取服务定义列表（关联查询）
// 通过代理配置表进行关联查询：
// HUB_GW_SERVICE_DEFINITION -> HUB_GW_PROXY_CONFIG -> HUB_GW_INSTANCE
func (dao *ServiceDefinitionDAO) GetServiceDefinitionsByInstance(ctx context.Context, gatewayInstanceId, tenantId string) ([]*models.ServiceDefinitionWithProxy, error) {
	if gatewayInstanceId == "" {
		return nil, errors.New("gatewayInstanceId不能为空")
	}

	// 构建关联查询SQL
	// 服务定义表通过proxyConfigId关联代理配置表，代理配置表通过gatewayInstanceId关联网关实例
	query := `
		SELECT 
			sd.tenantId,
			sd.serviceDefinitionId,
			sd.serviceName,
			sd.serviceDesc,
			sd.serviceType,
			sd.loadBalanceStrategy,
			sd.discoveryType,
			sd.discoveryConfig,
			sd.sessionAffinity,
			sd.stickySession,
			sd.maxRetries,
			sd.retryTimeoutMs,
			sd.enableCircuitBreaker,
			sd.healthCheckEnabled,
			sd.healthCheckPath,
			sd.healthCheckMethod,
			sd.healthCheckIntervalSeconds,
			sd.healthCheckTimeoutMs,
			sd.healthyThreshold,
			sd.unhealthyThreshold,
			sd.expectedStatusCodes,
			sd.healthCheckHeaders,
			sd.loadBalancerConfig,
			sd.serviceMetadata,
			
			pc.proxyConfigId,
			pc.proxyName,
			pc.proxyType,
			pc.proxyId,
			pc.configPriority,
			pc.proxyConfig,
			pc.customConfig as proxyCustomConfig,
			
			gi.gatewayInstanceId,
			gi.instanceName,
			gi.instanceDesc,
			sd.addTime,
			sd.addWho,
			sd.editTime,
			sd.editWho,
			sd.currentVersion,
			sd.activeFlag,
			sd.noteText
			
		FROM HUB_GW_SERVICE_DEFINITION sd
		INNER JOIN HUB_GW_PROXY_CONFIG pc ON sd.tenantId = pc.tenantId AND sd.proxyConfigId = pc.proxyConfigId
		INNER JOIN HUB_GW_INSTANCE gi ON pc.tenantId = gi.tenantId AND pc.gatewayInstanceId = gi.gatewayInstanceId
		WHERE gi.gatewayInstanceId = ?`

	args := []interface{}{gatewayInstanceId}

	// 如果提供了 tenantId，添加到查询条件中
	if tenantId != "" {
		query += ` AND gi.tenantId = ?`
		args = append(args, tenantId)
	}

	query += ` ORDER BY sd.addTime DESC`

	var serviceDefinitions []*models.ServiceDefinitionWithProxy
	err := dao.db.Query(ctx, &serviceDefinitions, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询网关实例关联的服务定义失败")
	}

	return serviceDefinitions, nil
}

// GetServiceDefinitionById 根据ID获取服务定义
func (dao *ServiceDefinitionDAO) GetServiceDefinitionById(ctx context.Context, serviceDefinitionId, tenantId string, activeFlag string) (*models.ServiceDefinition, error) {
	if serviceDefinitionId == "" {
		return nil, errors.New("serviceDefinitionId不能为空")
	}

	// 构建查询条件
	whereConditions := []string{"serviceDefinitionId = ?"}
	args := []interface{}{serviceDefinitionId}

	// 如果提供了 tenantId，添加到查询条件中
	if tenantId != "" {
		whereConditions = append(whereConditions, "tenantId = ?")
		args = append(args, tenantId)
	}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT * FROM HUB_GW_SERVICE_DEFINITION 
		WHERE %s
	`, whereClause)

	var serviceDefinition models.ServiceDefinition
	err := dao.db.QueryOne(ctx, &serviceDefinition, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil
		}
		return nil, huberrors.WrapError(err, "获取服务定义失败")
	}

	return &serviceDefinition, nil
}

// ListServiceDefinitions 分页查询服务定义列表
func (dao *ServiceDefinitionDAO) ListServiceDefinitions(ctx context.Context, tenantId string, activeFlag string, page, pageSize int, filters map[string]interface{}) ([]*models.ServiceDefinition, int, error) {
	// 提取 gatewayInstanceId 和 proxyConfigId
	gatewayInstanceId := ""
	proxyConfigId := ""
	if filters != nil {
		if gi, ok := filters["gatewayInstanceId"].(string); ok && gi != "" {
			gatewayInstanceId = gi
		}
		if pi, ok := filters["proxyConfigId"].(string); ok && pi != "" {
			proxyConfigId = pi
		}
	}

	// 构建代理配置ID列表
	proxyConfigIds := make([]string, 0)

	// 如果提供了 gatewayInstanceId，先从 HUB_GW_PROXY_CONFIG 查询对应的代理配置ID
	if gatewayInstanceId != "" {
		proxyConfigQuery := `
			SELECT DISTINCT proxyConfigId 
			FROM HUB_GW_PROXY_CONFIG 
			WHERE gatewayInstanceId = ?`
		proxyConfigArgs := []interface{}{gatewayInstanceId}

		// 如果提供了 tenantId，添加到查询条件中
		if tenantId != "" {
			proxyConfigQuery += ` AND tenantId = ?`
			proxyConfigArgs = append(proxyConfigArgs, tenantId)
		}

		// 只查询激活的代理配置
		proxyConfigQuery += ` AND activeFlag = 'Y'`

		var proxyConfigs []struct {
			ProxyConfigId string `db:"proxyConfigId"`
		}
		err := dao.db.Query(ctx, &proxyConfigs, proxyConfigQuery, proxyConfigArgs, true)
		if err != nil {
			// 如果查询失败，记录错误但不中断，尝试使用原始的 gatewayInstanceId 作为 proxyConfigId（兼容全局代理配置）
			proxyConfigIds = append(proxyConfigIds, gatewayInstanceId)
		} else {
			// 收集查询到的代理配置ID
			for _, pc := range proxyConfigs {
				if pc.ProxyConfigId != "" {
					proxyConfigIds = append(proxyConfigIds, pc.ProxyConfigId)
				}
			}
			// 同时保留原始的 gatewayInstanceId，可能是全局代理配置（proxyConfigId 直接等于 gatewayInstanceId）
			proxyConfigIds = append(proxyConfigIds, gatewayInstanceId)
		}
	}

	// 如果提供了 proxyConfigId，也添加到列表中（去重）
	if proxyConfigId != "" {
		found := false
		for _, id := range proxyConfigIds {
			if id == proxyConfigId {
				found = true
				break
			}
		}
		if !found {
			proxyConfigIds = append(proxyConfigIds, proxyConfigId)
		}
	}

	// 如果没有代理配置ID，返回空结果
	if len(proxyConfigIds) == 0 {
		// 如果没有提供任何筛选条件，返回空列表（或者可以查询所有，根据业务需求）
		return []*models.ServiceDefinition{}, 0, nil
	}

	// 构建查询条件
	whereConditions := []string{}
	params := []interface{}{}

	// 如果提供了 tenantId，添加到查询条件中
	if tenantId != "" {
		whereConditions = append(whereConditions, "tenantId = ?")
		params = append(params, tenantId)
	}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		params = append(params, activeFlag)
	} else {
		// 默认只查询激活的记录
		whereConditions = append(whereConditions, "activeFlag = 'Y'")
	}

	// 使用 IN 查询代理配置ID
	placeholders := make([]string, len(proxyConfigIds))
	for i, id := range proxyConfigIds {
		placeholders[i] = "?"
		params = append(params, id)
	}
	whereConditions = append(whereConditions, fmt.Sprintf("proxyConfigId IN (%s)", strings.Join(placeholders, ",")))

	// 添加其他筛选条件
	for key, value := range filters {
		if value != nil && value != "" {
			// 排除 gatewayInstanceId 和 proxyConfigId，因为已经通过 IN 查询处理了
			if key == "gatewayInstanceId" || key == "proxyConfigId" {
				continue
			}
			// 对于字符串类型的值，支持模糊查询
			if strValue, ok := value.(string); ok && (key == "serviceName" || key == "serviceDesc") {
				whereConditions = append(whereConditions, fmt.Sprintf("%s LIKE ?", key))
				params = append(params, "%"+strValue+"%")
			} else {
				whereConditions = append(whereConditions, fmt.Sprintf("%s = ?", key))
				params = append(params, value)
			}
		}
	}

	// 构建 WHERE 子句（必须至少有一个条件）
	whereClause := "WHERE " + strings.Join(whereConditions, " AND ")

	// 构建基础查询语句
	baseQuery := fmt.Sprintf("SELECT * FROM HUB_GW_SERVICE_DEFINITION %s ORDER BY addTime DESC", whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务定义总数失败")
	}

	// 如果没有记录，直接返回空列表
	if result.Count == 0 {
		return []*models.ServiceDefinition{}, 0, nil
	}

	// 创建分页信息
	paginationInfo := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, paginationInfo)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allParams := append(params, paginationArgs...)

	// 执行分页查询
	var serviceDefinitions []*models.ServiceDefinition
	err = dao.db.Query(ctx, &serviceDefinitions, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务定义列表失败")
	}

	return serviceDefinitions, result.Count, nil
}

// ListAllServiceDefinitions 查询所有服务定义（不依赖代理配置，用于日志查询等场景）
// 支持分页和筛选，但不强制要求 proxyConfigId
func (dao *ServiceDefinitionDAO) ListAllServiceDefinitions(ctx context.Context, tenantId string, activeFlag string, page, pageSize int, filters map[string]interface{}) ([]*models.ServiceDefinition, int, error) {
	// 构建查询条件
	whereConditions := []string{}
	params := []interface{}{}

	// 如果提供了 tenantId，添加到查询条件中
	if tenantId != "" {
		whereConditions = append(whereConditions, "tenantId = ?")
		params = append(params, tenantId)
	}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		params = append(params, activeFlag)
	} else {
		// 默认只查询激活的记录
		whereConditions = append(whereConditions, "activeFlag = 'Y'")
	}

	// 添加其他筛选条件（排除 gatewayInstanceId 和 proxyConfigId，允许查询所有服务）
	for key, value := range filters {
		if value != nil && value != "" {
			// 排除 gatewayInstanceId 和 proxyConfigId，允许查询所有服务
			if key == "gatewayInstanceId" || key == "proxyConfigId" {
				continue
			}
			// 对于字符串类型的值，支持模糊查询
			if strValue, ok := value.(string); ok && (key == "serviceName" || key == "serviceDesc") {
				whereConditions = append(whereConditions, fmt.Sprintf("%s LIKE ?", key))
				params = append(params, "%"+strValue+"%")
			} else {
				whereConditions = append(whereConditions, fmt.Sprintf("%s = ?", key))
				params = append(params, value)
			}
		}
	}

	// 构建 WHERE 子句
	var whereClause string
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	} else {
		whereClause = ""
	}

	// 构建基础查询语句
	baseQuery := fmt.Sprintf("SELECT * FROM HUB_GW_SERVICE_DEFINITION %s ORDER BY addTime DESC", whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务定义总数失败")
	}

	// 如果没有记录，直接返回空列表
	if result.Count == 0 {
		return []*models.ServiceDefinition{}, 0, nil
	}

	// 创建分页信息
	paginationInfo := sqlutils.NewPaginationInfo(page, pageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, paginationInfo)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allParams := append(params, paginationArgs...)

	// 执行分页查询
	var serviceDefinitions []*models.ServiceDefinition
	err = dao.db.Query(ctx, &serviceDefinitions, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务定义列表失败")
	}

	return serviceDefinitions, result.Count, nil
}

// GetServiceDefinitionsByProxyConfig 根据代理配置ID获取服务定义列表
func (dao *ServiceDefinitionDAO) GetServiceDefinitionsByProxyConfig(ctx context.Context, proxyConfigId, tenantId string, activeFlag string) ([]*models.ServiceDefinition, error) {
	if proxyConfigId == "" {
		return nil, errors.New("proxyConfigId不能为空")
	}

	// 构建查询条件
	whereConditions := []string{"proxyConfigId = ?"}
	args := []interface{}{proxyConfigId}

	// 如果提供了 tenantId，添加到查询条件中
	if tenantId != "" {
		whereConditions = append(whereConditions, "tenantId = ?")
		args = append(args, tenantId)
	}

	// 添加activeFlag条件（如果指定了activeFlag参数）
	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT * FROM HUB_GW_SERVICE_DEFINITION 
		WHERE %s
		ORDER BY addTime DESC
	`, whereClause)

	var serviceDefinitions []*models.ServiceDefinition
	err := dao.db.Query(ctx, &serviceDefinitions, query, args, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取代理配置关联的服务定义失败")
	}

	return serviceDefinitions, nil
}

// CountServiceDefinitionsByInstance 统计网关实例关联的服务定义数量
func (dao *ServiceDefinitionDAO) CountServiceDefinitionsByInstance(ctx context.Context, gatewayInstanceId, tenantId string, serviceActiveFlag, proxyActiveFlag string) (int, error) {
	if gatewayInstanceId == "" {
		return 0, errors.New("gatewayInstanceId不能为空")
	}

	// 构建查询条件
	whereConditions := []string{"pc.gatewayInstanceId = ?"}
	args := []interface{}{gatewayInstanceId}

	// 如果提供了 tenantId，添加到查询条件中
	if tenantId != "" {
		whereConditions = append(whereConditions, "pc.tenantId = ?")
		args = append(args, tenantId)
	}

	// 添加服务定义activeFlag条件（如果指定了serviceActiveFlag参数）
	if serviceActiveFlag != "" {
		whereConditions = append(whereConditions, "sd.activeFlag = ?")
		args = append(args, serviceActiveFlag)
	}

	// 添加代理配置activeFlag条件（如果指定了proxyActiveFlag参数）
	if proxyActiveFlag != "" {
		whereConditions = append(whereConditions, "pc.activeFlag = ?")
		args = append(args, proxyActiveFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT COUNT(*) as count
		FROM HUB_GW_SERVICE_DEFINITION sd
		INNER JOIN HUB_GW_PROXY_CONFIG pc ON sd.tenantId = pc.tenantId AND sd.proxyConfigId = pc.proxyConfigId
		WHERE %s
	`, whereClause)

	var result struct {
		Count int `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &result, query, args, true)
	if err != nil {
		return 0, huberrors.WrapError(err, "统计网关实例关联的服务定义数量失败")
	}

	return result.Count, nil
}
