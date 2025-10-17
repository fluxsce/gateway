package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0042/dao"
	"gateway/web/views/hub0042/models"

	"github.com/gin-gonic/gin"
)

// JvmQueryController JVM监控查询控制器
type JvmQueryController struct {
	dao *dao.JvmQueryDao
}

// NewJvmQueryController 创建JVM查询控制器实例
func NewJvmQueryController(db database.Database) *JvmQueryController {
	return &JvmQueryController{
		dao: dao.NewJvmQueryDao(db),
	}
}

// ===============================
// JVM资源监控查询
// ===============================

// QueryJvmResources 查询JVM资源列表
// @Summary 查询JVM资源列表
// @Description 支持分页、过滤、排序的JVM资源查询
// @Tags JVM监控
// @Accept json
// @Produce json
// @Param request body models.JvmResourceQueryRequest true "查询请求"
// @Success 200 {object} response.JsonData{data=models.JvmResourceListResponse}
// @Router /gateway/hub0042/queryJvmResources [post]
func (c *JvmQueryController) QueryJvmResources(ctx *gin.Context) {
	// 1. 解析请求参数
	var req models.JvmResourceQueryRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 2. 从上下文获取tenantId
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 3. 获取分页参数
	req.PageNum, req.PageSize = request.GetPaginationParams(ctx)

	// 4. 查询数据
	result, err := c.dao.QueryJvmResources(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询JVM资源列表失败", err)
		response.ErrorJSON(ctx, "查询JVM资源列表失败", constants.ED00009)
		return
	}

	// 5. 构建分页响应
	pageInfo := response.NewPageInfo(req.PageNum, req.PageSize, int(result.PageInfo.TotalCount))
	pageInfo.MainKey = "jvmResourceId"

	// 6. 返回结果
	response.PageJSON(ctx, result.List, pageInfo, constants.SD00002)
}

// GetJvmResourceDetail 获取JVM资源详情
// @Summary 获取JVM资源详情
// @Description 获取指定JVM实例的最新资源信息
// @Tags JVM监控
// @Accept json
// @Produce json
// @Param request body object{jvmResourceId=string} true "查询请求"
// @Success 200 {object} response.JsonData{data=models.JvmResourceResponse}
// @Router /gateway/hub0042/getJvmResourceDetail [post]
func (c *JvmQueryController) GetJvmResourceDetail(ctx *gin.Context) {
	// 1. 获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 2. 获取参数
	jvmResourceId := request.GetParam(ctx, "jvmResourceId")
	if jvmResourceId == "" {
		response.ErrorJSON(ctx, "jvmResourceId不能为空", constants.ED00006)
		return
	}

	// 3. 查询数据
	result, err := c.dao.GetJvmResourceDetail(ctx, tenantId, jvmResourceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询JVM资源详情失败", err, "jvmResourceId", jvmResourceId)
		response.ErrorJSON(ctx, "查询JVM资源详情失败", constants.ED00009)
		return
	}

	if result == nil {
		response.ErrorJSON(ctx, "JVM资源不存在", constants.ED00008)
		return
	}

	// 4. 返回结果
	response.SuccessJSON(ctx, result, constants.SD00002)
}

// ===============================
// GC快照查询
// ===============================

