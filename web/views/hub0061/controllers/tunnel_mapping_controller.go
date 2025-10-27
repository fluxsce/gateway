package controllers

import (
	"fmt"
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0061/dao"
	"gateway/web/views/hub0061/models"

	"github.com/gin-gonic/gin"
)

// TunnelMappingController 隧道映射控制器
type TunnelMappingController struct {
	tunnelMappingDAO *dao.TunnelMappingDAO
}

// NewTunnelMappingController 创建隧道映射控制器实例
func NewTunnelMappingController(db database.Database) *TunnelMappingController {
	return &TunnelMappingController{
		tunnelMappingDAO: dao.NewTunnelMappingDAO(db),
	}
}

// getCurrentUser 获取当前用户
func (c *TunnelMappingController) getCurrentUser(ctx *gin.Context) string {
	// 使用 request 工具类获取用户信息
	if userName := request.GetUserName(ctx); userName != "" {
		return userName
	}
	if userID := request.GetUserID(ctx); userID != "" {
		return userID
	}
	// 如果无法获取用户信息，返回默认用户
	return "admin"
}

// QueryTunnelMappings 查询隧道映射列表
func (c *TunnelMappingController) QueryTunnelMappings(ctx *gin.Context) {
	var req models.TunnelMappingQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("绑定查询参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "QUERY_TUNNEL_MAPPINGS")
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

	mappings, total, err := c.tunnelMappingDAO.QueryTunnelMappings(&req)
	if err != nil {
		logger.Error("查询隧道映射列表失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), "QUERY_TUNNEL_MAPPINGS")
		return
	}

	// 创建分页信息
	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)

	response.PageJSON(ctx, mappings, pageInfo, "QUERY_TUNNEL_MAPPINGS")
}

// GetTunnelMapping 获取隧道映射详情
func (c *TunnelMappingController) GetTunnelMapping(ctx *gin.Context) {
	type Request struct {
		TunnelMappingId string `json:"tunnelMappingId" binding:"required"`
	}

	var req Request
	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "GET_TUNNEL_MAPPING")
		return
	}

	mapping, err := c.tunnelMappingDAO.GetTunnelMapping(req.TunnelMappingId)
	if err != nil {
		logger.Error("获取隧道映射详情失败", "tunnelMappingId", req.TunnelMappingId, "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_MAPPING")
		return
	}

	response.SuccessJSON(ctx, mapping, "GET_TUNNEL_MAPPING")
}

