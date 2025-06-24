package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hub0003/dao"
	"gohub/web/views/hub0003/models"

	"github.com/gin-gonic/gin"
)

// TaskInfoController 定时任务运行时信息控制器
type TaskInfoController struct {
	dao *dao.TaskInfoDAO
}

// NewTaskInfoController 创建任务运行时信息控制器
func NewTaskInfoController(db database.Database) *TaskInfoController {
	return &TaskInfoController{
		dao: dao.NewTaskInfoDAO(db),
	}
}

// AddTaskInfo 添加任务运行时信息
func (c *TaskInfoController) AddTaskInfo(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加任务运行时信息", "controller", "TaskInfoController", "action", "AddTaskInfo")

	// 绑定请求参数
	var info models.TaskInfo
	if err := request.BindSafely(ctx, &info); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 验证必填字段
	if info.TaskConfigId == "" {
		response.ErrorJSON(ctx, "任务配置ID不能为空", constants.ED00007)
		return
	}
	if info.TaskId == "" {
		response.ErrorJSON(ctx, "任务ID不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID
	info.TenantId = tenantId

	// 调用DAO层添加任务信息
	err := c.dao.AddTaskInfo(ctx, &info, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加任务运行时信息失败", "error", err.Error(), 
			"taskInfoId", info.TaskInfoId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加任务运行时信息失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的信息数据返回给前端
	newInfo, err := c.dao.GetTaskInfo(info.TenantId, info.TaskInfoId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(), 
			"taskInfoId", info.TaskInfoId, "tenantId", tenantId)
		response.SuccessJSON(ctx, gin.H{"message": "任务运行时信息添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "任务运行时信息添加成功", "taskInfoId", info.TaskInfoId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newInfo, constants.SD00003)
}

// GetTaskInfo 获取任务运行时信息详情
func (c *TaskInfoController) GetTaskInfo(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取任务运行时信息", "controller", "TaskInfoController", "action", "GetTaskInfo")

	// 定义请求参数结构，支持多种查询方式
	var req struct {
		// 按任务信息ID查询
		TaskInfoId *string `json:"taskInfoId" form:"taskInfoId"`
		
		// 按任务ID查询
		TaskId *string `json:"taskId" form:"taskId"`
	}

	// 绑定请求参数
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 验证查询参数（至少提供一种查询方式）
	hasValidParam := false
	if req.TaskInfoId != nil && *req.TaskInfoId != "" {
		hasValidParam = true
	}
	if req.TaskId != nil && *req.TaskId != "" {
		hasValidParam = true
	}
	
	if !hasValidParam {
		response.ErrorJSON(ctx, "请提供taskInfoId或taskId中的任意一个", constants.ED00007)
		return
	}

	var info *models.TaskInfo
	var err error

	// 按任务信息ID查询
	if req.TaskInfoId != nil && *req.TaskInfoId != "" {
		info, err = c.dao.GetTaskInfo(tenantId, *req.TaskInfoId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取任务运行时信息失败", "error", err.Error(), 
				"taskInfoId", *req.TaskInfoId, "tenantId", tenantId)
			response.ErrorJSON(ctx, "获取任务运行时信息失败: "+err.Error(), constants.ED00009)
			return
		}
	} else if req.TaskId != nil && *req.TaskId != "" {
		// 按任务ID查询
		info, err = c.dao.GetTaskInfoByTaskId(tenantId, *req.TaskId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "根据任务ID获取运行时信息失败", "error", err.Error(), 
				"taskId", *req.TaskId, "tenantId", tenantId)
			response.ErrorJSON(ctx, "根据任务ID获取运行时信息失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	if info == nil {
		response.ErrorJSON(ctx, "任务运行时信息不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取任务运行时信息成功", "taskInfoId", info.TaskInfoId, "tenantId", tenantId)
	response.SuccessJSON(ctx, info, constants.SD00002)
}

// UpdateTaskInfo 更新任务运行时信息
func (c *TaskInfoController) UpdateTaskInfo(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新任务运行时信息", "controller", "TaskInfoController", "action", "UpdateTaskInfo")

	// 绑定请求参数
	var info models.TaskInfo
	if err := request.BindSafely(ctx, &info); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 验证必填字段
	if info.TaskInfoId == "" {
		response.ErrorJSON(ctx, "任务信息ID不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID
	info.TenantId = tenantId

	// 调用DAO层更新任务信息
	err := c.dao.UpdateTaskInfo(ctx, &info, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新任务运行时信息失败", "error", err.Error(), 
			"taskInfoId", info.TaskInfoId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新任务运行时信息失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的信息数据返回给前端
	updatedInfo, err := c.dao.GetTaskInfo(info.TenantId, info.TaskInfoId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(), 
			"taskInfoId", info.TaskInfoId, "tenantId", tenantId)
		response.SuccessJSON(ctx, gin.H{"message": "任务运行时信息更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "任务运行时信息更新成功", "taskInfoId", info.TaskInfoId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedInfo, constants.SD00003)
}

// UpdateTaskStatus 更新任务状态
func (c *TaskInfoController) UpdateTaskStatus(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新任务状态", "controller", "TaskInfoController", "action", "UpdateTaskStatus")

	// 定义请求参数结构
	var req struct {
		TaskId     string `json:"taskId" form:"taskId" binding:"required"`
		TaskStatus int    `json:"taskStatus" form:"taskStatus" binding:"required,min=1,max=5"`
	}

	// 绑定请求参数
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 调用DAO层更新任务状态
	err := c.dao.UpdateTaskStatus(ctx, tenantId, req.TaskId, req.TaskStatus, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新任务状态失败", "error", err.Error(), 
			"taskId", req.TaskId, "taskStatus", req.TaskStatus, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新任务状态失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "任务状态更新成功", "taskId", req.TaskId, 
		"taskStatus", req.TaskStatus, "tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "任务状态更新成功"}, constants.SD00003)
}

// QueryTaskInfos 查询任务运行时信息列表
func (c *TaskInfoController) QueryTaskInfos(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询任务运行时信息列表", "controller", "TaskInfoController", "action", "QueryTaskInfos")

	// 定义请求参数结构
	var req struct {
		Page       int  `json:"page" form:"page"`
		PageSize   int  `json:"pageSize" form:"pageSize"`
		TaskStatus *int `json:"taskStatus" form:"taskStatus"` // 可选的任务状态过滤
	}

	// 绑定请求参数
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	var infos []*models.TaskInfo
	var total int
	var err error

	// 根据任务状态过滤查询
	if req.TaskStatus != nil {
		infos, err = c.dao.GetTaskInfosByStatus(ctx, tenantId, *req.TaskStatus)
		if err != nil {
			logger.ErrorWithTrace(ctx, "根据状态查询任务信息失败", "error", err.Error(), 
				"tenantId", tenantId, "taskStatus", *req.TaskStatus)
			response.ErrorJSON(ctx, "根据状态查询任务信息失败: "+err.Error(), constants.ED00009)
			return
		}
		total = len(infos)

		// 手动分页
		start := (req.Page - 1) * req.PageSize
		end := start + req.PageSize
		if start >= len(infos) {
			infos = []*models.TaskInfo{}
		} else {
			if end > len(infos) {
				end = len(infos)
			}
			infos = infos[start:end]
		}
	} else {
		// 普通分页查询
		infos, total, err = c.dao.ListTaskInfos(ctx, tenantId, req.Page, req.PageSize)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询任务运行时信息列表失败", "error", err.Error(), 
				"tenantId", tenantId, "page", req.Page, "pageSize", req.PageSize)
			response.ErrorJSON(ctx, "查询任务运行时信息列表失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	// 构造响应数据
	responseData := gin.H{
		"page":      req.Page,
		"pageSize":  req.PageSize,
		"total":     total,
		"taskInfos": infos,
	}

	logger.InfoWithTrace(ctx, "查询任务运行时信息列表成功", "tenantId", tenantId, 
		"page", req.Page, "pageSize", req.PageSize, "total", total)
	response.SuccessJSON(ctx, responseData, constants.SD00002)
}

// DeleteTaskInfo 删除任务运行时信息
func (c *TaskInfoController) DeleteTaskInfo(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除任务运行时信息", "controller", "TaskInfoController", "action", "DeleteTaskInfo")

	var req struct {
		TaskInfoId string `json:"taskInfoId" form:"taskInfoId"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	if req.TaskInfoId == "" {
		response.ErrorJSON(ctx, "任务信息ID不能为空", constants.ED00007)
		return
	}

	// 调用DAO层删除任务信息
	err := c.dao.DeleteTaskInfo(ctx, tenantId, req.TaskInfoId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除任务运行时信息失败", "error", err.Error(), 
			"taskInfoId", req.TaskInfoId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除任务运行时信息失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "任务运行时信息删除成功", "taskInfoId", req.TaskInfoId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "任务运行时信息删除成功"}, constants.SD00003)
} 