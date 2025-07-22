package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hub0000/dao"
	"gohub/web/views/hub0000/models"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricQueryController 指标查询控制器
type MetricQueryController struct {
	dao *dao.MetricQueryDAO
}

// NewMetricQueryController 创建指标查询控制器
func NewMetricQueryController(db database.Database) *MetricQueryController {
	return &MetricQueryController{
		dao: dao.NewMetricQueryDAO(db),
	}
}

// QueryServerInfoList 查询服务器信息列表
func (c *MetricQueryController) QueryServerInfoList(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询服务器信息列表", "controller", "MetricQueryController", "action", "QueryServerInfoList")

	// 绑定请求参数
	var req models.ServerInfoQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.Page = page
	req.PageSize = pageSize

	// 验证时间字段格式
	if err := req.ValidateTimeFields(); err != nil {
		logger.ErrorWithTrace(ctx, "时间字段验证失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "时间格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调用DAO层查询服务器信息列表
	servers, total, err := c.dao.QueryServerInfoList(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务器信息列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询服务器信息列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应 - 直接返回结构体数据
	serverList := make([]interface{}, 0, len(servers))
	for _, server := range servers {
		if server != nil {
			serverList = append(serverList, server)
		}
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricServerId"

	logger.InfoWithTrace(ctx, "查询服务器信息列表成功", "tenantId", tenantId, "total", total, "page", page)
	response.PageJSON(ctx, serverList, pageInfo, constants.SD00002)
}

// GetServerInfoDetail 获取服务器信息详情
func (c *MetricQueryController) GetServerInfoDetail(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取服务器信息详情", "controller", "MetricQueryController", "action", "GetServerInfoDetail")

	// 定义请求参数
	var req struct {
		ServerId string `json:"serverId" form:"serverId" binding:"required"`
	}

	// 绑定请求参数
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层获取服务器信息详情
	serverInfo, err := c.dao.GetServerInfoDetail(ctx, tenantId, req.ServerId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务器信息详情失败", "error", err.Error(), "tenantId", tenantId, "serverId", req.ServerId)
		response.ErrorJSON(ctx, "获取服务器信息详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if serverInfo == nil {
		response.ErrorJSON(ctx, "服务器不存在", constants.ED00008)
		return
	}

	// 解析详细信息
	networkInfo, _ := serverInfo.GetNetworkInfo()
	systemInfo, _ := serverInfo.GetSystemInfo()
	hardwareInfo, _ := serverInfo.GetHardwareInfo()

	// 构建响应
	detailResponse := map[string]interface{}{
		"serverInfo":   serverInfo,
		"networkInfo":  networkInfo,
		"systemInfo":   systemInfo,
		"hardwareInfo": hardwareInfo,
		"timestamp":    time.Now(),
	}

	logger.InfoWithTrace(ctx, "获取服务器信息详情成功", "tenantId", tenantId, "serverId", req.ServerId)
	response.SuccessJSON(ctx, detailResponse, constants.SD00002)
}

// QueryCpuLogList 查询CPU性能日志列表
func (c *MetricQueryController) QueryCpuLogList(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询CPU性能日志列表", "controller", "MetricQueryController", "action", "QueryCpuLogList")

	// 绑定请求参数
	var req models.CpuLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.Page = page
	req.PageSize = pageSize

	// 验证时间字段格式
	if err := req.ValidateTimeFields(); err != nil {
		logger.ErrorWithTrace(ctx, "时间字段验证失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "时间格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调用DAO层查询CPU日志列表
	cpuLogs, total, err := c.dao.QueryCpuLogList(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询CPU性能日志列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询CPU性能日志列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应 - 直接返回结构体数据
	cpuLogList := make([]interface{}, 0, len(cpuLogs))
	for _, cpuLog := range cpuLogs {
		if cpuLog != nil {
			cpuLogList = append(cpuLogList, cpuLog)
		}
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricCpuLogId"

	logger.InfoWithTrace(ctx, "查询CPU性能日志列表成功", "tenantId", tenantId, "total", total, "page", page)
	response.PageJSON(ctx, cpuLogList, pageInfo, constants.SD00002)
}

// QueryMemoryLogList 查询内存性能日志列表
func (c *MetricQueryController) QueryMemoryLogList(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询内存性能日志列表", "controller", "MetricQueryController", "action", "QueryMemoryLogList")

	// 绑定请求参数
	var req models.MemoryLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.Page = page
	req.PageSize = pageSize

	// 验证时间字段格式
	if err := req.ValidateTimeFields(); err != nil {
		logger.ErrorWithTrace(ctx, "时间字段验证失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "时间格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调用DAO层查询内存日志列表
	memoryLogs, total, err := c.dao.QueryMemoryLogList(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询内存性能日志列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询内存性能日志列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应 - 直接返回结构体数据
	memoryLogList := make([]interface{}, 0, len(memoryLogs))
	for _, memoryLog := range memoryLogs {
		if memoryLog != nil {
			memoryLogList = append(memoryLogList, memoryLog)
		}
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricMemoryLogId"

	logger.InfoWithTrace(ctx, "查询内存性能日志列表成功", "tenantId", tenantId, "total", total, "page", page)
	response.PageJSON(ctx, memoryLogList, pageInfo, constants.SD00002)
}

// QueryDiskPartitionLogList 查询磁盘分区日志列表
func (c *MetricQueryController) QueryDiskPartitionLogList(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询磁盘分区日志列表", "controller", "MetricQueryController", "action", "QueryDiskPartitionLogList")

	// 绑定请求参数
	var req models.DiskPartitionLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.Page = page
	req.PageSize = pageSize

	// 验证时间字段格式
	if err := req.ValidateTimeFields(); err != nil {
		logger.ErrorWithTrace(ctx, "时间字段验证失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "时间格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调用DAO层查询磁盘分区日志列表
	diskPartitionLogs, total, err := c.dao.QueryDiskPartitionLogList(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询磁盘分区日志列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询磁盘分区日志列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应 - 直接返回结构体数据
	diskPartitionLogList := make([]interface{}, 0, len(diskPartitionLogs))
	for _, diskPartitionLog := range diskPartitionLogs {
		if diskPartitionLog != nil {
			diskPartitionLogList = append(diskPartitionLogList, diskPartitionLog)
		}
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricDiskPartitionLogId"

	logger.InfoWithTrace(ctx, "查询磁盘分区日志列表成功", "tenantId", tenantId, "total", total, "page", page)
	response.PageJSON(ctx, diskPartitionLogList, pageInfo, constants.SD00002)
}

// QueryDiskIoLogList 查询磁盘IO日志列表
func (c *MetricQueryController) QueryDiskIoLogList(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询磁盘IO日志列表", "controller", "MetricQueryController", "action", "QueryDiskIoLogList")

	// 绑定请求参数
	var req models.DiskIoLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.Page = page
	req.PageSize = pageSize

	// 验证时间字段格式
	if err := req.ValidateTimeFields(); err != nil {
		logger.ErrorWithTrace(ctx, "时间字段验证失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "时间格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调用DAO层查询磁盘IO日志列表
	diskIoLogs, total, err := c.dao.QueryDiskIoLogList(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询磁盘IO日志列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询磁盘IO日志列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应 - 直接返回结构体数据
	diskIoLogList := make([]interface{}, 0, len(diskIoLogs))
	for _, diskIoLog := range diskIoLogs {
		if diskIoLog != nil {
			diskIoLogList = append(diskIoLogList, diskIoLog)
		}
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricDiskIoLogId"

	logger.InfoWithTrace(ctx, "查询磁盘IO日志列表成功", "tenantId", tenantId, "total", total, "page", page)
	response.PageJSON(ctx, diskIoLogList, pageInfo, constants.SD00002)
}

// QueryNetworkLogList 查询网络日志列表
func (c *MetricQueryController) QueryNetworkLogList(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询网络日志列表", "controller", "MetricQueryController", "action", "QueryNetworkLogList")

	// 绑定请求参数
	var req models.NetworkLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.Page = page
	req.PageSize = pageSize

	// 验证时间字段格式
	if err := req.ValidateTimeFields(); err != nil {
		logger.ErrorWithTrace(ctx, "时间字段验证失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "时间格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调用DAO层查询网络日志列表
	networkLogs, total, err := c.dao.QueryNetworkLogList(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询网络日志列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询网络日志列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应 - 直接返回结构体数据
	networkLogList := make([]interface{}, 0, len(networkLogs))
	for _, networkLog := range networkLogs {
		if networkLog != nil {
			networkLogList = append(networkLogList, networkLog)
		}
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricNetworkLogId"

	logger.InfoWithTrace(ctx, "查询网络日志列表成功", "tenantId", tenantId, "total", total, "page", page)
	response.PageJSON(ctx, networkLogList, pageInfo, constants.SD00002)
}

// QueryProcessLogList 查询进程日志列表
func (c *MetricQueryController) QueryProcessLogList(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询进程日志列表", "controller", "MetricQueryController", "action", "QueryProcessLogList")

	// 绑定请求参数
	var req models.ProcessLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.Page = page
	req.PageSize = pageSize

	// 验证时间字段格式
	if err := req.ValidateTimeFields(); err != nil {
		logger.ErrorWithTrace(ctx, "时间字段验证失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "时间格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调用DAO层查询进程日志列表
	processLogs, total, err := c.dao.QueryProcessLogList(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询进程日志列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询进程日志列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应 - 直接返回结构体数据
	processLogList := make([]interface{}, 0, len(processLogs))
	for _, processLog := range processLogs {
		if processLog != nil {
			processLogList = append(processLogList, processLog)
		}
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricProcessLogId"

	logger.InfoWithTrace(ctx, "查询进程日志列表成功", "tenantId", tenantId, "total", total, "page", page)
	response.PageJSON(ctx, processLogList, pageInfo, constants.SD00002)
}

// QueryProcessStatsLogList 查询进程统计日志列表
func (c *MetricQueryController) QueryProcessStatsLogList(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询进程统计日志列表", "controller", "MetricQueryController", "action", "QueryProcessStatsLogList")

	// 绑定请求参数
	var req models.ProcessStatsLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.Page = page
	req.PageSize = pageSize

	// 验证时间字段格式
	if err := req.ValidateTimeFields(); err != nil {
		logger.ErrorWithTrace(ctx, "时间字段验证失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "时间格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调用DAO层查询进程统计日志列表
	processStatsLogs, total, err := c.dao.QueryProcessStatsLogList(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询进程统计日志列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询进程统计日志列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应 - 直接返回结构体数据
	processStatsLogList := make([]interface{}, 0, len(processStatsLogs))
	for _, processStatsLog := range processStatsLogs {
		if processStatsLog != nil {
			processStatsLogList = append(processStatsLogList, processStatsLog)
		}
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricProcessStatsLogId"

	logger.InfoWithTrace(ctx, "查询进程统计日志列表成功", "tenantId", tenantId, "total", total, "page", page)
	response.PageJSON(ctx, processStatsLogList, pageInfo, constants.SD00002)
}

// QueryTemperatureLogList 查询温度日志列表
func (c *MetricQueryController) QueryTemperatureLogList(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询温度日志列表", "controller", "MetricQueryController", "action", "QueryTemperatureLogList")

	// 绑定请求参数
	var req models.TemperatureLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.Page = page
	req.PageSize = pageSize

	// 验证时间字段格式
	if err := req.ValidateTimeFields(); err != nil {
		logger.ErrorWithTrace(ctx, "时间字段验证失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "时间格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调用DAO层查询温度日志列表
	temperatureLogs, total, err := c.dao.QueryTemperatureLogList(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询温度日志列表失败", "error", err.Error(), "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询温度日志列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应 - 直接返回结构体数据
	temperatureLogList := make([]interface{}, 0, len(temperatureLogs))
	for _, temperatureLog := range temperatureLogs {
		if temperatureLog != nil {
			temperatureLogList = append(temperatureLogList, temperatureLog)
		}
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricTemperatureLogId"

	logger.InfoWithTrace(ctx, "查询温度日志列表成功", "tenantId", tenantId, "total", total, "page", page)
	response.PageJSON(ctx, temperatureLogList, pageInfo, constants.SD00002)
}

 