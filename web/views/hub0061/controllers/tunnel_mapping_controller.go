package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/views/hub0061/dao"
	"gateway/web/views/hub0061/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 简化的工具函数
func ErrorResponse(ctx *gin.Context, statusCode int, message, detail string) {
	ctx.JSON(statusCode, gin.H{
		"success": false,
		"message": message,
		"detail":  detail,
	})
}

func SuccessResponse(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func GetCurrentUser(ctx *gin.Context) string {
	return "admin"
}

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

// QueryTunnelMappings 查询隧道映射列表
func (c *TunnelMappingController) QueryTunnelMappings(ctx *gin.Context) {
	var req models.TunnelMappingQueryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定查询参数失败", "error", err)
		ErrorResponse(ctx, http.StatusBadRequest, "参数格式错误", err.Error())
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
		ErrorResponse(ctx, http.StatusInternalServerError, "查询失败", err.Error())
		return
	}

	// 计算分页信息
	totalPages := (total + req.PageSize - 1) / req.PageSize

	response := map[string]interface{}{
		"list":       mappings,
		"total":      total,
		"pageIndex":  req.PageIndex,
		"pageSize":   req.PageSize,
		"totalPages": totalPages,
		"hasNext":    req.PageIndex < totalPages,
		"hasPrev":    req.PageIndex > 1,
	}

	SuccessResponse(ctx, response)
}

