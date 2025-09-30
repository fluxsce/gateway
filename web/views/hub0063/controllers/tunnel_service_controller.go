package controllers

import (
	"fmt"
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/response"
	"gateway/web/views/hub0063/dao"
	"gateway/web/views/hub0063/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TunnelServiceController 隧道服务控制器
type TunnelServiceController struct {
	tunnelServiceDAO *dao.TunnelServiceDAO
}

// NewTunnelServiceController 创建隧道服务控制器实例
func NewTunnelServiceController(db database.Database) *TunnelServiceController {
	return &TunnelServiceController{
		tunnelServiceDAO: dao.NewTunnelServiceDAO(db),
	}
}

// getCurrentUser 获取当前用户
func (c *TunnelServiceController) getCurrentUser(ctx *gin.Context) string {
	// 简化实现：返回默认用户
	// 实际项目中应该从session或JWT token中获取用户信息
	return "admin"
}

// QueryTunnelServices 查询隧道服务列表
func (c *TunnelServiceController) QueryTunnelServices(ctx *gin.Context) {
	var req models.TunnelServiceQueryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定查询参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "QUERY_TUNNEL_SERVICES")
		return
	}

	// 参数验证
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	services, total, err := c.tunnelServiceDAO.QueryTunnelServices(&req)
	if err != nil {
		logger.Error("查询隧道服务列表失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), "QUERY_TUNNEL_SERVICES")
		return
	}

	// 创建分页信息
	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)

	response.PageJSON(ctx, services, pageInfo, "QUERY_TUNNEL_SERVICES")
}

