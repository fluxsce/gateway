package dbloader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gohub/internal/gateway/handler/assertion"
	"gohub/internal/gateway/handler/filter"
	"gohub/internal/gateway/handler/router"
	"gohub/pkg/database"
	"gohub/pkg/logger"
)

// RouterConfigLoader 路由配置加载器
type RouterConfigLoader struct {
	db       database.Database
	tenantId string
}

// NewRouterConfigLoader 创建路由配置加载器
func NewRouterConfigLoader(db database.Database, tenantId string) *RouterConfigLoader {
	return &RouterConfigLoader{
		db:       db,
		tenantId: tenantId,
	}
}

// LoadRouterConfig 加载Router配置
func (loader *RouterConfigLoader) LoadRouterConfig(ctx context.Context, instanceId string) (*router.RouterConfig, error) {
	query := `
		SELECT tenantId, routerConfigId, gatewayInstanceId, routerName, defaultPriority,
		       enableRouteCache, routeCacheTtlSeconds, maxRoutes, routeMatchTimeout,
		       enableStrictMode, enableMetrics, enableTracing, caseSensitive, removeTrailingSlash,
		       enableGlobalFilters, filterExecutionMode, maxFilterChainDepth,
		       enableRoutePooling, routePoolSize, enableAsyncProcessing,
		       enableFallback, fallbackRoute, notFoundStatusCode, notFoundMessage,
		       routerMetadata, customConfig, activeFlag
		FROM HUB_GATEWAY_ROUTER_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY defaultPriority ASC
		LIMIT 1
	`

	var record RouterConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询Router配置失败: %w", err)
	}

	// 构建Router配置
	routerConfig := &router.RouterConfig{
		ID:              record.RouterConfigId,
		Enabled:         record.ActiveFlag == "Y",
		Name:            record.RouterName,
		DefaultPriority: record.DefaultPriority,
		EnableRouteCache: record.EnableRouteCache == "Y",
		RouteCacheTTL:   record.RouteCacheTtlSeconds,
	}

	// 加载路由配置
	routes, err := loader.LoadRoutes(ctx, instanceId)
	if err != nil {
		logger.Warn("加载路由配置失败", "error", err)
		routerConfig.Routes = []router.RouteConfig{}
	} else {
		routerConfig.Routes = routes
	}

	// 加载全局过滤器配置
	globalFilters, err := loader.LoadGlobalFilters(ctx, instanceId)
	if err != nil {
		logger.Warn("加载全局过滤器失败", "error", err)
	} else {
		routerConfig.FilterConfig = globalFilters
	}

	return routerConfig, nil
}

// LoadRoutes 加载路由配置
func (loader *RouterConfigLoader) LoadRoutes(ctx context.Context, instanceId string) ([]router.RouteConfig, error) {
	query := `
		SELECT tenantId, routeConfigId, gatewayInstanceId, routeName, routePath,
		       allowedMethods, allowedHosts, matchType, routePriority, stripPathPrefix,
		       rewritePath, enableWebsocket, timeoutMs, retryCount, retryIntervalMs,
		       serviceDefinitionId, logConfigId, routeMetadata, activeFlag
		FROM HUB_GATEWAY_ROUTE_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY routePriority ASC
	`

	var records []RouteConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由配置失败: %w", err)
	}

	var routes []router.RouteConfig
	for _, record := range records {
		// 构建路由配置
		routeConfig := router.RouteConfig{
			ID:       record.RouteConfigId,
			Name:     record.RouteName,
			Path:     record.RoutePath,
			Priority: record.RoutePriority,
			Enabled:  record.ActiveFlag == "Y",
		}

		// 设置服务ID
		if record.ServiceDefinitionId != nil {
			routeConfig.ServiceID = *record.ServiceDefinitionId
		}

		// 解析允许的方法
		if record.AllowedMethods != nil {
			var methods []string
			if err := json.Unmarshal([]byte(*record.AllowedMethods), &methods); err == nil {
				routeConfig.Methods = methods
			}
		}

		// 构建元数据
		metadata := make(map[string]interface{})
		metadata["match_type"] = record.MatchType
		metadata["strip_prefix"] = record.StripPathPrefix == "Y"
		metadata["enable_websocket"] = record.EnableWebsocket == "Y"
		metadata["timeout_ms"] = record.TimeoutMs
		metadata["retry_count"] = record.RetryCount
		metadata["retry_interval_ms"] = record.RetryIntervalMs

		if record.AllowedHosts != nil {
			metadata["allowed_hosts"] = strings.Split(*record.AllowedHosts, ",")
		}
		if record.RewritePath != nil {
			metadata["rewrite_path"] = *record.RewritePath
		}
		if record.RouteMetadata != nil {
			// 尝试解析JSON元数据
			var routeMetadata map[string]interface{}
			if err := json.Unmarshal([]byte(*record.RouteMetadata), &routeMetadata); err == nil {
				for k, v := range routeMetadata {
					metadata[k] = v
				}
			}
		}
		routeConfig.Metadata = metadata

		// 加载断言组配置
		assertionGroupConfig, err := loader.LoadRouteAssertionGroup(ctx, record.RouteConfigId)
		if err != nil {
			logger.Warn("加载路由断言组失败", 
				"routeId", record.RouteConfigId, 
				"error", err)
		} else if assertionGroupConfig != nil {
			routeConfig.AssertionGroupConfig = assertionGroupConfig
		}

		// 加载过滤器配置
		filters, err := loader.LoadRouteFilters(ctx, record.RouteConfigId)
		if err != nil {
			logger.Warn("加载路由过滤器失败", 
				"routeId", record.RouteConfigId, 
				"error", err)
		} else {
			routeConfig.FilterConfig = filters
		}

		routes = append(routes, routeConfig)
	}

	return routes, nil
}

