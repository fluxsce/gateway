package controllers

import (
	"strconv"

	"gateway/pkg/excel"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0020/models"
	hub0021models "gateway/web/views/hub0021/models"
	hub0022models "gateway/web/views/hub0022/models"
	hubcommon002models "gateway/web/views/hubcommon002/models"

	"github.com/gin-gonic/gin"
)

// ImportGatewayInstance 从 Excel 文件导入网关实例配置
//
// # 导入策略（Upsert）
//
// 每张 Sheet 独立处理，逐行执行 Upsert（存在则更新，不存在则插入）：
//   - 先通过主键 ID + tenantId 查询记录是否存在。
//   - 存在则调用 UpdateXxx，不存在则调用 AddXxx / CreateXxx。
//   - 操作失败时记录 Warn 日志并跳过该行（容错模式），不回滚已处理的行。
//   - 返回值 result 以 "inserted" / "updated" 为键分别统计各类操作数量，
//     供前端展示导入摘要（如：新增 3 条、更新 5 条）。
//
// # Sheet 读取顺序
//
// 按数据依赖关系顺序处理，确保被引用表先于引用方写入：
//  1. GatewayInstance（主表，其他表均引用其 gatewayInstanceId）
//  2. LogConfig（被 GatewayInstance 和 RouteConfig 引用）
//  3. RouteConfig → RouteAssertion → FilterConfig（路由级/实例级）
//  4. RouterConfig → ServiceDefinition
//  5. SecurityConfig → IpAccessConfig / UaAccessConfig / DomainAccessConfig / ApiAccessConfig
//  6. CorsConfig → AuthConfig → RateLimitConfig
//
// @Summary 导入网关实例
// @Description 从 Excel 文件（exportGatewayInstance 生成的格式）批量导入网关实例及其关联配置
// @Tags 网关实例管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel 文件"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0020/importGatewayInstance [post]
func (c *GatewayInstanceController) ImportGatewayInstance(ctx *gin.Context) {
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		response.ErrorJSON(ctx, "读取上传文件失败: "+err.Error(), constants.ED00006)
		return
	}
	defer file.Close()

	sheets, err := excel.Parse(file)
	if err != nil {
		response.ErrorJSON(ctx, "解析 Excel 失败: "+err.Error(), constants.ED00009)
		return
	}

	// ── 调试信息：确认 Excel 中 Sheet 是否存在/行数是否正常 ─────────────────────
	// 重点关注实例与路由两张表；如果它们不存在，说明导出模板或 Sheet 名不匹配。
	logSheetRowCount := func(sheetName string) {
		if rows, ok := sheets[sheetName]; ok {
			// rows[0] 为表头行；len(rows) <= 1 表示只有表头没有数据
			logger.InfoWithTrace(ctx, "Excel Sheet 行数", "sheet", sheetName, "rowCount", len(rows))
		} else {
			logger.WarnWithTrace(ctx, "Excel Sheet 缺失", "sheet", sheetName)
		}
	}
	logSheetRowCount(models.GatewayInstance{}.TableName())
	logSheetRowCount(hub0021models.RouteConfig{}.TableName())
	logSheetRowCount(hub0021models.RouteAssertion{}.TableName())
	logSheetRowCount(hub0021models.FilterConfig{}.TableName() + "_route")
	logSheetRowCount(hub0021models.FilterConfig{}.TableName() + "_instance")
	logSheetRowCount(hub0021models.ServiceDefinition{}.TableName())

	// inserted / updated 分别统计新增和更新行数，key 为实体名称
	inserted := map[string]int{}
	updated := map[string]int{}

	// ── 1. GatewayInstance ─────────────────────────────────────────────────
	// 主表，其他所有配置均通过 gatewayInstanceId 关联，需最先写入。
	if rows, ok := sheets[models.GatewayInstance{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			inst := parseGatewayInstanceRow(row, idx, tenantId)
			if inst.GatewayInstanceId == "" {
				continue
			}
			existing, getErr := c.gatewayInstanceDAO.GetGatewayInstanceById(ctx, inst.GatewayInstanceId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询网关实例失败，跳过", "id", inst.GatewayInstanceId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.gatewayInstanceDAO.UpdateGatewayInstance(ctx, inst, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新网关实例失败，跳过", "id", inst.GatewayInstanceId, "error", upErr)
					continue
				}
				updated["gatewayInstance"]++
			} else {
				if _, addErr := c.gatewayInstanceDAO.AddGatewayInstance(ctx, inst, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增网关实例失败，跳过", "id", inst.GatewayInstanceId, "error", addErr)
					continue
				}
				inserted["gatewayInstance"]++
			}
		}
	}

	// ── 2. LogConfig ───────────────────────────────────────────────────────
	// 日志配置，被 GatewayInstance.logConfigId 和 RouteConfig.logConfigId 引用，
	// 需在 GatewayInstance 之后、RouteConfig 之前写入。
	if rows, ok := sheets[models.LogConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			lc := parseLogConfigRow(row, idx, tenantId)
			if lc.LogConfigId == "" {
				continue
			}
			existing, getErr := c.logConfigDAO.GetLogConfigById(ctx, lc.LogConfigId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询日志配置失败，跳过", "id", lc.LogConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.logConfigDAO.UpdateLogConfig(ctx, lc, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新日志配置失败，跳过", "id", lc.LogConfigId, "error", upErr)
					continue
				}
				updated["logConfig"]++
			} else {
				if _, addErr := c.logConfigDAO.AddLogConfig(ctx, lc, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增日志配置失败，跳过", "id", lc.LogConfigId, "error", addErr)
					continue
				}
				inserted["logConfig"]++
			}
		}
	}

	// ── 3. RouteConfig ─────────────────────────────────────────────────────
	// 路由配置，依赖 GatewayInstance（和可选的 ServiceDefinition）。
	if rows, ok := sheets[hub0021models.RouteConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		_, hasRouteId := idx["routeConfigId"]
		_, hasRouteIdLower := idx["routeconfigid"]
		logger.InfoWithTrace(ctx, "RouteConfig HeaderIndex",
			"sheet", hub0021models.RouteConfig{}.TableName(),
			"hasRouteConfigId", hasRouteId,
			"hasRouteconfigid", hasRouteIdLower,
			"rowCount", len(rows),
		)
		skippedEmpty := 0
		routeConfigSuccess := 0
		routeConfigFail := 0
		for _, row := range rows[1:] {
			rc := parseRouteConfigRow(row, idx, tenantId)
			if rc.RouteConfigId == "" {
				skippedEmpty++
				continue
			}
			// 仅打印前 2 条，避免刷屏
			if skippedEmpty == 0 {
				logger.InfoWithTrace(ctx, "RouteConfig 解析结果示例",
					"routeConfigId", rc.RouteConfigId,
					"gatewayInstanceId", rc.GatewayInstanceId,
					"routeName", rc.RouteName,
				)
			}
			existing, getErr := c.routeConfigDAO.GetRouteConfigById(ctx, rc.RouteConfigId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询路由配置失败，跳过", "id", rc.RouteConfigId, "error", getErr)
				routeConfigFail++
				continue
			}
			if existing != nil {
				if upErr := c.routeConfigDAO.UpdateRouteConfig(ctx, rc, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新路由配置失败，跳过", "id", rc.RouteConfigId, "error", upErr)
					routeConfigFail++
					continue
				}
				updated["routeConfig"]++
				routeConfigSuccess++
				logger.InfoWithTrace(ctx, "RouteConfig 更新成功",
					"routeConfigId", rc.RouteConfigId,
					"gatewayInstanceId", rc.GatewayInstanceId,
					"routeName", rc.RouteName,
				)
			} else {
				if _, addErr := c.routeConfigDAO.AddRouteConfig(ctx, rc, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增路由配置失败，跳过", "id", rc.RouteConfigId, "error", addErr)
					routeConfigFail++
					continue
				}
				inserted["routeConfig"]++
				routeConfigSuccess++
				logger.InfoWithTrace(ctx, "RouteConfig 新增成功",
					"routeConfigId", rc.RouteConfigId,
					"gatewayInstanceId", rc.GatewayInstanceId,
					"routeName", rc.RouteName,
				)
			}
		}
		logger.InfoWithTrace(ctx, "RouteConfig 导入汇总", "totalRows", len(rows)-1, "skippedEmpty", skippedEmpty, "success", routeConfigSuccess, "fail", routeConfigFail)
		if skippedEmpty > 0 {
			logger.InfoWithTrace(ctx, "RouteConfig 跳过空 ID 数量", "skippedEmpty", skippedEmpty)
		}
	}

	// ── 4. RouteAssertion ──────────────────────────────────────────────────
	// 路由断言，依赖 RouteConfig.routeConfigId，需在 RouteConfig 之后写入。
	if rows, ok := sheets[hub0021models.RouteAssertion{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			ra := parseRouteAssertionRow(row, idx, tenantId)
			if ra.RouteAssertionId == "" {
				continue
			}
			existing, getErr := c.routeAssertionDAO.GetRouteAssertionById(ctx, ra.RouteAssertionId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询路由断言失败，跳过", "id", ra.RouteAssertionId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.routeAssertionDAO.UpdateRouteAssertion(ctx, ra, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新路由断言失败，跳过", "id", ra.RouteAssertionId, "error", upErr)
					continue
				}
				updated["routeAssertion"]++
			} else {
				if _, addErr := c.routeAssertionDAO.AddRouteAssertion(ctx, ra, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增路由断言失败，跳过", "id", ra.RouteAssertionId, "error", addErr)
					continue
				}
				inserted["routeAssertion"]++
			}
		}
	}

	// ── 5. FilterConfig（路由级 + 实例级）─────────────────────────────────
	// 过滤器配置存储在同一张表，导出时按作用域拆为两个 Sheet（_route / _instance）。
	// 导入时两个 Sheet 均写入同一 DAO。
	for _, sheetName := range []string{
		hub0021models.FilterConfig{}.TableName() + "_route",
		hub0021models.FilterConfig{}.TableName() + "_instance",
	} {
		if rows, ok := sheets[sheetName]; ok && len(rows) > 1 {
			idx := excel.HeaderIndex(rows[0])
			for _, row := range rows[1:] {
				fc := parseFilterConfigRow(row, idx, tenantId)
				if fc.FilterConfigId == "" {
					continue
				}
				existing, getErr := c.filterConfigDAO.GetFilterConfigById(ctx, fc.FilterConfigId, tenantId)
				if getErr != nil {
					logger.WarnWithTrace(ctx, "查询过滤器配置失败，跳过", "sheet", sheetName, "id", fc.FilterConfigId, "error", getErr)
					continue
				}
				if existing != nil {
					if upErr := c.filterConfigDAO.UpdateFilterConfig(ctx, fc, operatorId); upErr != nil {
						logger.WarnWithTrace(ctx, "更新过滤器配置失败，跳过", "sheet", sheetName, "id", fc.FilterConfigId, "error", upErr)
						continue
					}
					updated["filterConfig"]++
				} else {
					if _, addErr := c.filterConfigDAO.AddFilterConfig(ctx, fc, operatorId); addErr != nil {
						logger.WarnWithTrace(ctx, "新增过滤器配置失败，跳过", "sheet", sheetName, "id", fc.FilterConfigId, "error", addErr)
						continue
					}
					inserted["filterConfig"]++
				}
			}
		}
	}

	// ── 6. RouterConfig ────────────────────────────────────────────────────
	// 路由器配置，依赖 GatewayInstance。
	if rows, ok := sheets[hub0021models.RouterConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			rc := parseRouterConfigRow(row, idx, tenantId)
			if rc.RouterConfigId == "" {
				continue
			}
			existing, getErr := c.routerConfigDAO.GetRouterConfigById(ctx, rc.RouterConfigId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询Router配置失败，跳过", "id", rc.RouterConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.routerConfigDAO.UpdateRouterConfig(ctx, rc, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新Router配置失败，跳过", "id", rc.RouterConfigId, "error", upErr)
					continue
				}
				updated["routerConfig"]++
			} else {
				if _, addErr := c.routerConfigDAO.AddRouterConfig(ctx, rc, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增Router配置失败，跳过", "id", rc.RouterConfigId, "error", addErr)
					continue
				}
				inserted["routerConfig"]++
			}
		}
	}

	// ── 7. ServiceDefinition（hub0022 DAO）────────────────────────────────
	// 服务定义，被 RouteConfig.serviceDefinitionId 引用。
	if rows, ok := sheets[hub0021models.ServiceDefinition{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			sd := parseServiceDefinitionRow(row, idx, tenantId)
			if sd.ServiceDefinitionId == "" {
				continue
			}
			existing, getErr := c.svcDefDAO.GetServiceDefinitionById(ctx, sd.ServiceDefinitionId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询服务定义失败，跳过", "id", sd.ServiceDefinitionId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.svcDefDAO.UpdateServiceDefinition(ctx, sd, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新服务定义失败，跳过", "id", sd.ServiceDefinitionId, "error", upErr)
					continue
				}
				updated["serviceDefinition"]++
			} else {
				if _, addErr := c.svcDefDAO.CreateServiceDefinition(ctx, sd, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增服务定义失败，跳过", "id", sd.ServiceDefinitionId, "error", addErr)
					continue
				}
				inserted["serviceDefinition"]++
			}
		}
	}

	// ── 8. ProxyConfig（hub0022）──────────────────────────────────────────
	// 代理配置，被 ServiceDefinition.proxyConfigId 引用，需在 ServiceDefinition 之后写入。
	if rows, ok := sheets[hub0022models.ProxyConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			pc := parseProxyConfigRow(row, idx, tenantId)
			if pc.ProxyConfigId == "" {
				continue
			}
			existing, getErr := c.proxyConfigDAO.GetProxyConfigById(ctx, pc.ProxyConfigId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询代理配置失败，跳过", "id", pc.ProxyConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.proxyConfigDAO.UpdateProxyConfig(ctx, pc, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新代理配置失败，跳过", "id", pc.ProxyConfigId, "error", upErr)
					continue
				}
				updated["proxyConfig"]++
			} else {
				if _, addErr := c.proxyConfigDAO.CreateProxyConfig(ctx, pc, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增代理配置失败，跳过", "id", pc.ProxyConfigId, "error", addErr)
					continue
				}
				inserted["proxyConfig"]++
			}
		}
	}

	// ── 9. ServiceNode（hub0022）──────────────────────────────────────────
	// 服务节点，依赖 ServiceDefinition.serviceDefinitionId，需在 ServiceDefinition 之后写入。
	if rows, ok := sheets["HUB_GW_SERVICE_NODE"]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			sn := parseServiceNodeRow(row, idx, tenantId)
			if sn.ServiceNodeId == "" {
				continue
			}
			existing, getErr := c.serviceNodeDAO.GetServiceNodeById(ctx, sn.ServiceNodeId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询服务节点失败，跳过", "id", sn.ServiceNodeId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.serviceNodeDAO.UpdateServiceNode(ctx, sn, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新服务节点失败，跳过", "id", sn.ServiceNodeId, "error", upErr)
					continue
				}
				updated["serviceNode"]++
			} else {
				if _, addErr := c.serviceNodeDAO.CreateServiceNode(ctx, sn, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增服务节点失败，跳过", "id", sn.ServiceNodeId, "error", addErr)
					continue
				}
				inserted["serviceNode"]++
			}
		}
	}

	// ── 10. SecurityConfig ─────────────────────────────────────────────────
	// 安全配置，依赖 GatewayInstance 和 RouteConfig（可选）。
	// IpAccess / UA / Domain / ApiAccess 均依赖 SecurityConfig，必须在这些子配置之前写入。
	if rows, ok := sheets[hubcommon002models.SecurityConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			sc := parseSecurityConfigRow(row, idx, tenantId)
			if sc.SecurityConfigId == "" {
				continue
			}
			existing, getErr := c.securityConfigDAO.GetSecurityConfigById(ctx, sc.SecurityConfigId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询安全配置失败，跳过", "id", sc.SecurityConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.securityConfigDAO.UpdateSecurityConfig(ctx, sc, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新安全配置失败，跳过", "id", sc.SecurityConfigId, "error", upErr)
					continue
				}
				updated["securityConfig"]++
			} else {
				if _, addErr := c.securityConfigDAO.AddSecurityConfig(ctx, sc, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增安全配置失败，跳过", "id", sc.SecurityConfigId, "error", addErr)
					continue
				}
				inserted["securityConfig"]++
			}
		}
	}

	// ── 9. IpAccessConfig ──────────────────────────────────────────────────
	// IP 访问控制，依赖 SecurityConfig.securityConfigId。
	if rows, ok := sheets[hubcommon002models.IpAccessConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			ip := parseIpAccessConfigRow(row, idx, tenantId)
			if ip.IpAccessConfigId == "" {
				continue
			}
			existing, getErr := c.ipAccessConfigDAO.GetIpAccessConfigById(ctx, ip.IpAccessConfigId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询IP访问配置失败，跳过", "id", ip.IpAccessConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.ipAccessConfigDAO.UpdateIpAccessConfig(ctx, ip, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新IP访问配置失败，跳过", "id", ip.IpAccessConfigId, "error", upErr)
					continue
				}
				updated["ipAccessConfig"]++
			} else {
				if addErr := c.ipAccessConfigDAO.AddIpAccessConfig(ctx, ip, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增IP访问配置失败，跳过", "id", ip.IpAccessConfigId, "error", addErr)
					continue
				}
				inserted["ipAccessConfig"]++
			}
		}
	}

	// ── 10. UaAccessConfig ─────────────────────────────────────────────────
	// User-Agent 访问控制，依赖 SecurityConfig.securityConfigId。
	if rows, ok := sheets[hubcommon002models.UseragentAccessConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			ua := parseUaAccessConfigRow(row, idx, tenantId)
			if ua.UseragentAccessConfigId == "" {
				continue
			}
			existing, getErr := c.uaAccessConfigDAO.GetUseragentAccessConfigById(ctx, ua.UseragentAccessConfigId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询UA访问配置失败，跳过", "id", ua.UseragentAccessConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.uaAccessConfigDAO.UpdateUseragentAccessConfig(ctx, ua, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新UA访问配置失败，跳过", "id", ua.UseragentAccessConfigId, "error", upErr)
					continue
				}
				updated["uaAccessConfig"]++
			} else {
				if addErr := c.uaAccessConfigDAO.AddUseragentAccessConfig(ctx, ua, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增UA访问配置失败，跳过", "id", ua.UseragentAccessConfigId, "error", addErr)
					continue
				}
				inserted["uaAccessConfig"]++
			}
		}
	}

	// ── 11. DomainAccessConfig ─────────────────────────────────────────────
	// 域名访问控制，依赖 SecurityConfig.securityConfigId。
	if rows, ok := sheets[hubcommon002models.DomainAccessConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			d := parseDomainAccessConfigRow(row, idx, tenantId)
			if d.DomainAccessConfigId == "" {
				continue
			}
			existing, getErr := c.domainAccessConfigDAO.GetDomainAccessConfigById(ctx, d.DomainAccessConfigId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询域名访问配置失败，跳过", "id", d.DomainAccessConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.domainAccessConfigDAO.UpdateDomainAccessConfig(ctx, d, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新域名访问配置失败，跳过", "id", d.DomainAccessConfigId, "error", upErr)
					continue
				}
				updated["domainAccessConfig"]++
			} else {
				if addErr := c.domainAccessConfigDAO.AddDomainAccessConfig(ctx, d, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增域名访问配置失败，跳过", "id", d.DomainAccessConfigId, "error", addErr)
					continue
				}
				inserted["domainAccessConfig"]++
			}
		}
	}

	// ── 12. ApiAccessConfig ────────────────────────────────────────────────
	// API 路径访问控制，依赖 SecurityConfig.securityConfigId。
	if rows, ok := sheets[hubcommon002models.ApiAccessConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			a := parseApiAccessConfigRow(row, idx, tenantId)
			if a.ApiAccessConfigId == "" {
				continue
			}
			existing, getErr := c.apiAccessConfigDAO.GetApiAccessConfigById(ctx, a.ApiAccessConfigId, tenantId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询API访问配置失败，跳过", "id", a.ApiAccessConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.apiAccessConfigDAO.UpdateApiAccessConfig(ctx, a, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新API访问配置失败，跳过", "id", a.ApiAccessConfigId, "error", upErr)
					continue
				}
				updated["apiAccessConfig"]++
			} else {
				if addErr := c.apiAccessConfigDAO.AddApiAccessConfig(ctx, a, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增API访问配置失败，跳过", "id", a.ApiAccessConfigId, "error", addErr)
					continue
				}
				inserted["apiAccessConfig"]++
			}
		}
	}

	// ── 13. CorsConfig ─────────────────────────────────────────────────────
	// CORS 跨域配置，依赖 GatewayInstance 和 RouteConfig（可选）。
	if rows, ok := sheets[hubcommon002models.CorsConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			cc := parseCorsConfigRow(row, idx, tenantId)
			if cc.CorsConfigId == "" {
				continue
			}
			existing, getErr := c.corsConfigDAO.GetCorsConfig(tenantId, cc.CorsConfigId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询CORS配置失败，跳过", "id", cc.CorsConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.corsConfigDAO.UpdateCorsConfig(ctx, cc, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新CORS配置失败，跳过", "id", cc.CorsConfigId, "error", upErr)
					continue
				}
				updated["corsConfig"]++
			} else {
				if addErr := c.corsConfigDAO.AddCorsConfig(ctx, cc, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增CORS配置失败，跳过", "id", cc.CorsConfigId, "error", addErr)
					continue
				}
				inserted["corsConfig"]++
			}
		}
	}

	// ── 14. AuthConfig ─────────────────────────────────────────────────────
	// 认证配置，依赖 GatewayInstance 和 RouteConfig（可选）。
	if rows, ok := sheets[hubcommon002models.AuthConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			ac := parseAuthConfigRow(row, idx, tenantId)
			if ac.AuthConfigId == "" {
				continue
			}
			existing, getErr := c.authConfigDAO.GetAuthConfig(tenantId, ac.AuthConfigId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询认证配置失败，跳过", "id", ac.AuthConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.authConfigDAO.UpdateAuthConfig(ctx, ac, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新认证配置失败，跳过", "id", ac.AuthConfigId, "error", upErr)
					continue
				}
				updated["authConfig"]++
			} else {
				if addErr := c.authConfigDAO.AddAuthConfig(ctx, ac, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增认证配置失败，跳过", "id", ac.AuthConfigId, "error", addErr)
					continue
				}
				inserted["authConfig"]++
			}
		}
	}

	// ── 15. RateLimitConfig ────────────────────────────────────────────────
	// 限流配置，依赖 GatewayInstance 和 RouteConfig（可选）。
	if rows, ok := sheets[hubcommon002models.RateLimitConfig{}.TableName()]; ok && len(rows) > 1 {
		idx := excel.HeaderIndex(rows[0])
		for _, row := range rows[1:] {
			rl := parseRateLimitConfigRow(row, idx, tenantId)
			if rl.RateLimitConfigId == "" {
				continue
			}
			existing, getErr := c.rateLimitConfigDAO.GetRateLimitConfig(tenantId, rl.RateLimitConfigId)
			if getErr != nil {
				logger.WarnWithTrace(ctx, "查询限流配置失败，跳过", "id", rl.RateLimitConfigId, "error", getErr)
				continue
			}
			if existing != nil {
				if upErr := c.rateLimitConfigDAO.UpdateRateLimitConfig(ctx, rl, operatorId); upErr != nil {
					logger.WarnWithTrace(ctx, "更新限流配置失败，跳过", "id", rl.RateLimitConfigId, "error", upErr)
					continue
				}
				updated["rateLimitConfig"]++
			} else {
				if addErr := c.rateLimitConfigDAO.AddRateLimitConfig(ctx, rl, operatorId); addErr != nil {
					logger.WarnWithTrace(ctx, "新增限流配置失败，跳过", "id", rl.RateLimitConfigId, "error", addErr)
					continue
				}
				inserted["rateLimitConfig"]++
			}
		}
	}

	logger.InfoWithTrace(ctx, "导入统计结果", "inserted", inserted, "updated", updated)
	response.SuccessJSON(ctx, map[string]any{"inserted": inserted, "updated": updated}, constants.SD00002)
}

