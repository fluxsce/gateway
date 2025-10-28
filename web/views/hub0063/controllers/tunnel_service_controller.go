package controllers

import (
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0063/dao"
	"gateway/web/views/hub0063/models"

	"github.com/gin-gonic/gin"
)

type TunnelServiceController struct {
	serviceDAO *dao.TunnelServiceDAO
}

func NewTunnelServiceController(db database.Database) *TunnelServiceController {
	return &TunnelServiceController{
		serviceDAO: dao.NewTunnelServiceDAO(db),
	}
}

// getCurrentUser 获取当前用户
func (c *TunnelServiceController) getCurrentUser(ctx *gin.Context) string {
	if userName := request.GetUserName(ctx); userName != "" {
		return userName
	}
	if userID := request.GetUserID(ctx); userID != "" {
		return userID
	}
	return "admin"
}

// QueryTunnelServices 查询服务列表
func (c *TunnelServiceController) QueryTunnelServices(ctx *gin.Context) {
	var req models.TunnelServiceQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "QUERY_TUNNEL_SERVICES")
		return
	}

	// 参数验证和默认值
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}

	services, total, err := c.serviceDAO.QueryTunnelServices(&req)
	if err != nil {
		response.ErrorJSON(ctx, "查询服务列表失败: "+err.Error(), "QUERY_TUNNEL_SERVICES")
		return
	}

	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)
	response.PageJSON(ctx, services, pageInfo, "QUERY_TUNNEL_SERVICES")
}

// GetTunnelService 获取服务详情
func (c *TunnelServiceController) GetTunnelService(ctx *gin.Context) {
	tunnelServiceId := ctx.PostForm("tunnelServiceId")
	if tunnelServiceId == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "GET_TUNNEL_SERVICE")
		return
	}

	service, err := c.serviceDAO.GetTunnelService(tunnelServiceId)
	if err != nil {
		response.ErrorJSON(ctx, "获取服务详情失败: "+err.Error(), "GET_TUNNEL_SERVICE")
		return
	}

	response.SuccessJSON(ctx, service, "GET_TUNNEL_SERVICE")
}

// CreateTunnelService 创建服务
func (c *TunnelServiceController) CreateTunnelService(ctx *gin.Context) {
	var service models.TunnelService
	if err := request.Bind(ctx, &service); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CREATE_TUNNEL_SERVICE")
		return
	}

	// 参数验证
	if strings.TrimSpace(service.ServiceName) == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", "CREATE_TUNNEL_SERVICE")
		return
	}
	if strings.TrimSpace(service.TunnelClientId) == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "CREATE_TUNNEL_SERVICE")
		return
	}
	if strings.TrimSpace(service.ServiceType) == "" {
		response.ErrorJSON(ctx, "服务类型不能为空", "CREATE_TUNNEL_SERVICE")
		return
	}
	if strings.TrimSpace(service.LocalAddress) == "" {
		response.ErrorJSON(ctx, "本地地址不能为空", "CREATE_TUNNEL_SERVICE")
		return
	}
	if service.LocalPort <= 0 || service.LocalPort > 65535 {
		response.ErrorJSON(ctx, "本地端口必须在1-65535之间", "CREATE_TUNNEL_SERVICE")
		return
	}

	// 验证服务类型
	validTypes := map[string]bool{
		"tcp": true, "udp": true, "http": true, "https": true,
		"stcp": true, "sudp": true, "xtcp": true,
	}
	if !validTypes[service.ServiceType] {
		response.ErrorJSON(ctx, "无效的服务类型", "CREATE_TUNNEL_SERVICE")
		return
	}

	// 设置创建人
	currentUser := c.getCurrentUser(ctx)
	service.AddWho = currentUser
	service.EditWho = currentUser

	// 创建服务
	createdService, err := c.serviceDAO.CreateTunnelService(&service)
	if err != nil {
		response.ErrorJSON(ctx, "创建服务失败: "+err.Error(), "CREATE_TUNNEL_SERVICE")
		return
	}

	logger.Info("创建服务成功", "tunnelServiceId", createdService.TunnelServiceId, "serviceName", createdService.ServiceName, "user", currentUser)
	response.SuccessJSON(ctx, createdService, "CREATE_TUNNEL_SERVICE")
}

