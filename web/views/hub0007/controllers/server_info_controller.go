package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0007/dao"
	"gateway/web/views/hub0007/models"

	"github.com/gin-gonic/gin"
)

// ServerInfoController 系统节点信息控制器
type ServerInfoController struct {
	db            database.Database
	serverInfoDAO *dao.ServerInfoDAO
	metricDAO     *dao.MetricDAO
}

// NewServerInfoController 创建系统节点信息控制器
func NewServerInfoController(db database.Database) *ServerInfoController {
	return &ServerInfoController{
		db:            db,
		serverInfoDAO: dao.NewServerInfoDAO(db),
		metricDAO:     dao.NewMetricDAO(db),
	}
}

// QueryServerInfos 获取系统节点信息列表
// @Summary 获取系统节点信息列表
// @Description 分页获取系统节点信息列表，支持条件查询
// @Tags 系统节点监控
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param hostname query string false "主机名（模糊查询）"
// @Param osType query string false "操作系统类型"
// @Param serverType query string false "服务器类型（physical/virtual/unknown）"
// @Param ipAddress query string false "IP地址（模糊查询）"
// @Param serverLocation query string false "服务器位置（模糊查询）"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hub0007/queryServerInfos [post]
func (c *ServerInfoController) QueryServerInfos(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.ServerInfoQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定服务器信息查询条件失败，使用默认条件", "error", err.Error())
	}

	// 调用DAO获取服务器信息列表
	serverInfos, total, err := c.serverInfoDAO.ListServerInfos(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务器信息列表失败", err)
		// 使用统一的错误响应
		response.ErrorJSON(ctx, "获取服务器信息列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	serverInfoList := make([]map[string]interface{}, 0, len(serverInfos))
	for _, serverInfo := range serverInfos {
		serverInfoList = append(serverInfoList, serverInfo.ToMap())
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "metricServerId"

	// 使用统一的分页响应
	response.PageJSON(ctx, serverInfoList, pageInfo, constants.SD00002)
}

// GetServerInfo 获取单个系统节点信息详情
// @Summary 获取系统节点信息详情
// @Description 根据ID获取系统节点详细信息
// @Tags 系统节点监控
// @Produce json
// @Param metricServerId query string true "服务器ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0007/getServerInfo [post]
func (c *ServerInfoController) GetServerInfo(ctx *gin.Context) {
	metricServerId := request.GetParam(ctx, "metricServerId")
	tenantId := request.GetTenantID(ctx)

	// 验证参数
	if metricServerId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	// 调用DAO获取服务器信息进行校验
	serverInfo, err := c.serverInfoDAO.GetServerInfoById(ctx, metricServerId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务器信息失败", err)
		response.ErrorJSON(ctx, "获取服务器信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if serverInfo == nil {
		response.ErrorJSON(ctx, "服务器信息不存在", constants.ED00008)
		return
	}

	response.SuccessJSON(ctx, serverInfo, constants.SD00001)
}

// QueryCPUMetrics 查询CPU监控数据
// @Summary 查询CPU监控数据
// @Description 根据服务器ID和时间范围查询CPU监控数据
// @Tags 系统节点监控
// @Produce json
// @Param metricServerId query string true "服务器ID"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.JsonData
// @Router /api/hub0007/metrics/cpu [post]
func (c *ServerInfoController) QueryCPUMetrics(ctx *gin.Context) {
	metricServerId := request.GetParam(ctx, "metricServerId")
	startTime := request.GetParam(ctx, "startTime")
	endTime := request.GetParam(ctx, "endTime")
	tenantId := request.GetTenantID(ctx)

	if metricServerId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	metrics, err := c.metricDAO.QueryCPUMetrics(ctx, tenantId, metricServerId, startTime, endTime)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询CPU监控数据失败", err)
		response.ErrorJSON(ctx, "查询CPU监控数据失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, metrics, constants.SD00002)
}

// QueryMemoryMetrics 查询内存监控数据
// @Summary 查询内存监控数据
// @Description 根据服务器ID和时间范围查询内存监控数据
// @Tags 系统节点监控
// @Produce json
// @Param metricServerId query string true "服务器ID"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.JsonData
// @Router /api/hub0007/metrics/memory [post]
func (c *ServerInfoController) QueryMemoryMetrics(ctx *gin.Context) {
	metricServerId := request.GetParam(ctx, "metricServerId")
	startTime := request.GetParam(ctx, "startTime")
	endTime := request.GetParam(ctx, "endTime")
	tenantId := request.GetTenantID(ctx)

	if metricServerId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	metrics, err := c.metricDAO.QueryMemoryMetrics(ctx, tenantId, metricServerId, startTime, endTime)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询内存监控数据失败", err)
		response.ErrorJSON(ctx, "查询内存监控数据失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, metrics, constants.SD00002)
}

// QueryDiskMetrics 查询磁盘监控数据
// @Summary 查询磁盘监控数据
// @Description 根据服务器ID和时间范围查询磁盘监控数据
// @Tags 系统节点监控
// @Produce json
// @Param metricServerId query string true "服务器ID"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.JsonData
// @Router /api/hub0007/metrics/disk [post]
func (c *ServerInfoController) QueryDiskMetrics(ctx *gin.Context) {
	metricServerId := request.GetParam(ctx, "metricServerId")
	startTime := request.GetParam(ctx, "startTime")
	endTime := request.GetParam(ctx, "endTime")
	tenantId := request.GetTenantID(ctx)

	if metricServerId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	metrics, err := c.metricDAO.QueryDiskMetrics(ctx, tenantId, metricServerId, startTime, endTime)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询磁盘监控数据失败", err)
		response.ErrorJSON(ctx, "查询磁盘监控数据失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, metrics, constants.SD00002)
}

// QueryNetworkMetrics 查询网络监控数据
// @Summary 查询网络监控数据
// @Description 根据服务器ID和时间范围查询网络监控数据
// @Tags 系统节点监控
// @Produce json
// @Param metricServerId query string true "服务器ID"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.JsonData
// @Router /api/hub0007/metrics/network [post]
func (c *ServerInfoController) QueryNetworkMetrics(ctx *gin.Context) {
	metricServerId := request.GetParam(ctx, "metricServerId")
	startTime := request.GetParam(ctx, "startTime")
	endTime := request.GetParam(ctx, "endTime")
	tenantId := request.GetTenantID(ctx)

	if metricServerId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	metrics, err := c.metricDAO.QueryNetworkMetrics(ctx, tenantId, metricServerId, startTime, endTime)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询网络监控数据失败", err)
		response.ErrorJSON(ctx, "查询网络监控数据失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, metrics, constants.SD00002)
}

// QueryDiskIOMetrics 查询磁盘IO监控数据
// @Summary 查询磁盘IO监控数据
// @Description 根据服务器ID和时间范围查询磁盘IO监控数据
// @Tags 系统节点监控
// @Produce json
// @Param metricServerId query string true "服务器ID"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.JsonData
// @Router /api/hub0007/metrics/diskio [post]
func (c *ServerInfoController) QueryDiskIOMetrics(ctx *gin.Context) {
	metricServerId := request.GetParam(ctx, "metricServerId")
	startTime := request.GetParam(ctx, "startTime")
	endTime := request.GetParam(ctx, "endTime")
	tenantId := request.GetTenantID(ctx)

	if metricServerId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	metrics, err := c.metricDAO.QueryDiskIOMetrics(ctx, tenantId, metricServerId, startTime, endTime)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询磁盘IO监控数据失败", err)
		response.ErrorJSON(ctx, "查询磁盘IO监控数据失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, metrics, constants.SD00002)
}

// QueryProcessMetrics 查询进程监控数据
// @Summary 查询进程监控数据
// @Description 根据服务器ID和时间范围查询进程监控数据
// @Tags 系统节点监控
// @Produce json
// @Param metricServerId query string true "服务器ID"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.JsonData
// @Router /api/hub0007/metrics/process [post]
func (c *ServerInfoController) QueryProcessMetrics(ctx *gin.Context) {
	metricServerId := request.GetParam(ctx, "metricServerId")
	startTime := request.GetParam(ctx, "startTime")
	endTime := request.GetParam(ctx, "endTime")
	tenantId := request.GetTenantID(ctx)

	if metricServerId == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", constants.ED00006)
		return
	}

	metrics, err := c.metricDAO.QueryProcessMetrics(ctx, tenantId, metricServerId, startTime, endTime)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询进程监控数据失败", err)
		response.ErrorJSON(ctx, "查询进程监控数据失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, metrics, constants.SD00002)
}
