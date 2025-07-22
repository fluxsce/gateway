package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/utils/random"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	hub0003dao "gateway/web/views/hub0003/dao"
	hub0003models "gateway/web/views/hub0003/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SchedulerConfigController 调度器配置控制器
type SchedulerConfigController struct {
	dao *hub0003dao.SchedulerDao
}

// NewSchedulerConfigController 创建调度器配置控制器
func NewSchedulerConfigController(db database.Database) *SchedulerConfigController {
	return &SchedulerConfigController{
		dao: hub0003dao.NewSchedulerDao(db),
	}
}

// AddSchedulerConfig 添加调度器配置
// @Summary 添加调度器配置
// @Description 添加新的调度器配置
// @Tags 调度器管理
// @Accept json
// @Produce json
// @Param data body models.TimerScheduler true "调度器配置信息"
// @Success 200 {object} response.Response
// @Router /gateway/hub0003/scheduler/add [post]
func (c *SchedulerConfigController) AddSchedulerConfig(ctx *gin.Context) {
	// 解析请求参数
	var scheduler hub0003models.TimerScheduler
	if err := request.BindSafely(ctx, &scheduler); err != nil {
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
	if scheduler.SchedulerName == "" {
		response.ErrorJSON(ctx, "调度器名称不能为空", constants.ED00007)
		return
	}

	// 强制设置从上下文获取的租户ID和操作人信息
	scheduler.TenantId = tenantId
	scheduler.AddWho = operatorId
	scheduler.EditWho = operatorId

	// 生成调度器ID (32位长度限制)
	if scheduler.SchedulerId == "" {
		// 使用UUID去掉连字符，确保长度为32位
		scheduler.SchedulerId = strings.ReplaceAll(uuid.New().String(), "-", "")
	}

	// 设置默认值
	now := time.Now()
	scheduler.AddTime = now
	scheduler.EditTime = now
	scheduler.CurrentVersion = 1
	scheduler.ActiveFlag = "Y"

	// 生成OprSeqFlag
	scheduler.OprSeqFlag = random.Generate32BitRandomString()

	// 如果未指定实例ID，使用UUID生成 (注意长度限制)
	if scheduler.SchedulerInstanceId == "" {
		// 生成不超过字段长度限制的实例ID
		uuidStr := strings.ReplaceAll(uuid.New().String(), "-", "")
		scheduler.SchedulerInstanceId = "INST_" + uuidStr[:24] // INST_ + 24位UUID = 28位
	}

	// 如果未指定状态，默认为停止状态
	if scheduler.SchedulerStatus == 0 {
		scheduler.SchedulerStatus = 1 // 停止状态
	}

	// 设置默认配置
	if scheduler.MaxWorkers == 0 {
		scheduler.MaxWorkers = 5
	}
	if scheduler.QueueSize == 0 {
		scheduler.QueueSize = 100
	}
	if scheduler.DefaultTimeoutSeconds == 0 {
		scheduler.DefaultTimeoutSeconds = 1800 // 30分钟
	}
	if scheduler.DefaultRetries == 0 {
		scheduler.DefaultRetries = 3
	}

	// 添加到数据库
	_, err := c.dao.Add(ctx, &scheduler)
	if err != nil {
		response.ErrorJSON(ctx, "添加调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的调度器信息
	newScheduler, err := c.dao.GetById(ctx, tenantId, scheduler.SchedulerId)
	if err != nil {
		// 即使查询失败，也返回成功但只带有调度器ID
		response.SuccessJSON(ctx, gin.H{
			"schedulerId": scheduler.SchedulerId,
			"tenantId":    tenantId,
			"message":     "调度器配置创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	response.SuccessJSON(ctx, newScheduler, constants.SD00003)
}

// GetSchedulerConfig 获取调度器配置
// @Summary 获取调度器配置
// @Description 根据ID获取调度器配置详情
// @Tags 调度器管理
// @Accept json
// @Produce json
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /gateway/hub0003/scheduler/get [post]
func (c *SchedulerConfigController) GetSchedulerConfig(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		SchedulerId string `json:"schedulerId" form:"schedulerId" query:"schedulerId"`
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
	if params.SchedulerId == "" {
		response.ErrorJSON(ctx, "调度器ID不能为空", constants.ED00007)
		return
	}

	// 从数据库查询
	scheduler, err := c.dao.GetById(ctx, tenantId, params.SchedulerId)
	if err != nil {
		response.ErrorJSON(ctx, "获取调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, scheduler, constants.SD00001)
}

// UpdateSchedulerConfig 更新调度器配置
// @Summary 更新调度器配置
// @Description 更新调度器配置信息
// @Tags 调度器管理
// @Accept json
// @Produce json
// @Param data body models.TimerScheduler true "调度器配置信息"
// @Success 200 {object} response.Response
// @Router /gateway/hub0003/scheduler/update [post]
func (c *SchedulerConfigController) UpdateSchedulerConfig(ctx *gin.Context) {
	// 解析请求参数
	var scheduler hub0003models.TimerScheduler
	if err := request.BindSafely(ctx, &scheduler); err != nil {
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
	if scheduler.SchedulerId == "" {
		response.ErrorJSON(ctx, "调度器ID不能为空", constants.ED00007)
		return
	}

	// 查询原记录
	currentScheduler, err := c.dao.GetById(ctx, tenantId, scheduler.SchedulerId)
	if err != nil {
		response.ErrorJSON(ctx, "获取原调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentScheduler == nil {
		response.ErrorJSON(ctx, "调度器配置不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	schedulerId := currentScheduler.SchedulerId
	tenantIdValue := currentScheduler.TenantId
	addTime := currentScheduler.AddTime
	addWho := currentScheduler.AddWho

	// 强制设置从上下文获取的租户ID和操作人信息
	scheduler.TenantId = tenantIdValue // 强制使用数据库中的租户ID
	scheduler.EditWho = operatorId
	scheduler.EditTime = time.Now()

	// 强制恢复不可修改的字段，防止前端恶意修改
	scheduler.SchedulerId = schedulerId
	scheduler.AddTime = addTime
	scheduler.AddWho = addWho

	// 更新OprSeqFlag
	scheduler.OprSeqFlag = random.Generate32BitRandomString()

	// 更新数据库
	_, err = c.dao.Update(ctx, &scheduler)
	if err != nil {
		response.ErrorJSON(ctx, "更新调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新数据
	updatedScheduler, err := c.dao.GetById(ctx, tenantId, scheduler.SchedulerId)
	if err != nil {
		response.ErrorJSON(ctx, "获取更新后的调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, updatedScheduler, constants.SD00004)
}

// DeleteSchedulerConfig 删除调度器配置
// @Summary 删除调度器配置
// @Description 删除调度器配置
// @Tags 调度器管理
// @Accept json
// @Produce json
// @Param data body object true "删除参数"
// @Success 200 {object} response.Response
// @Router /gateway/hub0003/scheduler/delete [post]
func (c *SchedulerConfigController) DeleteSchedulerConfig(ctx *gin.Context) {
	// 解析请求参数
	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)
	schedulerId := request.GetParam(ctx, "schedulerId")
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
	if schedulerId == "" {
		response.ErrorJSON(ctx, "调度器ID不能为空", constants.ED00007)
		return
	}

	// 删除记录
	_, err := c.dao.Delete(ctx, tenantId, schedulerId, operatorId)
	if err != nil {
		response.ErrorJSON(ctx, "删除调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"schedulerId": schedulerId,
		"message":     "调度器配置删除成功",
	}, constants.SD00005)
}

// QuerySchedulerConfigs 查询调度器配置列表
// @Summary 查询调度器配置列表
// @Description 根据条件查询调度器配置列表
// @Tags 调度器管理
// @Accept json
// @Produce json
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /gateway/hub0003/scheduler/query [post]
func (c *SchedulerConfigController) QuerySchedulerConfigs(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		SchedulerName   string `json:"schedulerName" form:"schedulerName" query:"schedulerName"`
		SchedulerStatus int    `json:"schedulerStatus" form:"schedulerStatus" query:"schedulerStatus"`
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
		"tenantId":        tenantId,
		"schedulerName":   params.SchedulerName,
		"schedulerStatus": params.SchedulerStatus,
	}

	// 查询数据
	schedulers, total, err := c.dao.Query(ctx, queryParams, page, pageSize)
	if err != nil {
		response.ErrorJSON(ctx, "查询调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建分页信息
	pageInfo := response.PageInfo{
		PageIndex:      page,
		PageSize:       pageSize,
		TotalCount:     int(total),
		TotalPageIndex: int((total + int64(pageSize) - 1) / int64(pageSize)),
		CurPageCount:   len(schedulers),
	}

	response.PageJSON(ctx, schedulers, pageInfo, constants.SD00002)
}

// UpdateSchedulerStatus 更新调度器状态
// @Summary 更新调度器状态
// @Description 更新调度器运行状态
// @Tags 调度器管理
// @Accept json
// @Produce json
// @Param data body object true "状态更新参数"
// @Success 200 {object} response.Response
// @Router /gateway/hub0003/scheduler/update-status [post]
func (c *SchedulerConfigController) UpdateSchedulerStatus(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		SchedulerId     string `json:"schedulerId" form:"schedulerId" query:"schedulerId"`
		SchedulerStatus int    `json:"schedulerStatus" form:"schedulerStatus" query:"schedulerStatus"`
	}
	if err := request.BindSafely(ctx, &params); err != nil {
		response.ErrorJSON(ctx, "参数解析失败: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
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
	if params.SchedulerId == "" {
		response.ErrorJSON(ctx, "调度器ID不能为空", constants.ED00007)
		return
	}
	if params.SchedulerStatus <= 0 {
		response.ErrorJSON(ctx, "调度器状态无效", constants.ED00007)
		return
	}

	// 更新状态
	_, err := c.dao.UpdateStatus(ctx, tenantId, params.SchedulerId, params.SchedulerStatus, operatorId)
	if err != nil {
		response.ErrorJSON(ctx, "更新调度器状态失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新数据
	updatedScheduler, err := c.dao.GetById(ctx, tenantId, params.SchedulerId)
	if err != nil {
		response.ErrorJSON(ctx, "获取更新后的调度器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, updatedScheduler, constants.SD00004)
}
