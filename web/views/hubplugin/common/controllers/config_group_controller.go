package controllers

import (
	"gohub/pkg/database"
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

// ConfigGroupController 工具配置分组控制器
type ConfigGroupController struct {
	dao *dao.ConfigGroupDao
}

// NewConfigGroupController 创建工具配置分组控制器
func NewConfigGroupController(db database.Database) *ConfigGroupController {
	return &ConfigGroupController{
		dao: dao.NewConfigGroupDao(db),
	}
}

// AddConfigGroup 添加工具配置分组
// @Summary 添加工具配置分组
// @Description 添加新的工具配置分组
// @Tags SFTP配置分组管理
// @Accept json
// @Produce json
// @Param data body models.ToolConfigGroup true "配置分组信息"
// @Success 200 {object} response.Response
// @Router /api/sftp/config-group/add [post]
func (c *ConfigGroupController) AddConfigGroup(ctx *gin.Context) {
	// 解析请求参数
	var configGroup models.ToolConfigGroup
	if err := request.BindSafely(ctx, &configGroup); err != nil {
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
	if configGroup.GroupName == "" {
		response.ErrorJSON(ctx, "分组名称不能为空", constants.ED00007)
		return
	}

	// 强制设置从上下文获取的租户ID和操作人信息
	configGroup.TenantId = tenantId
	configGroup.AddWho = operatorId
	configGroup.EditWho = operatorId

	// 生成配置分组ID (32位长度限制)
	if configGroup.ConfigGroupId == "" {
		// 使用UUID去掉连字符，确保长度为32位
		configGroup.ConfigGroupId = strings.ReplaceAll(uuid.New().String(), "-", "")
	}

	// 设置默认值
	now := time.Now()
	configGroup.AddTime = now
	configGroup.EditTime = now
	configGroup.CurrentVersion = 1
	configGroup.ActiveFlag = "Y"

	// 生成OprSeqFlag
	configGroup.OprSeqFlag = random.Generate32BitRandomString()

	// 设置默认排序
	if configGroup.SortOrder == nil {
		defaultSort := 100
		configGroup.SortOrder = &defaultSort
	}

	// 设置默认分组级别
	if configGroup.GroupLevel == nil {
		defaultLevel := 1
		configGroup.GroupLevel = &defaultLevel
	}

	// 设置默认分组类型
	if configGroup.GroupType == nil {
		defaultType := "SFTP"
		configGroup.GroupType = &defaultType
	}

	// 设置默认访问级别
	if configGroup.AccessLevel == nil {
		defaultAccess := "private"
		configGroup.AccessLevel = &defaultAccess
	}

	// 添加到数据库
	_, err := c.dao.Add(ctx, &configGroup)
	if err != nil {
		response.ErrorJSON(ctx, "添加配置分组失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的配置分组信息
	newConfigGroup, err := c.dao.GetById(ctx, tenantId, configGroup.ConfigGroupId)
	if err != nil {
		// 即使查询失败，也返回成功但只带有分组ID
		response.SuccessJSON(ctx, gin.H{
			"configGroupId": configGroup.ConfigGroupId,
			"tenantId":      tenantId,
			"message":       "配置分组创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	response.SuccessJSON(ctx, newConfigGroup, constants.SD00003)
}

// GetConfigGroup 获取工具配置分组
// @Summary 获取工具配置分组
// @Description 根据ID获取工具配置分组详情
// @Tags SFTP配置分组管理
// @Accept json
// @Produce json
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /api/sftp/config-group/get [post]
func (c *ConfigGroupController) GetConfigGroup(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		ConfigGroupId string `json:"configGroupId" form:"configGroupId" query:"configGroupId"`
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
	if params.ConfigGroupId == "" {
		response.ErrorJSON(ctx, "配置分组ID不能为空", constants.ED00007)
		return
	}

	// 从数据库查询
	configGroup, err := c.dao.GetById(ctx, tenantId, params.ConfigGroupId)
	if err != nil {
		response.ErrorJSON(ctx, "获取配置分组失败: "+err.Error(), constants.ED00009)
		return
	}

	if configGroup == nil {
		response.ErrorJSON(ctx, "配置分组不存在", constants.ED00008)
		return
	}

	response.SuccessJSON(ctx, configGroup, constants.SD00001)
}

// UpdateConfigGroup 更新工具配置分组
// @Summary 更新工具配置分组
// @Description 更新工具配置分组信息
// @Tags SFTP配置分组管理
// @Accept json
// @Produce json
// @Param data body models.ToolConfigGroup true "配置分组信息"
// @Success 200 {object} response.Response
// @Router /api/sftp/config-group/update [post]
func (c *ConfigGroupController) UpdateConfigGroup(ctx *gin.Context) {
	// 解析请求参数
	var configGroup models.ToolConfigGroup
	if err := request.BindSafely(ctx, &configGroup); err != nil {
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
	if configGroup.ConfigGroupId == "" {
		response.ErrorJSON(ctx, "配置分组ID不能为空", constants.ED00007)
		return
	}

	// 查询原记录
	currentConfigGroup, err := c.dao.GetById(ctx, tenantId, configGroup.ConfigGroupId)
	if err != nil {
		response.ErrorJSON(ctx, "获取原配置分组失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentConfigGroup == nil {
		response.ErrorJSON(ctx, "配置分组不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	configGroupId := currentConfigGroup.ConfigGroupId
	tenantIdValue := currentConfigGroup.TenantId
	addTime := currentConfigGroup.AddTime
	addWho := currentConfigGroup.AddWho

	// 强制设置从上下文获取的租户ID和操作人信息
	configGroup.TenantId = tenantIdValue // 强制使用数据库中的租户ID
	configGroup.EditWho = operatorId
	configGroup.EditTime = time.Now()

	// 强制恢复不可修改的字段，防止前端恶意修改
	configGroup.ConfigGroupId = configGroupId
	configGroup.AddTime = addTime
	configGroup.AddWho = addWho

	// 更新OprSeqFlag
	configGroup.OprSeqFlag = random.Generate32BitRandomString()

	// 更新数据库
	_, err = c.dao.Update(ctx, &configGroup)
	if err != nil {
		response.ErrorJSON(ctx, "更新配置分组失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新数据
	updatedConfigGroup, err := c.dao.GetById(ctx, tenantId, configGroup.ConfigGroupId)
	if err != nil {
		response.ErrorJSON(ctx, "获取更新后的配置分组失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, updatedConfigGroup, constants.SD00004)
}

// DeleteConfigGroup 删除工具配置分组
// @Summary 删除工具配置分组
// @Description 删除工具配置分组
// @Tags SFTP配置分组管理
// @Accept json
// @Produce json
// @Param data body object true "删除参数"
// @Success 200 {object} response.Response
// @Router /api/sftp/config-group/delete [post]
func (c *ConfigGroupController) DeleteConfigGroup(ctx *gin.Context) {
	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)
	configGroupId := request.GetParam(ctx, "configGroupId")

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
	if configGroupId == "" {
		response.ErrorJSON(ctx, "配置分组ID不能为空", constants.ED00007)
		return
	}

	// 删除记录
	_, err := c.dao.Delete(ctx, tenantId, configGroupId, operatorId)
	if err != nil {
		response.ErrorJSON(ctx, "删除配置分组失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"configGroupId": configGroupId,
		"message":       "配置分组删除成功",
	}, constants.SD00005)
}

// QueryConfigGroups 查询工具配置分组列表
// @Summary 查询工具配置分组列表
// @Description 根据条件查询工具配置分组列表
// @Tags SFTP配置分组管理
// @Accept json
// @Produce json
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /api/sftp/config-group/query [post]
func (c *ConfigGroupController) QueryConfigGroups(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		GroupName     string `json:"groupName" form:"groupName" query:"groupName"`
		GroupType     string `json:"groupType" form:"groupType" query:"groupType"`
		ParentGroupId string `json:"parentGroupId" form:"parentGroupId" query:"parentGroupId"`
		Page          int    `json:"page" form:"page" query:"page"`
		PageSize      int    `json:"pageSize" form:"pageSize" query:"pageSize"`
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

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// 构建查询条件，强制使用从上下文获取的租户ID
	queryParams := map[string]interface{}{
		"tenantId":      tenantId,
		"groupName":     params.GroupName,
		"groupType":     params.GroupType,
		"parentGroupId": params.ParentGroupId,
	}

	// 查询数据
	configGroups, total, err := c.dao.Query(ctx, queryParams, params.Page, params.PageSize)
	if err != nil {
		response.ErrorJSON(ctx, "查询配置分组失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建分页信息
	pageInfo := response.PageInfo{
		PageIndex:      params.Page,
		PageSize:       params.PageSize,
		TotalCount:     int(total),
		TotalPageIndex: int((total + int64(params.PageSize) - 1) / int64(params.PageSize)),
		CurPageCount:   len(configGroups),
	}

	response.PageJSON(ctx, configGroups, pageInfo, constants.SD00002)
} 