package controllers

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"gateway/pkg/excel"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0020/models"
	hub0021models "gateway/web/views/hub0021/models"
	hub0022dao "gateway/web/views/hub0022/dao"
	hub0022models "gateway/web/views/hub0022/models"
	hubcommon002models "gateway/web/views/hubcommon002/models"

	"github.com/gin-gonic/gin"
)

// ExportGatewayInstance 导出完整网关实例配置为 Excel 文件
func (c *GatewayInstanceController) ExportGatewayInstance(ctx *gin.Context) {
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	if gatewayInstanceId == "" {
		response.ErrorJSON(ctx, "gatewayInstanceId不能为空", constants.ED00006)
		return
	}
	tenantId := request.GetTenantID(ctx)

	instance, err := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例失败", err)
		response.ErrorJSON(ctx, "获取网关实例失败: "+err.Error(), constants.ED00009)
		return
	}
	if instance == nil {
		response.ErrorJSON(ctx, "网关实例不存在", constants.ED00008)
		return
	}

	sheets, err := c.buildSheets(ctx, instance, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建导出数据失败", err)
		response.ErrorJSON(ctx, "构建导出数据失败: "+err.Error(), constants.ED00009)
		return
	}

	filename := fmt.Sprintf("GatewayInstance_%s_%s.xlsx",
		instance.InstanceName,
		time.Now().Format("20060102150405"))
	tmpPath := filepath.Join(os.TempDir(), filename)
	// 无论 Build 成功与否都清理临时文件，避免 Build 中途失败留下残留
	defer os.Remove(tmpPath)

	result, err := excel.Build(tmpPath, sheets...)
	if err != nil {
		logger.ErrorWithTrace(ctx, "生成 Excel 失败", err)
		response.ErrorJSON(ctx, "生成 Excel 失败: "+err.Error(), constants.ED00009)
		return
	}

	file, err := os.Open(result.Path)
	if err != nil {
		logger.ErrorWithTrace(ctx, "打开临时文件失败", err)
		response.ErrorJSON(ctx, "读取导出文件失败: "+err.Error(), constants.ED00009)
		return
	}
	defer file.Close()

	encoded := url.PathEscape(filename)
	ctx.Writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Writer.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encoded))
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", result.Size))
	ctx.Writer.WriteHeader(200)
	io.Copy(ctx.Writer, file) //nolint:errcheck
}