// ─── 行解析辅助函数 ────────────────────────────────────────────────────────

func parseGatewayInstanceRow(row []string, idx map[string]int, tenantId string) *models.GatewayInstance {
	g := &models.GatewayInstance{
		TenantId:                     tenantId,
		GatewayInstanceId:            excel.GetCell(row, idx, "gatewayInstanceId"),
		InstanceName:                 excel.GetCell(row, idx, "instanceName"),
		InstanceDesc:                 excel.GetCell(row, idx, "instanceDesc"),
		BindAddress:                  excel.GetCell(row, idx, "bindAddress"),
		TlsEnabled:                   excel.GetCell(row, idx, "tlsEnabled"),
		CertStorageType:              excel.GetCell(row, idx, "certStorageType"),
		CertFilePath:                 excel.GetCell(row, idx, "certFilePath"),
		KeyFilePath:                  excel.GetCell(row, idx, "keyFilePath"),
		CertContent:                  excel.GetCell(row, idx, "certContent"),
		KeyContent:                   excel.GetCell(row, idx, "keyContent"),
		CertChainContent:             excel.GetCell(row, idx, "certChainContent"),
		CertPassword:                 excel.GetCell(row, idx, "certPassword"),
		KeepAliveEnabled:             excel.GetCell(row, idx, "keepAliveEnabled"),
		TcpKeepAliveEnabled:          excel.GetCell(row, idx, "tcpKeepAliveEnabled"),
		EnableHttp2:                  excel.GetCell(row, idx, "enableHttp2"),
		TlsVersion:                   excel.GetCell(row, idx, "tlsVersion"),
		TlsCipherSuites:              excel.GetCell(row, idx, "tlsCipherSuites"),
		DisableGeneralOptionsHandler: excel.GetCell(row, idx, "disableGeneralOptionsHandler"),
		LogConfigId:                  excel.GetCell(row, idx, "logConfigId"),
		HealthStatus:                 excel.GetCell(row, idx, "healthStatus"),
		InstanceMetadata:             excel.GetCell(row, idx, "instanceMetadata"),
		Reserved1:                    excel.GetCell(row, idx, "reserved1"),
		Reserved2:                    excel.GetCell(row, idx, "reserved2"),
		ExtProperty:                  excel.GetCell(row, idx, "extProperty"),
		ActiveFlag:                   strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
		NoteText:                     excel.GetCell(row, idx, "noteText"),
	}
	g.HttpPort = atoiPtr(excel.GetCell(row, idx, "httpPort"))
	g.HttpsPort = atoiPtr(excel.GetCell(row, idx, "httpsPort"))
	g.MaxConnections = atoiSafe(excel.GetCell(row, idx, "maxConnections"))
	g.ReadTimeoutMs = atoiSafe(excel.GetCell(row, idx, "readTimeoutMs"))
	g.WriteTimeoutMs = atoiSafe(excel.GetCell(row, idx, "writeTimeoutMs"))
	g.IdleTimeoutMs = atoiSafe(excel.GetCell(row, idx, "idleTimeoutMs"))
	g.MaxHeaderBytes = atoiSafe(excel.GetCell(row, idx, "maxHeaderBytes"))
	g.MaxWorkers = atoiSafe(excel.GetCell(row, idx, "maxWorkers"))
	g.GracefulShutdownTimeoutMs = atoiSafe(excel.GetCell(row, idx, "gracefulShutdownTimeoutMs"))
	g.Reserved3 = atoiPtr(excel.GetCell(row, idx, "reserved3"))
	g.Reserved4 = atoiPtr(excel.GetCell(row, idx, "reserved4"))
	return g
}