// CreateTunnelMapping 创建隧道映射
func (c *TunnelMappingController) CreateTunnelMapping(ctx *gin.Context) {
	var mapping models.TunnelMapping
	if err := request.Bind(ctx, &mapping); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CREATE_TUNNEL_MAPPING")
		return
	}

	// 参数验证
	if strings.TrimSpace(mapping.MappingName) == "" {
		response.ErrorJSON(ctx, "映射名称不能为空", "CREATE_TUNNEL_MAPPING")
		return
	}
	if strings.TrimSpace(mapping.MappingType) == "" {
		response.ErrorJSON(ctx, "映射类型不能为空", "CREATE_TUNNEL_MAPPING")
		return
	}
	if strings.TrimSpace(mapping.TunnelServerId) == "" {
		response.ErrorJSON(ctx, "隧道服务器ID不能为空", "CREATE_TUNNEL_MAPPING")
		return
	}

	// 验证映射类型
	validTypes := []string{"PORT", "DOMAIN", "SUBDOMAIN"}
	isValidType := false
	for _, t := range validTypes {
		if mapping.MappingType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		response.ErrorJSON(ctx, "无效的映射类型", "TUNNEL_MAPPING_OPERATION")
		return
	}

	// 根据映射类型验证特定参数
	switch mapping.MappingType {
	case "PORT":
		if mapping.ExternalPort == nil || *mapping.ExternalPort <= 0 || *mapping.ExternalPort > 65535 {
			response.ErrorJSON(ctx, "外部端口必须在1-65535之间", "CREATE_TUNNEL_MAPPING")
			return
		}
		if mapping.InternalPort == nil || *mapping.InternalPort <= 0 || *mapping.InternalPort > 65535 {
			response.ErrorJSON(ctx, "内部端口必须在1-65535之间", "CREATE_TUNNEL_MAPPING")
			return
		}
		if mapping.Protocol == nil || (*mapping.Protocol != "TCP" && *mapping.Protocol != "UDP") {
			response.ErrorJSON(ctx, "协议类型必须是TCP或UDP", "CREATE_TUNNEL_MAPPING")
			return
		}

		// 检查端口是否已被占用
		portExists, err := c.tunnelMappingDAO.CheckPortExists(*mapping.ExternalPort, *mapping.Protocol, "")
		if err != nil {
			logger.Error("检查端口是否存在失败", "error", err)
			response.ErrorJSON(ctx, "检查失败: "+err.Error(), "CREATE_TUNNEL_MAPPING")
			return
		}
		if portExists {
			response.ErrorJSON(ctx, fmt.Sprintf("端口 %d (%s) 已被占用", *mapping.ExternalPort, *mapping.Protocol), "CREATE_TUNNEL_MAPPING")
			return
		}

	case "DOMAIN", "SUBDOMAIN":
		if mapping.ExternalDomain == nil || strings.TrimSpace(*mapping.ExternalDomain) == "" {
			response.ErrorJSON(ctx, "外部域名不能为空", "CREATE_TUNNEL_MAPPING")
			return
		}
		if mapping.InternalHost == nil || strings.TrimSpace(*mapping.InternalHost) == "" {
			response.ErrorJSON(ctx, "内部主机不能为空", "CREATE_TUNNEL_MAPPING")
			return
		}
		if mapping.InternalPort2 == nil || *mapping.InternalPort2 <= 0 || *mapping.InternalPort2 > 65535 {
			response.ErrorJSON(ctx, "内部端口必须在1-65535之间", "CREATE_TUNNEL_MAPPING")
			return
		}

		// 检查域名是否已被占用
		domainExists, err := c.tunnelMappingDAO.CheckDomainExists(*mapping.ExternalDomain, "")
		if err != nil {
			logger.Error("检查域名是否存在失败", "error", err)
			response.ErrorJSON(ctx, "检查失败: "+err.Error(), "CREATE_TUNNEL_MAPPING")
			return
		}
		if domainExists {
			response.ErrorJSON(ctx, fmt.Sprintf("域名 %s 已被占用", *mapping.ExternalDomain), "CREATE_TUNNEL_MAPPING")
			return
		}
	}

	// 生成ID和设置审计字段
	mapping.TunnelMappingId = random.Generate32BitRandomString()
	mapping.AddWho = c.getCurrentUser(ctx)
	mapping.EditWho = mapping.AddWho
	mapping.OprSeqFlag = random.Generate32BitRandomString()

	err := c.tunnelMappingDAO.CreateTunnelMapping(&mapping)
	if err != nil {
		logger.Error("创建隧道映射失败", "error", err)
		response.ErrorJSON(ctx, "创建失败: "+err.Error(), "CREATE_TUNNEL_MAPPING")
		return
	}

	logger.Info("创建隧道映射成功", "tunnelMappingId", mapping.TunnelMappingId, "mappingName", mapping.MappingName)
	response.SuccessJSON(ctx, mapping, "CREATE_TUNNEL_MAPPING")
}