// GetTunnelMapping 获取隧道映射详情
func (c *TunnelMappingController) GetTunnelMapping(ctx *gin.Context) {
	type Request struct {
		TunnelMappingId string `json:"tunnelMappingId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		ErrorResponse(ctx, http.StatusBadRequest, "参数格式错误", err.Error())
		return
	}

	mapping, err := c.tunnelMappingDAO.GetTunnelMapping(req.TunnelMappingId)
	if err != nil {
		logger.Error("获取隧道映射详情失败", "tunnelMappingId", req.TunnelMappingId, "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "获取失败", err.Error())
		return
	}

	SuccessResponse(ctx, mapping)
}

// CreateTunnelMapping 创建隧道映射
func (c *TunnelMappingController) CreateTunnelMapping(ctx *gin.Context) {
	var mapping models.TunnelMapping
	if err := ctx.ShouldBindJSON(&mapping); err != nil {
		logger.Error("绑定参数失败", "error", err)
		ErrorResponse(ctx, http.StatusBadRequest, "参数格式错误", err.Error())
		return
	}

	// 参数验证
	if strings.TrimSpace(mapping.MappingName) == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "映射名称不能为空")
		return
	}
	if strings.TrimSpace(mapping.MappingType) == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "映射类型不能为空")
		return
	}
	if strings.TrimSpace(mapping.TunnelServerId) == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "隧道服务器ID不能为空")
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
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "无效的映射类型")
		return
	}

	// 根据映射类型验证特定参数
	switch mapping.MappingType {
	case "PORT":
		if mapping.ExternalPort == nil || *mapping.ExternalPort <= 0 || *mapping.ExternalPort > 65535 {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "外部端口必须在1-65535之间")
			return
		}
		if mapping.InternalPort == nil || *mapping.InternalPort <= 0 || *mapping.InternalPort > 65535 {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "内部端口必须在1-65535之间")
			return
		}
		if mapping.Protocol == nil || (*mapping.Protocol != "TCP" && *mapping.Protocol != "UDP") {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "协议类型必须是TCP或UDP")
			return
		}

		// 检查端口是否已被占用
		portExists, err := c.tunnelMappingDAO.CheckPortExists(*mapping.ExternalPort, *mapping.Protocol, "")
		if err != nil {
			logger.Error("检查端口是否存在失败", "error", err)
			ErrorResponse(ctx, http.StatusInternalServerError, "检查失败", err.Error())
			return
		}
		if portExists {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", fmt.Sprintf("端口 %d (%s) 已被占用", *mapping.ExternalPort, *mapping.Protocol))
			return
		}

	case "DOMAIN", "SUBDOMAIN":
		if mapping.ExternalDomain == nil || strings.TrimSpace(*mapping.ExternalDomain) == "" {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "外部域名不能为空")
			return
		}
		if mapping.InternalHost == nil || strings.TrimSpace(*mapping.InternalHost) == "" {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "内部主机不能为空")
			return
		}
		if mapping.InternalPort2 == nil || *mapping.InternalPort2 <= 0 || *mapping.InternalPort2 > 65535 {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "内部端口必须在1-65535之间")
			return
		}

		// 检查域名是否已被占用
		domainExists, err := c.tunnelMappingDAO.CheckDomainExists(*mapping.ExternalDomain, "")
		if err != nil {
			logger.Error("检查域名是否存在失败", "error", err)
			ErrorResponse(ctx, http.StatusInternalServerError, "检查失败", err.Error())
			return
		}
		if domainExists {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", fmt.Sprintf("域名 %s 已被占用", *mapping.ExternalDomain))
			return
		}
	}

	// 生成ID和设置审计字段
	mapping.TunnelMappingId = uuid.New().String()
	mapping.AddWho = GetCurrentUser(ctx)
	mapping.EditWho = mapping.AddWho
	mapping.OprSeqFlag = uuid.New().String()

	err := c.tunnelMappingDAO.CreateTunnelMapping(&mapping)
	if err != nil {
		logger.Error("创建隧道映射失败", "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "创建失败", err.Error())
		return
	}

	logger.Info("创建隧道映射成功", "tunnelMappingId", mapping.TunnelMappingId, "mappingName", mapping.MappingName)
	SuccessResponse(ctx, map[string]interface{}{
		"tunnelMappingId": mapping.TunnelMappingId,
		"message":         "创建成功",
	})
}

// UpdateTunnelMapping 更新隧道映射
func (c *TunnelMappingController) UpdateTunnelMapping(ctx *gin.Context) {
	var mapping models.TunnelMapping
	if err := ctx.ShouldBindJSON(&mapping); err != nil {
		logger.Error("绑定参数失败", "error", err)
		ErrorResponse(ctx, http.StatusBadRequest, "参数格式错误", err.Error())
		return
	}

	// 参数验证
	if strings.TrimSpace(mapping.TunnelMappingId) == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "隧道映射ID不能为空")
		return
	}
	if strings.TrimSpace(mapping.MappingName) == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "映射名称不能为空")
		return
	}
	if strings.TrimSpace(mapping.MappingType) == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "映射类型不能为空")
		return
	}
	if strings.TrimSpace(mapping.TunnelServerId) == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "隧道服务器ID不能为空")
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
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "无效的映射类型")
		return
	}

	// 根据映射类型验证特定参数
	switch mapping.MappingType {
	case "PORT":
		if mapping.ExternalPort == nil || *mapping.ExternalPort <= 0 || *mapping.ExternalPort > 65535 {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "外部端口必须在1-65535之间")
			return
		}
		if mapping.InternalPort == nil || *mapping.InternalPort <= 0 || *mapping.InternalPort > 65535 {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "内部端口必须在1-65535之间")
			return
		}
		if mapping.Protocol == nil || (*mapping.Protocol != "TCP" && *mapping.Protocol != "UDP") {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "协议类型必须是TCP或UDP")
			return
		}

		// 检查端口是否已被占用（排除自己）
		portExists, err := c.tunnelMappingDAO.CheckPortExists(*mapping.ExternalPort, *mapping.Protocol, mapping.TunnelMappingId)
		if err != nil {
			logger.Error("检查端口是否存在失败", "error", err)
			ErrorResponse(ctx, http.StatusInternalServerError, "检查失败", err.Error())
			return
		}
		if portExists {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", fmt.Sprintf("端口 %d (%s) 已被占用", *mapping.ExternalPort, *mapping.Protocol))
			return
		}

	case "DOMAIN", "SUBDOMAIN":
		if mapping.ExternalDomain == nil || strings.TrimSpace(*mapping.ExternalDomain) == "" {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "外部域名不能为空")
			return
		}
		if mapping.InternalHost == nil || strings.TrimSpace(*mapping.InternalHost) == "" {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "内部主机不能为空")
			return
		}
		if mapping.InternalPort2 == nil || *mapping.InternalPort2 <= 0 || *mapping.InternalPort2 > 65535 {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "内部端口必须在1-65535之间")
			return
		}

		// 检查域名是否已被占用（排除自己）
		domainExists, err := c.tunnelMappingDAO.CheckDomainExists(*mapping.ExternalDomain, mapping.TunnelMappingId)
		if err != nil {
			logger.Error("检查域名是否存在失败", "error", err)
			ErrorResponse(ctx, http.StatusInternalServerError, "检查失败", err.Error())
			return
		}
		if domainExists {
			ErrorResponse(ctx, http.StatusBadRequest, "参数错误", fmt.Sprintf("域名 %s 已被占用", *mapping.ExternalDomain))
			return
		}
	}

	// 设置审计字段
	mapping.EditWho = GetCurrentUser(ctx)

	err := c.tunnelMappingDAO.UpdateTunnelMapping(&mapping)
	if err != nil {
		logger.Error("更新隧道映射失败", "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "更新失败", err.Error())
		return
	}

	logger.Info("更新隧道映射成功", "tunnelMappingId", mapping.TunnelMappingId, "mappingName", mapping.MappingName)
	SuccessResponse(ctx, map[string]interface{}{
		"message": "更新成功",
	})
}

// DeleteTunnelMapping 删除隧道映射
func (c *TunnelMappingController) DeleteTunnelMapping(ctx *gin.Context) {
	type Request struct {
		TunnelMappingId string `json:"tunnelMappingId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		ErrorResponse(ctx, http.StatusBadRequest, "参数格式错误", err.Error())
		return
	}

	editWho := GetCurrentUser(ctx)
	err := c.tunnelMappingDAO.DeleteTunnelMapping(req.TunnelMappingId, editWho)
	if err != nil {
		logger.Error("删除隧道映射失败", "tunnelMappingId", req.TunnelMappingId, "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "删除失败", err.Error())
		return
	}

	logger.Info("删除隧道映射成功", "tunnelMappingId", req.TunnelMappingId)
	SuccessResponse(ctx, map[string]interface{}{
		"message": "删除成功",
	})
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		ErrorResponse(ctx, http.StatusBadRequest, "参数格式错误", err.Error())
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
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "无效的映射状态")
		return
	}

	err := c.tunnelMappingDAO.UpdateTunnelMappingStatus(req.TunnelMappingId, req.Status, req.ActiveConnections, req.ErrorMessage)
	if err != nil {
		logger.Error("更新隧道映射状态失败", "tunnelMappingId", req.TunnelMappingId, "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "更新失败", err.Error())
		return
	}

	logger.Info("更新隧道映射状态成功", "tunnelMappingId", req.TunnelMappingId, "status", req.Status)
	SuccessResponse(ctx, map[string]interface{}{
		"message": "状态更新成功",
	})
}