// GetTunnelService 获取隧道服务详情
func (c *TunnelServiceController) GetTunnelService(ctx *gin.Context) {
	type Request struct {
		TunnelServiceId string `json:"tunnelServiceId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "GET_TUNNEL_SERVICE")
		return
	}

	service, err := c.tunnelServiceDAO.GetTunnelService(req.TunnelServiceId)
	if err != nil {
		logger.Error("获取隧道服务详情失败", "tunnelServiceId", req.TunnelServiceId, "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_SERVICE")
		return
	}

	response.SuccessJSON(ctx, service, "GET_TUNNEL_SERVICE")
}

// CreateTunnelService 创建隧道服务
func (c *TunnelServiceController) CreateTunnelService(ctx *gin.Context) {
	var service models.TunnelService
	if err := ctx.ShouldBindJSON(&service); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CREATE_TUNNEL_SERVICE")
		return
	}

	// 参数验证
	if strings.TrimSpace(service.ServiceName) == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", "CREATE_TUNNEL_SERVICE")
		return
	}
	if strings.TrimSpace(service.ServiceType) == "" {
		response.ErrorJSON(ctx, "服务类型不能为空", "CREATE_TUNNEL_SERVICE")
		return
	}
	if strings.TrimSpace(service.TunnelClientId) == "" {
		response.ErrorJSON(ctx, "隧道客户端ID不能为空", "CREATE_TUNNEL_SERVICE")
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
	validTypes := []string{"TCP", "UDP", "HTTP", "HTTPS", "STCP", "SUDP", "XTCP"}
	isValidType := false
	for _, t := range validTypes {
		if service.ServiceType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		response.ErrorJSON(ctx, "无效的服务类型", "CREATE_TUNNEL_SERVICE")
		return
	}

	// 检查服务名称是否重复（在同一客户端下）
	exists, err := c.tunnelServiceDAO.CheckServiceNameExists(service.TunnelClientId, service.ServiceName, "")
	if err != nil {
		logger.Error("检查服务名称是否存在失败", "error", err)
		response.ErrorJSON(ctx, "检查失败: "+err.Error(), "CREATE_TUNNEL_SERVICE")
		return
	}
	if exists {
		response.ErrorJSON(ctx, "在该客户端下服务名称已存在", "CREATE_TUNNEL_SERVICE")
		return
	}

	// 检查远程端口是否被占用（如果指定了远程端口）
	if service.RemotePort != nil && *service.RemotePort > 0 {
		if *service.RemotePort > 65535 {
			response.ErrorJSON(ctx, "远程端口必须在1-65535之间", "CREATE_TUNNEL_SERVICE")
			return
		}

		portExists, err := c.tunnelServiceDAO.CheckRemotePortExists(*service.RemotePort, "")
		if err != nil {
			logger.Error("检查远程端口是否存在失败", "error", err)
			response.ErrorJSON(ctx, "检查失败: "+err.Error(), "CREATE_TUNNEL_SERVICE")
			return
		}
		if portExists {
			response.ErrorJSON(ctx, fmt.Sprintf("远程端口 %d 已被占用", *service.RemotePort), "CREATE_TUNNEL_SERVICE")
			return
		}
	}

	// 生成ID和设置审计字段
	service.TunnelServiceId = uuid.New().String()
	service.AddWho = c.getCurrentUser(ctx)
	service.EditWho = service.AddWho
	service.OprSeqFlag = uuid.New().String()

	err = c.tunnelServiceDAO.CreateTunnelService(&service)
	if err != nil {
		logger.Error("创建隧道服务失败", "error", err)
		response.ErrorJSON(ctx, "创建失败: "+err.Error(), "CREATE_TUNNEL_SERVICE")
		return
	}

	logger.Info("创建隧道服务成功", "tunnelServiceId", service.TunnelServiceId, "serviceName", service.ServiceName)
	result := map[string]interface{}{
		"tunnelServiceId": service.TunnelServiceId,
		"message":         "创建成功",
	}
	response.SuccessJSON(ctx, result, "CREATE_TUNNEL_SERVICE")
}

// UpdateTunnelService 更新隧道服务
func (c *TunnelServiceController) UpdateTunnelService(ctx *gin.Context) {
	var service models.TunnelService
	if err := ctx.ShouldBindJSON(&service); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_SERVICE")
		return
	}

	// 参数验证
	if strings.TrimSpace(service.TunnelServiceId) == "" {
		response.ErrorJSON(ctx, "隧道服务ID不能为空", "UPDATE_TUNNEL_SERVICE")
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
	if strings.TrimSpace(service.TunnelClientId) == "" {
		response.ErrorJSON(ctx, "隧道客户端ID不能为空", "UPDATE_TUNNEL_SERVICE")
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

	// 验证服务类型
	validTypes := []string{"TCP", "UDP", "HTTP", "HTTPS", "STCP", "SUDP", "XTCP"}
	isValidType := false
	for _, t := range validTypes {
		if service.ServiceType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		response.ErrorJSON(ctx, "无效的服务类型", "UPDATE_TUNNEL_SERVICE")
		return
	}

	// 检查服务名称是否重复（在同一客户端下）
	exists, err := c.tunnelServiceDAO.CheckServiceNameExists(service.TunnelClientId, service.ServiceName, service.TunnelServiceId)
	if err != nil {
		logger.Error("检查服务名称是否存在失败", "error", err)
		response.ErrorJSON(ctx, "检查失败: "+err.Error(), "UPDATE_TUNNEL_SERVICE")
		return
	}
	if exists {
		response.ErrorJSON(ctx, "在该客户端下服务名称已存在", "UPDATE_TUNNEL_SERVICE")
		return
	}

	// 检查远程端口是否被占用（如果指定了远程端口）
	if service.RemotePort != nil && *service.RemotePort > 0 {
		if *service.RemotePort > 65535 {
			response.ErrorJSON(ctx, "远程端口必须在1-65535之间", "UPDATE_TUNNEL_SERVICE")
			return
		}

		portExists, err := c.tunnelServiceDAO.CheckRemotePortExists(*service.RemotePort, service.TunnelServiceId)
		if err != nil {
			logger.Error("检查远程端口是否存在失败", "error", err)
			response.ErrorJSON(ctx, "检查失败: "+err.Error(), "UPDATE_TUNNEL_SERVICE")
			return
		}
		if portExists {
			response.ErrorJSON(ctx, fmt.Sprintf("远程端口 %d 已被占用", *service.RemotePort), "UPDATE_TUNNEL_SERVICE")
			return
		}
	}

	// 设置审计字段
	service.EditWho = c.getCurrentUser(ctx)

	err = c.tunnelServiceDAO.UpdateTunnelService(&service)
	if err != nil {
		logger.Error("更新隧道服务失败", "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "UPDATE_TUNNEL_SERVICE")
		return
	}

	logger.Info("更新隧道服务成功", "tunnelServiceId", service.TunnelServiceId, "serviceName", service.ServiceName)
	result := map[string]interface{}{
		"message": "更新成功",
	}
	response.SuccessJSON(ctx, result, "UPDATE_TUNNEL_SERVICE")
}

// DeleteTunnelService 删除隧道服务
func (c *TunnelServiceController) DeleteTunnelService(ctx *gin.Context) {
	type Request struct {
		TunnelServiceId string `json:"tunnelServiceId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "DELETE_TUNNEL_SERVICE")
		return
	}

	editWho := c.getCurrentUser(ctx)
	err := c.tunnelServiceDAO.DeleteTunnelService(req.TunnelServiceId, editWho)
	if err != nil {
		logger.Error("删除隧道服务失败", "tunnelServiceId", req.TunnelServiceId, "error", err)
		response.ErrorJSON(ctx, "删除失败: "+err.Error(), "DELETE_TUNNEL_SERVICE")
		return
	}

	logger.Info("删除隧道服务成功", "tunnelServiceId", req.TunnelServiceId)
	result := map[string]interface{}{
		"message": "删除成功",
	}
	response.SuccessJSON(ctx, result, "DELETE_TUNNEL_SERVICE")
}

