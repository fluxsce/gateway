package controllers

import (
	"strings"

	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0062/dao"
	"gateway/web/views/hub0062/models"

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

// QueryTunnelServices 查询服务列表
func (c *TunnelServiceController) QueryTunnelServices(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 绑定查询条件
	var req models.TunnelServiceQueryRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		logger.Warn("绑定服务查询条件失败，使用默认条件", "error", err.Error())
	}
	req.PageIndex = page
	req.PageSize = pageSize

	// 设置租户ID过滤（如果请求中没有指定，则使用当前用户的租户ID）
	if req.TenantId == "" {
		req.TenantId = request.GetTenantID(ctx)
	}

	services, total, err := c.serviceDAO.QueryTunnelServices(&req)
	if err != nil {
		response.ErrorJSON(ctx, "查询服务列表失败: "+err.Error(), "QUERY_TUNNEL_SERVICES")
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "tunnelServiceId"
	response.PageJSON(ctx, services, pageInfo, "QUERY_TUNNEL_SERVICES")
}

// GetTunnelService 获取服务详情
func (c *TunnelServiceController) GetTunnelService(ctx *gin.Context) {
	tunnelServiceId := request.GetParam(ctx, "tunnelServiceId")
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
	var service types.TunnelService
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

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "租户ID不能为空", "CREATE_TUNNEL_SERVICE")
		return
	}
	service.TenantId = tenantId

	// 使用工具类获取操作人ID
	operatorId := request.GetOperatorID(ctx)
	service.AddWho = operatorId
	service.EditWho = operatorId

	// 创建服务
	createdService, err := c.serviceDAO.CreateTunnelService(&service)
	if err != nil {
		response.ErrorJSON(ctx, "创建服务失败: "+err.Error(), "CREATE_TUNNEL_SERVICE")
		return
	}

	logger.Info("创建服务成功", "tunnelServiceId", createdService.TunnelServiceId, "serviceName", createdService.ServiceName, "user", operatorId)
	response.SuccessJSON(ctx, createdService, "CREATE_TUNNEL_SERVICE")
}

// UpdateTunnelService 更新服务
func (c *TunnelServiceController) UpdateTunnelService(ctx *gin.Context) {
	var service types.TunnelService
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

	// 使用工具类获取操作人ID
	operatorId := request.GetOperatorID(ctx)
	service.EditWho = operatorId

	// 更新服务
	updatedService, err := c.serviceDAO.UpdateTunnelService(&service)
	if err != nil {
		response.ErrorJSON(ctx, "更新服务失败: "+err.Error(), "UPDATE_TUNNEL_SERVICE")
		return
	}

	logger.Info("更新服务成功", "tunnelServiceId", updatedService.TunnelServiceId, "serviceName", updatedService.ServiceName, "user", operatorId)
	response.SuccessJSON(ctx, updatedService, "UPDATE_TUNNEL_SERVICE")
}

// DeleteTunnelService 删除服务
func (c *TunnelServiceController) DeleteTunnelService(ctx *gin.Context) {
	tunnelServiceId := request.GetParam(ctx, "tunnelServiceId")
	if tunnelServiceId == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "DELETE_TUNNEL_SERVICE")
		return
	}

	// 使用工具类获取操作人ID
	operatorId := request.GetOperatorID(ctx)

	// 删除服务
	deletedService, err := c.serviceDAO.DeleteTunnelService(tunnelServiceId, operatorId)
	if err != nil {
		response.ErrorJSON(ctx, "删除服务失败: "+err.Error(), "DELETE_TUNNEL_SERVICE")
		return
	}

	logger.Info("删除服务成功", "tunnelServiceId", tunnelServiceId, "user", operatorId)
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

// RegisterService 注册服务到隧道管理器
func (c *TunnelServiceController) RegisterService(ctx *gin.Context) {
	tunnelServiceId := request.GetParam(ctx, "tunnelServiceId")
	if tunnelServiceId == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "REGISTER_SERVICE")
		return
	}

	// 调用DAO层注册服务
	service, err := c.serviceDAO.RegisterService(ctx.Request.Context(), tunnelServiceId)
	if err != nil {
		response.ErrorJSON(ctx, "注册服务失败: "+err.Error(), "REGISTER_SERVICE")
		return
	}

	logger.Info("服务注册成功", "tunnelServiceId", tunnelServiceId)
	response.SuccessJSON(ctx, service, "REGISTER_SERVICE")
}

// UnregisterService 从隧道管理器注销服务
func (c *TunnelServiceController) UnregisterService(ctx *gin.Context) {
	tunnelServiceId := request.GetParam(ctx, "tunnelServiceId")
	if tunnelServiceId == "" {
		response.ErrorJSON(ctx, "服务ID不能为空", "UNREGISTER_SERVICE")
		return
	}

	// 调用DAO层注销服务
	service, err := c.serviceDAO.UnregisterService(ctx.Request.Context(), tunnelServiceId)
	if err != nil {
		response.ErrorJSON(ctx, "注销服务失败: "+err.Error(), "UNREGISTER_SERVICE")
		return
	}

	logger.Info("服务注销成功", "tunnelServiceId", tunnelServiceId)
	response.SuccessJSON(ctx, service, "UNREGISTER_SERVICE")
}
