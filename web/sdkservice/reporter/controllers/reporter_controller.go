package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/sdkservice/reporter/dao"
	"gateway/web/sdkservice/reporter/models"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReporterController JVM监控数据上报控制器
type ReporterController struct {
	dao *dao.JvmReporterDao
	db  database.Database
}

// NewReporterController 创建Reporter控制器实例
func NewReporterController(db database.Database) (*ReporterController, error) {
	if db == nil {
		logger.Error("数据库连接为空，无法创建Reporter控制器")
		return nil, nil
	}

	return &ReporterController{
		dao: dao.NewJvmReporterDao(db),
		db:  db,
	}, nil
}

// ReportJvmData 接收JVM监控数据上报
// @Summary JVM监控数据上报
// @Description 后端Java应用上报JVM运行时监控数据
// @Tags JVM Monitor Reporter
// @Accept json
// @Produce json
// @Param request body models.JvmReportRequest true "JVM监控数据"
// @Success 200 {object} response.JsonData "上报成功"
// @Failure 400 {object} response.JsonData "请求参数错误"
// @Failure 500 {object} response.JsonData "服务器内部错误"
// @Router /gateway/sdk/reporter/jvm/report [post]
func (c *ReporterController) ReportJvmData(ctx *gin.Context) {
	// 1. 从认证中间件设置的上下文中获取已验证的信息（认证通过才能到这里，所以一定存在）
	tenantId := ctx.GetString("tenantId")
	serviceGroupId := ctx.GetString("serviceGroupId")
	groupName := ctx.GetString("groupName")

	logger.Info("收到JVM监控数据上报请求",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"groupName", groupName,
		"clientIP", ctx.ClientIP())

	// 2. 解析请求参数（使用request工具类的绑定方法）
	var req models.JvmReportRequest
	if err := request.BindJSON(ctx, &req); err != nil {
		logger.Error("解析JVM监控数据失败",
			"error", err,
			"tenantId", tenantId)
		response.ErrorJSON(ctx, "请求参数格式错误: "+err.Error(), "INVALID_REQUEST_FORMAT", http.StatusBadRequest)
		return
	}

	// 3. 验证必要字段
	if req.JvmResourceId == "" {
		logger.Warn("JVM资源ID为空", "tenantId", tenantId)
		response.ErrorJSON(ctx, "jvmResourceId不能为空（应由应用端生成唯一标识）", "MISSING_JVM_RESOURCE_ID", http.StatusBadRequest)
		return
	}

	if req.ApplicationName == "" {
		logger.Warn("应用名称为空", "tenantId", tenantId)
		response.ErrorJSON(ctx, "应用名称不能为空", "MISSING_APPLICATION_NAME", http.StatusBadRequest)
		return
	}

	// 4. 记录上报的基本信息
	logger.Info("JVM监控数据详情",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"groupName", groupName,
		"jvmResourceId", req.JvmResourceId,
		"applicationName", req.ApplicationName,
		"hostName", req.HostName,
		"hostIpAddress", req.HostIpAddress,
		"collectionTime", req.JvmResourceInfo.CollectionTime,
		"healthy", req.JvmResourceInfo.Healthy,
		"healthGrade", req.JvmResourceInfo.HealthGrade,
		"requiresAttention", req.JvmResourceInfo.RequiresAttention)

	// 5. 保存监控数据到数据库（传入serviceGroupId和groupName）
	if err := c.dao.SaveJvmMonitoringData(tenantId, serviceGroupId, groupName, &req); err != nil {
		logger.Error("保存JVM监控数据失败",
			"error", err,
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"groupName", groupName,
			"jvmResourceId", req.JvmResourceId,
			"applicationName", req.ApplicationName)
		response.ErrorJSON(ctx, "保存监控数据失败: "+err.Error(), "SAVE_JVM_DATA_FAILED", http.StatusInternalServerError)
		return
	}

	// 6. 检查是否需要告警
	if req.JvmResourceInfo.RequiresAttention {
		logger.Warn("JVM状态需要关注！",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"groupName", groupName,
			"jvmResourceId", req.JvmResourceId,
			"applicationName", req.ApplicationName,
			"healthGrade", req.JvmResourceInfo.HealthGrade,
			"summary", req.JvmResourceInfo.Summary)

		// TODO: 触发告警逻辑
	}

	// 检查死锁
	if req.JvmResourceInfo.ThreadInfo != nil &&
		req.JvmResourceInfo.ThreadInfo.DeadlockInfo != nil &&
		req.JvmResourceInfo.ThreadInfo.DeadlockInfo.HasDeadlock {
		logger.Error("检测到JVM死锁！",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"groupName", groupName,
			"jvmResourceId", req.JvmResourceId,
			"applicationName", req.ApplicationName,
			"deadlockThreadCount", req.JvmResourceInfo.ThreadInfo.DeadlockInfo.DeadlockThreadCount,
			"severity", req.JvmResourceInfo.ThreadInfo.DeadlockInfo.Severity)

		// TODO: 触发紧急告警
	}

	// 7. 返回成功响应
	responseData := map[string]interface{}{
		"jvmResourceId":   req.JvmResourceId,
		"applicationName": req.ApplicationName,
		"serviceGroupId":  serviceGroupId,
		"groupName":       groupName,
		"collectionTime":  req.JvmResourceInfo.CollectionTime,
		"healthy":         req.JvmResourceInfo.Healthy,
		"healthGrade":     req.JvmResourceInfo.HealthGrade,
		"message":         "JVM监控数据已成功接收并保存",
	}

	response.SuccessJSON(ctx, responseData, "REPORT_JVM_DATA_SUCCESS")

	logger.Info("JVM监控数据上报成功",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"groupName", groupName,
		"jvmResourceId", req.JvmResourceId,
		"applicationName", req.ApplicationName)
}