// UpdateTunnelMappingTraffic 更新隧道映射流量统计
func (c *TunnelMappingController) UpdateTunnelMappingTraffic(ctx *gin.Context) {
	type Request struct {
		TunnelMappingId string `json:"tunnelMappingId" binding:"required"`
		TotalRequests   int64  `json:"totalRequests"`
		TotalTraffic    int64  `json:"totalTraffic"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		ErrorResponse(ctx, http.StatusBadRequest, "参数格式错误", err.Error())
		return
	}

	err := c.tunnelMappingDAO.UpdateTunnelMappingTraffic(req.TunnelMappingId, req.TotalRequests, req.TotalTraffic)
	if err != nil {
		logger.Error("更新隧道映射流量统计失败", "tunnelMappingId", req.TunnelMappingId, "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "更新失败", err.Error())
		return
	}

	logger.Info("更新隧道映射流量统计成功", "tunnelMappingId", req.TunnelMappingId)
	SuccessResponse(ctx, map[string]interface{}{
		"message": "流量统计更新成功",
	})
}

// GetTunnelMappingStats 获取隧道映射统计信息
func (c *TunnelMappingController) GetTunnelMappingStats(ctx *gin.Context) {
	stats, err := c.tunnelMappingDAO.GetTunnelMappingStats()
	if err != nil {
		logger.Error("获取隧道映射统计信息失败", "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "获取失败", err.Error())
		return
	}

	SuccessResponse(ctx, stats)
}

// GetMappingTypeOptions 获取映射类型选项
func (c *TunnelMappingController) GetMappingTypeOptions(ctx *gin.Context) {
	options := c.tunnelMappingDAO.GetMappingTypeOptions()
	SuccessResponse(ctx, options)
}

// GetMappingStatusOptions 获取映射状态选项
func (c *TunnelMappingController) GetMappingStatusOptions(ctx *gin.Context) {
	options := c.tunnelMappingDAO.GetMappingStatusOptions()
	SuccessResponse(ctx, options)
}

// GetProtocolOptions 获取协议选项
func (c *TunnelMappingController) GetProtocolOptions(ctx *gin.Context) {
	options := c.tunnelMappingDAO.GetProtocolOptions()
	SuccessResponse(ctx, options)
}

// CheckPortAvailable 检查端口是否可用
func (c *TunnelMappingController) CheckPortAvailable(ctx *gin.Context) {
	type Request struct {
		ExternalPort    int    `json:"externalPort" binding:"required"`
		Protocol        string `json:"protocol" binding:"required"`
		TunnelMappingId string `json:"tunnelMappingId"` // 编辑时排除自己
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		ErrorResponse(ctx, http.StatusBadRequest, "参数格式错误", err.Error())
		return
	}

	// 端口范围验证
	if req.ExternalPort <= 0 || req.ExternalPort > 65535 {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "端口必须在1-65535之间")
		return
	}

	// 协议验证
	if req.Protocol != "TCP" && req.Protocol != "UDP" {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "协议必须是TCP或UDP")
		return
	}

	exists, err := c.tunnelMappingDAO.CheckPortExists(req.ExternalPort, req.Protocol, req.TunnelMappingId)
	if err != nil {
		logger.Error("检查端口是否存在失败", "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "检查失败", err.Error())
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

	SuccessResponse(ctx, result)
}

// CheckDomainAvailable 检查域名是否可用
func (c *TunnelMappingController) CheckDomainAvailable(ctx *gin.Context) {
	type Request struct {
		Domain          string `json:"domain" binding:"required"`
		TunnelMappingId string `json:"tunnelMappingId"` // 编辑时排除自己
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		ErrorResponse(ctx, http.StatusBadRequest, "参数格式错误", err.Error())
		return
	}

	// 域名格式简单验证
	if strings.TrimSpace(req.Domain) == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "参数错误", "域名不能为空")
		return
	}

	exists, err := c.tunnelMappingDAO.CheckDomainExists(req.Domain, req.TunnelMappingId)
	if err != nil {
		logger.Error("检查域名是否存在失败", "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "检查失败", err.Error())
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

	SuccessResponse(ctx, result)
}

// GetPortUsageList 获取端口使用列表
func (c *TunnelMappingController) GetPortUsageList(ctx *gin.Context) {
	portList, err := c.tunnelMappingDAO.GetPortUsageList()
	if err != nil {
		logger.Error("获取端口使用列表失败", "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "获取失败", err.Error())
		return
	}

	SuccessResponse(ctx, portList)
}

// GetDomainUsageList 获取域名使用列表
func (c *TunnelMappingController) GetDomainUsageList(ctx *gin.Context) {
	domainList, err := c.tunnelMappingDAO.GetDomainUsageList()
	if err != nil {
		logger.Error("获取域名使用列表失败", "error", err)
		ErrorResponse(ctx, http.StatusInternalServerError, "获取失败", err.Error())
		return
	}

	SuccessResponse(ctx, domainList)
}