// UpdateTunnelService 更新服务
func (c *TunnelServiceController) UpdateTunnelService(ctx *gin.Context) {
	var service models.TunnelService
	if err := request.Bind(ctx, &service); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_SERVICE")
		return
	}

	// 参数验证
	if strings.TrimSpace(service.TunnelServiceId) == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "UPDATE_TUNNEL_SERVICE")
		return
	}
	if strings.TrimSpace(service.ServiceName) == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", "UPDATE_TUNNEL_SERVICE")
		return
	}
	if strings.TrimSpace(service.ServiceType) == "" {
		response.ErrorJSON(ctx, "服务类型不能为空", "UPDATE_TUNNEL_SERVICE")
		return
	}
	if strings.TrimSpace(service.LocalAddress) == "" {
		response.ErrorJSON(ctx, "本地地址不能为空", "UPDATE_TUNNEL_SERVICE")
		return
	}
	if service.LocalPort <= 0 || service.LocalPort > 65535 {
		response.ErrorJSON(ctx, "本地端口必须在1-65535之间", "UPDATE_TUNNEL_SERVICE")
		return
	}

	// 设置修改人
	currentUser := c.getCurrentUser(ctx)
	service.EditWho = currentUser

	// 更新服务
	updatedService, err := c.serviceDAO.UpdateTunnelService(&service)
	if err != nil {
		response.ErrorJSON(ctx, "更新服务失败: "+err.Error(), "UPDATE_TUNNEL_SERVICE")
		return
	}

	logger.Info("更新服务成功", "tunnelServiceId", updatedService.TunnelServiceId, "serviceName", updatedService.ServiceName, "user", currentUser)
	response.SuccessJSON(ctx, updatedService, "UPDATE_TUNNEL_SERVICE")
}

// DeleteTunnelService 删除服务
func (c *TunnelServiceController) DeleteTunnelService(ctx *gin.Context) {
	tunnelServiceId := ctx.PostForm("tunnelServiceId")
	if tunnelServiceId == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "DELETE_TUNNEL_SERVICE")
		return
	}

	currentUser := c.getCurrentUser(ctx)

	// 删除服务
	deletedService, err := c.serviceDAO.DeleteTunnelService(tunnelServiceId, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "删除服务失败: "+err.Error(), "DELETE_TUNNEL_SERVICE")
		return
	}

	logger.Info("删除服务成功", "tunnelServiceId", tunnelServiceId, "user", currentUser)
	response.SuccessJSON(ctx, deletedService, "DELETE_TUNNEL_SERVICE")
}

// GetServiceStats 获取服务统计信息
func (c *TunnelServiceController) GetServiceStats(ctx *gin.Context) {
	stats, err := c.serviceDAO.GetServiceStats()
	if err != nil {
		response.ErrorJSON(ctx, "获取服务统计信息失败: "+err.Error(), "GET_SERVICE_STATS")
		return
	}

	response.SuccessJSON(ctx, stats, "GET_SERVICE_STATS")
}

// GetServiceTypeOptions 获取服务类型选项
func (c *TunnelServiceController) GetServiceTypeOptions(ctx *gin.Context) {
	options := []gin.H{
		{"value": "tcp", "label": "TCP"},
		{"value": "udp", "label": "UDP"},
		{"value": "http", "label": "HTTP"},
		{"value": "https", "label": "HTTPS"},
		{"value": "stcp", "label": "STCP（安全TCP）"},
		{"value": "sudp", "label": "SUDP（安全UDP）"},
		{"value": "xtcp", "label": "XTCP（P2P TCP）"},
	}

	response.SuccessJSON(ctx, options, "GET_SERVICE_TYPE_OPTIONS")
}

// GetServicesByClient 按客户端查询服务列表
func (c *TunnelServiceController) GetServicesByClient(ctx *gin.Context) {
	var req models.ServicesByClientRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "GET_SERVICES_BY_CLIENT")
		return
	}

	if strings.TrimSpace(req.TunnelClientId) == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "GET_SERVICES_BY_CLIENT")
		return
	}

	services, err := c.serviceDAO.GetServicesByClient(&req)
	if err != nil {
		response.ErrorJSON(ctx, "查询客户端服务列表失败: "+err.Error(), "GET_SERVICES_BY_CLIENT")
		return
	}

	response.SuccessJSON(ctx, services, "GET_SERVICES_BY_CLIENT")
}

