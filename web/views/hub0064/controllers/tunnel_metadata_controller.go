package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/response"
	"gateway/web/views/hub0064/models"

	"github.com/gin-gonic/gin"
)

// TunnelMetadataController 隧道配置元数据控制器
type TunnelMetadataController struct {
	db database.Database
}

// NewTunnelMetadataController 创建隧道配置元数据控制器实例
func NewTunnelMetadataController(db database.Database) *TunnelMetadataController {
	return &TunnelMetadataController{
		db: db,
	}
}

// GetAllMetadata 获取所有元数据
func (c *TunnelMetadataController) GetAllMetadata(ctx *gin.Context) {
	metadata := &models.MetadataResponse{
		ServerStatusOptions: []models.StatusOption{
			{Value: "ONLINE", Label: "在线"},
			{Value: "OFFLINE", Label: "离线"},
			{Value: "MAINTENANCE", Label: "维护中"},
		},
		ClientStatusOptions: []models.StatusOption{
			{Value: "CONNECTED", Label: "已连接"},
			{Value: "DISCONNECTED", Label: "已断开"},
			{Value: "RECONNECTING", Label: "重连中"},
		},
		ServiceTypeOptions: []models.ServiceTypeOption{
			{Value: "TCP", Label: "TCP"},
			{Value: "UDP", Label: "UDP"},
			{Value: "HTTP", Label: "HTTP"},
			{Value: "HTTPS", Label: "HTTPS"},
			{Value: "STCP", Label: "STCP（安全TCP）"},
			{Value: "SUDP", Label: "SUDP（安全UDP）"},
			{Value: "XTCP", Label: "XTCP（P2P TCP）"},
		},
		ServiceStatusOptions: []models.StatusOption{
			{Value: "ACTIVE", Label: "活跃"},
			{Value: "INACTIVE", Label: "非活跃"},
			{Value: "ERROR", Label: "错误"},
		},
		ProtocolOptions: []models.ProtocolOption{
			{Value: "TCP", Label: "TCP"},
			{Value: "UDP", Label: "UDP"},
		},
		MappingTypeOptions: []models.MappingTypeOption{
			{Value: "PORT", Label: "端口映射"},
			{Value: "DOMAIN", Label: "域名映射"},
			{Value: "SUBDOMAIN", Label: "子域名映射"},
		},
		MappingStatusOptions: []models.StatusOption{
			{Value: "ACTIVE", Label: "活跃"},
			{Value: "INACTIVE", Label: "非活跃"},
			{Value: "ERROR", Label: "错误"},
		},
	}

	response.SuccessJSON(ctx, metadata, "GET_ALL_METADATA")
}

// GetServerStatusOptions 获取服务器状态选项
func (c *TunnelMetadataController) GetServerStatusOptions(ctx *gin.Context) {
	options := []models.StatusOption{
		{Value: "ONLINE", Label: "在线"},
		{Value: "OFFLINE", Label: "离线"},
		{Value: "MAINTENANCE", Label: "维护中"},
	}
	response.SuccessJSON(ctx, options, "GET_SERVER_STATUS_OPTIONS")
}

// GetClientStatusOptions 获取客户端状态选项
func (c *TunnelMetadataController) GetClientStatusOptions(ctx *gin.Context) {
	options := []models.StatusOption{
		{Value: "CONNECTED", Label: "已连接"},
		{Value: "DISCONNECTED", Label: "已断开"},
		{Value: "RECONNECTING", Label: "重连中"},
	}
	response.SuccessJSON(ctx, options, "GET_CLIENT_STATUS_OPTIONS")
}

// GetServiceTypeOptions 获取服务类型选项
func (c *TunnelMetadataController) GetServiceTypeOptions(ctx *gin.Context) {
	options := []models.ServiceTypeOption{
		{Value: "TCP", Label: "TCP"},
		{Value: "UDP", Label: "UDP"},
		{Value: "HTTP", Label: "HTTP"},
		{Value: "HTTPS", Label: "HTTPS"},
		{Value: "STCP", Label: "STCP（安全TCP）"},
		{Value: "SUDP", Label: "SUDP（安全UDP）"},
		{Value: "XTCP", Label: "XTCP（P2P TCP）"},
	}
	response.SuccessJSON(ctx, options, "GET_SERVICE_TYPE_OPTIONS")
}