func parseLogConfigRow(row []string, idx map[string]int, tenantId string) *models.LogConfig {
	lc := &models.LogConfig{
		TenantId:                   tenantId,
		LogConfigId:                excel.GetCell(row, idx, "logConfigId"),
		ConfigName:                 excel.GetCell(row, idx, "configName"),
		ConfigDesc:                 excel.GetCell(row, idx, "configDesc"),
		LogFormat:                  excel.GetCell(row, idx, "logFormat"),
		RecordRequestBody:          excel.GetCell(row, idx, "recordRequestBody"),
		RecordResponseBody:         excel.GetCell(row, idx, "recordResponseBody"),
		RecordHeaders:              excel.GetCell(row, idx, "recordHeaders"),
		OutputTargets:              excel.GetCell(row, idx, "outputTargets"),
		FileConfig:                 excel.GetCell(row, idx, "fileConfig"),
		DatabaseConfig:             excel.GetCell(row, idx, "databaseConfig"),
		MongoConfig:                excel.GetCell(row, idx, "mongoConfig"),
		ElasticsearchConfig:        excel.GetCell(row, idx, "elasticsearchConfig"),
		ClickhouseConfig:           excel.GetCell(row, idx, "clickhouseConfig"),
		EnableAsyncLogging:         excel.GetCell(row, idx, "enableAsyncLogging"),
		EnableBatchProcessing:      excel.GetCell(row, idx, "enableBatchProcessing"),
		RotationPattern:            excel.GetCell(row, idx, "rotationPattern"),
		EnableFileRotation:         excel.GetCell(row, idx, "enableFileRotation"),
		EnableSensitiveDataMasking: excel.GetCell(row, idx, "enableSensitiveDataMasking"),
		SensitiveFields:            excel.GetCell(row, idx, "sensitiveFields"),
		MaskingPattern:             excel.GetCell(row, idx, "maskingPattern"),
		Reserved1:                  excel.GetCell(row, idx, "reserved1"),
		Reserved2:                  excel.GetCell(row, idx, "reserved2"),
		ExtProperty:                excel.GetCell(row, idx, "extProperty"),
		ActiveFlag:                 strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
		NoteText:                   excel.GetCell(row, idx, "noteText"),
	}
	lc.MaxBodySizeBytes = atoiSafe(excel.GetCell(row, idx, "maxBodySizeBytes"))
	lc.AsyncQueueSize = atoiSafe(excel.GetCell(row, idx, "asyncQueueSize"))
	lc.AsyncFlushIntervalMs = atoiSafe(excel.GetCell(row, idx, "asyncFlushIntervalMs"))
	lc.BatchSize = atoiSafe(excel.GetCell(row, idx, "batchSize"))
	lc.BatchTimeoutMs = atoiSafe(excel.GetCell(row, idx, "batchTimeoutMs"))
	lc.LogRetentionDays = atoiSafe(excel.GetCell(row, idx, "logRetentionDays"))
	lc.ConfigPriority = atoiSafe(excel.GetCell(row, idx, "configPriority"))
	lc.BufferSize = atoiSafe(excel.GetCell(row, idx, "bufferSize"))
	lc.FlushThreshold = atoiSafe(excel.GetCell(row, idx, "flushThreshold"))
	lc.MaxFileSizeMB = atoiPtr(excel.GetCell(row, idx, "maxFileSizeMB"))
	lc.MaxFileCount = atoiPtr(excel.GetCell(row, idx, "maxFileCount"))
	lc.Reserved3 = atoiPtr(excel.GetCell(row, idx, "reserved3"))
	lc.Reserved4 = atoiPtr(excel.GetCell(row, idx, "reserved4"))
	return lc
}