// UpdateTunnelMapping 更新隧道映射
func (c *TunnelMappingController) UpdateTunnelMapping(ctx *gin.Context) {
	var mapping models.TunnelMapping
	if err := request.Bind(ctx, &mapping); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	// 参数验证
	if strings.TrimSpace(mapping.TunnelMappingId) == "" {
		response.ErrorJSON(ctx, "隧道映射ID不能为空", "UPDATE_TUNNEL_MAPPING")
		return
	}
	if strings.TrimSpace(mapping.MappingName) == "" {
		response.ErrorJSON(ctx, "映射名称不能为空", "UPDATE_TUNNEL_MAPPING")
		return
	}
	if strings.TrimSpace(mapping.MappingType) == "" {
		response.ErrorJSON(ctx, "映射类型不能为空", "UPDATE_TUNNEL_MAPPING")
		return
	}
	if strings.TrimSpace(mapping.TunnelServerId) == "" {
		response.ErrorJSON(ctx, "隧道服务器ID不能为空", "UPDATE_TUNNEL_MAPPING")
		return
	}

	// 验证映射类型
	validTypes := []string{"PORT", "DOMAIN", "SUBDOMAIN"}
	isValidType := false
	for _, t := range validTypes {
		if mapping.MappingType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		response.ErrorJSON(ctx, "无效的映射类型", "UPDATE_TUNNEL_MAPPING")
		return
	}

	// 根据映射类型验证特定参数
	switch mapping.MappingType {
	case "PORT":
		if mapping.ExternalPort == nil || *mapping.ExternalPort <= 0 || *mapping.ExternalPort > 65535 {
			response.ErrorJSON(ctx, "外部端口必须在1-65535之间", "UPDATE_TUNNEL_MAPPING")
			return
		}
		if mapping.InternalPort == nil || *mapping.InternalPort <= 0 || *mapping.InternalPort > 65535 {
			response.ErrorJSON(ctx, "内部端口必须在1-65535之间", "UPDATE_TUNNEL_MAPPING")
			return
		}
		if mapping.Protocol == nil || (*mapping.Protocol != "TCP" && *mapping.Protocol != "UDP") {
			response.ErrorJSON(ctx, "协议类型必须是TCP或UDP", "UPDATE_TUNNEL_MAPPING")
			return
		}

		// 检查端口是否已被占用（排除自己）
		portExists, err := c.tunnelMappingDAO.CheckPortExists(*mapping.ExternalPort, *mapping.Protocol, mapping.TunnelMappingId)
		if err != nil {
			logger.Error("检查端口是否存在失败", "error", err)
			response.ErrorJSON(ctx, "检查失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
			return
		}
		if portExists {
			response.ErrorJSON(ctx, fmt.Sprintf("端口 %d (%s) 已被占用", *mapping.ExternalPort, *mapping.Protocol), "UPDATE_TUNNEL_MAPPING")
			return
		}

	case "DOMAIN", "SUBDOMAIN":
		if mapping.ExternalDomain == nil || strings.TrimSpace(*mapping.ExternalDomain) == "" {
			response.ErrorJSON(ctx, "外部域名不能为空", "UPDATE_TUNNEL_MAPPING")
			return
		}
		if mapping.InternalHost == nil || strings.TrimSpace(*mapping.InternalHost) == "" {
			response.ErrorJSON(ctx, "内部主机不能为空", "UPDATE_TUNNEL_MAPPING")
			return
		}
		if mapping.InternalPort2 == nil || *mapping.InternalPort2 <= 0 || *mapping.InternalPort2 > 65535 {
			response.ErrorJSON(ctx, "内部端口必须在1-65535之间", "UPDATE_TUNNEL_MAPPING")
			return
		}

		// 检查域名是否已被占用（排除自己）
		domainExists, err := c.tunnelMappingDAO.CheckDomainExists(*mapping.ExternalDomain, mapping.TunnelMappingId)
		if err != nil {
			logger.Error("检查域名是否存在失败", "error", err)
			response.ErrorJSON(ctx, "检查失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
			return
		}
		if domainExists {
			response.ErrorJSON(ctx, fmt.Sprintf("域名 %s 已被占用", *mapping.ExternalDomain), "UPDATE_TUNNEL_MAPPING")
			return
		}
	}

	// 设置审计字段
	mapping.EditWho = c.getCurrentUser(ctx)

	err := c.tunnelMappingDAO.UpdateTunnelMapping(&mapping)
	if err != nil {
		logger.Error("更新隧道映射失败", "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	logger.Info("更新隧道映射成功", "tunnelMappingId", mapping.TunnelMappingId, "mappingName", mapping.MappingName)
	response.SuccessJSON(ctx, mapping, "UPDATE_TUNNEL_MAPPING")
}

// DeleteTunnelMapping 删除隧道映射
func (c *TunnelMappingController) DeleteTunnelMapping(ctx *gin.Context) {
	type Request struct {
		TunnelMappingId string `json:"tunnelMappingId" binding:"required"`
	}

	var req Request
	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	editWho := c.getCurrentUser(ctx)
	err := c.tunnelMappingDAO.DeleteTunnelMapping(req.TunnelMappingId, editWho)
	if err != nil {
		logger.Error("删除隧道映射失败", "tunnelMappingId", req.TunnelMappingId, "error", err)
		response.ErrorJSON(ctx, "删除失败: "+err.Error(), "DELETE_TUNNEL_MAPPING")
		return
	}

	logger.Info("删除隧道映射成功", "tunnelMappingId", req.TunnelMappingId)
	response.SuccessJSON(ctx, gin.H{"message": "删除成功"}, "DELETE_TUNNEL_MAPPING")
}

// UpdateTunnelMappingStatus 更新隧道映射状态
func (c *TunnelMappingController) UpdateTunnelMappingStatus(ctx *gin.Context) {
	type Request struct {
		TunnelMappingId   string  `json:"tunnelMappingId" binding:"required"`
		Status            string  `json:"status" binding:"required"`
		ActiveConnections int     `json:"activeConnections"`
		ErrorMessage      *string `json:"errorMessage"`
	}

	var req Request
	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
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
		response.ErrorJSON(ctx, "无效的映射状态", "UPDATE_TUNNEL_MAPPING_STATUS")
		return
	}

	err := c.tunnelMappingDAO.UpdateTunnelMappingStatus(req.TunnelMappingId, req.Status, req.ActiveConnections, req.ErrorMessage)
	if err != nil {
		logger.Error("更新隧道映射状态失败", "tunnelMappingId", req.TunnelMappingId, "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	logger.Info("更新隧道映射状态成功", "tunnelMappingId", req.TunnelMappingId, "status", req.Status)
	response.SuccessJSON(ctx, gin.H{"message": "状态更新成功"}, "UPDATE_TUNNEL_MAPPING_STATUS")
}

// UpdateTunnelMappingTraffic 更新隧道映射流量统计
func (c *TunnelMappingController) UpdateTunnelMappingTraffic(ctx *gin.Context) {
	type Request struct {
		TunnelMappingId string `json:"tunnelMappingId" binding:"required"`
		TotalRequests   int64  `json:"totalRequests"`
		TotalTraffic    int64  `json:"totalTraffic"`
	}

	var req Request
	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	err := c.tunnelMappingDAO.UpdateTunnelMappingTraffic(req.TunnelMappingId, req.TotalRequests, req.TotalTraffic)
	if err != nil {
		logger.Error("更新隧道映射流量统计失败", "tunnelMappingId", req.TunnelMappingId, "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	logger.Info("更新隧道映射流量统计成功", "tunnelMappingId", req.TunnelMappingId)
	response.SuccessJSON(ctx, gin.H{"message": "流量统计更新成功"}, "UPDATE_TUNNEL_MAPPING_TRAFFIC")
}

// GetTunnelMappingStats 获取隧道映射统计信息
func (c *TunnelMappingController) GetTunnelMappingStats(ctx *gin.Context) {
	stats, err := c.tunnelMappingDAO.GetTunnelMappingStats()
	if err != nil {
		logger.Error("获取隧道映射统计信息失败", "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	response.SuccessJSON(ctx, stats, "GET_TUNNEL_MAPPING_STATS")
}

// GetMappingTypeOptions 获取映射类型选项
func (c *TunnelMappingController) GetMappingTypeOptions(ctx *gin.Context) {
	options := c.tunnelMappingDAO.GetMappingTypeOptions()
	response.SuccessJSON(ctx, options, "GET_OPTIONS")
}

// GetMappingStatusOptions 获取映射状态选项
func (c *TunnelMappingController) GetMappingStatusOptions(ctx *gin.Context) {
	options := c.tunnelMappingDAO.GetMappingStatusOptions()
	response.SuccessJSON(ctx, options, "GET_OPTIONS")
}

// GetProtocolOptions 获取协议选项
func (c *TunnelMappingController) GetProtocolOptions(ctx *gin.Context) {
	options := c.tunnelMappingDAO.GetProtocolOptions()
	response.SuccessJSON(ctx, options, "GET_OPTIONS")
}

// CheckPortAvailable 检查端口是否可用
func (c *TunnelMappingController) CheckPortAvailable(ctx *gin.Context) {
	type Request struct {
		ExternalPort    int    `json:"externalPort" binding:"required"`
		Protocol        string `json:"protocol" binding:"required"`
		TunnelMappingId string `json:"tunnelMappingId"` // 编辑时排除自己
	}

	var req Request
	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	// 端口范围验证
	if req.ExternalPort <= 0 || req.ExternalPort > 65535 {
		response.ErrorJSON(ctx, "端口必须在1-65535之间", "CHECK_PORT_AVAILABLE")
		return
	}

	// 协议验证
	if req.Protocol != "TCP" && req.Protocol != "UDP" {
		response.ErrorJSON(ctx, "协议必须是TCP或UDP", "CHECK_PORT_AVAILABLE")
		return
	}

	exists, err := c.tunnelMappingDAO.CheckPortExists(req.ExternalPort, req.Protocol, req.TunnelMappingId)
	if err != nil {
		logger.Error("检查端口是否存在失败", "error", err)
		response.ErrorJSON(ctx, "检查失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	result := map[string]interface{}{
		"available": !exists,
		"port":      req.ExternalPort,
		"protocol":  req.Protocol,
	}

	if exists {
		result["message"] = fmt.Sprintf("端口 %d (%s) 已被占用", req.ExternalPort, req.Protocol)
	} else {
		result["message"] = fmt.Sprintf("端口 %d (%s) 可用", req.ExternalPort, req.Protocol)
	}

	response.SuccessJSON(ctx, result, "CHECK_AVAILABILITY")
}

// CheckDomainAvailable 检查域名是否可用
func (c *TunnelMappingController) CheckDomainAvailable(ctx *gin.Context) {
	type Request struct {
		Domain          string `json:"domain" binding:"required"`
		TunnelMappingId string `json:"tunnelMappingId"` // 编辑时排除自己
	}

	var req Request
	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	// 域名格式简单验证
	if strings.TrimSpace(req.Domain) == "" {
		response.ErrorJSON(ctx, "域名不能为空", "CHECK_DOMAIN_AVAILABLE")
		return
	}

	exists, err := c.tunnelMappingDAO.CheckDomainExists(req.Domain, req.TunnelMappingId)
	if err != nil {
		logger.Error("检查域名是否存在失败", "error", err)
		response.ErrorJSON(ctx, "检查失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
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

	response.SuccessJSON(ctx, result, "CHECK_AVAILABILITY")
}

// GetPortUsageList 获取端口使用列表
func (c *TunnelMappingController) GetPortUsageList(ctx *gin.Context) {
	portList, err := c.tunnelMappingDAO.GetPortUsageList()
	if err != nil {
		logger.Error("获取端口使用列表失败", "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	response.SuccessJSON(ctx, portList, "GET_PORT_USAGE_LIST")
}

// GetDomainUsageList 获取域名使用列表
func (c *TunnelMappingController) GetDomainUsageList(ctx *gin.Context) {
	domainList, err := c.tunnelMappingDAO.GetDomainUsageList()
	if err != nil {
		logger.Error("获取域名使用列表失败", "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "TUNNEL_MAPPING_OPERATION")
		return
	}

	response.SuccessJSON(ctx, domainList, "GET_DOMAIN_USAGE_LIST")
}