// LoadRouteAssertionGroup 加载路由断言组配置
func (loader *RouterConfigLoader) LoadRouteAssertionGroup(ctx context.Context, routeId string) (*assertion.AssertionGroupConfig, error) {
	query := `
		SELECT tenantId, routeAssertionId, routeConfigId, assertionName, assertionType,
		       assertionOperator, fieldName, expectedValue, patternValue, caseSensitive,
		       assertionOrder, isRequired, assertionDesc, activeFlag
		FROM HUB_GATEWAY_ROUTE_ASSERTION 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY assertionOrder ASC
	`

	var records []RouteAssertionRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, routeId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由断言失败: %w", err)
	}

	if len(records) == 0 {
		return nil, nil
	}

	// 先查询路由元数据，获取断言组配置
	metadataQuery := `
		SELECT routeName, routeMetadata
		FROM HUB_GATEWAY_ROUTE_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		LIMIT 1
	`
	var metadataRecord struct {
		RouteName     string  `db:"routeName"`
		RouteMetadata *string `db:"routeMetadata"`
	}
	err = loader.db.QueryOne(ctx, &metadataRecord, metadataQuery, []interface{}{loader.tenantId, routeId}, true)
	
	// 默认所有断言都必须满足
	allRequired := true
	
	// 从元数据中获取断言组配置
	if err == nil && metadataRecord.RouteMetadata != nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(*metadataRecord.RouteMetadata), &metadata); err == nil {
			// 检查是否有断言组配置
			if assertionGroupSettings, ok := metadata["assertion_group"]; ok {
				if settings, ok := assertionGroupSettings.(map[string]interface{}); ok {
					// 获取 all_required 配置
					if allRequiredValue, ok := settings["all_required"]; ok {
						if boolValue, ok := allRequiredValue.(bool); ok {
							allRequired = boolValue
						}
					}
				}
			}
		}
	}

	// 创建断言组配置
	assertionGroupConfig := &assertion.AssertionGroupConfig{
		ID:               routeId + "_assertions",
		AllRequired:      allRequired, // 使用从元数据中获取的值
		AssertionConfigs: make([]assertion.AssertionConfig, 0),
		Description:      metadataRecord.RouteName + " - 路由断言组",
	}

	// 转换数据库记录为断言配置
	for _, record := range records {
		assertionConfig := assertion.AssertionConfig{
			ID:          record.RouteAssertionId,
			Type:        strings.ToLower(record.AssertionType), // 转换为小写
			Name:        strings.ToLower(record.AssertionType),
			Operator:    strings.ToLower(record.AssertionOperator), // 转换为小写
			CaseSensitive: record.CaseSensitive == "Y",
		}

		// 根据断言类型设置值
		switch strings.ToUpper(record.AssertionType) {
		case "PATH":
			if record.ExpectedValue != nil {
				assertionConfig.Value = *record.ExpectedValue
			}
			if record.PatternValue != nil {
				assertionConfig.Pattern = *record.PatternValue
			}
		case "HEADER":
			if record.FieldName != nil {
				assertionConfig.Name = *record.FieldName
			}
			if record.ExpectedValue != nil {
				assertionConfig.Value = *record.ExpectedValue
			}
		case "QUERY":
			if record.FieldName != nil {
				assertionConfig.Name = *record.FieldName
			}
			if record.ExpectedValue != nil {
				assertionConfig.Value = *record.ExpectedValue
			}
		case "METHOD":
			if record.ExpectedValue != nil {
				assertionConfig.Value = *record.ExpectedValue
			}
		default:
			if record.ExpectedValue != nil {
				assertionConfig.Value = *record.ExpectedValue
			}
		}

		if record.AssertionDesc != nil {
			assertionConfig.Description = *record.AssertionDesc
		}

		assertionGroupConfig.AssertionConfigs = append(assertionGroupConfig.AssertionConfigs, assertionConfig)
	}

	return assertionGroupConfig, nil
}

