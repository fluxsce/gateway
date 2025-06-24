package dao

import (
	"context"
	"errors"
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/utils/huberrors"
	"gohub/web/views/hub0021/models"
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
// HUB_GATEWAY_SERVICE_DEFINITION -> HUB_GATEWAY_PROXY_CONFIG -> HUB_GATEWAY_INSTANCE
func (dao *ServiceDefinitionDAO) GetServiceDefinitionsByInstance(ctx context.Context, gatewayInstanceId, tenantId string) ([]*models.ServiceDefinitionWithProxy, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return nil, errors.New("gatewayInstanceId和tenantId不能为空")
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
			
		FROM HUB_GATEWAY_SERVICE_DEFINITION sd
		INNER JOIN HUB_GATEWAY_PROXY_CONFIG pc ON sd.tenantId = pc.tenantId AND sd.proxyConfigId = pc.proxyConfigId
		INNER JOIN HUB_GATEWAY_INSTANCE gi ON pc.tenantId = gi.tenantId AND pc.gatewayInstanceId = gi.gatewayInstanceId
		WHERE gi.gatewayInstanceId = ? 
		  AND gi.tenantId = ? 
		ORDER BY sd.addTime DESC
	`

	var serviceDefinitions []*models.ServiceDefinitionWithProxy
	err := dao.db.Query(ctx, &serviceDefinitions, query, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询网关实例关联的服务定义失败")
	}

	return serviceDefinitions, nil
}

// GetServiceDefinitionById 根据ID获取服务定义
func (dao *ServiceDefinitionDAO) GetServiceDefinitionById(ctx context.Context, serviceDefinitionId, tenantId string) (*models.ServiceDefinition, error) {
	if serviceDefinitionId == "" || tenantId == "" {
		return nil, errors.New("serviceDefinitionId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GATEWAY_SERVICE_DEFINITION 
		WHERE serviceDefinitionId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var serviceDefinition models.ServiceDefinition
	err := dao.db.QueryOne(ctx, &serviceDefinition, query, []interface{}{serviceDefinitionId, tenantId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil
		}
		return nil, huberrors.WrapError(err, "获取服务定义失败")
	}

	return &serviceDefinition, nil
}

// ListServiceDefinitions 分页查询服务定义列表
func (dao *ServiceDefinitionDAO) ListServiceDefinitions(ctx context.Context, tenantId string, page, pageSize int, filters map[string]interface{}) ([]*models.ServiceDefinition, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	// 构建查询条件
	whereClause := "WHERE tenantId = ? AND activeFlag = 'Y'"
	params := []interface{}{tenantId}

	// 添加筛选条件
	for key, value := range filters {
		if value != nil && value != "" {
			// 对于字符串类型的值，支持模糊查询
			if strValue, ok := value.(string); ok && (key == "serviceName" || key == "serviceDesc") {
				whereClause += fmt.Sprintf(" AND %s LIKE ?", key)
				params = append(params, "%"+strValue+"%")
			} else {
				whereClause += fmt.Sprintf(" AND %s = ?", key)
				params = append(params, value)
			}
		}
	}

	// 计算总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM HUB_GATEWAY_SERVICE_DEFINITION %s", whereClause)
	var result struct {
		Count int `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务定义总数失败")
	}
	total := result.Count

	// 如果没有记录，直接返回空列表
	if total == 0 {
		return []*models.ServiceDefinition{}, 0, nil
	}

	// 计算分页参数
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// 查询数据
	query := fmt.Sprintf(`
		SELECT * FROM HUB_GATEWAY_SERVICE_DEFINITION 
		%s 
		ORDER BY addTime DESC 
		LIMIT ? OFFSET ?
	`, whereClause)
	params = append(params, pageSize, offset)

	var serviceDefinitions []*models.ServiceDefinition
	err = dao.db.Query(ctx, &serviceDefinitions, query, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务定义列表失败")
	}

	return serviceDefinitions, total, nil
}

// GetServiceDefinitionsByProxyConfig 根据代理配置ID获取服务定义列表
func (dao *ServiceDefinitionDAO) GetServiceDefinitionsByProxyConfig(ctx context.Context, proxyConfigId, tenantId string) ([]*models.ServiceDefinition, error) {
	if proxyConfigId == "" || tenantId == "" {
		return nil, errors.New("proxyConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_GATEWAY_SERVICE_DEFINITION 
		WHERE proxyConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC
	`

	var serviceDefinitions []*models.ServiceDefinition
	err := dao.db.Query(ctx, &serviceDefinitions, query, []interface{}{proxyConfigId, tenantId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "获取代理配置关联的服务定义失败")
	}

	return serviceDefinitions, nil
}

// CountServiceDefinitionsByInstance 统计网关实例关联的服务定义数量
func (dao *ServiceDefinitionDAO) CountServiceDefinitionsByInstance(ctx context.Context, gatewayInstanceId, tenantId string) (int, error) {
	if gatewayInstanceId == "" || tenantId == "" {
		return 0, errors.New("gatewayInstanceId和tenantId不能为空")
	}

	query := `
		SELECT COUNT(*) as count
		FROM HUB_GATEWAY_SERVICE_DEFINITION sd
		INNER JOIN HUB_GATEWAY_PROXY_CONFIG pc ON sd.tenantId = pc.tenantId AND sd.proxyConfigId = pc.proxyConfigId
		WHERE pc.gatewayInstanceId = ? 
		  AND pc.tenantId = ? 
		  AND sd.activeFlag = 'Y' 
		  AND pc.activeFlag = 'Y'
	`

	var result struct {
		Count int `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{gatewayInstanceId, tenantId}, true)
	if err != nil {
		return 0, huberrors.WrapError(err, "统计网关实例关联的服务定义数量失败")
	}

	return result.Count, nil
} 