// GetHealthStatus 获取应用健康状态（示例查询接口）
// @Summary 查询应用健康状态
// @Description 查询指定应用的最新JVM健康状态
// @Tags JVM Monitor Reporter
// @Accept json
// @Produce json
// @Param applicationName query string true "应用名称"
// @Param instanceId query string false "实例ID（可选）"
// @Success 200 {object} response.JsonData "查询成功"
// @Failure 400 {object} response.JsonData "请求参数错误"
// @Failure 500 {object} response.JsonData "服务器内部错误"
// @Router /gateway/sdk/reporter/jvm/health [get]
func (c *ReporterController) GetHealthStatus(ctx *gin.Context) {
	// 1. 获取租户ID（使用request工具类）
	tenantId := request.GetParam(ctx, "tenantId", "default")

	// 2. 获取查询参数（使用request工具类）
	jvmResourceId := request.GetParam(ctx, "jvmResourceId")
	applicationName := request.GetParam(ctx, "applicationName")

	// 必须提供jvmResourceId或applicationName之一
	if jvmResourceId == "" && applicationName == "" {
		response.ErrorJSON(ctx, "必须提供jvmResourceId或applicationName参数", "MISSING_QUERY_PARAMS", http.StatusBadRequest)
		return
	}

	logger.Info("查询JVM健康状态",
		"tenantId", tenantId,
		"jvmResourceId", jvmResourceId,
		"applicationName", applicationName)

	// 3. 构建查询SQL（优先使用jvmResourceId快速检索）
	var queryType string

	if jvmResourceId != "" {
		// 使用jvmResourceId快速检索
		queryType = "byJvmResourceId"
		logger.Info("使用jvmResourceId快速检索", "jvmResourceId", jvmResourceId)
	} else {
		// 使用applicationName检索
		queryType = "byApplicationName"
	}

	// 由于Database接口限制，返回模拟数据
	logger.Warn("查询功能暂未完全实现，需要使用原生SQL查询",
		"tenantId", tenantId,
		"queryType", queryType,
		"jvmResourceId", jvmResourceId,
		"applicationName", applicationName)

	// 返回成功但数据为空（实际应该查询数据库）
	var resultData map[string]interface{}
	if jvmResourceId != "" {
		resultData = map[string]interface{}{
			"jvmResourceId":     jvmResourceId,
			"applicationName":   applicationName,
			"collectionTime":    "暂无数据",
			"healthy":           true,
			"healthGrade":       "UNKNOWN",
			"requiresAttention": false,
			"summary":           "查询功能开发中，可通过SQL直接查询：SELECT * FROM HUB_MONITOR_JVM_RESOURCE WHERE tenantId=? AND jvmResourceId=? ORDER BY collectionTime DESC LIMIT 1",
			"uptimeMs":          0,
		}
	} else {
		resultData = map[string]interface{}{
			"applicationName":   applicationName,
			"collectionTime":    "暂无数据",
			"healthy":           true,
			"healthGrade":       "UNKNOWN",
			"requiresAttention": false,
			"summary":           "查询功能开发中，可通过SQL直接查询：SELECT * FROM HUB_MONITOR_JVM_RESOURCE WHERE tenantId=? AND applicationName=? ORDER BY collectionTime DESC LIMIT 10",
			"uptimeMs":          0,
		}
	}

	response.SuccessJSON(ctx, []map[string]interface{}{resultData}, "QUERY_HEALTH_STATUS_SUCCESS")

	logger.Info("JVM健康状态查询完成（模拟数据）",
		"tenantId", tenantId,
		"queryType", queryType,
		"jvmResourceId", jvmResourceId,
		"applicationName", applicationName)
}

