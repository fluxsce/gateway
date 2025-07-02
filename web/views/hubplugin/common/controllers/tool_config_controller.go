package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/pkg/utils/crypto"
	"gohub/pkg/utils/random"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hubplugin/common/dao"
	"gohub/web/views/hubplugin/common/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ToolConfigController 工具配置控制器
type ToolConfigController struct {
	dao *dao.ToolConfigDao
}

// NewToolConfigController 创建工具配置控制器
func NewToolConfigController(db database.Database) *ToolConfigController {
	return &ToolConfigController{
		dao: dao.NewToolConfigDao(db),
	}
}

// AddToolConfig 添加工具配置
// @Summary 添加工具配置
// @Description 添加新的工具配置
// @Tags SFTP配置管理
// @Accept json
// @Produce json
// @Param data body models.ToolConfig true "工具配置信息"
// @Success 200 {object} response.Response
// @Router /api/sftp/add [post]
func (c *ToolConfigController) AddToolConfig(ctx *gin.Context) {
	// 解析请求参数
	var toolConfig models.ToolConfig
	if err := request.BindSafely(ctx, &toolConfig); err != nil {
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 参数验证
	if toolConfig.ToolName == "" {
		response.ErrorJSON(ctx, "工具名称不能为空", constants.ED00007)
		return
	}
	if toolConfig.ConfigName == "" {
		response.ErrorJSON(ctx, "配置名称不能为空", constants.ED00007)
		return
	}
	if toolConfig.HostAddress == nil || *toolConfig.HostAddress == "" {
		response.ErrorJSON(ctx, "主机地址不能为空", constants.ED00007)
		return
	}

	// 强制设置从上下文获取的租户ID和操作人信息
	toolConfig.TenantId = tenantId
	toolConfig.AddWho = operatorId
	toolConfig.EditWho = operatorId

	// 生成工具配置ID (32位长度限制)
	if toolConfig.ToolConfigId == "" {
		// 使用UUID去掉连字符，确保长度为32位
		toolConfig.ToolConfigId = strings.ReplaceAll(uuid.New().String(), "-", "")
	}

	// 设置默认值
	now := time.Now()
	toolConfig.AddTime = now
	toolConfig.EditTime = now
	toolConfig.CurrentVersion = 1
	toolConfig.ActiveFlag = "Y"

	// 生成OprSeqFlag
	toolConfig.OprSeqFlag = random.Generate32BitRandomString()

	// 设置默认工具类型
	if toolConfig.ToolType == "" {
		toolConfig.ToolType = "SFTP"
	}

	// 设置默认端口
	if toolConfig.PortNumber == nil {
		defaultPort := 22
		toolConfig.PortNumber = &defaultPort
	}

	// 设置默认认证类型
	if toolConfig.AuthType == nil {
		defaultAuth := "password"
		toolConfig.AuthType = &defaultAuth
	}

	// 设置默认协议类型
	if toolConfig.ProtocolType == nil {
		defaultProtocol := "SFTP"
		toolConfig.ProtocolType = &defaultProtocol
	}

	// 设置默认配置状态
	if toolConfig.ConfigStatus == "" {
		toolConfig.ConfigStatus = "Y"
	}

	// 设置默认值
	if toolConfig.DefaultFlag == "" {
		toolConfig.DefaultFlag = "N"
	}

	// 设置默认优先级
	if toolConfig.PriorityLevel == nil {
		defaultPriority := 100
		toolConfig.PriorityLevel = &defaultPriority
	}

	// 加密密码字段
	if toolConfig.PasswordEncrypted != nil && *toolConfig.PasswordEncrypted != "" {
		encryptedPassword, err := crypto.EncryptString(*toolConfig.PasswordEncrypted)
		if err != nil {
			logger.Error("密码加密失败", "error", err, "toolConfigId", toolConfig.ToolConfigId)
			response.ErrorJSON(ctx, "密码加密失败: "+err.Error(), constants.ED00009)
			return
		}
		toolConfig.PasswordEncrypted = &encryptedPassword
		logger.Info("密码加密成功", "toolConfigId", toolConfig.ToolConfigId)
	}

	// 加密私钥内容字段
	if toolConfig.KeyFileContent != nil && *toolConfig.KeyFileContent != "" {
		encryptedKeyContent, err := crypto.EncryptString(*toolConfig.KeyFileContent)
		if err != nil {
			logger.Error("私钥内容加密失败", "error", err, "toolConfigId", toolConfig.ToolConfigId)
			response.ErrorJSON(ctx, "私钥内容加密失败: "+err.Error(), constants.ED00009)
			return
		}
		toolConfig.KeyFileContent = &encryptedKeyContent
		logger.Info("私钥内容加密成功", "toolConfigId", toolConfig.ToolConfigId)
	}

	// 添加到数据库
	_, err := c.dao.Add(ctx, &toolConfig)
	if err != nil {
		logger.Error("添加工具配置失败", "error", err, "toolConfigId", toolConfig.ToolConfigId)
		response.ErrorJSON(ctx, "添加工具配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的工具配置信息
	newToolConfig, err := c.dao.GetById(ctx, tenantId, toolConfig.ToolConfigId)
	if err != nil {
		// 即使查询失败，也返回成功但只带有配置ID
		response.SuccessJSON(ctx, gin.H{
			"toolConfigId": toolConfig.ToolConfigId,
			"tenantId":     tenantId,
			"message":      "工具配置创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	response.SuccessJSON(ctx, newToolConfig, constants.SD00003)
}

// GetToolConfig 获取工具配置
// @Summary 获取工具配置
// @Description 根据ID获取工具配置详情
// @Tags SFTP配置管理
// @Accept json
// @Produce json
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /api/sftp/get [post]
func (c *ToolConfigController) GetToolConfig(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		ToolConfigId string `json:"toolConfigId" form:"toolConfigId" query:"toolConfigId"`
	}
	if err := request.BindSafely(ctx, &params); err != nil {
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 参数验证
	if params.ToolConfigId == "" {
		response.ErrorJSON(ctx, "工具配置ID不能为空", constants.ED00007)
		return
	}

	// 从数据库查询
	toolConfig, err := c.dao.GetById(ctx, tenantId, params.ToolConfigId)
	if err != nil {
		response.ErrorJSON(ctx, "获取工具配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if toolConfig == nil {
		response.ErrorJSON(ctx, "工具配置不存在", constants.ED00008)
		return
	}

	// 直接返回数据库中的加密数据，不进行脱敏处理
	response.SuccessJSON(ctx, toolConfig, constants.SD00001)
}

// UpdateToolConfig 更新工具配置
// @Summary 更新工具配置
// @Description 更新工具配置信息
// @Tags SFTP配置管理
// @Accept json
// @Produce json
// @Param data body models.ToolConfig true "工具配置信息"
// @Success 200 {object} response.Response
// @Router /api/sftp/update [post]
func (c *ToolConfigController) UpdateToolConfig(ctx *gin.Context) {
	// 解析请求参数
	var toolConfig models.ToolConfig
	if err := request.BindSafely(ctx, &toolConfig); err != nil {
		logger.Error("工具配置更新参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		logger.Error("工具配置更新无法获取租户信息")
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		logger.Error("工具配置更新无法获取操作人信息")
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 参数验证
	if toolConfig.ToolConfigId == "" {
		logger.Error("工具配置更新缺少配置ID")
		response.ErrorJSON(ctx, "工具配置ID不能为空", constants.ED00007)
		return
	}

	logger.Info("开始更新工具配置", "tenantId", tenantId, "toolConfigId", toolConfig.ToolConfigId, "operatorId", operatorId)

	// 查询原记录
	currentToolConfig, err := c.dao.GetById(ctx, tenantId, toolConfig.ToolConfigId)
	if err != nil {
		logger.Error("获取原工具配置失败", "error", err, "tenantId", tenantId, "toolConfigId", toolConfig.ToolConfigId)
		response.ErrorJSON(ctx, "获取原工具配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentToolConfig == nil {
		logger.Error("工具配置不存在", "tenantId", tenantId, "toolConfigId", toolConfig.ToolConfigId)
		response.ErrorJSON(ctx, "工具配置不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	toolConfigId := currentToolConfig.ToolConfigId
	tenantIdValue := currentToolConfig.TenantId
	addTime := currentToolConfig.AddTime
	addWho := currentToolConfig.AddWho

	// 处理密码字段：与数据库中的密码比对，不一致则加密保存
	if toolConfig.PasswordEncrypted != nil && *toolConfig.PasswordEncrypted != "" {
		var currentPassword string
		if currentToolConfig.PasswordEncrypted != nil {
			currentPassword = *currentToolConfig.PasswordEncrypted
		}
		
		// 直接比对密码是否与数据库中的不同
		if *toolConfig.PasswordEncrypted != currentPassword {
			// 密码不同，需要加密保存
			logger.Info("检测到密码变更，进行加密处理", "toolConfigId", toolConfig.ToolConfigId)
			encryptedPassword, err := crypto.EncryptString(*toolConfig.PasswordEncrypted)
			if err != nil {
				logger.Error("密码加密失败", "error", err, "toolConfigId", toolConfig.ToolConfigId)
				response.ErrorJSON(ctx, "密码加密失败: "+err.Error(), constants.ED00009)
				return
			}
			toolConfig.PasswordEncrypted = &encryptedPassword
			logger.Info("密码加密成功", "toolConfigId", toolConfig.ToolConfigId)
		} else {
			// 密码相同，保持不变
			logger.Debug("密码未变更，保持原值", "toolConfigId", toolConfig.ToolConfigId)
		}
	} else {
		// 前端未传递密码或密码为空，保持原有密码
		logger.Debug("未传递密码字段，保持原有密码", "toolConfigId", toolConfig.ToolConfigId)
		toolConfig.PasswordEncrypted = currentToolConfig.PasswordEncrypted
	}

	// 处理私钥内容字段：与数据库中的私钥比对，不一致则加密保存
	if toolConfig.KeyFileContent != nil && *toolConfig.KeyFileContent != "" {
		var currentKeyContent string
		if currentToolConfig.KeyFileContent != nil {
			currentKeyContent = *currentToolConfig.KeyFileContent
		}
		
		// 直接比对私钥内容是否与数据库中的不同
		if *toolConfig.KeyFileContent != currentKeyContent {
			// 私钥内容不同，需要加密保存
			logger.Info("检测到私钥内容变更，进行加密处理", "toolConfigId", toolConfig.ToolConfigId)
			encryptedKeyContent, err := crypto.EncryptString(*toolConfig.KeyFileContent)
			if err != nil {
				logger.Error("私钥内容加密失败", "error", err, "toolConfigId", toolConfig.ToolConfigId)
				response.ErrorJSON(ctx, "私钥内容加密失败: "+err.Error(), constants.ED00009)
				return
			}
			toolConfig.KeyFileContent = &encryptedKeyContent
			logger.Info("私钥内容加密成功", "toolConfigId", toolConfig.ToolConfigId)
		} else {
			// 私钥内容相同，保持不变
			logger.Debug("私钥内容未变更，保持原值", "toolConfigId", toolConfig.ToolConfigId)
		}
	} else {
		// 前端未传递私钥内容或私钥内容为空，保持原有私钥内容
		logger.Debug("未传递私钥内容字段，保持原有私钥内容", "toolConfigId", toolConfig.ToolConfigId)
		toolConfig.KeyFileContent = currentToolConfig.KeyFileContent
	}

	// 强制设置从上下文获取的租户ID和操作人信息
	toolConfig.TenantId = tenantIdValue // 强制使用数据库中的租户ID
	toolConfig.EditWho = operatorId
	toolConfig.EditTime = time.Now()

	// 强制恢复不可修改的字段，防止前端恶意修改
	toolConfig.ToolConfigId = toolConfigId
	toolConfig.AddTime = addTime
	toolConfig.AddWho = addWho

	// 更新OprSeqFlag
	toolConfig.OprSeqFlag = random.Generate32BitRandomString()

	// 更新数据库
	updateResult, err := c.dao.Update(ctx, &toolConfig)
	if err != nil {
		logger.Error("更新工具配置失败", "error", err, "toolConfigId", toolConfig.ToolConfigId)
		response.ErrorJSON(ctx, "更新工具配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.Info("工具配置更新成功", "toolConfigId", toolConfig.ToolConfigId, "updateResult", updateResult)

	// 查询最新数据
	updatedToolConfig, err := c.dao.GetById(ctx, tenantId, toolConfig.ToolConfigId)
	if err != nil {
		logger.Error("获取更新后的工具配置失败", "error", err, "toolConfigId", toolConfig.ToolConfigId)
		response.ErrorJSON(ctx, "获取更新后的工具配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 直接返回数据库中的数据，不进行脱敏处理
	logger.Info("工具配置更新流程完成", "toolConfigId", toolConfig.ToolConfigId)
	response.SuccessJSON(ctx, updatedToolConfig, constants.SD00004)
}

// DeleteToolConfig 删除工具配置
// @Summary 删除工具配置
// @Description 删除工具配置
// @Tags SFTP配置管理
// @Accept json
// @Produce json
// @Param data body object true "删除参数"
// @Success 200 {object} response.Response
// @Router /api/sftp/delete [post]
func (c *ToolConfigController) DeleteToolConfig(ctx *gin.Context) {
	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)
	toolConfigId := request.GetParam(ctx, "toolConfigId")

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 参数验证
	if toolConfigId == "" {
		response.ErrorJSON(ctx, "工具配置ID不能为空", constants.ED00007)
		return
	}

	// 删除记录
	_, err := c.dao.Delete(ctx, tenantId, toolConfigId, operatorId)
	if err != nil {
		response.ErrorJSON(ctx, "删除工具配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"toolConfigId": toolConfigId,
		"message":      "工具配置删除成功",
	}, constants.SD00005)
}

// QueryToolConfigs 查询工具配置列表
// @Summary 查询工具配置列表
// @Description 根据条件查询工具配置列表
// @Tags SFTP配置管理
// @Accept json
// @Produce json
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /api/sftp/query [post]
func (c *ToolConfigController) QueryToolConfigs(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		ToolName      string `json:"toolName" form:"toolName" query:"toolName"`
		ToolType      string `json:"toolType" form:"toolType" query:"toolType"`
		ConfigName    string `json:"configName" form:"configName" query:"configName"`
		ConfigGroupId string `json:"configGroupId" form:"configGroupId" query:"configGroupId"`
		HostAddress   string `json:"hostAddress" form:"hostAddress" query:"hostAddress"`
	}
	if err := request.BindSafely(ctx, &params); err != nil {
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 构建查询条件，强制使用从上下文获取的租户ID
	queryParams := map[string]interface{}{
		"tenantId":      tenantId,
		"toolName":      params.ToolName,
		"toolType":      params.ToolType,
		"configName":    params.ConfigName,
		"configGroupId": params.ConfigGroupId,
		"hostAddress":   params.HostAddress,
	}

	// 查询数据
	toolConfigs, total, err := c.dao.Query(ctx, queryParams, page, pageSize)
	if err != nil {
		response.ErrorJSON(ctx, "查询工具配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 直接返回数据库中的数据，不进行脱敏处理

	// 构建分页信息
	pageInfo := response.PageInfo{
		PageIndex:      page,
		PageSize:       pageSize,
		TotalCount:     int(total),
		TotalPageIndex: int((total + int64(pageSize) - 1) / int64(pageSize)),
		CurPageCount:   len(toolConfigs),
	}

	response.PageJSON(ctx, toolConfigs, pageInfo, constants.SD00002)
}
 