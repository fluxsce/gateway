package gatewaylogroutes

import (
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0023/controllers"
	"gateway/web/views/hub0023/dao"
	"gateway/web/views/hub0023/models"

	"github.com/gin-gonic/gin"
)

// pickEffectiveGatewayLogQueryType 在实例解析结果与可用连接之间取实际使用的查询类型。
func pickEffectiveGatewayLogQueryType(
	resolved string,
	mongoCtl *controllers.MongoQueryController,
	chCtl *controllers.ClickHouseQueryController,
) string {
	switch resolved {
	case "mongo":
		if mongoCtl != nil {
			return "mongo"
		}
		logger.Warn("网关日志按实例配置应为MongoDB查询，但Mongo连接不可用，回退关系数据库")
		return "database"
	case "clickhouse":
		if chCtl != nil {
			return "clickhouse"
		}
		logger.Warn("网关日志按实例配置应为ClickHouse查询，但ClickHouse连接不可用，回退关系数据库")
		return "database"
	default:
		return "database"
	}
}

// countGatewayLogsDatabase 使用关系库 DAO 统计网关日志条数（供分发层使用，不放在 GatewayLogController）。
func countGatewayLogsDatabase(c *gin.Context, db database.Database) {
	var req models.GatewayAccessLogQueryRequest
	if err := request.Bind(c, &req); err != nil {
		logger.ErrorWithTrace(c, "网关日志统计参数解析失败", "error", err)
		response.ErrorJSON(c, "参数解析错误: "+err.Error(), constants.ED00006)
		return
	}
	req.TenantId = request.GetTenantID(c)
	req.PageIndex = 1
	req.PageSize = 1
	gatewayLogDAO := dao.NewGatewayLogDAO(db)
	_, total, err := gatewayLogDAO.Query(c, &req)
	if err != nil {
		logger.ErrorWithTrace(c, "网关日志统计失败", "error", err)
		response.ErrorJSON(c, "统计失败: "+err.Error(), constants.ED00009)
		return
	}
	response.SuccessJSON(c, gin.H{
		"count":      total,
		"collection": models.GatewayAccessLog{}.TableName(),
	}, constants.SD00002)
}

// dispatchGatewayLogQuery 按实例日志配置分发网关日志列表查询。
func dispatchGatewayLogQuery(
	db database.Database,
	mongoCtl *controllers.MongoQueryController,
	chCtl *controllers.ClickHouseQueryController,
	dbCtl *controllers.GatewayLogController,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		gid := strings.TrimSpace(request.GetParam(c, "gatewayInstanceId"))
		tenantID := request.GetTenantID(c)
		resolved := dao.ResolveGatewayLogQueryType(c.Request.Context(), db, tenantID, gid)
		switch pickEffectiveGatewayLogQueryType(resolved, mongoCtl, chCtl) {
		case "mongo":
			mongoCtl.QueryGatewayLogs(c)
		case "clickhouse":
			chCtl.QueryGatewayLogs(c)
		default:
			dbCtl.Query(c)
		}
	}
}

// dispatchGatewayLogGet 按实例日志配置分发网关日志详情查询。
func dispatchGatewayLogGet(
	db database.Database,
	mongoCtl *controllers.MongoQueryController,
	chCtl *controllers.ClickHouseQueryController,
	dbCtl *controllers.GatewayLogController,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		gid := strings.TrimSpace(request.GetParam(c, "gatewayInstanceId"))
		tenantID := request.GetTenantID(c)
		resolved := dao.ResolveGatewayLogQueryType(c.Request.Context(), db, tenantID, gid)
		switch pickEffectiveGatewayLogQueryType(resolved, mongoCtl, chCtl) {
		case "mongo":
			mongoCtl.GetGatewayLog(c)
		case "clickhouse":
			chCtl.GetGatewayLog(c)
		default:
			dbCtl.Get(c)
		}
	}
}

// dispatchGatewayLogAccessDetail 按实例日志配置分发网关日志主表详情（不查询后端追踪日志）。
func dispatchGatewayLogAccessDetail(
	db database.Database,
	mongoCtl *controllers.MongoQueryController,
	chCtl *controllers.ClickHouseQueryController,
	dbCtl *controllers.GatewayLogController,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		gid := strings.TrimSpace(request.GetParam(c, "gatewayInstanceId"))
		tenantID := request.GetTenantID(c)
		resolved := dao.ResolveGatewayLogQueryType(c.Request.Context(), db, tenantID, gid)
		switch pickEffectiveGatewayLogQueryType(resolved, mongoCtl, chCtl) {
		case "mongo":
			mongoCtl.GetGatewayLogAccessDetail(c)
		case "clickhouse":
			chCtl.GetGatewayLogAccessDetail(c)
		default:
			dbCtl.GetAccessDetail(c)
		}
	}
}

// dispatchGatewayLogCount 按实例日志配置分发网关日志统计。
func dispatchGatewayLogCount(
	db database.Database,
	mongoCtl *controllers.MongoQueryController,
	chCtl *controllers.ClickHouseQueryController,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		gid := strings.TrimSpace(request.GetParam(c, "gatewayInstanceId"))
		tenantID := request.GetTenantID(c)
		resolved := dao.ResolveGatewayLogQueryType(c.Request.Context(), db, tenantID, gid)
		switch pickEffectiveGatewayLogQueryType(resolved, mongoCtl, chCtl) {
		case "mongo":
			mongoCtl.CountGatewayLogs(c)
		case "clickhouse":
			chCtl.CountGatewayLogs(c)
		default:
			countGatewayLogsDatabase(c, db)
		}
	}
}

// dispatchGatewayMonitoringOverview 按实例日志配置分发监控概览查询。
func dispatchGatewayMonitoringOverview(
	db database.Database,
	mongoCtl *controllers.MongoQueryController,
	chCtl *controllers.ClickHouseQueryController,
	dbCtl *controllers.GatewayLogController,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		gid := strings.TrimSpace(request.GetParam(c, "gatewayInstanceId"))
		tenantID := request.GetTenantID(c)
		resolved := dao.ResolveGatewayLogQueryType(c.Request.Context(), db, tenantID, gid)
		switch pickEffectiveGatewayLogQueryType(resolved, mongoCtl, chCtl) {
		case "mongo":
			mongoCtl.GetGatewayMonitoringOverview(c)
		case "clickhouse":
			chCtl.GetGatewayMonitoringOverview(c)
		default:
			dbCtl.GetMonitoringOverview(c)
		}
	}
}

// dispatchGatewayMonitoringChartData 按实例日志配置分发监控图表数据查询。
func dispatchGatewayMonitoringChartData(
	db database.Database,
	mongoCtl *controllers.MongoQueryController,
	chCtl *controllers.ClickHouseQueryController,
	dbCtl *controllers.GatewayLogController,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		gid := strings.TrimSpace(request.GetParam(c, "gatewayInstanceId"))
		tenantID := request.GetTenantID(c)
		resolved := dao.ResolveGatewayLogQueryType(c.Request.Context(), db, tenantID, gid)
		switch pickEffectiveGatewayLogQueryType(resolved, mongoCtl, chCtl) {
		case "mongo":
			mongoCtl.GetGatewayMonitoringChartData(c)
		case "clickhouse":
			chCtl.GetGatewayMonitoringChartData(c)
		default:
			dbCtl.GetMonitoringChartData(c)
		}
	}
}