// buildSheets 查询所有关联数据并构造 excel.Sheet 列表
func (c *GatewayInstanceController) buildSheets(
	ctx *gin.Context,
	instance *models.GatewayInstance,
	gatewayInstanceId, tenantId string,
) ([]excel.Sheet, error) {
	// ── 预加载多处复用的数据 ──────────────────────────────────────────────────
	securityConfigs, _ := c.securityConfigDAO.ListSecurityConfigsByGatewayInstance(ctx, gatewayInstanceId, tenantId)

	routes, err := c.routeConfigDAO.GetRouteConfigsByGatewayInstance(ctx, gatewayInstanceId, tenantId, "")
	if err != nil {
		return nil, fmt.Errorf("获取路由配置失败: %w", err)
	}

	// 代理配置需在服务定义之前查询，服务定义按 proxyConfigId 过滤
	proxyConfigs, _, err := c.proxyConfigDAO.ListProxyConfigs(ctx, tenantId, gatewayInstanceId, 1, 10000)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取代理配置失败", "error", err)
	}

	// ── t_gateway_instance ────────────────────────────────────────────────────
	gatewayInstanceSheet := excel.Sheet{
		Name: models.GatewayInstance{}.TableName(),
		Headers: []string{
			"gatewayInstanceId", "instanceName", "instanceDesc",
			"bindAddress", "httpPort", "httpsPort", "tlsEnabled",
			"certStorageType", "certFilePath", "keyFilePath",
			"certContent", "keyContent", "certChainContent", "certPassword",
			"maxConnections", "readTimeoutMs", "writeTimeoutMs", "idleTimeoutMs",
			"maxHeaderBytes", "maxWorkers", "keepAliveEnabled", "tcpKeepAliveEnabled",
			"gracefulShutdownTimeoutMs", "enableHttp2", "tlsVersion", "tlsCipherSuites",
			"disableGeneralOptionsHandler",
			"logConfigId", "healthStatus", "lastHeartbeatTime", "instanceMetadata",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
		Rows: [][]any{{
			instance.GatewayInstanceId, instance.InstanceName, instance.InstanceDesc,
			instance.BindAddress, instance.HttpPort, instance.HttpsPort, instance.TlsEnabled,
			instance.CertStorageType, instance.CertFilePath, instance.KeyFilePath,
			instance.CertContent, instance.KeyContent, instance.CertChainContent, instance.CertPassword,
			instance.MaxConnections, instance.ReadTimeoutMs, instance.WriteTimeoutMs, instance.IdleTimeoutMs,
			instance.MaxHeaderBytes, instance.MaxWorkers, instance.KeepAliveEnabled, instance.TcpKeepAliveEnabled,
			instance.GracefulShutdownTimeoutMs, instance.EnableHttp2, instance.TlsVersion, instance.TlsCipherSuites,
			instance.DisableGeneralOptionsHandler,
			instance.LogConfigId, instance.HealthStatus, instance.LastHeartbeatTime, instance.InstanceMetadata,
			instance.Reserved1, instance.Reserved2, instance.Reserved3, instance.Reserved4, instance.Reserved5,
			instance.ExtProperty,
			instance.OprSeqFlag, instance.CurrentVersion, instance.ActiveFlag, instance.NoteText,
			instance.AddTime, instance.AddWho, instance.EditTime, instance.EditWho,
		}},
	}

	// ── t_log_config ──────────────────────────────────────────────────────────
	logConfigSheet := excel.Sheet{
		Name: models.LogConfig{}.TableName(),
		Headers: []string{
			"logConfigId", "configName", "configDesc",
			"logFormat", "recordRequestBody", "recordResponseBody", "recordHeaders", "maxBodySizeBytes",
			"outputTargets", "fileConfig", "databaseConfig",
			"mongoConfig", "elasticsearchConfig", "clickhouseConfig",
			"enableAsyncLogging", "asyncQueueSize", "asyncFlushIntervalMs",
			"enableBatchProcessing", "batchSize", "batchTimeoutMs",
			"logRetentionDays", "enableFileRotation", "maxFileSizeMB", "maxFileCount", "rotationPattern",
			"enableSensitiveDataMasking", "sensitiveFields", "maskingPattern",
			"bufferSize", "flushThreshold", "configPriority",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	if instance.LogConfigId != "" {
		lc, err := c.logConfigDAO.GetLogConfigById(ctx, instance.LogConfigId, tenantId)
		if err != nil {
			logger.WarnWithTrace(ctx, "获取日志配置失败", "error", err)
		} else if lc != nil {
			logConfigSheet.Rows = append(logConfigSheet.Rows, []any{
				lc.LogConfigId, lc.ConfigName, lc.ConfigDesc,
				lc.LogFormat, lc.RecordRequestBody, lc.RecordResponseBody, lc.RecordHeaders, lc.MaxBodySizeBytes,
				lc.OutputTargets, lc.FileConfig, lc.DatabaseConfig,
				lc.MongoConfig, lc.ElasticsearchConfig, lc.ClickhouseConfig,
				lc.EnableAsyncLogging, lc.AsyncQueueSize, lc.AsyncFlushIntervalMs,
				lc.EnableBatchProcessing, lc.BatchSize, lc.BatchTimeoutMs,
				lc.LogRetentionDays, lc.EnableFileRotation, lc.MaxFileSizeMB, lc.MaxFileCount, lc.RotationPattern,
				lc.EnableSensitiveDataMasking, lc.SensitiveFields, lc.MaskingPattern,
				lc.BufferSize, lc.FlushThreshold, lc.ConfigPriority,
				lc.Reserved1, lc.Reserved2, lc.Reserved3, lc.Reserved4, lc.Reserved5,
				lc.ExtProperty,
				lc.OprSeqFlag, lc.CurrentVersion, lc.ActiveFlag, lc.NoteText,
				lc.AddTime, lc.AddWho, lc.EditTime, lc.EditWho,
			})
		}
	}

	// ── t_route_config ────────────────────────────────────────────────────────
	routeConfigSheet := excel.Sheet{
		Name: hub0021models.RouteConfig{}.TableName(),
		Headers: []string{
			"routeConfigId", "gatewayInstanceId", "routeName", "routePath",
			"allowedMethods", "allowedHosts", "matchType", "routePriority",
			"stripPathPrefix", "rewritePath", "enableWebsocket",
			"timeoutMs", "retryCount", "retryIntervalMs",
			"serviceDefinitionId", "logConfigId", "routeMetadata",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	for _, rc := range routes {
		routeConfigSheet.Rows = append(routeConfigSheet.Rows, []any{
			rc.RouteConfigId, rc.GatewayInstanceId, rc.RouteName, rc.RoutePath,
			rc.AllowedMethods, rc.AllowedHosts, rc.MatchType, rc.RoutePriority,
			rc.StripPathPrefix, rc.RewritePath, rc.EnableWebsocket,
			rc.TimeoutMs, rc.RetryCount, rc.RetryIntervalMs,
			rc.ServiceDefinitionId, rc.LogConfigId, rc.RouteMetadata,
			rc.Reserved1, rc.Reserved2, rc.Reserved3, rc.Reserved4, rc.Reserved5,
			rc.ExtProperty,
			rc.OprSeqFlag, rc.CurrentVersion, rc.ActiveFlag, rc.NoteText,
			rc.AddTime, rc.AddWho, rc.EditTime, rc.EditWho,
		})
	}

	// ── t_route_assertion ─────────────────────────────────────────────────────
	routeAssertionSheet := excel.Sheet{
		Name: hub0021models.RouteAssertion{}.TableName(),
		Headers: []string{
			"routeAssertionId", "routeConfigId",
			"assertionName", "assertionType", "assertionOperator",
			"fieldName", "expectedValue", "patternValue",
			"caseSensitive", "assertionOrder", "isRequired", "assertionDesc",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	for _, rc := range routes {
		assertions, _ := c.routeAssertionDAO.GetRouteAssertionsByRouteId(ctx, rc.RouteConfigId, tenantId)
		for _, a := range assertions {
			routeAssertionSheet.Rows = append(routeAssertionSheet.Rows, []any{
				a.RouteAssertionId, a.RouteConfigId,
				a.AssertionName, a.AssertionType, a.AssertionOperator,
				a.FieldName, a.ExpectedValue, a.PatternValue,
				a.CaseSensitive, a.AssertionOrder, a.IsRequired, a.AssertionDesc,
				a.Reserved1, a.Reserved2, a.Reserved3, a.Reserved4, a.Reserved5,
				a.ExtProperty,
				a.OprSeqFlag, a.CurrentVersion, a.ActiveFlag, a.NoteText,
				a.AddTime, a.AddWho, a.EditTime, a.EditWho,
			})
		}
	}

	// ── t_filter_config（路由级）──────────────────────────────────────────────
	routeFiltersSheet := filterSheet(hub0021models.FilterConfig{}.TableName() + "_route")
	for _, rc := range routes {
		filters, _ := c.filterConfigDAO.GetFilterConfigsByRoute(ctx, rc.RouteConfigId, tenantId, "")
		for _, f := range filters {
			routeFiltersSheet.Rows = append(routeFiltersSheet.Rows, filterRow(f))
		}
	}

	// ── t_filter_config（实例级）─────────────────────────────────────────────
	instanceFiltersSheet := filterSheet(hub0021models.FilterConfig{}.TableName() + "_instance")
	instanceFilters, err := c.filterConfigDAO.GetFilterConfigsByGatewayInstance(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取实例级过滤器失败", "error", err)
	}
	for _, f := range instanceFilters {
		instanceFiltersSheet.Rows = append(instanceFiltersSheet.Rows, filterRow(f))
	}

	// ── t_router_config ───────────────────────────────────────────────────────
	routerConfigSheet := excel.Sheet{
		Name: hub0021models.RouterConfig{}.TableName(),
		Headers: []string{
			"routerConfigId", "gatewayInstanceId", "routerName", "routerDesc",
			"defaultPriority", "enableRouteCache", "routeCacheTtlSeconds",
			"maxRoutes", "routeMatchTimeout",
			"enableStrictMode", "enableMetrics", "enableTracing",
			"caseSensitive", "removeTrailingSlash",
			"enableGlobalFilters", "filterExecutionMode", "maxFilterChainDepth",
			"enableRoutePooling", "routePoolSize", "enableAsyncProcessing",
			"enableFallback", "fallbackRoute", "notFoundStatusCode", "notFoundMessage",
			"routerMetadata", "customConfig",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	routerRows, _, err := c.routerConfigDAO.ListRouterConfigs(ctx, tenantId, gatewayInstanceId, 1, 100)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取Router配置失败", "error", err)
	}
	for _, rc := range routerRows {
		routerConfigSheet.Rows = append(routerConfigSheet.Rows, []any{
			rc.RouterConfigId, rc.GatewayInstanceId, rc.RouterName, rc.RouterDesc,
			rc.DefaultPriority, rc.EnableRouteCache, rc.RouteCacheTtlSeconds,
			rc.MaxRoutes, rc.RouteMatchTimeout,
			rc.EnableStrictMode, rc.EnableMetrics, rc.EnableTracing,
			rc.CaseSensitive, rc.RemoveTrailingSlash,
			rc.EnableGlobalFilters, rc.FilterExecutionMode, rc.MaxFilterChainDepth,
			rc.EnableRoutePooling, rc.RoutePoolSize, rc.EnableAsyncProcessing,
			rc.EnableFallback, rc.FallbackRoute, rc.NotFoundStatusCode, rc.NotFoundMessage,
			rc.RouterMetadata, rc.CustomConfig,
			rc.Reserved1, rc.Reserved2, rc.Reserved3, rc.Reserved4, rc.Reserved5,
			rc.ExtProperty,
			rc.OprSeqFlag, rc.CurrentVersion, rc.ActiveFlag, rc.NoteText,
			rc.AddTime, rc.AddWho, rc.EditTime, rc.EditWho,
		})
	}

	// ── t_service_definition ──────────────────────────────────────────────────
	serviceDefSheet := excel.Sheet{
		Name: hub0021models.ServiceDefinition{}.TableName(),
		Headers: []string{
			"serviceDefinitionId", "serviceName", "serviceDesc", "serviceType",
			"proxyConfigId", "loadBalanceStrategy",
			"discoveryType", "discoveryConfig",
			"sessionAffinity", "stickySession", "maxRetries", "retryTimeoutMs", "enableCircuitBreaker",
			"healthCheckEnabled", "healthCheckPath", "healthCheckMethod",
			"healthCheckIntervalSeconds", "healthCheckTimeoutMs",
			"healthyThreshold", "unhealthyThreshold", "expectedStatusCodes", "healthCheckHeaders",
			"loadBalancerConfig", "serviceMetadata",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	// 服务定义参照 hub0022 的 service_definition_controller.go 查询方式：
	// 你的数据里 service_definition.proxyConfigId 实际保存的是 gatewayInstanceId，
	// 因此这里直接使用 gatewayInstanceId 作为过滤条件（同时兼容性地追加 proxyConfigs 的 proxyConfigId）。
	proxyIds := []string{gatewayInstanceId}
	seen := map[string]struct{}{gatewayInstanceId: {}}
	for _, pc := range proxyConfigs {
		if pc == nil || pc.ProxyConfigId == "" {
			continue
		}
		if _, ok := seen[pc.ProxyConfigId]; ok {
			continue
		}
		seen[pc.ProxyConfigId] = struct{}{}
		proxyIds = append(proxyIds, pc.ProxyConfigId)
	}

	sds, _, err := c.svcDefDAO.ListServiceDefinitions(ctx, tenantId, 1, 10000,
		&hub0022dao.ServiceDefinitionQueryFilter{ProxyConfigIds: proxyIds})
	if err != nil {
		logger.WarnWithTrace(ctx, "获取服务定义失败", "error", err)
		sds = nil
	}
	for _, sd := range sds {
		serviceDefSheet.Rows = append(serviceDefSheet.Rows, []any{
			sd.ServiceDefinitionId, sd.ServiceName, sd.ServiceDesc, sd.ServiceType,
			sd.ProxyConfigId, sd.LoadBalanceStrategy,
			sd.DiscoveryType, sd.DiscoveryConfig,
			sd.SessionAffinity, sd.StickySession, sd.MaxRetries, sd.RetryTimeoutMs, sd.EnableCircuitBreaker,
			sd.HealthCheckEnabled, sd.HealthCheckPath, sd.HealthCheckMethod,
			sd.HealthCheckIntervalSeconds, sd.HealthCheckTimeoutMs,
			sd.HealthyThreshold, sd.UnhealthyThreshold, sd.ExpectedStatusCodes, sd.HealthCheckHeaders,
			sd.LoadBalancerConfig, sd.ServiceMetadata,
			sd.Reserved1, sd.Reserved2, sd.Reserved3, sd.Reserved4, sd.Reserved5,
			sd.ExtProperty,
			sd.OprSeqFlag, sd.CurrentVersion, sd.ActiveFlag, sd.NoteText,
			sd.AddTime, sd.AddWho, sd.EditTime, sd.EditWho,
		})
	}

	// ── t_proxy_config ────────────────────────────────────────────────────────
	proxyConfigSheet := excel.Sheet{
		Name: hub0022models.ProxyConfig{}.TableName(),
		Headers: []string{
			"proxyConfigId", "gatewayInstanceId", "proxyName", "proxyType",
			"proxyId", "configPriority",
			"proxyConfig", "customConfig",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	for _, pc := range proxyConfigs {
		proxyConfigSheet.Rows = append(proxyConfigSheet.Rows, []any{
			pc.ProxyConfigId, pc.GatewayInstanceId, pc.ProxyName, pc.ProxyType,
			pc.ProxyId, pc.ConfigPriority,
			pc.ProxyConfig, pc.CustomConfig,
			pc.Reserved1, pc.Reserved2, pc.Reserved3, pc.Reserved4, pc.Reserved5,
			pc.ExtProperty,
			pc.OprSeqFlag, pc.CurrentVersion, pc.ActiveFlag, pc.NoteText,
			pc.AddTime, pc.AddWho, pc.EditTime, pc.EditWho,
		})
	}

	// ── t_service_node ────────────────────────────────────────────────────────
	serviceNodeSheet := excel.Sheet{
		Name: "HUB_GW_SERVICE_NODE",
		Headers: []string{
			"serviceNodeId", "serviceDefinitionId", "nodeId",
			"nodeUrl", "nodeHost", "nodePort", "nodeProtocol",
			"nodeWeight", "healthStatus", "nodeMetadata",
			"nodeStatus", "lastHealthCheckTime", "healthCheckResult",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	// 收集所有涉及的 serviceDefinitionId：
	// 来源1：路由直接关联的服务定义
	// 来源2：服务定义 Sheet（hub0022 按 proxyConfigId/gatewayInstanceId 过滤出的服务定义）
	svcDefIdSetForNode := map[string]struct{}{}
	for _, r := range routes {
		if r.ServiceDefinitionId != "" {
			svcDefIdSetForNode[r.ServiceDefinitionId] = struct{}{}
		}
	}
	for _, sd := range sds {
		svcDefIdSetForNode[sd.ServiceDefinitionId] = struct{}{}
	}
	for svcDefId := range svcDefIdSetForNode {
		nodes, nodeErr := c.serviceNodeDAO.GetServiceNodesByService(ctx, svcDefId, tenantId)
		if nodeErr != nil {
			logger.WarnWithTrace(ctx, "获取服务节点失败", "serviceDefinitionId", svcDefId, "error", nodeErr)
			continue
		}
		for _, n := range nodes {
			serviceNodeSheet.Rows = append(serviceNodeSheet.Rows, []any{
				n.ServiceNodeId, n.ServiceDefinitionId, n.NodeId,
				n.NodeUrl, n.NodeHost, n.NodePort, n.NodeProtocol,
				n.NodeWeight, n.HealthStatus, n.NodeMetadata,
				n.NodeStatus, n.LastHealthCheckTime, n.HealthCheckResult,
				n.Reserved1, n.Reserved2, n.Reserved3, n.Reserved4, n.Reserved5,
				n.ExtProperty,
				n.OprSeqFlag, n.CurrentVersion, n.ActiveFlag, n.NoteText,
				n.AddTime, n.AddWho, n.EditTime, n.EditWho,
			})
		}
	}

	// ── t_security_config ─────────────────────────────────────────────────────
	securityConfigSheet := excel.Sheet{
		Name: hubcommon002models.SecurityConfig{}.TableName(),
		Headers: []string{
			"securityConfigId", "gatewayInstanceId", "routeConfigId",
			"configName", "configDesc", "configPriority", "customConfigJson",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	for _, sc := range securityConfigs {
		securityConfigSheet.Rows = append(securityConfigSheet.Rows, []any{
			sc.SecurityConfigId, sc.GatewayInstanceId, sc.RouteConfigId,
			sc.ConfigName, sc.ConfigDesc, sc.ConfigPriority, sc.CustomConfigJson,
			sc.Reserved1, sc.Reserved2, sc.Reserved3, sc.Reserved4, sc.Reserved5,
			sc.ExtProperty,
			sc.OprSeqFlag, sc.CurrentVersion, sc.ActiveFlag, sc.NoteText,
			sc.AddTime, sc.AddWho, sc.EditTime, sc.EditWho,
		})
	}

	// ── t_ip_access_config ────────────────────────────────────────────────────
	ipAccessSheet := excel.Sheet{
		Name: hubcommon002models.IpAccessConfig{}.TableName(),
		Headers: []string{
			"ipAccessConfigId", "securityConfigId", "configName", "defaultPolicy",
			"whitelistIps", "blacklistIps", "whitelistCidrs", "blacklistCidrs",
			"trustXForwardedFor", "trustXRealIp",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	for _, sc := range securityConfigs {
		rows, _, _ := c.ipAccessConfigDAO.ListIpAccessConfigs(ctx, tenantId,
			&hubcommon002models.IpAccessConfigQuery{SecurityConfigId: sc.SecurityConfigId}, 1, 1000)
		for _, ip := range rows {
			ipAccessSheet.Rows = append(ipAccessSheet.Rows, []any{
				ip.IpAccessConfigId, ip.SecurityConfigId, ip.ConfigName, ip.DefaultPolicy,
				ip.WhitelistIps, ip.BlacklistIps, ip.WhitelistCidrs, ip.BlacklistCidrs,
				ip.TrustXForwardedFor, ip.TrustXRealIp,
				ip.Reserved1, ip.Reserved2, ip.Reserved3, ip.Reserved4, ip.Reserved5,
				ip.ExtProperty,
				ip.OprSeqFlag, ip.CurrentVersion, ip.ActiveFlag, ip.NoteText,
				ip.AddTime, ip.AddWho, ip.EditTime, ip.EditWho,
			})
		}
	}

	// ── t_useragent_access_config ─────────────────────────────────────────────
	uaAccessSheet := excel.Sheet{
		Name: hubcommon002models.UseragentAccessConfig{}.TableName(),
		Headers: []string{
			"useragentAccessConfigId", "securityConfigId", "configName", "defaultPolicy",
			"whitelistPatterns", "blacklistPatterns", "blockEmptyUserAgent",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	for _, sc := range securityConfigs {
		rows, _, _ := c.uaAccessConfigDAO.ListUseragentAccessConfigs(ctx, tenantId,
			&hubcommon002models.UseragentAccessConfigQuery{SecurityConfigId: sc.SecurityConfigId}, 1, 1000)
		for _, ua := range rows {
			uaAccessSheet.Rows = append(uaAccessSheet.Rows, []any{
				ua.UseragentAccessConfigId, ua.SecurityConfigId, ua.ConfigName, ua.DefaultPolicy,
				ua.WhitelistPatterns, ua.BlacklistPatterns, ua.BlockEmptyUserAgent,
				ua.Reserved1, ua.Reserved2, ua.Reserved3, ua.Reserved4, ua.Reserved5,
				ua.ExtProperty,
				ua.OprSeqFlag, ua.CurrentVersion, ua.ActiveFlag, ua.NoteText,
				ua.AddTime, ua.AddWho, ua.EditTime, ua.EditWho,
			})
		}
	}

	// ── t_domain_access_config ────────────────────────────────────────────────
	domainAccessSheet := excel.Sheet{
		Name: hubcommon002models.DomainAccessConfig{}.TableName(),
		Headers: []string{
			"domainAccessConfigId", "securityConfigId", "configName", "defaultPolicy",
			"whitelistDomains", "blacklistDomains", "allowSubdomains",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	for _, sc := range securityConfigs {
		rows, _, _ := c.domainAccessConfigDAO.ListDomainAccessConfigs(ctx, tenantId,
			&hubcommon002models.DomainAccessConfigQuery{SecurityConfigId: sc.SecurityConfigId}, 1, 1000)
		for _, d := range rows {
			domainAccessSheet.Rows = append(domainAccessSheet.Rows, []any{
				d.DomainAccessConfigId, d.SecurityConfigId, d.ConfigName, d.DefaultPolicy,
				d.WhitelistDomains, d.BlacklistDomains, d.AllowSubdomains,
				d.Reserved1, d.Reserved2, d.Reserved3, d.Reserved4, d.Reserved5,
				d.ExtProperty,
				d.OprSeqFlag, d.CurrentVersion, d.ActiveFlag, d.NoteText,
				d.AddTime, d.AddWho, d.EditTime, d.EditWho,
			})
		}
	}

	// ── t_api_access_config ───────────────────────────────────────────────────
	apiAccessSheet := excel.Sheet{
		Name: hubcommon002models.ApiAccessConfig{}.TableName(),
		Headers: []string{
			"apiAccessConfigId", "securityConfigId", "configName", "defaultPolicy",
			"whitelistPaths", "blacklistPaths", "allowedMethods", "blockedMethods",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	for _, sc := range securityConfigs {
		rows, _, _ := c.apiAccessConfigDAO.ListApiAccessConfigs(ctx, tenantId,
			&hubcommon002models.ApiAccessConfigQuery{SecurityConfigId: sc.SecurityConfigId}, 1, 1000)
		for _, a := range rows {
			apiAccessSheet.Rows = append(apiAccessSheet.Rows, []any{
				a.ApiAccessConfigId, a.SecurityConfigId, a.ConfigName, a.DefaultPolicy,
				a.WhitelistPaths, a.BlacklistPaths, a.AllowedMethods, a.BlockedMethods,
				a.Reserved1, a.Reserved2, a.Reserved3, a.Reserved4, a.Reserved5,
				a.ExtProperty,
				a.OprSeqFlag, a.CurrentVersion, a.ActiveFlag, a.NoteText,
				a.AddTime, a.AddWho, a.EditTime, a.EditWho,
			})
		}
	}

	// ── t_cors_config ─────────────────────────────────────────────────────────
	corsConfigSheet := excel.Sheet{
		Name: hubcommon002models.CorsConfig{}.TableName(),
		Headers: []string{
			"corsConfigId", "gatewayInstanceId", "routeConfigId",
			"configName", "allowOrigins", "allowMethods", "allowHeaders", "exposeHeaders",
			"allowCredentials", "maxAgeSeconds", "configPriority",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	corsRows, _, err := c.corsConfigDAO.ListCorsConfigsByGatewayInstance(ctx, tenantId, gatewayInstanceId, 1, 1000)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取CORS配置失败", "error", err)
	}
	for _, cc := range corsRows {
		corsConfigSheet.Rows = append(corsConfigSheet.Rows, []any{
			cc.CorsConfigId, cc.GatewayInstanceId, cc.RouteConfigId,
			cc.ConfigName, cc.AllowOrigins, cc.AllowMethods, cc.AllowHeaders, cc.ExposeHeaders,
			cc.AllowCredentials, cc.MaxAgeSeconds, cc.ConfigPriority,
			cc.Reserved1, cc.Reserved2, cc.Reserved3, cc.Reserved4, cc.Reserved5,
			cc.ExtProperty,
			cc.OprSeqFlag, cc.CurrentVersion, cc.ActiveFlag, cc.NoteText,
			cc.AddTime, cc.AddWho, cc.EditTime, cc.EditWho,
		})
	}

	// ── t_auth_config ─────────────────────────────────────────────────────────
	authConfigSheet := excel.Sheet{
		Name: hubcommon002models.AuthConfig{}.TableName(),
		Headers: []string{
			"authConfigId", "gatewayInstanceId", "routeConfigId",
			"authName", "authType", "authStrategy", "authConfig",
			"exemptPaths", "exemptHeaders", "failureStatusCode", "failureMessage", "configPriority",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	ac, err := c.authConfigDAO.GetAuthConfigByGatewayInstance(tenantId, gatewayInstanceId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取认证配置失败", "error", err)
	} else if ac != nil {
		authConfigSheet.Rows = append(authConfigSheet.Rows, []any{
			ac.AuthConfigId, ac.GatewayInstanceId, ac.RouteConfigId,
			ac.AuthName, ac.AuthType, ac.AuthStrategy, ac.AuthConfig,
			ac.ExemptPaths, ac.ExemptHeaders, ac.FailureStatusCode, ac.FailureMessage, ac.ConfigPriority,
			ac.Reserved1, ac.Reserved2, ac.Reserved3, ac.Reserved4, ac.Reserved5,
			ac.ExtProperty,
			ac.OprSeqFlag, ac.CurrentVersion, ac.ActiveFlag, ac.NoteText,
			ac.AddTime, ac.AddWho, ac.EditTime, ac.EditWho,
		})
	}

	// ── t_rate_limit_config ───────────────────────────────────────────────────
	rateLimitSheet := excel.Sheet{
		Name: hubcommon002models.RateLimitConfig{}.TableName(),
		Headers: []string{
			"rateLimitConfigId", "gatewayInstanceId", "routeConfigId",
			"limitName", "algorithm", "keyStrategy",
			"limitRate", "burstCapacity", "timeWindowSeconds",
			"rejectionStatusCode", "rejectionMessage", "configPriority", "customConfig",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
	rl, err := c.rateLimitConfigDAO.GetRateLimitConfigByGatewayInstance(tenantId, gatewayInstanceId)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取限流配置失败", "error", err)
	} else if rl != nil {
		rateLimitSheet.Rows = append(rateLimitSheet.Rows, []any{
			rl.RateLimitConfigId, rl.GatewayInstanceId, rl.RouteConfigId,
			rl.LimitName, rl.Algorithm, rl.KeyStrategy,
			rl.LimitRate, rl.BurstCapacity, rl.TimeWindowSeconds,
			rl.RejectionStatusCode, rl.RejectionMessage, rl.ConfigPriority, rl.CustomConfig,
			rl.Reserved1, rl.Reserved2, rl.Reserved3, rl.Reserved4, rl.Reserved5,
			rl.ExtProperty,
			rl.OprSeqFlag, rl.CurrentVersion, rl.ActiveFlag, rl.NoteText,
			rl.AddTime, rl.AddWho, rl.EditTime, rl.EditWho,
		})
	}

	return []excel.Sheet{
		gatewayInstanceSheet,
		logConfigSheet,
		routeConfigSheet,
		routeAssertionSheet,
		routeFiltersSheet,
		instanceFiltersSheet,
		routerConfigSheet,
		serviceDefSheet,
		proxyConfigSheet,
		serviceNodeSheet,
		securityConfigSheet,
		ipAccessSheet,
		uaAccessSheet,
		domainAccessSheet,
		apiAccessSheet,
		corsConfigSheet,
		authConfigSheet,
		rateLimitSheet,
	}, nil
}

// filterSheet 构造过滤器类 Sheet（路由级/实例级共用列头）
func filterSheet(name string) excel.Sheet {
	return excel.Sheet{
		Name: name,
		Headers: []string{
			"filterConfigId", "gatewayInstanceId", "routeConfigId",
			"filterName", "filterType", "filterAction",
			"filterOrder", "filterConfig", "filterDesc", "configId",
			"reserved1", "reserved2", "reserved3", "reserved4", "reserved5",
			"extProperty",
			"oprSeqFlag", "currentVersion", "activeFlag", "noteText",
			"addTime", "addWho", "editTime", "editWho",
		},
	}
}

func filterRow(fc *hub0021models.FilterConfig) []any {
	return []any{
		fc.FilterConfigId, fc.GatewayInstanceId, fc.RouteConfigId,
		fc.FilterName, fc.FilterType, fc.FilterAction,
		fc.FilterOrder, fc.FilterConfig, fc.FilterDesc, fc.ConfigId,
		fc.Reserved1, fc.Reserved2, fc.Reserved3, fc.Reserved4, fc.Reserved5,
		fc.ExtProperty,
		fc.OprSeqFlag, fc.CurrentVersion, fc.ActiveFlag, fc.NoteText,
		fc.AddTime, fc.AddWho, fc.EditTime, fc.EditWho,
	}
}