func parseRouteConfigRow(row []string, idx map[string]int, tenantId string) *hub0021models.RouteConfig {
	rc := &hub0021models.RouteConfig{
		TenantId:            tenantId,
		RouteConfigId:       excel.GetCell(row, idx, "routeConfigId"),
		GatewayInstanceId:   excel.GetCell(row, idx, "gatewayInstanceId"),
		RouteName:           excel.GetCell(row, idx, "routeName"),
		RoutePath:           excel.GetCell(row, idx, "routePath"),
		AllowedMethods:      excel.GetCell(row, idx, "allowedMethods"),
		AllowedHosts:        excel.GetCell(row, idx, "allowedHosts"),
		StripPathPrefix:     excel.GetCell(row, idx, "stripPathPrefix"),
		RewritePath:         excel.GetCell(row, idx, "rewritePath"),
		EnableWebsocket:     excel.GetCell(row, idx, "enableWebsocket"),
		ServiceDefinitionId: excel.GetCell(row, idx, "serviceDefinitionId"),
		LogConfigId:         excel.GetCell(row, idx, "logConfigId"),
		RouteMetadata:       excel.GetCell(row, idx, "routeMetadata"),
		ActiveFlag:          strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
		NoteText:            excel.GetCell(row, idx, "noteText"),
	}
	rc.MatchType = atoiSafe(excel.GetCell(row, idx, "matchType"))
	rc.RoutePriority = atoiSafe(excel.GetCell(row, idx, "routePriority"))
	rc.TimeoutMs = atoiSafe(excel.GetCell(row, idx, "timeoutMs"))
	rc.RetryCount = atoiSafe(excel.GetCell(row, idx, "retryCount"))
	rc.RetryIntervalMs = atoiSafe(excel.GetCell(row, idx, "retryIntervalMs"))
	return rc
}

