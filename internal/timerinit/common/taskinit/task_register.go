package taskinit

import (
	"context"
	"fmt"

	"gateway/internal/timerinit/common/dao"
	"gateway/internal/types/timertypes"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/timer"
)

// TaskRegister 通用任务注册器
// 提供统一的任务注册和管理功能
type TaskRegister struct {
	daoManager *dao.DAOManager
}

// NewTaskRegister 创建任务注册器实例
func NewTaskRegister(db database.Database) *TaskRegister {
	return &TaskRegister{
		daoManager: dao.NewDAOManager(db),
	}
}

// RegisterTasks 注册指定类型的任务
// 参数:
//
//	factory: 任务执行器工厂
//	tenantIds: 需要初始化的租户ID列表，如果为空则初始化所有租户
//
// 返回:
//
//	error: 初始化失败时返回错误信息
func (r *TaskRegister) RegisterTasks(ctx context.Context, factory TaskExecutorFactory, tenantIds ...string) error {
	executorType := factory.GetExecutorType()
	logger.Info("开始注册任务", "executorType", executorType)

	// 检查数据库连接
	if err := r.daoManager.HealthCheck(ctx); err != nil {
		return fmt.Errorf("数据库连接检查失败: %w", err)
	}

	// 如果没有指定租户ID，获取所有有该类型任务的租户
	if len(tenantIds) == 0 {
		var err error
		tenantIds, err = r.getTaskTenants(ctx, executorType)
		if err != nil {
			return fmt.Errorf("获取任务租户列表失败: %w", err)
		}
	}

	if len(tenantIds) == 0 {
		logger.Info("未找到任何需要注册的任务", "executorType", executorType)
		return nil
	}

	// 创建基础任务初始化器
	initializer := NewBaseTaskInitializer(r.daoManager, factory)

	// 为每个租户初始化任务
	var initErrors []error
	successCount := 0

	for _, tenantId := range tenantIds {
		if err := initializer.InitializeTasks(ctx, tenantId); err != nil {
			logger.Error("租户任务初始化失败",
				"tenantId", tenantId,
				"executorType", executorType,
				"error", err)
			initErrors = append(initErrors, fmt.Errorf("租户 %s: %w", tenantId, err))
			// 继续处理其他租户，不因单个租户失败而停止
			continue
		}
		successCount++
		logger.Info("租户任务初始化成功", "tenantId", tenantId, "executorType", executorType)
	}

	// 记录最终结果
	if len(initErrors) > 0 {
		logger.Warn("部分租户任务初始化失败",
			"executorType", executorType,
			"failedCount", len(initErrors),
			"totalCount", len(tenantIds))
	}

	logger.Info("任务注册完成",
		"executorType", executorType,
		"totalTenants", len(tenantIds),
		"successTenants", successCount,
		"failedTenants", len(initErrors))

	return nil
}

// RegisterTasksForTenant 为指定租户注册指定类型的任务
// 参数:
//
//	factory: 任务执行器工厂
//	tenantId: 租户ID
//
// 返回:
//
//	error: 初始化失败时返回错误信息
func (r *TaskRegister) RegisterTasksForTenant(ctx context.Context, factory TaskExecutorFactory, tenantId string) error {
	executorType := factory.GetExecutorType()
	logger.Info("开始为租户注册任务", "tenantId", tenantId, "executorType", executorType)

	// 检查数据库连接
	if err := r.daoManager.HealthCheck(ctx); err != nil {
		return fmt.Errorf("数据库连接检查失败: %w", err)
	}

	// 创建基础任务初始化器
	initializer := NewBaseTaskInitializer(r.daoManager, factory)

	// 初始化指定租户的任务
	if err := initializer.InitializeTasks(ctx, tenantId); err != nil {
		return fmt.Errorf("租户任务初始化失败: %w", err)
	}

	logger.Info("租户任务注册完成", "tenantId", tenantId, "executorType", executorType)
	return nil
}