// QueryGCSnapshots 查询GC快照列表
// @Summary 查询GC快照列表
// @Description 查询指定JVM实例的GC快照历史记录
// @Tags JVM监控-GC
// @Accept json
// @Produce json
// @Param request body models.GCSnapshotQueryRequest true "查询请求"
// @Success 200 {object} response.JsonData{data=models.GCSnapshotListResponse}
// @Router /gateway/hub0042/queryGCSnapshots [post]
func (c *JvmQueryController) QueryGCSnapshots(ctx *gin.Context) {
	var req models.GCSnapshotQueryRequest

	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	if req.JvmResourceId == "" {
		response.ErrorJSON(ctx, "jvmResourceId不能为空", constants.ED00006)
		return
	}

	list, err := c.dao.QueryGCSnapshots(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询GC快照列表失败", err)
		response.ErrorJSON(ctx, "查询GC快照列表失败", constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, list, constants.SD00002)
}

// GetLatestGCSnapshot 获取最新GC快照
// @Summary 获取最新GC快照
// @Description 获取指定JVM实例的最新GC快照数据
// @Tags JVM监控-GC
// @Accept json
// @Produce json
// @Param request body object{jvmResourceId=string} true "查询请求"
// @Success 200 {object} response.JsonData{data=models.GCSnapshotResponse}
// @Router /gateway/hub0042/getLatestGCSnapshot [post]
func (c *JvmQueryController) GetLatestGCSnapshot(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	jvmResourceId := request.GetParam(ctx, "jvmResourceId")
	if jvmResourceId == "" {
		response.ErrorJSON(ctx, "jvmResourceId不能为空", constants.ED00006)
		return
	}

	result, err := c.dao.GetLatestGCSnapshot(ctx, tenantId, jvmResourceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询最新GC快照失败", err, "jvmResourceId", jvmResourceId)
		response.ErrorJSON(ctx, "查询最新GC快照失败", constants.ED00009)
		return
	}

	if result == nil {
		response.ErrorJSON(ctx, "GC快照不存在", constants.ED00008)
		return
	}

	response.SuccessJSON(ctx, result, constants.SD00002)
}

// ===============================
// 内存监控查询
// ===============================

// QueryMemory 查询内存记录
// @Summary 查询内存记录
// @Description 查询指定JVM实例的内存使用历史记录
// @Tags JVM监控-内存
// @Accept json
// @Produce json
// @Param request body models.MemoryQueryRequest true "查询请求"
// @Success 200 {object} response.JsonData{data=models.MemoryListResponse}
// @Router /gateway/hub0042/queryMemory [post]
func (c *JvmQueryController) QueryMemory(ctx *gin.Context) {
	var req models.MemoryQueryRequest

	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	if req.JvmResourceId == "" {
		response.ErrorJSON(ctx, "jvmResourceId不能为空", constants.ED00006)
		return
	}

	list, err := c.dao.QueryMemory(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询内存记录失败", err)
		response.ErrorJSON(ctx, "查询内存记录失败", constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, list, constants.SD00002)
}

// QueryMemoryPools 查询内存池记录
// @Summary 查询内存池记录
// @Description 查询指定JVM实例的内存池使用历史记录
// @Tags JVM监控-内存
// @Accept json
// @Produce json
// @Param request body models.MemoryPoolQueryRequest true "查询请求"
// @Success 200 {object} response.JsonData{data=models.MemoryPoolListResponse}
// @Router /gateway/hub0042/queryMemoryPools [post]
func (c *JvmQueryController) QueryMemoryPools(ctx *gin.Context) {
	var req models.MemoryPoolQueryRequest

	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	if req.JvmResourceId == "" {
		response.ErrorJSON(ctx, "jvmResourceId不能为空", constants.ED00006)
		return
	}

	list, err := c.dao.QueryMemoryPools(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询内存池记录失败", err)
		response.ErrorJSON(ctx, "查询内存池记录失败", constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, list, constants.SD00002)
}

// ===============================
// 线程监控查询
// ===============================

// QueryThreads 查询线程记录
// @Summary 查询线程记录
// @Description 查询指定JVM实例的线程监控历史记录
// @Tags JVM监控-线程
// @Accept json
// @Produce json
// @Param request body models.ThreadQueryRequest true "查询请求"
// @Success 200 {object} response.JsonData{data=models.ThreadListResponse}
// @Router /gateway/hub0042/queryThreads [post]
func (c *JvmQueryController) QueryThreads(ctx *gin.Context) {
	var req models.ThreadQueryRequest

	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	if req.JvmResourceId == "" {
		response.ErrorJSON(ctx, "jvmResourceId不能为空", constants.ED00006)
		return
	}

	list, err := c.dao.QueryThreads(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询线程记录失败", err)
		response.ErrorJSON(ctx, "查询线程记录失败", constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, list, constants.SD00002)
}

// QueryThreadStates 查询线程状态记录
// @Summary 查询线程状态记录
// @Description 查询指定JVM实例的线程状态统计历史记录
// @Tags JVM监控-线程
// @Accept json
// @Produce json
// @Param request body models.ThreadStateQueryRequest true "查询请求"
// @Success 200 {object} response.JsonData{data=[]models.ThreadStateResponse}
// @Router /gateway/hub0042/queryThreadStates [post]
func (c *JvmQueryController) QueryThreadStates(ctx *gin.Context) {
	var req models.ThreadStateQueryRequest

	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	if req.JvmResourceId == "" {
		response.ErrorJSON(ctx, "jvmResourceId不能为空", constants.ED00006)
		return
	}

	list, err := c.dao.QueryThreadStates(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询线程状态记录失败", err)
		response.ErrorJSON(ctx, "查询线程状态记录失败", constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, list, constants.SD00002)
}

// QueryDeadlocks 查询死锁记录
// @Summary 查询死锁记录
// @Description 查询JVM实例的死锁检测历史记录
// @Tags JVM监控-线程
// @Accept json
// @Produce json
// @Param request body models.DeadlockQueryRequest true "查询请求"
// @Success 200 {object} response.JsonData{data=models.DeadlockListResponse}
// @Router /gateway/hub0042/queryDeadlocks [post]
func (c *JvmQueryController) QueryDeadlocks(ctx *gin.Context) {
	var req models.DeadlockQueryRequest

	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	list, err := c.dao.QueryDeadlocks(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询死锁记录失败", err)
		response.ErrorJSON(ctx, "查询死锁记录失败", constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, list, constants.SD00002)
}

// ===============================
// 类加载监控查询
// ===============================

// QueryClassLoading 查询类加载记录
// @Summary 查询类加载记录
// @Description 查询指定JVM实例的类加载监控历史记录
// @Tags JVM监控-类加载
// @Accept json
// @Produce json
// @Param request body models.ClassLoadingQueryRequest true "查询请求"
// @Success 200 {object} response.JsonData{data=models.ClassLoadingListResponse}
// @Router /gateway/hub0042/queryClassLoading [post]
func (c *JvmQueryController) QueryClassLoading(ctx *gin.Context) {
	var req models.ClassLoadingQueryRequest

	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	if req.JvmResourceId == "" {
		response.ErrorJSON(ctx, "jvmResourceId不能为空", constants.ED00006)
		return
	}

	list, err := c.dao.QueryClassLoading(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询类加载记录失败", err)
		response.ErrorJSON(ctx, "查询类加载记录失败", constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, list, constants.SD00002)
}

// ===============================
// 统计和概览
// ===============================

// GetJvmOverview 获取JVM监控概览
// @Summary 获取JVM监控概览
// @Description 获取JVM实例的整体统计概览信息
// @Tags JVM监控-概览
// @Accept json
// @Produce json
// @Param request body models.JvmOverviewRequest false "查询请求"
// @Success 200 {object} response.JsonData{data=models.JvmOverviewResponse}
// @Router /gateway/hub0042/getJvmOverview [post]
func (c *JvmQueryController) GetJvmOverview(ctx *gin.Context) {
	var req models.JvmOverviewRequest

	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	result, err := c.dao.GetJvmOverview(ctx, tenantId, req.ApplicationName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询JVM概览失败", err)
		response.ErrorJSON(ctx, "查询JVM概览失败", constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, result, constants.SD00002)
}

// ==========================================
// 应用监控数据查询方法
// ==========================================

// QueryAppMonitorData 查询应用监控数据列表
func (c *JvmQueryController) QueryAppMonitorData(ctx *gin.Context) {
	var req models.QueryAppMonitorDataRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 获取分页参数
	req.PageIndex, req.PageSize = request.GetPaginationParams(ctx)

	result, total, err := c.dao.QueryAppMonitorData(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询应用监控数据失败", err)
		response.ErrorJSON(ctx, "查询应用监控数据失败", constants.ED00009)
		return
	}

	// 构建分页响应
	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)
	pageInfo.MainKey = "appDataId"

	// 返回结果
	response.PageJSON(ctx, result, pageInfo, constants.SD00002)
}

// GetAppMonitorDataDetail 获取应用监控数据详情
func (c *JvmQueryController) GetAppMonitorDataDetail(ctx *gin.Context) {
	var req struct {
		AppDataId string `json:"appDataId" form:"appDataId" query:"appDataId"` // 应用监控数据ID
	}
	if err := request.BindSafely(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	if req.AppDataId == "" {
		response.ErrorJSON(ctx, "appDataId不能为空", constants.ED00006)
		return
	}

	result, err := c.dao.GetAppMonitorDataDetail(ctx, tenantId, req.AppDataId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取应用监控数据详情失败", err)
		response.ErrorJSON(ctx, "获取应用监控数据详情失败", constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, result, constants.SD00002)
}