// GetServiceStatusOptions 获取服务状态选项
func (c *TunnelMetadataController) GetServiceStatusOptions(ctx *gin.Context) {
	options := []models.StatusOption{
		{Value: "ACTIVE", Label: "活跃"},
		{Value: "INACTIVE", Label: "非活跃"},
		{Value: "ERROR", Label: "错误"},
	}
	response.SuccessJSON(ctx, options, "GET_SERVICE_STATUS_OPTIONS")
}

// GetProtocolOptions 获取协议选项
func (c *TunnelMetadataController) GetProtocolOptions(ctx *gin.Context) {
	options := []models.ProtocolOption{
		{Value: "TCP", Label: "TCP"},
		{Value: "UDP", Label: "UDP"},
	}
	response.SuccessJSON(ctx, options, "GET_PROTOCOL_OPTIONS")
}

// GetMappingTypeOptions 获取映射类型选项
func (c *TunnelMetadataController) GetMappingTypeOptions(ctx *gin.Context) {
	options := []models.MappingTypeOption{
		{Value: "PORT", Label: "端口映射"},
		{Value: "DOMAIN", Label: "域名映射"},
		{Value: "SUBDOMAIN", Label: "子域名映射"},
	}
	response.SuccessJSON(ctx, options, "GET_MAPPING_TYPE_OPTIONS")
}

// GetMappingStatusOptions 获取映射状态选项
func (c *TunnelMetadataController) GetMappingStatusOptions(ctx *gin.Context) {
	options := []models.StatusOption{
		{Value: "ACTIVE", Label: "活跃"},
		{Value: "INACTIVE", Label: "非活跃"},
		{Value: "ERROR", Label: "错误"},
	}
	response.SuccessJSON(ctx, options, "GET_MAPPING_STATUS_OPTIONS")
}

// GetTunnelServerList 获取隧道服务器列表（用于下拉选择）
func (c *TunnelMetadataController) GetTunnelServerList(ctx *gin.Context) {
	// 模拟数据
	options := []models.TunnelServerOption{
		{
			Value:  "server-001",
			Label:  "主服务器 (192.168.1.100:7000)",
			Status: "ONLINE",
		},
		{
			Value:  "server-002",
			Label:  "备用服务器 (192.168.1.101:7001)",
			Status: "OFFLINE",
		},
	}
	response.SuccessJSON(ctx, options, "GET_TUNNEL_SERVER_LIST")
}

// GetTunnelClientList 获取隧道客户端列表（用于下拉选择）
func (c *TunnelMetadataController) GetTunnelClientList(ctx *gin.Context) {
	// 模拟数据
	options := []models.TunnelClientOption{
		{
			Value:              "client-001",
			Label:              "客户端1 (CONNECTED)",
			Status:             "CONNECTED",
			RegisteredServices: 3,
			ActiveProxies:      2,
		},
		{
			Value:              "client-002",
			Label:              "客户端2 (DISCONNECTED)",
			Status:             "DISCONNECTED",
			RegisteredServices: 1,
			ActiveProxies:      0,
		},
	}
	response.SuccessJSON(ctx, options, "GET_TUNNEL_CLIENT_LIST")
}

// GetTunnelClientsByServerId 根据服务器ID获取客户端列表
func (c *TunnelMetadataController) GetTunnelClientsByServerId(ctx *gin.Context) {
	type Request struct {
		TunnelServerId string `json:"tunnelServerId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "GET_TUNNEL_CLIENTS_BY_SERVER_ID")
		return
	}

	// 模拟数据 - 根据服务器ID过滤
	var options []models.TunnelClientOption
	if req.TunnelServerId == "server-001" {
		options = []models.TunnelClientOption{
			{
				Value:              "client-001",
				Label:              "客户端1 (CONNECTED)",
				Status:             "CONNECTED",
				RegisteredServices: 3,
				ActiveProxies:      2,
			},
			{
				Value:              "client-002",
				Label:              "客户端2 (DISCONNECTED)",
				Status:             "DISCONNECTED",
				RegisteredServices: 1,
				ActiveProxies:      0,
			},
		}
	} else {
		options = []models.TunnelClientOption{}
	}

	response.SuccessJSON(ctx, options, "GET_TUNNEL_CLIENTS_BY_SERVER_ID")
}