// GetTaskStatus 获取指定类型任务的状态
// 参数:
//
//	executorType: 执行器类型
//	tenantId: 租户ID
//
// 返回:
//
//	map[string]interface{}: 状态统计信息
//	error: 获取失败时返回错误信息
func (r *TaskRegister) GetTaskStatus(ctx context.Context, executorType, tenantId string) (map[string]interface{}, error) {
	// 构建查询条件，按执行器类型过滤
	activeFlag := "Y"
	query := &dao.TimerTaskQuery{
		TenantId:     &tenantId,
		ExecutorType: &executorType,
		ActiveFlag:   &activeFlag,
	}

	// 查询指定类型的任务
	result, err := r.daoManager.GetTaskDAO().QueryTasks(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("获取任务统计失败: %w", err)
	}

	// 统计任务状态分布
	statusMap := make(map[string]int64)
	totalTasks := int64(len(result.Tasks))

	for _, task := range result.Tasks {
		switch task.TaskStatus {
		case timertypes.TaskStatusPending:
			statusMap["pending"]++
		case timertypes.TaskStatusRunning:
			statusMap["running"]++
		case timertypes.TaskStatusCompleted:
			statusMap["completed"]++
		case timertypes.TaskStatusFailed:
			statusMap["failed"]++
		case timertypes.TaskStatusCancelled:
			statusMap["cancelled"]++
		}
	}

	// 构建返回结果
	taskStats := make(map[string]interface{})
	taskStats["tenant_id"] = tenantId
	taskStats["executor_type"] = executorType
	taskStats["total_tasks"] = totalTasks
	taskStats["status_distribution"] = statusMap

	return taskStats, nil
}

// getTaskTenants 获取有指定类型任务的租户列表
func (r *TaskRegister) getTaskTenants(ctx context.Context, executorType string) ([]string, error) {
	// 查询所有指定类型的活动任务的租户
	activeFlag := "Y"

	query := &dao.TimerTaskQuery{
		ExecutorType:   &executorType,
		ActiveFlag:     &activeFlag,
		OrderBy:        "tenantId",
		OrderDirection: "ASC",
	}

	result, err := r.daoManager.GetTaskDAO().QueryTasks(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询任务租户失败: %w", err)
	}

	// 提取唯一的租户ID列表
	tenantMap := make(map[string]bool)
	for _, task := range result.Tasks {
		tenantMap[task.TenantId] = true
	}

	tenantIds := make([]string, 0, len(tenantMap))
	for tenantId := range tenantMap {
		tenantIds = append(tenantIds, tenantId)
	}

	return tenantIds, nil
}

// ReloadTasks 重新加载指定类型的任务
// 停止现有任务并重新从数据库加载配置
// 参数:
//
//	factory: 任务执行器工厂
//	tenantIds: 需要重新加载的租户ID列表，如果为空则重新加载所有租户
//
// 返回:
//
//	error: 重新加载失败时返回错误信息
func (r *TaskRegister) ReloadTasks(ctx context.Context, factory TaskExecutorFactory, tenantIds ...string) error {
	executorType := factory.GetExecutorType()
	logger.Info("开始重新加载任务", "executorType", executorType)

	// TODO: 实现任务重新加载逻辑
	// 1. 停止指定租户的所有指定类型任务
	// 2. 清理定时器池中的相关调度器
	// 3. 重新注册任务

	// 暂时使用简单的重新注册方式
	return r.RegisterTasks(ctx, factory, tenantIds...)
}

// StopTasks 停止指定类型的任务
// 停止指定租户的所有指定类型任务
// 参数:
//
//	executorType: 执行器类型
//	tenantIds: 需要停止的租户ID列表，如果为空则停止所有租户的任务
//
// 返回:
//
//	error: 停止失败时返回错误信息
func (r *TaskRegister) StopTasks(ctx context.Context, executorType string, tenantIds ...string) error {
	logger.Info("开始停止任务", "executorType", executorType)

	// 获取全局定时器池
	timerPool := timer.GetTimerPool()

	// 如果没有指定租户ID，获取所有有该类型任务的租户
	if len(tenantIds) == 0 {
		var err error
		tenantIds, err = r.getTaskTenants(ctx, executorType)
		if err != nil {
			return fmt.Errorf("获取任务租户列表失败: %w", err)
		}
	}

	if len(tenantIds) == 0 {
		logger.Info("未找到任何需要停止的任务", "executorType", executorType)
		return nil
	}

	// 停止每个租户的相关调度器
	var stopErrors []error
	stoppedCount := 0

	for _, tenantId := range tenantIds {
		// 生成调度器ID
		schedulerId := fmt.Sprintf("%s_scheduler_%s", executorType, tenantId)

		// 尝试获取调度器
		scheduler, err := timerPool.GetScheduler(schedulerId)
		if err != nil {
			// 调度器不存在，跳过
			logger.Debug("调度器不存在，跳过停止", "schedulerId", schedulerId)
			continue
		}

		// 停止调度器
		if scheduler.IsRunning() {
			if err := scheduler.Stop(); err != nil {
				logger.Error("停止调度器失败", "schedulerId", schedulerId, "error", err)
				stopErrors = append(stopErrors, fmt.Errorf("停止调度器 %s 失败: %w", schedulerId, err))
				continue
			}
			logger.Info("调度器已停止", "schedulerId", schedulerId)
		}

		// 移除调度器
		if err := timerPool.RemoveScheduler(schedulerId); err != nil {
			logger.Warn("移除调度器失败", "schedulerId", schedulerId, "error", err)
			stopErrors = append(stopErrors, fmt.Errorf("移除调度器 %s 失败: %w", schedulerId, err))
		} else {
			logger.Info("调度器已移除", "schedulerId", schedulerId)
		}

		stoppedCount++
	}

	// 记录停止结果
	if len(stopErrors) > 0 {
		logger.Warn("部分调度器停止失败",
			"executorType", executorType,
			"failedCount", len(stopErrors),
			"totalCount", len(tenantIds))
		return fmt.Errorf("部分调度器停止失败: %v", stopErrors)
	}

	logger.Info("任务停止完成",
		"executorType", executorType,
		"stoppedCount", stoppedCount,
		"totalCount", len(tenantIds))
	return nil
}