func parseRouteAssertionRow(row []string, idx map[string]int, tenantId string) *hub0021models.RouteAssertion {
	ra := &hub0021models.RouteAssertion{
		TenantId:          tenantId,
		RouteAssertionId:  excel.GetCell(row, idx, "routeAssertionId"),
		RouteConfigId:     excel.GetCell(row, idx, "routeConfigId"),
		AssertionName:     excel.GetCell(row, idx, "assertionName"),
		AssertionType:     excel.GetCell(row, idx, "assertionType"),
		AssertionOperator: excel.GetCell(row, idx, "assertionOperator"),
		FieldName:         excel.GetCell(row, idx, "fieldName"),
		ExpectedValue:     excel.GetCell(row, idx, "expectedValue"),
		PatternValue:      excel.GetCell(row, idx, "patternValue"),
		CaseSensitive:     excel.GetCell(row, idx, "caseSensitive"),
		IsRequired:        excel.GetCell(row, idx, "isRequired"),
		AssertionDesc:     excel.GetCell(row, idx, "assertionDesc"),
		ActiveFlag:        strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
		NoteText:          excel.GetCell(row, idx, "noteText"),
	}
	ra.AssertionOrder = atoiSafe(excel.GetCell(row, idx, "assertionOrder"))
	return ra
}