// UpdateTunnelServiceStatus 更新隧道服务状态
func (c *TunnelServiceController) UpdateTunnelServiceStatus(ctx *gin.Context) {
	type Request struct {
		TunnelServiceId string `json:"tunnelServiceId" binding:"required"`
		Status          string `json:"status" binding:"required"`
		ConnectionCount int    `json:"connectionCount"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_SERVICE_STATUS")
		return
	}

	// 状态验证
	validStatuses := []string{"ACTIVE", "INACTIVE", "ERROR"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		response.ErrorJSON(ctx, "无效的服务状态", "UPDATE_TUNNEL_SERVICE_STATUS")
		return
	}

	err := c.tunnelServiceDAO.UpdateTunnelServiceStatus(req.TunnelServiceId, req.Status, req.ConnectionCount)
	if err != nil {
		logger.Error("更新隧道服务状态失败", "tunnelServiceId", req.TunnelServiceId, "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "UPDATE_TUNNEL_SERVICE_STATUS")
		return
	}

	logger.Info("更新隧道服务状态成功", "tunnelServiceId", req.TunnelServiceId, "status", req.Status)
	result := map[string]interface{}{
		"message": "状态更新成功",
	}
	response.SuccessJSON(ctx, result, "UPDATE_TUNNEL_SERVICE_STATUS")
}

// UpdateTunnelServiceTraffic 更新隧道服务流量统计
func (c *TunnelServiceController) UpdateTunnelServiceTraffic(ctx *gin.Context) {
	type Request struct {
		TunnelServiceId  string `json:"tunnelServiceId" binding:"required"`
		TotalConnections int64  `json:"totalConnections"`
		TotalTraffic     int64  `json:"totalTraffic"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_SERVICE_TRAFFIC")
		return
	}

	err := c.tunnelServiceDAO.UpdateTunnelServiceTraffic(req.TunnelServiceId, req.TotalConnections, req.TotalTraffic)
	if err != nil {
		logger.Error("更新隧道服务流量统计失败", "tunnelServiceId", req.TunnelServiceId, "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "UPDATE_TUNNEL_SERVICE_TRAFFIC")
		return
	}

	logger.Info("更新隧道服务流量统计成功", "tunnelServiceId", req.TunnelServiceId)
	result := map[string]interface{}{
		"message": "流量统计更新成功",
	}
	response.SuccessJSON(ctx, result, "UPDATE_TUNNEL_SERVICE_TRAFFIC")
}

// GetTunnelServiceStats 获取隧道服务统计信息
func (c *TunnelServiceController) GetTunnelServiceStats(ctx *gin.Context) {
	stats, err := c.tunnelServiceDAO.GetTunnelServiceStats()
	if err != nil {
		logger.Error("获取隧道服务统计信息失败", "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_SERVICE_STATS")
		return
	}

	response.SuccessJSON(ctx, stats, "GET_TUNNEL_SERVICE_STATS")
}

// GetServiceTypeOptions 获取服务类型选项
func (c *TunnelServiceController) GetServiceTypeOptions(ctx *gin.Context) {
	options := c.tunnelServiceDAO.GetServiceTypeOptions()
	response.SuccessJSON(ctx, options, "GET_SERVICE_TYPE_OPTIONS")
}

// GetServiceStatusOptions 获取服务状态选项
func (c *TunnelServiceController) GetServiceStatusOptions(ctx *gin.Context) {
	options := c.tunnelServiceDAO.GetServiceStatusOptions()
	response.SuccessJSON(ctx, options, "GET_SERVICE_STATUS_OPTIONS")
}

// GetTunnelServicesByClientId 根据客户端ID获取服务列表
func (c *TunnelServiceController) GetTunnelServicesByClientId(ctx *gin.Context) {
	type Request struct {
		TunnelClientId string `json:"tunnelClientId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "GET_TUNNEL_SERVICES_BY_CLIENT_ID")
		return
	}

	services, err := c.tunnelServiceDAO.GetTunnelServicesByClientId(req.TunnelClientId)
	if err != nil {
		logger.Error("根据客户端ID获取服务列表失败", "tunnelClientId", req.TunnelClientId, "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_SERVICES_BY_CLIENT_ID")
		return
	}

	// 转换为下拉选项格式
	options := make([]map[string]interface{}, 0, len(services))
	for _, service := range services {
		option := map[string]interface{}{
			"value":           service.TunnelServiceId,
			"label":           fmt.Sprintf("%s (%s:%d)", service.ServiceName, service.ServiceType, service.LocalPort),
			"serviceType":     service.ServiceType,
			"status":          service.ServiceStatus,
			"connectionCount": service.ConnectionCount,
		}
		options = append(options, option)
	}

	response.SuccessJSON(ctx, options, "GET_TUNNEL_SERVICES_BY_CLIENT_ID")
}

// CheckRemotePortAvailable 检查远程端口是否可用
func (c *TunnelServiceController) CheckRemotePortAvailable(ctx *gin.Context) {
	type Request struct {
		RemotePort      int    `json:"remotePort" binding:"required"`
		TunnelServiceId string `json:"tunnelServiceId"` // 编辑时排除自己
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CHECK_REMOTE_PORT_AVAILABLE")
		return
	}

	// 端口范围验证
	if req.RemotePort <= 0 || req.RemotePort > 65535 {
		response.ErrorJSON(ctx, "端口必须在1-65535之间", "CHECK_REMOTE_PORT_AVAILABLE")
		return
	}

	exists, err := c.tunnelServiceDAO.CheckRemotePortExists(req.RemotePort, req.TunnelServiceId)
	if err != nil {
		logger.Error("检查远程端口是否存在失败", "error", err)
		response.ErrorJSON(ctx, "检查失败: "+err.Error(), "CHECK_REMOTE_PORT_AVAILABLE")
		return
	}

	result := map[string]interface{}{
		"available": !exists,
		"port":      req.RemotePort,
	}

	if exists {
		result["message"] = fmt.Sprintf("端口 %d 已被占用", req.RemotePort)
	} else {
		result["message"] = fmt.Sprintf("端口 %d 可用", req.RemotePort)
	}

	response.SuccessJSON(ctx, result, "CHECK_REMOTE_PORT_AVAILABLE")
}

// CheckCustomDomainAvailable 检查自定义域名是否可用
func (c *TunnelServiceController) CheckCustomDomainAvailable(ctx *gin.Context) {
	type Request struct {
		Domain          string `json:"domain" binding:"required"`
		TunnelServiceId string `json:"tunnelServiceId"` // 编辑时排除自己
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CHECK_CUSTOM_DOMAIN_AVAILABLE")
		return
	}

	// 域名格式简单验证
	if strings.TrimSpace(req.Domain) == "" {
		response.ErrorJSON(ctx, "域名不能为空", "CHECK_CUSTOM_DOMAIN_AVAILABLE")
		return
	}

	exists, err := c.tunnelServiceDAO.CheckCustomDomainExists(req.Domain, req.TunnelServiceId)
	if err != nil {
		logger.Error("检查自定义域名是否存在失败", "error", err)
		response.ErrorJSON(ctx, "检查失败: "+err.Error(), "CHECK_CUSTOM_DOMAIN_AVAILABLE")
		return
	}

	result := map[string]interface{}{
		"available": !exists,
		"domain":    req.Domain,
	}

	if exists {
		result["message"] = fmt.Sprintf("域名 %s 已被占用", req.Domain)
	} else {
		result["message"] = fmt.Sprintf("域名 %s 可用", req.Domain)
	}

	response.SuccessJSON(ctx, result, "CHECK_CUSTOM_DOMAIN_AVAILABLE")
}