// GracefulShutdown 优雅关闭所有调度器
// 停止所有正在运行的调度器并等待任务完成
// 参数:
//
//	ctx: 上下文，用于控制关闭超时
//
// 返回:
//
//	error: 关闭失败时返回错误信息
func (r *TaskRegister) GracefulShutdown(ctx context.Context) error {
	logger.Info("开始优雅关闭所有任务调度器")

	// 获取全局定时器池
	timerPool := timer.GetTimerPool()

	// 获取所有调度器ID
	schedulerIds := timerPool.ListSchedulers()
	if len(schedulerIds) == 0 {
		logger.Info("没有运行中的调度器需要关闭")
		return nil
	}

	logger.Info("准备关闭调度器", "count", len(schedulerIds))

	// 停止所有调度器
	stopErrors := timerPool.StopAllSchedulers()
	if len(stopErrors) > 0 {
		logger.Warn("部分调度器停止时出现错误", "errorCount", len(stopErrors))
		for id, err := range stopErrors {
			logger.Error("调度器停止失败", "schedulerId", id, "error", err)
		}
	}

	// 移除所有调度器
	var removeErrors []error
	for _, id := range schedulerIds {
		if err := timerPool.RemoveScheduler(id); err != nil {
			removeErrors = append(removeErrors, fmt.Errorf("移除调度器 %s 失败: %w", id, err))
		}
	}

	if len(removeErrors) > 0 {
		logger.Error("部分调度器移除失败", "errorCount", len(removeErrors))
		return fmt.Errorf("移除调度器失败: %v", removeErrors)
	}

	logger.Info("所有任务调度器已优雅关闭", "closedCount", len(schedulerIds))
	return nil
}

// GetAllSchedulerStatus 获取所有调度器的状态信息
// 返回:
//
//	map[string]*SchedulerStatusInfo: 调度器ID -> 状态信息的映射
//	error: 获取失败时返回错误信息
func (r *TaskRegister) GetAllSchedulerStatus() (map[string]*SchedulerStatusInfo, error) {
	// 获取全局定时器池
	timerPool := timer.GetTimerPool()

	// 获取所有调度器信息
	allInfos := timerPool.GetAllSchedulerInfo()

	statusMap := make(map[string]*SchedulerStatusInfo)
	for _, info := range allInfos {
		statusMap[info.ID] = &SchedulerStatusInfo{
			ID:         info.ID,
			Name:       info.Config.Name,
			IsRunning:  info.IsRunning,
			TaskCount:  info.TaskCount,
			MaxWorkers: info.Config.MaxWorkers,
			QueueSize:  info.Config.QueueSize,
		}
	}

	return statusMap, nil
}

// SchedulerStatusInfo 调度器状态信息
type SchedulerStatusInfo struct {
	ID         string `json:"id"`         // 调度器ID
	Name       string `json:"name"`       // 调度器名称
	IsRunning  bool   `json:"isRunning"`  // 是否正在运行
	TaskCount  int    `json:"taskCount"`  // 任务数量
	MaxWorkers int    `json:"maxWorkers"` // 最大工作线程数
	QueueSize  int    `json:"queueSize"`  // 队列大小
}