// ReportApplicationData 接收应用监控数据上报
// @Summary 应用监控数据上报
// @Description 后端Java应用上报应用层监控数据（线程池、连接池、自定义指标等）
// @Tags Application Monitor Reporter
// @Accept json
// @Produce json
// @Param request body models.AppReportRequest true "应用监控数据"
// @Success 200 {object} response.JsonData "上报成功"
// @Failure 400 {object} response.JsonData "请求参数错误"
// @Failure 500 {object} response.JsonData "服务器内部错误"
// @Router /gateway/sdk/reporter/application/report [post]
func (c *ReporterController) ReportApplicationData(ctx *gin.Context) {
	// 1. 从认证中间件设置的上下文中获取已验证的信息（认证通过才能到这里，所以一定存在）
	tenantId := ctx.GetString("tenantId")
	serviceGroupId := ctx.GetString("serviceGroupId")
	groupName := ctx.GetString("groupName")

	logger.Info("收到应用监控数据上报请求",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"groupName", groupName,
		"clientIP", ctx.ClientIP())

	// 2. 解析请求参数（使用request工具类的绑定方法）
	var req models.AppReportRequest
	if err := request.BindJSON(ctx, &req); err != nil {
		logger.Error("解析应用监控数据失败",
			"error", err,
			"tenantId", tenantId)
		response.ErrorJSON(ctx, "请求参数格式错误: "+err.Error(), "INVALID_REQUEST_FORMAT", http.StatusBadRequest)
		return
	}

	// 3. 验证必要字段
	if req.JvmResourceId == "" {
		logger.Warn("JVM资源ID为空", "tenantId", tenantId)
		response.ErrorJSON(ctx, "jvmResourceId不能为空（应由应用端生成唯一标识）", "MISSING_JVM_RESOURCE_ID", http.StatusBadRequest)
		return
	}

	if req.ApplicationName == "" {
		logger.Warn("应用名称为空", "tenantId", tenantId)
		response.ErrorJSON(ctx, "应用名称不能为空", "MISSING_APPLICATION_NAME", http.StatusBadRequest)
		return
	}

	if req.ApplicationResourceInfo.ThirdPartyMonitorData == nil || len(req.ApplicationResourceInfo.ThirdPartyMonitorData) == 0 {
		logger.Warn("监控数据为空", "tenantId", tenantId)
		response.ErrorJSON(ctx, "监控数据不能为空", "MISSING_MONITOR_DATA", http.StatusBadRequest)
		return
	}

	// 4. 记录上报的基本信息
	logger.Info("应用监控数据详情",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"groupName", groupName,
		"jvmResourceId", req.JvmResourceId,
		"applicationName", req.ApplicationName,
		"hostName", req.HostName,
		"hostIpAddress", req.HostIpAddress,
		"dataCount", len(req.ApplicationResourceInfo.ThirdPartyMonitorData))

	// 5. 保存监控数据到数据库（传入serviceGroupId和groupName）
	if err := c.dao.SaveAppMonitoringData(tenantId, serviceGroupId, groupName, &req); err != nil {
		logger.Error("保存应用监控数据失败",
			"error", err,
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"groupName", groupName,
			"jvmResourceId", req.JvmResourceId,
			"applicationName", req.ApplicationName)
		response.ErrorJSON(ctx, "保存监控数据失败: "+err.Error(), "SAVE_APP_DATA_FAILED", http.StatusInternalServerError)
		return
	}

	// 6. 检查是否需要告警
	alertCount := 0
	for _, data := range req.ApplicationResourceInfo.ThirdPartyMonitorData {
		if data.RequiresAttention() {
			alertCount++
			logger.Warn("应用监控数据需要关注！",
				"tenantId", tenantId,
				"serviceGroupId", serviceGroupId,
				"groupName", groupName,
				"jvmResourceId", req.JvmResourceId,
				"applicationName", req.ApplicationName,
				"dataType", data.DataType,
				"dataName", data.DataName,
				"healthGrade", data.HealthGrade)

			// TODO: 触发告警逻辑
		}
	}

	// 7. 返回成功响应
	responseData := map[string]interface{}{
		"jvmResourceId":   req.JvmResourceId,
		"applicationName": req.ApplicationName,
		"serviceGroupId":  serviceGroupId,
		"groupName":       groupName,
		"dataCount":       len(req.ApplicationResourceInfo.ThirdPartyMonitorData),
		"alertCount":      alertCount,
		"message":         "应用监控数据已成功接收并保存",
	}

	response.SuccessJSON(ctx, responseData, "REPORT_APP_DATA_SUCCESS")

	logger.Info("应用监控数据上报成功",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"groupName", groupName,
		"jvmResourceId", req.JvmResourceId,
		"applicationName", req.ApplicationName,
		"dataCount", len(req.ApplicationResourceInfo.ThirdPartyMonitorData),
		"alertCount", alertCount)
}