func parseFilterConfigRow(row []string, idx map[string]int, tenantId string) *hub0021models.FilterConfig {
	return &hub0021models.FilterConfig{
		TenantId:          tenantId,
		FilterConfigId:    excel.GetCell(row, idx, "filterConfigId"),
		GatewayInstanceId: excel.GetCell(row, idx, "gatewayInstanceId"),
		RouteConfigId:     excel.GetCell(row, idx, "routeConfigId"),
		FilterName:        excel.GetCell(row, idx, "filterName"),
		FilterType:        excel.GetCell(row, idx, "filterType"),
		FilterAction:      excel.GetCell(row, idx, "filterAction"),
		FilterOrder:       atoiSafe(excel.GetCell(row, idx, "filterOrder")),
		FilterConfig:      excel.GetCell(row, idx, "filterConfig"),
		FilterDesc:        excel.GetCell(row, idx, "filterDesc"),
		ActiveFlag:        strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
		NoteText:          excel.GetCell(row, idx, "noteText"),
	}
}

func parseRouterConfigRow(row []string, idx map[string]int, tenantId string) *hub0021models.RouterConfig {
	rc := &hub0021models.RouterConfig{
		TenantId:              tenantId,
		RouterConfigId:        excel.GetCell(row, idx, "routerConfigId"),
		GatewayInstanceId:     excel.GetCell(row, idx, "gatewayInstanceId"),
		RouterName:            excel.GetCell(row, idx, "routerName"),
		RouterDesc:            excel.GetCell(row, idx, "routerDesc"),
		EnableRouteCache:      excel.GetCell(row, idx, "enableRouteCache"),
		EnableTracing:         excel.GetCell(row, idx, "enableTracing"),
		CaseSensitive:         excel.GetCell(row, idx, "caseSensitive"),
		RemoveTrailingSlash:   excel.GetCell(row, idx, "removeTrailingSlash"),
		EnableGlobalFilters:   excel.GetCell(row, idx, "enableGlobalFilters"),
		FilterExecutionMode:   excel.GetCell(row, idx, "filterExecutionMode"),
		EnableRoutePooling:    excel.GetCell(row, idx, "enableRoutePooling"),
		EnableAsyncProcessing: excel.GetCell(row, idx, "enableAsyncProcessing"),
		EnableFallback:        excel.GetCell(row, idx, "enableFallback"),
		FallbackRoute:         excel.GetCell(row, idx, "fallbackRoute"),
		NotFoundMessage:       excel.GetCell(row, idx, "notFoundMessage"),
		ActiveFlag:            strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
		NoteText:              excel.GetCell(row, idx, "noteText"),
	}
	rc.DefaultPriority = atoiSafe(excel.GetCell(row, idx, "defaultPriority"))
	rc.RouteCacheTtlSeconds = atoiSafe(excel.GetCell(row, idx, "routeCacheTtlSeconds"))
	rc.NotFoundStatusCode = atoiSafe(excel.GetCell(row, idx, "notFoundStatusCode"))
	return rc
}

func parseServiceDefinitionRow(row []string, idx map[string]int, tenantId string) *hub0022models.ServiceDefinition {
	sd := &hub0022models.ServiceDefinition{
		TenantId:             tenantId,
		ServiceDefinitionId:  excel.GetCell(row, idx, "serviceDefinitionId"),
		ServiceName:          excel.GetCell(row, idx, "serviceName"),
		ServiceDesc:          excel.GetCell(row, idx, "serviceDesc"),
		ProxyConfigId:        excel.GetCell(row, idx, "proxyConfigId"),
		LoadBalanceStrategy:  excel.GetCell(row, idx, "loadBalanceStrategy"),
		DiscoveryType:        excel.GetCell(row, idx, "discoveryType"),
		DiscoveryConfig:      excel.GetCell(row, idx, "discoveryConfig"),
		SessionAffinity:      excel.GetCell(row, idx, "sessionAffinity"),
		StickySession:        excel.GetCell(row, idx, "stickySession"),
		EnableCircuitBreaker: excel.GetCell(row, idx, "enableCircuitBreaker"),
		HealthCheckEnabled:   excel.GetCell(row, idx, "healthCheckEnabled"),
		HealthCheckPath:      excel.GetCell(row, idx, "healthCheckPath"),
		HealthCheckMethod:    excel.GetCell(row, idx, "healthCheckMethod"),
		ExpectedStatusCodes:  excel.GetCell(row, idx, "expectedStatusCodes"),
		HealthCheckHeaders:   excel.GetCell(row, idx, "healthCheckHeaders"),
		LoadBalancerConfig:   excel.GetCell(row, idx, "loadBalancerConfig"),
		ServiceMetadata:      excel.GetCell(row, idx, "serviceMetadata"),
		ActiveFlag:           strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
		NoteText:             excel.GetCell(row, idx, "noteText"),
	}
	sd.ServiceType = atoiSafe(excel.GetCell(row, idx, "serviceType"))
	sd.MaxRetries = atoiSafe(excel.GetCell(row, idx, "maxRetries"))
	sd.RetryTimeoutMs = atoiSafe(excel.GetCell(row, idx, "retryTimeoutMs"))
	if v := atoiSafe(excel.GetCell(row, idx, "healthCheckIntervalSeconds")); v != 0 {
		sd.HealthCheckIntervalSeconds = &v
	}
	if v := atoiSafe(excel.GetCell(row, idx, "healthCheckTimeoutMs")); v != 0 {
		sd.HealthCheckTimeoutMs = &v
	}
	if v := atoiSafe(excel.GetCell(row, idx, "healthyThreshold")); v != 0 {
		sd.HealthyThreshold = &v
	}
	if v := atoiSafe(excel.GetCell(row, idx, "unhealthyThreshold")); v != 0 {
		sd.UnhealthyThreshold = &v
	}
	return sd
}

// ─── Security 行解析 ───────────────────────────────────────────────────────

func parseSecurityConfigRow(row []string, idx map[string]int, tenantId string) *hubcommon002models.SecurityConfig {
	gwId := excel.GetCell(row, idx, "gatewayInstanceId")
	routeId := excel.GetCell(row, idx, "routeConfigId")
	configDesc := excel.GetCell(row, idx, "configDesc")
	customCfg := excel.GetCell(row, idx, "customConfigJson")
	noteText := excel.GetCell(row, idx, "noteText")
	sc := &hubcommon002models.SecurityConfig{
		TenantId:         tenantId,
		SecurityConfigId: excel.GetCell(row, idx, "securityConfigId"),
		ConfigName:       excel.GetCell(row, idx, "configName"),
		ConfigPriority:   atoiSafe(excel.GetCell(row, idx, "configPriority")),
		ActiveFlag:       strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
	}
	if gwId != "" {
		sc.GatewayInstanceId = &gwId
	}
	if routeId != "" {
		sc.RouteConfigId = &routeId
	}
	if configDesc != "" {
		sc.ConfigDesc = &configDesc
	}
	if customCfg != "" {
		sc.CustomConfigJson = &customCfg
	}
	if noteText != "" {
		sc.NoteText = &noteText
	}
	return sc
}

