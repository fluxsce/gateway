package controllers

import (
	"context"
	"fmt"
	"gohub/internal/timerinit/sftp"
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/pkg/timer"
	"gohub/pkg/utils/random"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	hub0003dao "gohub/web/views/hub0003/dao"
	hub0003models "gohub/web/views/hub0003/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskConfigController 任务配置控制器
type TaskConfigController struct {
	dao *hub0003dao.TaskDao
	db  database.Database
}

// NewTaskConfigController 创建任务配置控制器
func NewTaskConfigController(db database.Database) *TaskConfigController {
	return &TaskConfigController{
		dao: hub0003dao.NewTaskDao(db),
		db:  db,
	}
}

// BasicTaskExecutor 基本任务执行器实现
type BasicTaskExecutor struct{}

// Execute 执行任务
func (e *BasicTaskExecutor) Execute(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	// 这里可以根据实际需求实现任务执行逻辑
	// 目前只是一个占位实现
	result := &timer.ExecuteResult{
		Success: true,
		Message: "任务执行成功",
	}
	return result, nil
}

// GetName 获取执行器名称
func (e *BasicTaskExecutor) GetName() string {
	return "BasicTaskExecutor"
}

// Close 关闭执行器
func (e *BasicTaskExecutor) Close() error {
	return nil
}

// AddTaskConfig 添加任务配置
// @Summary 添加任务配置
// @Description 添加新的任务配置
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param data body models.TimerTask true "任务配置信息"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/task/add [post]
func (c *TaskConfigController) AddTaskConfig(ctx *gin.Context) {
	// 解析请求参数
	var task hub0003models.TimerTask
	if err := request.BindSafely(ctx, &task); err != nil {
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
	if task.SchedulerId == nil || *task.SchedulerId == "" {
		response.ErrorJSON(ctx, "调度器ID不能为空", constants.ED00007)
		return
	}
	if task.TaskName == "" {
		response.ErrorJSON(ctx, "任务名称不能为空", constants.ED00007)
		return
	}
	if (task.CronExpression == nil || *task.CronExpression == "") && (task.IntervalSeconds == nil || *task.IntervalSeconds == 0) {
		response.ErrorJSON(ctx, "必须指定Cron表达式或固定频率", constants.ED00007)
		return
	}

	// 强制设置从上下文获取的租户ID和操作人信息
	task.TenantId = tenantId
	task.AddWho = operatorId
	task.EditWho = operatorId

	// 生成任务ID (32位长度限制)
	if task.TaskId == "" {
		// 使用UUID去掉连字符，确保长度为32位
		task.TaskId = strings.ReplaceAll(uuid.New().String(), "-", "")
	}

	// 设置默认值
	now := time.Now()
	task.AddTime = now
	task.EditTime = now
	task.CurrentVersion = 1
	task.ActiveFlag = "Y"

	// 生成OprSeqFlag
	task.OprSeqFlag = random.Generate32BitRandomString()

	// 如果未指定状态，默认为停止状态
	if task.TaskStatus == 0 {
		task.TaskStatus = 1 // 停止状态
	}

	// 设置默认超时时间和重试次数
	if task.TimeoutSeconds == 0 {
		task.TimeoutSeconds = 1800 // 30分钟
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}

	// 添加到数据库
	_, err := c.dao.Add(ctx, &task)
	if err != nil {
		response.ErrorJSON(ctx, "添加任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的任务信息
	newTask, err := c.dao.GetById(ctx, tenantId, task.TaskId)
	if err != nil {
		// 即使查询失败，也返回成功但只带有任务ID
		response.SuccessJSON(ctx, gin.H{
			"taskId":   task.TaskId,
			"tenantId": tenantId,
			"message":  "任务配置创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	response.SuccessJSON(ctx, newTask, constants.SD00003)
}

// GetTaskConfig 获取任务配置
// @Summary 获取任务配置
// @Description 根据ID获取任务配置详情
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/task/get [post]
func (c *TaskConfigController) GetTaskConfig(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		TaskId string `json:"taskId" form:"taskId" query:"taskId"`
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
	if params.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}

	// 从数据库查询
	task, err := c.dao.GetById(ctx, tenantId, params.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "获取任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, task, constants.SD00001)
}

// UpdateTaskConfig 更新任务配置
// @Summary 更新任务配置
// @Description 更新任务配置信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param data body models.TimerTask true "任务配置信息"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/task/update [post]
func (c *TaskConfigController) UpdateTaskConfig(ctx *gin.Context) {
	// 解析请求参数
	var task hub0003models.TimerTask
	if err := request.BindSafely(ctx, &task); err != nil {
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
	if task.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}

	// 查询原记录
	currentTask, err := c.dao.GetById(ctx, tenantId, task.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "获取原任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentTask == nil {
		response.ErrorJSON(ctx, "任务配置不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	taskId := currentTask.TaskId
	tenantIdValue := currentTask.TenantId
	addTime := currentTask.AddTime
	addWho := currentTask.AddWho

	// 强制设置从上下文获取的租户ID和操作人信息
	task.TenantId = tenantIdValue // 强制使用数据库中的租户ID
	task.EditWho = operatorId
	task.EditTime = time.Now()

	// 强制恢复不可修改的字段，防止前端恶意修改
	task.TaskId = taskId
	task.AddTime = addTime
	task.AddWho = addWho

	// 更新OprSeqFlag
	task.OprSeqFlag = random.Generate32BitRandomString()

	// 更新数据库
	_, err = c.dao.Update(ctx, &task)
	if err != nil {
		response.ErrorJSON(ctx, "更新任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新数据
	updatedTask, err := c.dao.GetById(ctx, tenantId, task.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "获取更新后的任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, updatedTask, constants.SD00004)
}

// DeleteTaskConfig 删除任务配置
// @Summary 删除任务配置
// @Description 删除任务配置
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param data body object true "删除参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/task/delete [post]
func (c *TaskConfigController) DeleteTaskConfig(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		TaskId string `json:"taskId" form:"taskId" query:"taskId"`
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
	if params.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}

	// 删除记录
	_, err := c.dao.Delete(ctx, tenantId, params.TaskId, operatorId)
	if err != nil {
		response.ErrorJSON(ctx, "删除任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"taskId":  params.TaskId,
		"message": "任务配置删除成功",
	}, constants.SD00005)
}

// QueryTaskConfigs 查询任务配置列表
// @Summary 查询任务配置列表
// @Description 根据条件查询任务配置列表
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param data body object true "查询参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/task/query [post]
func (c *TaskConfigController) QueryTaskConfigs(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		SchedulerId string `json:"schedulerId" form:"schedulerId" query:"schedulerId"`
		TaskName    string `json:"taskName" form:"taskName" query:"taskName"`
		TaskStatus  int    `json:"taskStatus" form:"taskStatus" query:"taskStatus"`
		Page        int    `json:"page" form:"page" query:"page"`
		PageSize    int    `json:"pageSize" form:"pageSize" query:"pageSize"`
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
		"tenantId":    tenantId,
		"schedulerId": params.SchedulerId,
		"taskName":    params.TaskName,
		"taskStatus":  params.TaskStatus,
	}

	// 查询数据
	tasks, total, err := c.dao.Query(ctx, queryParams, params.Page, params.PageSize)
	if err != nil {
		response.ErrorJSON(ctx, "查询任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建分页信息
	pageInfo := response.PageInfo{
		PageIndex:      params.Page,
		PageSize:       params.PageSize,
		TotalCount:     int(total),
		TotalPageIndex: int((total + int64(params.PageSize) - 1) / int64(params.PageSize)),
		CurPageCount:   len(tasks),
	}

	response.PageJSON(ctx, tasks, pageInfo, constants.SD00002)
}

// UpdateTaskStatus 更新任务状态
// @Summary 更新任务状态
// @Description 更新任务运行状态
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param data body object true "状态更新参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/task/update-status [post]
func (c *TaskConfigController) UpdateTaskStatus(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		TaskId     string `json:"taskId" form:"taskId" query:"taskId"`
		TaskStatus int    `json:"taskStatus" form:"taskStatus" query:"taskStatus"`
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
	if params.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}
	if params.TaskStatus <= 0 {
		response.ErrorJSON(ctx, "任务状态无效", constants.ED00007)
		return
	}

	// 获取原任务配置
	task, err := c.dao.GetById(ctx, tenantId, params.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "获取任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if task == nil {
		response.ErrorJSON(ctx, "任务配置不存在", constants.ED00008)
		return
	}

	// 更新状态
	task.TaskStatus = params.TaskStatus
	task.EditTime = time.Now()
	task.EditWho = operatorId

	// 更新OprSeqFlag
	task.OprSeqFlag = random.Generate32BitRandomString()

	// 更新数据库
	_, err = c.dao.Update(ctx, task)
	if err != nil {
		response.ErrorJSON(ctx, "更新任务状态失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新数据
	updatedTask, err := c.dao.GetById(ctx, tenantId, task.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "获取更新后的任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, updatedTask, constants.SD00004)
}

// StartTask 启动任务
// @Summary 启动任务
// @Description 启动指定的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param data body object true "启动任务参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/task/start [post]
func (c *TaskConfigController) StartTask(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		TaskId      string `json:"taskId" form:"taskId" query:"taskId"`
		SchedulerId string `json:"schedulerId" form:"schedulerId" query:"schedulerId"`
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
	if params.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}
	if params.SchedulerId == "" {
		response.ErrorJSON(ctx, "调度器ID不能为空", constants.ED00007)
		return
	}

	// 获取任务配置
	task, err := c.dao.GetById(ctx, tenantId, params.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "获取任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if task == nil {
		response.ErrorJSON(ctx, "任务不存在", constants.ED00008)
		return
	}

	// 验证执行器类型
	if task.ExecutorType == nil || *task.ExecutorType == "" {
		response.ErrorJSON(ctx, "任务执行器类型不能为空", constants.ED00007)
		return
	}

	// 根据任务的执行器类型进行任务注册
	switch *task.ExecutorType {
	case "SFTP_TRANSFER":
		// 注册单个SFTP任务（包括调度器创建、任务注册、启动等完整流程）
		if err := sftp.RegisterSFTPTaskById(ctx, c.db, tenantId, task.TaskId); err != nil {
			response.ErrorJSON(ctx, "注册SFTP任务失败: "+err.Error(), constants.ED00009)
			return
		}
		logger.Info("SFTP任务注册成功", "taskId", task.TaskId, "tenantId", tenantId)
	case "HTTP_REQUEST":
		// 这里可以添加HTTP请求任务的注册逻辑
		// TODO: 实现HTTP请求任务注册
		response.ErrorJSON(ctx, "HTTP请求任务注册功能尚未实现", constants.ED00009)
		return
	default:
		response.ErrorJSON(ctx, fmt.Sprintf("不支持的执行器类型: %s", *task.ExecutorType), constants.ED00007)
		return
	}

	// 获取全局定时器池
	timerPool := timer.GetTimerPool()

	// 获取调度器（注册完成后应该已存在）
	scheduler, err := timerPool.GetScheduler(params.SchedulerId)
	if err != nil {
		response.ErrorJSON(ctx, "获取调度器失败: "+err.Error(), constants.ED00009)
		return
	}

	// 启动调度器（如果还没有启动）
	if !scheduler.IsRunning() {
		err = scheduler.Start()
		if err != nil {
			response.ErrorJSON(ctx, "启动调度器失败: "+err.Error(), constants.ED00009)
			return
		}
		logger.Info("调度器启动成功", "schedulerId", params.SchedulerId)
	}

	// 启动任务
	err = scheduler.StartTask(task.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "启动任务失败: "+err.Error(), constants.ED00009)
		return
	}
	logger.Info("任务启动成功", "taskId", task.TaskId)

	// 更新任务状态为运行中
	task.TaskStatus = 2 // 2-运行中
	task.EditTime = time.Now()
	task.EditWho = operatorId

	// 更新OprSeqFlag
	task.OprSeqFlag = random.Generate32BitRandomString()

	// 更新数据库
	_, err = c.dao.Update(ctx, task)
	if err != nil {
		response.ErrorJSON(ctx, "更新任务状态失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新数据
	updatedTask, err := c.dao.GetById(ctx, tenantId, task.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "获取更新后的任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, updatedTask, constants.SD00004)
}

// StopTask 停止任务
// @Summary 停止任务
// @Description 停止指定的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param data body object true "停止任务参数"
// @Success 200 {object} response.Response
// @Router /gohub/hub0003/task/stop [post]
func (c *TaskConfigController) StopTask(ctx *gin.Context) {
	// 解析请求参数
	var params struct {
		TaskId      string `json:"taskId" form:"taskId" query:"taskId"`
		SchedulerId string `json:"schedulerId" form:"schedulerId" query:"schedulerId"`
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
	if params.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}
	if params.SchedulerId == "" {
		response.ErrorJSON(ctx, "调度器ID不能为空", constants.ED00007)
		return
	}

	// 获取任务配置
	task, err := c.dao.GetById(ctx, tenantId, params.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "获取任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if task == nil {
		response.ErrorJSON(ctx, "任务不存在", constants.ED00008)
		return
	}

	// 获取全局定时器池
	timerPool := timer.GetTimerPool()

	// 获取调度器
	scheduler, err := timerPool.GetScheduler(params.SchedulerId)
	if err != nil {
		response.ErrorJSON(ctx, "获取调度器失败: "+err.Error(), constants.ED00009)
		return
	}

	// 停止任务
	err = scheduler.StopTask(task.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "停止任务失败: "+err.Error(), constants.ED00009)
		return
	}
	// 更新任务状态为已停止
	task.TaskStatus = 1 // 1-已停止
	task.EditTime = time.Now()
	task.EditWho = operatorId

	// 更新OprSeqFlag
	task.OprSeqFlag = random.Generate32BitRandomString()

	// 更新数据库
	_, err = c.dao.Update(ctx, task)
	if err != nil {
		response.ErrorJSON(ctx, "更新任务状态失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新数据
	updatedTask, err := c.dao.GetById(ctx, tenantId, task.TaskId)
	if err != nil {
		response.ErrorJSON(ctx, "获取更新后的任务配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, updatedTask, constants.SD00004)
} 