// LoadRouteFilters 加载路由过滤器配置
func (loader *RouterConfigLoader) LoadRouteFilters(ctx context.Context, routeId string) ([]filter.FilterConfig, error) {
	query := `
		SELECT tenantId, filterConfigId, filterName, filterType, filterAction,
		       filterOrder, filterConfig, configId, activeFlag
		FROM HUB_GATEWAY_FILTER_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY filterOrder ASC
	`

	var records []FilterConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, routeId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由过滤器失败: %w", err)
	}

	var filters []filter.FilterConfig
	for _, record := range records {
		// 解析过滤器配置
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(record.FilterConfig), &config); err != nil {
			logger.Error("解析过滤器配置失败", 
				"filterId", record.FilterConfigId, 
				"error", err)
			continue
		}

		// 修复配置层级问题：直接使用解析的配置，不再嵌套
		// 如果配置中有特定的子配置对象（如headerConfig），将其提升到顶层
		flatConfig := make(map[string]interface{})
		for key, value := range config {
			if subConfig, ok := value.(map[string]interface{}); ok && 
			   (key == "headerConfig" || key == "queryConfig" || key == "bodyConfig" || key == "urlConfig") {
				// 将子配置的内容提升到顶层
				for subKey, subValue := range subConfig {
					flatConfig[subKey] = subValue
				}
			} else {
				flatConfig[key] = value
			}
		}

		// 构建过滤器配置
		filterCfg := filter.FilterConfig{
			ID:      record.FilterConfigId,
			Name:    record.FilterName,
			Type:    record.FilterType,
			Enabled: record.ActiveFlag == "Y",
			Order:   record.FilterOrder,
			Action:  record.FilterAction,
			Config:  flatConfig, // 使用扁平化的配置
		}

		filters = append(filters, filterCfg)
	}

	return filters, nil
}

// LoadGlobalFilters 加载全局过滤器配置
func (loader *RouterConfigLoader) LoadGlobalFilters(ctx context.Context, instanceId string) ([]filter.FilterConfig, error) {
	query := `
		SELECT tenantId, filterConfigId, filterName, filterType, filterAction,
		       filterOrder, filterConfig, configId, activeFlag
		FROM HUB_GATEWAY_FILTER_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND routeConfigId IS NULL AND activeFlag = 'Y'
		ORDER BY filterOrder ASC
	`

	var records []FilterConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询全局过滤器失败: %w", err)
	}

	var filters []filter.FilterConfig
	for _, record := range records {
		// 解析过滤器配置
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(record.FilterConfig), &config); err != nil {
			logger.Error("解析全局过滤器配置失败", 
				"filterId", record.FilterConfigId, 
				"error", err)
			continue
		}

		// 修复配置层级问题：直接使用解析的配置，不再嵌套
		flatConfig := make(map[string]interface{})
		for key, value := range config {
			if subConfig, ok := value.(map[string]interface{}); ok && 
			   (key == "headerConfig" || key == "queryConfig" || key == "bodyConfig" || key == "urlConfig") {
				// 将子配置的内容提升到顶层
				for subKey, subValue := range subConfig {
					flatConfig[subKey] = subValue
				}
			} else {
				flatConfig[key] = value
			}
		}

		// 构建过滤器配置
		filterCfg := filter.FilterConfig{
			ID:      record.FilterConfigId,
			Name:    record.FilterName,
			Type:    record.FilterType,
			Enabled: record.ActiveFlag == "Y",
			Order:   record.FilterOrder,
			Action:  record.FilterAction,
			Config:  flatConfig, // 使用扁平化的配置
		}

		filters = append(filters, filterCfg)
	}

	return filters, nil
} 