func parseIpAccessConfigRow(row []string, idx map[string]int, tenantId string) *hubcommon002models.IpAccessConfig {
	ip := &hubcommon002models.IpAccessConfig{
		TenantId:           tenantId,
		IpAccessConfigId:   excel.GetCell(row, idx, "ipAccessConfigId"),
		SecurityConfigId:   excel.GetCell(row, idx, "securityConfigId"),
		ConfigName:         excel.GetCell(row, idx, "configName"),
		DefaultPolicy:      strOrDefault(excel.GetCell(row, idx, "defaultPolicy"), "allow"),
		TrustXForwardedFor: strOrDefault(excel.GetCell(row, idx, "trustXForwardedFor"), "N"),
		TrustXRealIp:       strOrDefault(excel.GetCell(row, idx, "trustXRealIp"), "N"),
		ActiveFlag:         strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
	}
	if v := excel.GetCell(row, idx, "whitelistIps"); v != "" {
		ip.WhitelistIps = &v
	}
	if v := excel.GetCell(row, idx, "blacklistIps"); v != "" {
		ip.BlacklistIps = &v
	}
	if v := excel.GetCell(row, idx, "whitelistCidrs"); v != "" {
		ip.WhitelistCidrs = &v
	}
	if v := excel.GetCell(row, idx, "blacklistCidrs"); v != "" {
		ip.BlacklistCidrs = &v
	}
	if v := excel.GetCell(row, idx, "noteText"); v != "" {
		ip.NoteText = &v
	}
	return ip
}

func parseUaAccessConfigRow(row []string, idx map[string]int, tenantId string) *hubcommon002models.UseragentAccessConfig {
	ua := &hubcommon002models.UseragentAccessConfig{
		TenantId:                tenantId,
		UseragentAccessConfigId: excel.GetCell(row, idx, "useragentAccessConfigId"),
		SecurityConfigId:        excel.GetCell(row, idx, "securityConfigId"),
		ConfigName:              excel.GetCell(row, idx, "configName"),
		DefaultPolicy:           strOrDefault(excel.GetCell(row, idx, "defaultPolicy"), "allow"),
		BlockEmptyUserAgent:     strOrDefault(excel.GetCell(row, idx, "blockEmptyUserAgent"), "N"),
		ActiveFlag:              strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
	}
	if v := excel.GetCell(row, idx, "whitelistPatterns"); v != "" {
		ua.WhitelistPatterns = &v
	}
	if v := excel.GetCell(row, idx, "blacklistPatterns"); v != "" {
		ua.BlacklistPatterns = &v
	}
	if v := excel.GetCell(row, idx, "noteText"); v != "" {
		ua.NoteText = &v
	}
	return ua
}

func parseDomainAccessConfigRow(row []string, idx map[string]int, tenantId string) *hubcommon002models.DomainAccessConfig {
	d := &hubcommon002models.DomainAccessConfig{
		TenantId:             tenantId,
		DomainAccessConfigId: excel.GetCell(row, idx, "domainAccessConfigId"),
		SecurityConfigId:     excel.GetCell(row, idx, "securityConfigId"),
		ConfigName:           excel.GetCell(row, idx, "configName"),
		DefaultPolicy:        strOrDefault(excel.GetCell(row, idx, "defaultPolicy"), "allow"),
		AllowSubdomains:      strOrDefault(excel.GetCell(row, idx, "allowSubdomains"), "N"),
		ActiveFlag:           strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
	}
	if v := excel.GetCell(row, idx, "whitelistDomains"); v != "" {
		d.WhitelistDomains = &v
	}
	if v := excel.GetCell(row, idx, "blacklistDomains"); v != "" {
		d.BlacklistDomains = &v
	}
	if v := excel.GetCell(row, idx, "noteText"); v != "" {
		d.NoteText = &v
	}
	return d
}

func parseApiAccessConfigRow(row []string, idx map[string]int, tenantId string) *hubcommon002models.ApiAccessConfig {
	a := &hubcommon002models.ApiAccessConfig{
		TenantId:          tenantId,
		ApiAccessConfigId: excel.GetCell(row, idx, "apiAccessConfigId"),
		SecurityConfigId:  excel.GetCell(row, idx, "securityConfigId"),
		ConfigName:        excel.GetCell(row, idx, "configName"),
		DefaultPolicy:     strOrDefault(excel.GetCell(row, idx, "defaultPolicy"), "allow"),
		ActiveFlag:        strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
	}
	if v := excel.GetCell(row, idx, "whitelistPaths"); v != "" {
		a.WhitelistPaths = &v
	}
	if v := excel.GetCell(row, idx, "blacklistPaths"); v != "" {
		a.BlacklistPaths = &v
	}
	if v := excel.GetCell(row, idx, "allowedMethods"); v != "" {
		a.AllowedMethods = &v
	}
	if v := excel.GetCell(row, idx, "blockedMethods"); v != "" {
		a.BlockedMethods = &v
	}
	if v := excel.GetCell(row, idx, "noteText"); v != "" {
		a.NoteText = &v
	}
	return a
}

func parseCorsConfigRow(row []string, idx map[string]int, tenantId string) *hubcommon002models.CorsConfig {
	gwId := excel.GetCell(row, idx, "gatewayInstanceId")
	routeId := excel.GetCell(row, idx, "routeConfigId")
	cc := &hubcommon002models.CorsConfig{
		TenantId:         tenantId,
		CorsConfigId:     excel.GetCell(row, idx, "corsConfigId"),
		ConfigName:       excel.GetCell(row, idx, "configName"),
		AllowMethods:     excel.GetCell(row, idx, "allowMethods"),
		AllowCredentials: strOrDefault(excel.GetCell(row, idx, "allowCredentials"), "N"),
		MaxAgeSeconds:    atoiSafe(excel.GetCell(row, idx, "maxAgeSeconds")),
		ConfigPriority:   atoiSafe(excel.GetCell(row, idx, "configPriority")),
		ActiveFlag:       strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
	}
	if gwId != "" {
		cc.GatewayInstanceId = &gwId
	}
	if routeId != "" {
		cc.RouteConfigId = &routeId
	}
	if v := excel.GetCell(row, idx, "allowOrigins"); v != "" {
		cc.AllowOrigins = &v
	}
	if v := excel.GetCell(row, idx, "allowHeaders"); v != "" {
		cc.AllowHeaders = &v
	}
	if v := excel.GetCell(row, idx, "exposeHeaders"); v != "" {
		cc.ExposeHeaders = &v
	}
	if v := excel.GetCell(row, idx, "noteText"); v != "" {
		cc.NoteText = &v
	}
	return cc
}