// AllocateRemotePort 分配远程端口
func (c *TunnelServiceController) AllocateRemotePort(ctx *gin.Context) {
	var req models.AllocatePortRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "ALLOCATE_REMOTE_PORT")
		return
	}

	if strings.TrimSpace(req.TunnelServiceId) == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "ALLOCATE_REMOTE_PORT")
		return
	}

	result, err := c.serviceDAO.AllocateRemotePort(req.TunnelServiceId, req.PreferredPort)
	if err != nil {
		response.ErrorJSON(ctx, "分配远程端口失败: "+err.Error(), "ALLOCATE_REMOTE_PORT")
		return
	}

	logger.Info("分配远程端口成功", "tunnelServiceId", req.TunnelServiceId, "remotePort", result.RemotePort)
	response.SuccessJSON(ctx, result, "ALLOCATE_REMOTE_PORT")
}

// ReleaseRemotePort 释放远程端口
func (c *TunnelServiceController) ReleaseRemotePort(ctx *gin.Context) {
	var req models.ReleasePortRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "RELEASE_REMOTE_PORT")
		return
	}

	if strings.TrimSpace(req.TunnelServiceId) == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "RELEASE_REMOTE_PORT")
		return
	}

	err := c.serviceDAO.ReleaseRemotePort(req.TunnelServiceId)
	if err != nil {
		response.ErrorJSON(ctx, "释放远程端口失败: "+err.Error(), "RELEASE_REMOTE_PORT")
		return
	}

	logger.Info("释放远程端口成功", "tunnelServiceId", req.TunnelServiceId)
	response.SuccessJSON(ctx, gin.H{
		"tunnelServiceId": req.TunnelServiceId,
		"message":         "端口已释放",
	}, "RELEASE_REMOTE_PORT")
}

// GetServiceConnections 获取服务连接列表
func (c *TunnelServiceController) GetServiceConnections(ctx *gin.Context) {
	tunnelServiceId := ctx.PostForm("tunnelServiceId")
	if tunnelServiceId == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "GET_SERVICE_CONNECTIONS")
		return
	}

	// TODO: 实现从 hub0065 模块查询连接列表
	// 这里暂时返回空列表，等 hub0065 模块开发完成后再实现
	response.SuccessJSON(ctx, []interface{}{}, "GET_SERVICE_CONNECTIONS")
}

// GetServiceTraffic 获取服务流量统计
func (c *TunnelServiceController) GetServiceTraffic(ctx *gin.Context) {
	var req models.ServiceTrafficRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "GET_SERVICE_TRAFFIC")
		return
	}

	if strings.TrimSpace(req.TunnelServiceId) == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "GET_SERVICE_TRAFFIC")
		return
	}

	// TODO: 实现从 hub0066 模块查询流量统计
	// 这里暂时返回模拟数据
	response.SuccessJSON(ctx, models.ServiceTrafficResponse{
		TunnelServiceId:   req.TunnelServiceId,
		ServiceName:       "service-name",
		TotalConnections:  0,
		ActiveConnections: 0,
		TotalTraffic:      0,
		AvgResponseTime:   0,
		TrafficByHour:     []int64{},
	}, "GET_SERVICE_TRAFFIC")
}

// EnableService 启用服务
func (c *TunnelServiceController) EnableService(ctx *gin.Context) {
	tunnelServiceId := ctx.PostForm("tunnelServiceId")
	if tunnelServiceId == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "ENABLE_SERVICE")
		return
	}

	currentUser := c.getCurrentUser(ctx)

	err := c.serviceDAO.EnableService(tunnelServiceId, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "启用服务失败: "+err.Error(), "ENABLE_SERVICE")
		return
	}

	logger.Info("启用服务成功", "tunnelServiceId", tunnelServiceId, "user", currentUser)
	response.SuccessJSON(ctx, gin.H{
		"tunnelServiceId": tunnelServiceId,
		"message":         "服务已启用",
	}, "ENABLE_SERVICE")
}

// DisableService 禁用服务
func (c *TunnelServiceController) DisableService(ctx *gin.Context) {
	tunnelServiceId := ctx.PostForm("tunnelServiceId")
	if tunnelServiceId == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "DISABLE_SERVICE")
		return
	}

	currentUser := c.getCurrentUser(ctx)

	err := c.serviceDAO.DisableService(tunnelServiceId, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "禁用服务失败: "+err.Error(), "DISABLE_SERVICE")
		return
	}

	logger.Info("禁用服务成功", "tunnelServiceId", tunnelServiceId, "user", currentUser)
	response.SuccessJSON(ctx, gin.H{
		"tunnelServiceId": tunnelServiceId,
		"message":         "服务已禁用",
	}, "DISABLE_SERVICE")
}
