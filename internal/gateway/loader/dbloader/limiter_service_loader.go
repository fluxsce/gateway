package dbloader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gateway/internal/gateway/handler/limiter"
	"gateway/internal/gateway/handler/proxy"
	"gateway/internal/gateway/handler/service"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
)

// LimiterServiceLoader 限流和服务配置加载器
type LimiterServiceLoader struct {
	db       database.Database
	tenantId string
}

// NewLimiterServiceLoader 创建限流和服务配置加载器
func NewLimiterServiceLoader(db database.Database, tenantId string) *LimiterServiceLoader {
	return &LimiterServiceLoader{
		db:       db,
		tenantId: tenantId,
	}
}

// LoadRateLimitConfig 加载实例级别限流配置
func (loader *LimiterServiceLoader) LoadRateLimitConfig(ctx context.Context, instanceId string) (*limiter.RateLimitConfig, error) {
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, rateLimitConfigId, limitName, algorithm, keyStrategy, 
		       limitRate, burstCapacity, timeWindowSeconds, rejectionStatusCode, 
		       rejectionMessage, customConfig
		FROM HUB_GW_RATE_LIMIT_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
	`

	// 创建分页信息（只取第一条记录）
	pagination := sqlutils.NewPaginationInfo(1, 1)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(loader.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("构建分页查询失败: %w", err)
	}

	// 合并查询参数
	allArgs := append([]interface{}{loader.tenantId, instanceId}, paginationArgs...)

	// 执行查询
	var records []RateLimitConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询限流配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

	// 构建限流配置
	rateLimitConf := &limiter.RateLimitConfig{
		ID:      record.RateLimitConfigId,
		Name:    record.LimitName,
		Enabled: true,
	}

	// 解析限流算法类型
	// 支持的算法：
	// - TOKEN_BUCKET: 令牌桶算法（平滑突发流量）
	// - LEAKY_BUCKET: 漏桶算法（固定速率处理）
	// - SLIDING_WINDOW: 滑动窗口算法（更精确的时间窗口）
	// - FIXED_WINDOW: 固定窗口算法（简单高效）
	// - NONE: 无限制（不进行限流）
	switch record.Algorithm {
	case "TOKEN_BUCKET":
		rateLimitConf.Algorithm = limiter.AlgorithmTokenBucket
	case "LEAKY_BUCKET":
		rateLimitConf.Algorithm = limiter.AlgorithmLeakyBucket
	case "SLIDING_WINDOW":
		rateLimitConf.Algorithm = limiter.AlgorithmSlidingWindow
	case "FIXED_WINDOW":
		rateLimitConf.Algorithm = limiter.AlgorithmFixedWindow
	case "NONE":
		rateLimitConf.Algorithm = limiter.AlgorithmNone
	default:
		// 默认使用令牌桶算法
		rateLimitConf.Algorithm = limiter.AlgorithmTokenBucket
	}

	// 设置限流参数
	rateLimitConf.Rate = record.LimitRate
	rateLimitConf.Burst = record.BurstCapacity
	rateLimitConf.WindowSize = record.TimeWindowSeconds
	rateLimitConf.KeyStrategy = record.KeyStrategy
	rateLimitConf.ErrorStatusCode = record.RejectionStatusCode
	rateLimitConf.ErrorMessage = record.RejectionMessage

	// 解析自定义配置
	if record.CustomConfig != "" {
		var customConfig map[string]interface{}
		if err := json.Unmarshal([]byte(record.CustomConfig), &customConfig); err == nil {
			rateLimitConf.CustomConfig = customConfig
		}
	}

	return rateLimitConf, nil
}

// LoadRouteRateLimitConfig 加载路由级别限流配置
func (loader *LimiterServiceLoader) LoadRouteRateLimitConfig(ctx context.Context, routeId string) (*limiter.RateLimitConfig, error) {
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, rateLimitConfigId, limitName, algorithm, keyStrategy, 
		       limitRate, burstCapacity, timeWindowSeconds, rejectionStatusCode, 
		       rejectionMessage, customConfig
		FROM HUB_GW_RATE_LIMIT_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
	`

	// 创建分页信息（只取第一条记录）
	pagination := sqlutils.NewPaginationInfo(1, 1)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(loader.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("构建分页查询失败: %w", err)
	}

	// 合并查询参数
	allArgs := append([]interface{}{loader.tenantId, routeId}, paginationArgs...)

	// 执行查询
	var records []RateLimitConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由限流配置失败: %w", err)
	}

	// 如果没有找到记录，返回database.ErrRecordNotFound（保持原有逻辑）
	if len(records) == 0 {
		return nil, database.ErrRecordNotFound
	}

	record := records[0]

	// 构建限流配置
	rateLimitConf := &limiter.RateLimitConfig{
		ID:      record.RateLimitConfigId,
		Name:    record.LimitName,
		Enabled: true,
	}

	// 解析限流算法类型（小写格式，用于路由级配置）
	// 支持的算法：
	// - token-bucket: 令牌桶算法（平滑突发流量）
	// - leaky-bucket: 漏桶算法（固定速率处理）
	// - sliding-window: 滑动窗口算法（更精确的时间窗口）
	// - fixed-window: 固定窗口算法（简单高效）
	// - none: 无限制（不进行限流）
	switch record.Algorithm {
	case "token-bucket":
		rateLimitConf.Algorithm = limiter.AlgorithmTokenBucket
	case "leaky-bucket":
		rateLimitConf.Algorithm = limiter.AlgorithmLeakyBucket
	case "sliding-window":
		rateLimitConf.Algorithm = limiter.AlgorithmSlidingWindow
	case "fixed-window":
		rateLimitConf.Algorithm = limiter.AlgorithmFixedWindow
	case "none":
		rateLimitConf.Algorithm = limiter.AlgorithmNone
	default:
		// 默认使用令牌桶算法
		rateLimitConf.Algorithm = limiter.AlgorithmTokenBucket
	}

	// 设置限流参数
	rateLimitConf.Rate = record.LimitRate
	rateLimitConf.Burst = record.BurstCapacity
	rateLimitConf.WindowSize = record.TimeWindowSeconds
	rateLimitConf.KeyStrategy = record.KeyStrategy
	rateLimitConf.ErrorStatusCode = record.RejectionStatusCode
	rateLimitConf.ErrorMessage = record.RejectionMessage

	// 解析自定义配置
	if record.CustomConfig != "" {
		var customConfig map[string]interface{}
		if err := json.Unmarshal([]byte(record.CustomConfig), &customConfig); err == nil {
			rateLimitConf.CustomConfig = customConfig
		}
	}

	return rateLimitConf, nil
}