func parseAuthConfigRow(row []string, idx map[string]int, tenantId string) *hubcommon002models.AuthConfig {
	gwId := excel.GetCell(row, idx, "gatewayInstanceId")
	routeId := excel.GetCell(row, idx, "routeConfigId")
	ac := &hubcommon002models.AuthConfig{
		TenantId:          tenantId,
		AuthConfigId:      excel.GetCell(row, idx, "authConfigId"),
		AuthName:          excel.GetCell(row, idx, "authName"),
		AuthType:          excel.GetCell(row, idx, "authType"),
		AuthStrategy:      excel.GetCell(row, idx, "authStrategy"),
		AuthConfig:        excel.GetCell(row, idx, "authConfig"),
		FailureStatusCode: atoiSafe(excel.GetCell(row, idx, "failureStatusCode")),
		FailureMessage:    excel.GetCell(row, idx, "failureMessage"),
		ConfigPriority:    atoiSafe(excel.GetCell(row, idx, "configPriority")),
		ActiveFlag:        strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
	}
	if gwId != "" {
		ac.GatewayInstanceId = &gwId
	}
	if routeId != "" {
		ac.RouteConfigId = &routeId
	}
	if v := excel.GetCell(row, idx, "exemptPaths"); v != "" {
		ac.ExemptPaths = &v
	}
	if v := excel.GetCell(row, idx, "exemptHeaders"); v != "" {
		ac.ExemptHeaders = &v
	}
	if v := excel.GetCell(row, idx, "noteText"); v != "" {
		ac.NoteText = &v
	}
	return ac
}

func parseRateLimitConfigRow(row []string, idx map[string]int, tenantId string) *hubcommon002models.RateLimitConfig {
	gwId := excel.GetCell(row, idx, "gatewayInstanceId")
	routeId := excel.GetCell(row, idx, "routeConfigId")
	rl := &hubcommon002models.RateLimitConfig{
		TenantId:            tenantId,
		RateLimitConfigId:   excel.GetCell(row, idx, "rateLimitConfigId"),
		LimitName:           excel.GetCell(row, idx, "limitName"),
		Algorithm:           excel.GetCell(row, idx, "algorithm"),
		KeyStrategy:         excel.GetCell(row, idx, "keyStrategy"),
		LimitRate:           atoiSafe(excel.GetCell(row, idx, "limitRate")),
		BurstCapacity:       atoiSafe(excel.GetCell(row, idx, "burstCapacity")),
		TimeWindowSeconds:   atoiSafe(excel.GetCell(row, idx, "timeWindowSeconds")),
		RejectionStatusCode: atoiSafe(excel.GetCell(row, idx, "rejectionStatusCode")),
		RejectionMessage:    excel.GetCell(row, idx, "rejectionMessage"),
		ConfigPriority:      atoiSafe(excel.GetCell(row, idx, "configPriority")),
		CustomConfig:        excel.GetCell(row, idx, "customConfig"),
		ActiveFlag:          strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
	}
	if gwId != "" {
		rl.GatewayInstanceId = &gwId
	}
	if routeId != "" {
		rl.RouteConfigId = &routeId
	}
	if v := excel.GetCell(row, idx, "noteText"); v != "" {
		rl.NoteText = &v
	}
	return rl
}

// ─── 通用转换工具 ──────────────────────────────────────────────────────────

func atoiSafe(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// atoiPtr 将字符串解析为 *int：空字符串或解析失败返回 nil，否则返回值的指针。
// 用于 *int 模型字段（如 httpPort、httpsPort、reserved3/4 等）。
func atoiPtr(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

func strOrDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func parseProxyConfigRow(row []string, idx map[string]int, tenantId string) *hub0022models.ProxyConfig {
	return &hub0022models.ProxyConfig{
		TenantId:          tenantId,
		ProxyConfigId:     excel.GetCell(row, idx, "proxyConfigId"),
		GatewayInstanceId: excel.GetCell(row, idx, "gatewayInstanceId"),
		ProxyName:         excel.GetCell(row, idx, "proxyName"),
		ProxyType:         excel.GetCell(row, idx, "proxyType"),
		ProxyId:           excel.GetCell(row, idx, "proxyId"),
		ConfigPriority:    atoiSafe(excel.GetCell(row, idx, "configPriority")),
		ProxyConfig:       excel.GetCell(row, idx, "proxyConfig"),
		CustomConfig:      excel.GetCell(row, idx, "customConfig"),
		Reserved1:         excel.GetCell(row, idx, "reserved1"),
		Reserved2:         excel.GetCell(row, idx, "reserved2"),
		Reserved3:         atoiPtr(excel.GetCell(row, idx, "reserved3")),
		Reserved4:         atoiPtr(excel.GetCell(row, idx, "reserved4")),
		ExtProperty:       excel.GetCell(row, idx, "extProperty"),
		OprSeqFlag:        excel.GetCell(row, idx, "oprSeqFlag"),
		CurrentVersion:    atoiSafe(excel.GetCell(row, idx, "currentVersion")),
		ActiveFlag:        strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
		NoteText:          excel.GetCell(row, idx, "noteText"),
		AddWho:            excel.GetCell(row, idx, "addWho"),
		EditWho:           excel.GetCell(row, idx, "editWho"),
	}
}

func parseServiceNodeRow(row []string, idx map[string]int, tenantId string) *hub0022models.ServiceNodeModel {
	return &hub0022models.ServiceNodeModel{
		TenantId:            tenantId,
		ServiceNodeId:       excel.GetCell(row, idx, "serviceNodeId"),
		ServiceDefinitionId: excel.GetCell(row, idx, "serviceDefinitionId"),
		NodeId:              excel.GetCell(row, idx, "nodeId"),
		NodeUrl:             excel.GetCell(row, idx, "nodeUrl"),
		NodeHost:            excel.GetCell(row, idx, "nodeHost"),
		NodePort:            atoiSafe(excel.GetCell(row, idx, "nodePort")),
		NodeProtocol:        excel.GetCell(row, idx, "nodeProtocol"),
		NodeWeight:          atoiSafe(excel.GetCell(row, idx, "nodeWeight")),
		HealthStatus:        excel.GetCell(row, idx, "healthStatus"),
		NodeMetadata:        excel.GetCell(row, idx, "nodeMetadata"),
		NodeStatus:          atoiSafe(excel.GetCell(row, idx, "nodeStatus")),
		HealthCheckResult:   excel.GetCell(row, idx, "healthCheckResult"),
		Reserved1:           excel.GetCell(row, idx, "reserved1"),
		Reserved2:           excel.GetCell(row, idx, "reserved2"),
		Reserved3:           atoiSafe(excel.GetCell(row, idx, "reserved3")),
		Reserved4:           atoiSafe(excel.GetCell(row, idx, "reserved4")),
		ExtProperty:         excel.GetCell(row, idx, "extProperty"),
		OprSeqFlag:          excel.GetCell(row, idx, "oprSeqFlag"),
		CurrentVersion:      atoiSafe(excel.GetCell(row, idx, "currentVersion")),
		ActiveFlag:          strOrDefault(excel.GetCell(row, idx, "activeFlag"), "Y"),
		NoteText:            excel.GetCell(row, idx, "noteText"),
		AddWho:              excel.GetCell(row, idx, "addWho"),
		EditWho:             excel.GetCell(row, idx, "editWho"),
	}
}