// LoadProxyConfig 加载代理配置
func (loader *LimiterServiceLoader) LoadProxyConfig(ctx context.Context, instanceId string) (*proxy.ProxyConfig, error) {
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, proxyConfigId, proxyName, proxyType, proxyConfig, customConfig
		FROM HUB_GW_PROXY_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
	`

	// 创建分页信息（只取第一条记录）
	pagination := sqlutils.NewPaginationInfo(1, 1)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(loader.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("构建分页查询失败: %w", err)
	}

	// 合并查询参数
	allArgs := append([]interface{}{loader.tenantId, instanceId}, paginationArgs...)

	// 执行查询
	var records []ProxyConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询代理配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

	// 构建代理配置
	proxyConf := &proxy.ProxyConfig{
		ID:      record.ProxyConfigId,
		Name:    record.ProxyName,
		Enabled: true,
	}

	// 解析代理类型
	switch record.ProxyType {
	case "HTTP":
		proxyConf.Type = proxy.ProxyTypeHTTP
	case "TCP":
		proxyConf.Type = proxy.ProxyTypeTCP
	case "UDP":
		proxyConf.Type = proxy.ProxyTypeUDP
	case "WEBSOCKET":
		proxyConf.Type = proxy.ProxyTypeWebSocket
	default:
		proxyConf.Type = proxy.ProxyTypeHTTP
	}

	// 解析代理配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(record.ProxyConfig), &config); err == nil {
		proxyConf.Config = config
	}

	// 解析自定义配置
	if record.CustomConfig != nil {
		var customConfig map[string]interface{}
		if err := json.Unmarshal([]byte(*record.CustomConfig), &customConfig); err == nil {
			proxyConf.CustomConfig = customConfig
		}
	}

	// 查询关联的服务配置 - 从服务定义表中查找关联到这个代理的服务
	serviceQuery := `
		SELECT serviceDefinitionId 
		FROM HUB_GW_SERVICE_DEFINITION 
		WHERE tenantId = ? AND proxyConfigId = ? AND activeFlag = 'Y'
	`
	var serviceIds []struct {
		ServiceDefinitionId string `db:"serviceDefinitionId"`
	}
	err = loader.db.Query(ctx, &serviceIds, serviceQuery, []interface{}{loader.tenantId, record.ProxyConfigId}, true)
	if err != nil && err != database.ErrRecordNotFound {
		return nil, fmt.Errorf("查询代理关联的服务失败: %w", err)
	}

	if len(serviceIds) > 0 {
		proxyConf.Service = make([]*service.ServiceConfig, 0, len(serviceIds))
		for _, record := range serviceIds {
			serviceConfig, err := loader.LoadServiceConfig(ctx, record.ServiceDefinitionId)
			if err != nil {
				return nil, fmt.Errorf("加载服务配置失败: %w", err)
			}
			if serviceConfig != nil {
				proxyConf.Service = append(proxyConf.Service, serviceConfig)
			}
		}
	}

	return proxyConf, nil
}

// LoadServiceConfig 加载服务配置
func (loader *LimiterServiceLoader) LoadServiceConfig(ctx context.Context, serviceId string) (*service.ServiceConfig, error) {
	query := `
		SELECT tenantId, serviceDefinitionId, serviceName, serviceDesc, serviceType,
		       loadBalanceStrategy, discoveryType, discoveryConfig, sessionAffinity,
		       stickySession, maxRetries, retryTimeoutMs, enableCircuitBreaker,
		       healthCheckEnabled, healthCheckPath, healthCheckMethod,
		       healthCheckIntervalSeconds, healthCheckTimeoutMs, healthyThreshold,
		       unhealthyThreshold, expectedStatusCodes, healthCheckHeaders,
		       loadBalancerConfig, serviceMetadata, activeFlag
		FROM HUB_GW_SERVICE_DEFINITION 
		WHERE tenantId = ? AND serviceDefinitionId = ? AND activeFlag = 'Y'
	`

	var record ServiceConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, serviceId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询服务配置失败: %w", err)
	}

	// 构建服务配置
	serviceConf := &service.ServiceConfig{
		ID:   record.ServiceDefinitionId,
		Name: record.ServiceName,
	}

	// 解析负载均衡策略
	switch record.LoadBalanceStrategy {
	case "ROUND_ROBIN":
		serviceConf.Strategy = service.RoundRobin
	case "RANDOM":
		serviceConf.Strategy = service.Random
	case "WEIGHTED_ROUND_ROBIN":
		serviceConf.Strategy = service.WeightedRoundRobin
	case "LEAST_CONNECTION":
		serviceConf.Strategy = service.LeastConn
	case "IP_HASH":
		serviceConf.Strategy = service.IPHash
	case "CONSISTENT_HASH":
		serviceConf.Strategy = service.ConsistentHash
	default:
		serviceConf.Strategy = service.RoundRobin
	}

	// 设置负载均衡器配置
	lbConfig := &service.LoadBalancerConfig{
		ID:              fmt.Sprintf("lb-%s", record.ServiceDefinitionId),
		Strategy:        serviceConf.Strategy,
		SessionAffinity: record.SessionAffinity == "Y",
		StickySession:   record.StickySession == "Y",
		MaxRetries:      record.MaxRetries,
		RetryTimeout:    time.Duration(record.RetryTimeoutMs) * time.Millisecond,
		CircuitBreaker:  record.EnableCircuitBreaker == "Y",
	}

	// 设置健康检查配置
	if record.HealthCheckEnabled == "Y" {
		healthCheck := &service.HealthConfig{
			ID:                  fmt.Sprintf("hc-%s", record.ServiceDefinitionId),
			Enabled:             true,
			Path:                record.HealthCheckPath,
			Method:              record.HealthCheckMethod,
			ExpectedStatusCodes: []int{200},
		}

		if record.HealthCheckIntervalSeconds != nil {
			healthCheck.Interval = time.Duration(*record.HealthCheckIntervalSeconds) * time.Second
		}
		if record.HealthCheckTimeoutMs != nil {
			healthCheck.Timeout = time.Duration(*record.HealthCheckTimeoutMs) * time.Millisecond
		}
		if record.HealthyThreshold != nil {
			healthCheck.HealthyThreshold = int(*record.HealthyThreshold)
		}
		if record.UnhealthyThreshold != nil {
			healthCheck.UnhealthyThreshold = int(*record.UnhealthyThreshold)
		}

		// 解析期望的状态码
		if record.ExpectedStatusCodes != "" {
			healthCheck.ExpectedStatusCodes = parseStatusCodes(record.ExpectedStatusCodes)
		}

		// 解析健康检查头
		if record.HealthCheckHeaders != nil {
			var headers map[string]string
			if err := json.Unmarshal([]byte(*record.HealthCheckHeaders), &headers); err == nil {
				healthCheck.Headers = headers
			}
		}

		lbConfig.HealthCheck = healthCheck
	}

	serviceConf.LoadBalancer = lbConfig

	// 解析服务元数据
	if record.ServiceMetadata != nil && *record.ServiceMetadata != "" {
		var serviceMetadata map[string]string
		if err := json.Unmarshal([]byte(*record.ServiceMetadata), &serviceMetadata); err == nil {
			// 直接解析成功，使用解析的数据
			serviceConf.ServiceMetadata = serviceMetadata
		} else {
			// 如果直接解析失败（可能包含嵌套数据），则先解析为interface{}再提取扁平字段
			var rawMetadata map[string]interface{}
			if err := json.Unmarshal([]byte(*record.ServiceMetadata), &rawMetadata); err == nil {
				serviceMetadata = make(map[string]string)
				// 只提取字符串类型的扁平字段
				for key, value := range rawMetadata {
					if strValue, ok := value.(string); ok {
						serviceMetadata[key] = strValue
					}
				}
				serviceConf.ServiceMetadata = serviceMetadata
			}
		}

		// 添加服务发现相关的元数据（从数据库其他字段获取）
		if serviceConf.ServiceMetadata != nil {
			if record.DiscoveryType != nil && *record.DiscoveryType != "" {
				serviceConf.ServiceMetadata["discoveryType"] = *record.DiscoveryType
				serviceConf.ServiceMetadata["discovery_type"] = *record.DiscoveryType // 兼容下划线格式
			}

			if record.DiscoveryConfig != nil && *record.DiscoveryConfig != "" {
				serviceConf.ServiceMetadata["discoveryConfig"] = *record.DiscoveryConfig
				serviceConf.ServiceMetadata["discovery_config"] = *record.DiscoveryConfig // 兼容下划线格式
			}
			//租户id默认传入避免后续使用没有对应字段
			serviceConf.ServiceMetadata["tenantId"] = record.TenantId
		}
	}

	// 加载服务节点
	nodes, err := loader.LoadServiceNodes(ctx, serviceId)
	if err != nil {
		return nil, fmt.Errorf("加载服务节点失败: %w", err)
	}
	serviceConf.Nodes = nodes

	return serviceConf, nil
}

// LoadServiceNodes 加载服务节点
func (loader *LimiterServiceLoader) LoadServiceNodes(ctx context.Context, serviceId string) ([]*service.NodeConfig, error) {
	query := `
		SELECT tenantId, serviceNodeId, serviceDefinitionId, nodeId, nodeUrl,
		       nodeHost, nodePort, nodeProtocol, nodeWeight, healthStatus,
		       nodeMetadata, nodeStatus, activeFlag
		FROM HUB_GW_SERVICE_NODE 
		WHERE tenantId = ? AND serviceDefinitionId = ? AND activeFlag = 'Y'
		ORDER BY nodeWeight DESC
	`

	var records []ServiceNodeRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, serviceId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询服务节点失败: %w", err)
	}

	var nodes []*service.NodeConfig
	for _, record := range records {
		node := &service.NodeConfig{
			ID:      record.ServiceNodeId,
			URL:     record.NodeUrl,
			Weight:  record.NodeWeight,
			Health:  record.HealthStatus == "Y",
			Enabled: record.ActiveFlag == "Y",
		}

		// 解析节点元数据
		if record.NodeMetadata != nil {
			var metadata map[string]string
			if err := json.Unmarshal([]byte(*record.NodeMetadata), &metadata); err == nil {
				node.Metadata = metadata
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// parseStatusCodes 解析逗号分隔的状态码字符串或JSON数组
// 功能特性：
// - 优先尝试解析JSON数组格式（如 ["200","201","202"]）
// - 如果JSON解析失败，则按逗号分割字符串
// - 将字符串转换为整数
// - 过滤掉无效的状态码
// - 返回清理后的整数切片
//
// 参数:
//
//	str: 要解析的字符串（JSON数组或逗号分隔字符串）
//
// 返回:
//
//	[]int: 解析后的状态码切片
//
// 示例:
//
//	parseStatusCodes(`["200","201","202"]`)
//	// 返回: [200, 201, 202]
//	parseStatusCodes("200, 201 , 202,, 404 ")
//	// 返回: [200, 201, 202, 404]
func parseStatusCodes(str string) []int {
	if str == "" {
		return []int{}
	}

	// 去除前后空白字符
	str = strings.TrimSpace(str)

	// 优先尝试解析JSON数组
	if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
		var jsonArray []string
		if err := json.Unmarshal([]byte(str), &jsonArray); err == nil {
			// JSON解析成功，转换为整数
			var result []int
			for _, item := range jsonArray {
				trimmed := strings.TrimSpace(item)
				if trimmed != "" {
					var statusCode int
					if _, err := fmt.Sscanf(trimmed, "%d", &statusCode); err == nil {
						result = append(result, statusCode)
					}
				}
			}
			return result
		}
	}

	// JSON解析失败，按逗号分割
	parts := strings.Split(str, ",")

	// 清理和转换
	var result []int
	for _, part := range parts {
		// 去除前后空白字符
		trimmed := strings.TrimSpace(part)
		// 转换为整数
		if trimmed != "" {
			var statusCode int
			if _, err := fmt.Sscanf(trimmed, "%d", &statusCode); err == nil {
				result = append(result, statusCode)
			}
		}
	}

	return result
}
