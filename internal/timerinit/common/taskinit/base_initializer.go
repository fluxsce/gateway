package taskinit

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/timerinit/common/dao"
	"gateway/internal/types/timertypes"
	"gateway/pkg/logger"
	"gateway/pkg/timer"
)

// TaskExecutorFactory 任务执行器工厂接口
// 不同类型的任务需要实现此接口来创建特定的执行器
type TaskExecutorFactory interface {
	// CreateExecutor 创建任务执行器
	CreateExecutor(ctx context.Context, task *timertypes.TimerTask) (timer.TaskExecutor, error)

	// GetExecutorType 获取执行器类型
	GetExecutorType() string
}

// BaseTaskInitializer 基础任务初始化器
// 包含通用的定时器任务初始化逻辑
type BaseTaskInitializer struct {
	daoManager *dao.DAOManager
	timerPool  *timer.TimerPool
	factory    TaskExecutorFactory
}

// NewBaseTaskInitializer 创建基础任务初始化器实例
// 基础任务初始化器是任务初始化的核心组件，负责协调任务的创建、配置和调度
// 参数:
//
//	daoManager: 数据访问对象管理器，用于数据库操作
//	factory: 任务执行器工厂，用于创建特定类型的任务执行器
//
// 返回:
//
//	*BaseTaskInitializer: 初始化器实例
func NewBaseTaskInitializer(daoManager *dao.DAOManager, factory TaskExecutorFactory) *BaseTaskInitializer {
	return &BaseTaskInitializer{
		daoManager: daoManager,
		timerPool:  timer.GetTimerPool(),
		factory:    factory,
	}
}

// InitializeTasks 初始化指定租户的任务
// 这是任务初始化的主入口方法，负责查询、转换和初始化指定租户下的所有相关任务
// 支持批量初始化，提供详细的成功/失败统计信息
// 参数:
//
//	ctx: 上下文对象，用于控制请求生命周期和传递元数据
//	tenantId: 租户ID，用于多租户环境下的任务隔离
//
// 返回:
//
//	error: 初始化过程中的错误信息，如果部分任务失败会返回汇总错误
func (init *BaseTaskInitializer) InitializeTasks(ctx context.Context, tenantId string) error {
	// 查询任务列表
	tasks, err := init.getTasks(ctx, tenantId)
	if err != nil {
		return fmt.Errorf("查询任务失败: %w", err)
	}

	if len(tasks) == 0 {
		logger.Info("未找到需要初始化的任务",
			"tenantId", tenantId,
			"executorType", init.factory.GetExecutorType())
		return nil
	}

	logger.Info("开始初始化任务",
		"tenantId", tenantId,
		"executorType", init.factory.GetExecutorType(),
		"taskCount", len(tasks))

	// 初始化每个任务
	var initErrors []error
	successCount := 0

	for _, task := range tasks {
		if err := init.initializeSingleTask(ctx, task); err != nil {
			logger.Error("任务初始化失败",
				"taskId", task.TaskId,
				"taskName", task.TaskName,
				"error", err)
			initErrors = append(initErrors, fmt.Errorf("任务 %s: %w", task.TaskId, err))
			continue
		}
		successCount++
	}

	logger.Info("任务初始化完成",
		"tenantId", tenantId,
		"executorType", init.factory.GetExecutorType(),
		"totalCount", len(tasks),
		"successCount", successCount,
		"failedCount", len(initErrors))

	// 如果有失败的任务，返回汇总错误
	if len(initErrors) > 0 {
		return fmt.Errorf("部分任务初始化失败 (%d/%d): %v", len(initErrors), len(tasks), initErrors)
	}

	return nil
}

// InitializeSingleTask 初始化单个任务（公开方法）
// 提供公开的接口用于初始化单个指定的任务
// 参数:
//
//	ctx: 上下文对象，用于控制请求生命周期
//	task: 要初始化的数据库任务对象
//
// 返回:
//
//	error: 初始化过程中的错误信息
func (init *BaseTaskInitializer) InitializeSingleTask(ctx context.Context, task *timertypes.TimerTask) error {
	logger.Info("开始初始化单个任务",
		"taskId", task.TaskId,
		"taskName", task.TaskName,
		"tenantId", task.TenantId,
		"executorType", init.factory.GetExecutorType())

	// 调用私有的初始化方法
	if err := init.initializeSingleTask(ctx, task); err != nil {
		logger.Error("单个任务初始化失败",
			"taskId", task.TaskId,
			"taskName", task.TaskName,
			"tenantId", task.TenantId,
			"error", err)
		return err
	}

	logger.Info("单个任务初始化完成",
		"taskId", task.TaskId,
		"taskName", task.TaskName,
		"tenantId", task.TenantId)

	return nil
}

// getTasks 查询指定类型的任务
// 根据租户ID和执行器类型查询数据库中的活动任务列表
// 只查询状态为活动的任务，并按任务名称排序
// 参数:
//
//	ctx: 上下文对象，用于控制请求生命周期
//	tenantId: 租户ID，用于多租户隔离
//
// 返回:
//
//	[]*timertypes.TimerTask: 查询到的任务列表
//	error: 查询过程中的错误信息
func (init *BaseTaskInitializer) getTasks(ctx context.Context, tenantId string) ([]*timertypes.TimerTask, error) {
	executorType := init.factory.GetExecutorType()
	activeFlag := "Y"

	query := &dao.TimerTaskQuery{
		TenantId:       &tenantId,
		ExecutorType:   &executorType,
		ActiveFlag:     &activeFlag,
		OrderBy:        "taskName",
		OrderDirection: "ASC",
	}

	// 执行查询
	result, err := init.daoManager.GetTaskDAO().QueryTasks(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	// 转换为指针切片
	tasks := make([]*timertypes.TimerTask, len(result.Tasks))
	for i := range result.Tasks {
		tasks[i] = &result.Tasks[i]
	}

	return tasks, nil
}

// initializeSingleTask 初始化单个任务
// 完成单个任务的完整初始化流程：创建执行器、转换配置、添加到调度器、启动任务
// 这是任务初始化的核心逻辑，确保任务能够正确地被调度和执行
// 参数:
//
//	ctx: 上下文对象，用于控制请求生命周期
//	task: 要初始化的数据库任务对象
//
// 返回:
//
//	error: 初始化过程中的错误信息
func (init *BaseTaskInitializer) initializeSingleTask(ctx context.Context, task *timertypes.TimerTask) error {
	// 创建任务执行器
	executor, err := init.factory.CreateExecutor(ctx, task)
	if err != nil {
		return fmt.Errorf("创建任务执行器失败: %w", err)
	}

	// 转换为定时器任务配置
	timerConfig, err := init.convertToTimerConfig(task)
	if err != nil {
		return fmt.Errorf("转换定时器配置失败: %w", err)
	}

	// 获取或创建调度器
	scheduler, err := init.getOrCreateScheduler(ctx, task.TenantId, task.SchedulerId)
	if err != nil {
		return fmt.Errorf("获取调度器失败: %w", err)
	}

	// 将任务添加到调度器
	if err := scheduler.AddTask(timerConfig, executor); err != nil {
		return fmt.Errorf("添加任务到调度器失败: %w", err)
	}

	// 如果任务已启用且调度器正在运行，启动任务
	if task.TaskStatus == timertypes.TaskStatusPending && timerConfig.Enabled {
		if scheduler.IsRunning() {
			if err := scheduler.StartTask(timerConfig.ID); err != nil {
				logger.Warn("启动任务失败",
					"taskId", task.TaskId,
					"error", err)
			}
		}
	}

	return nil
}

// convertToTimerConfig 转换为定时器任务配置
// 将数据库中的TimerTask对象转换为timer系统使用的TaskConfig对象
// 包含完整的任务配置信息，如调度规则、执行参数、时间限制等
// 参数:
//
//	task: 数据库中的任务对象，包含所有持久化的任务信息
//
// 返回:
//
//	*timer.TaskConfig: 转换后的任务配置对象
//	error: 转换过程中的错误信息
func (init *BaseTaskInitializer) convertToTimerConfig(task *timertypes.TimerTask) (*timer.TaskConfig, error) {
	// 创建基础任务配置
	config := &timer.TaskConfig{
		ID:            task.TaskId,
		Name:          task.TaskName,
		Priority:      init.convertPriority(task.TaskPriority),
		Enabled:       task.IsActive(),
		MaxRetries:    task.MaxRetries,
		Timeout:       time.Duration(task.TimeoutSeconds) * time.Second,
		RetryInterval: time.Duration(task.RetryIntervalSeconds) * time.Second,
	}

	// 设置任务描述
	if task.TaskDescription != nil {
		config.Description = *task.TaskDescription
	}

	// 设置任务参数
	if task.TaskParams != nil {
		config.Params = *task.TaskParams
	}

	// 设置创建和更新时间信息
	config.CreatedAt = task.AddTime
	config.UpdatedAt = task.EditTime

	// 设置任务运行状态信息
	config.Status = init.convertTaskStatus(task.TaskStatus)
	config.RunCount = task.RunCount
	config.FailureCount = task.FailureCount

	// 设置时间信息
	if task.NextRunTime != nil {
		config.NextRunTime = task.NextRunTime
	}
	if task.LastRunTime != nil {
		config.LastRunTime = task.LastRunTime
	}

	// 设置调度配置
	if err := init.setScheduleConfig(config, task); err != nil {
		return nil, fmt.Errorf("设置调度配置失败: %w", err)
	}

	return config, nil
}

// setScheduleConfig 设置调度配置
// 根据数据库任务的调度类型和参数，设置TaskConfig的调度相关配置
// 支持Cron表达式、固定间隔、延迟执行等多种调度模式
// 参数:
//
//	config: 要设置的任务配置对象
//	task: 数据库中的任务对象，包含调度参数
//
// 返回:
//
//	error: 配置设置过程中的错误信息
func (init *BaseTaskInitializer) setScheduleConfig(config *timer.TaskConfig, task *timertypes.TimerTask) error {
	// 设置调度类型
	config.ScheduleType = init.convertScheduleType(task.ScheduleType)

	// 设置时间范围限制
	if task.StartTime != nil {
		config.StartTime = task.StartTime
	}
	if task.EndTime != nil {
		config.EndTime = task.EndTime
	}

	// 根据调度类型设置相应的调度参数
	switch task.ScheduleType {
	case timertypes.ScheduleTypeCron:
		// Cron表达式调度：需要有效的Cron表达式
		if task.CronExpression != nil {
			config.CronExpr = *task.CronExpression
		} else {
			return fmt.Errorf("Cron调度类型必须提供Cron表达式")
		}
	case timertypes.ScheduleTypeInterval:
		// 固定间隔调度：需要正数的间隔秒数
		if task.IntervalSeconds != nil {
			config.Interval = time.Duration(*task.IntervalSeconds) * time.Second
		} else {
			return fmt.Errorf("间隔调度类型必须提供间隔时间")
		}
	case timertypes.ScheduleTypeDelay:
		// 延迟执行调度：需要正数的延迟秒数
		if task.DelaySeconds != nil {
			config.Delay = time.Duration(*task.DelaySeconds) * time.Second
		} else {
			return fmt.Errorf("延迟调度类型必须提供延迟时间")
		}
	case timertypes.ScheduleTypeOneTime:
		// 一次性任务：不需要额外参数
		break
	default:
		return fmt.Errorf("不支持的调度类型: %d", task.ScheduleType)
	}

	return nil
}

// convertPriority 转换任务优先级
// 将数据库中的整数优先级值转换为timer系统的TaskPriority枚举
// 参数:
//
//	priority: 数据库中的优先级整数值
//
// 返回:
//
//	timer.TaskPriority: 对应的任务优先级枚举值
func (init *BaseTaskInitializer) convertPriority(priority int) timer.TaskPriority {
	switch priority {
	case timertypes.TaskPriorityLow:
		return timer.TaskPriorityLow
	case timertypes.TaskPriorityHigh:
		return timer.TaskPriorityHigh
	default:
		return timer.TaskPriorityNormal
	}
}

// convertScheduleType 转换调度类型
// 将数据库中的整数调度类型值转换为timer系统的ScheduleType枚举
// 参数:
//
//	scheduleType: 数据库中的调度类型整数值
//
// 返回:
//
//	timer.ScheduleType: 对应的调度类型枚举值
func (init *BaseTaskInitializer) convertScheduleType(scheduleType int) timer.ScheduleType {
	switch scheduleType {
	case timertypes.ScheduleTypeOneTime:
		return timer.ScheduleTypeOnce
	case timertypes.ScheduleTypeInterval:
		return timer.ScheduleTypeInterval
	case timertypes.ScheduleTypeCron:
		return timer.ScheduleTypeCron
	case timertypes.ScheduleTypeDelay:
		return timer.ScheduleTypeDelay
	default:
		return timer.ScheduleTypeOnce
	}
}

// convertTaskStatus 转换任务状态
// 将数据库中的整数任务状态值转换为timer系统的TaskStatus枚举
// 参数:
//
//	status: 数据库中的任务状态整数值
//
// 返回:
//
//	timer.TaskStatus: 对应的任务状态枚举值
func (init *BaseTaskInitializer) convertTaskStatus(status int) timer.TaskStatus {
	switch status {
	case timertypes.TaskStatusPending:
		return timer.TaskStatusPending
	case timertypes.TaskStatusRunning:
		return timer.TaskStatusRunning
	case timertypes.TaskStatusCompleted:
		return timer.TaskStatusCompleted
	case timertypes.TaskStatusFailed:
		return timer.TaskStatusFailed
	case timertypes.TaskStatusCancelled:
		return timer.TaskStatusCancelled
	default:
		return timer.TaskStatusPending
	}
}

// getOrCreateScheduler 获取或创建调度器
// 根据租户ID和调度器ID获取已存在的调度器，如果不存在则创建新的调度器
// 调度器是任务执行的核心组件，负责任务的调度和执行管理
// 参数:
//
//	ctx: 上下文对象，用于控制请求生命周期
//	tenantId: 租户ID，用于多租户隔离
//	schedulerId: 调度器ID指针，可能为空
//
// 返回:
//
//	timer.TaskScheduler: 获取或创建的调度器实例
//	error: 操作过程中的错误信息
func (init *BaseTaskInitializer) getOrCreateScheduler(ctx context.Context, tenantId string, schedulerId *string) (timer.TaskScheduler, error) {
	// 确定调度器ID：优先使用指定的ID，否则生成默认ID
	var schedId string
	if schedulerId != nil && *schedulerId != "" {
		schedId = *schedulerId
	} else {
		// 生成默认调度器ID：执行器类型_scheduler_租户ID
		schedId = fmt.Sprintf("%s_scheduler_%s", init.factory.GetExecutorType(), tenantId)
	}

	// 尝试从全局定时器池中获取已存在的调度器
	scheduler, err := init.timerPool.GetScheduler(schedId)
	if err == nil {
		// 调度器已存在，直接返回复用
		logger.Info("使用已存在的调度器", "schedulerId", schedId)
		return scheduler, nil
	}

	// 调度器不存在，需要创建新的调度器实例
	logger.Info("创建新的调度器", "schedulerId", schedId, "tenantId", tenantId)
	return init.createNewScheduler(ctx, tenantId, schedId)
}

// createNewScheduler 创建新的调度器
// 根据配置参数创建新的调度器实例，并将其注册到全局定时器池中
// 新创建的调度器会自动启动，开始处理任务调度
// 调度器配置优先从数据库加载，如果数据库中没有配置则使用默认值
// 参数:
//
//	ctx: 上下文对象，用于控制请求生命周期
//	tenantId: 租户ID，用于标识和隔离
//	schedulerId: 调度器的唯一标识符
//
// 返回:
//
//	timer.TaskScheduler: 创建的调度器实例
//	error: 创建过程中的错误信息
func (init *BaseTaskInitializer) createNewScheduler(ctx context.Context, tenantId, schedulerId string) (timer.TaskScheduler, error) {
	// 尝试从数据库加载调度器配置
	config, err := init.loadSchedulerConfig(ctx, tenantId, schedulerId)
	if err != nil {
		// 如果数据库加载失败，使用默认配置
		logger.Warn("从数据库加载调度器配置失败，使用默认配置",
			"schedulerId", schedulerId,
			"tenantId", tenantId,
			"error", err)
		config = init.createDefaultSchedulerConfig(tenantId, schedulerId)
	}

	// 通过定时器池创建调度器实例
	scheduler, err := init.timerPool.CreateScheduler(config)
	if err != nil {
		return nil, fmt.Errorf("创建调度器失败: %w", err)
	}

	// 启动调度器，开始任务调度循环
	if err := scheduler.Start(); err != nil {
		// 启动失败时记录警告，但不返回错误，因为调度器可以稍后手动启动
		logger.Warn("启动调度器失败", "schedulerId", schedulerId, "error", err)
	}

	return scheduler, nil
}

// loadSchedulerConfig 从数据库加载调度器配置
// 根据租户ID和调度器ID查询数据库中的调度器配置信息
// 支持多租户环境下的调度器配置隔离和个性化设置
// 参数:
//
//	ctx: 上下文对象，用于控制请求生命周期
//	tenantId: 租户ID，用于多租户隔离
//	schedulerId: 调度器ID，用于唯一标识调度器
//
// 返回:
//
//	*timer.SchedulerConfig: 从数据库加载的调度器配置
//	error: 加载过程中的错误信息
func (init *BaseTaskInitializer) loadSchedulerConfig(ctx context.Context, tenantId, schedulerId string) (*timer.SchedulerConfig, error) {
	// 查询调度器配置
	schedulerData, err := init.daoManager.GetSchedulerDAO().GetSchedulerById(ctx, tenantId, schedulerId)
	if err != nil {
		return nil, fmt.Errorf("查询调度器配置失败: %w", err)
	}

	if schedulerData == nil {
		return nil, fmt.Errorf("未找到调度器配置: schedulerId=%s, tenantId=%s", schedulerId, tenantId)
	}

	// 转换数据库配置为调度器配置对象
	config := &timer.SchedulerConfig{
		ID:       schedulerId,
		Name:     schedulerData.SchedulerName,
		TenantId: tenantId,                           // 设置租户ID
		Tasks:    make(map[string]*timer.TaskConfig), // 初始化任务映射
	}

	// 设置工作线程数，使用数据库配置或默认值
	if schedulerData.MaxWorkers > 0 {
		config.MaxWorkers = schedulerData.MaxWorkers
	} else {
		config.MaxWorkers = 10 // 默认值
	}

	// 设置队列大小，使用数据库配置或默认值
	if schedulerData.QueueSize > 0 {
		config.QueueSize = schedulerData.QueueSize
	} else {
		config.QueueSize = 100 // 默认值
	}

	// 设置默认超时时间，使用数据库配置或默认值
	if schedulerData.DefaultTimeoutSeconds > 0 {
		config.DefaultTimeout = time.Duration(schedulerData.DefaultTimeoutSeconds) * time.Second
	} else {
		config.DefaultTimeout = 30 * time.Minute // 默认30分钟
	}

	// 设置默认重试次数，使用数据库配置或默认值
	if schedulerData.DefaultRetries >= 0 {
		config.DefaultRetries = schedulerData.DefaultRetries
	} else {
		config.DefaultRetries = 3 // 默认重试3次
	}

	// 设置调度间隔，使用默认值（数据库表中没有此字段）
	config.ScheduleInterval = 10 * time.Second // 默认10秒

	logger.Info("从数据库加载调度器配置成功",
		"schedulerId", schedulerId,
		"tenantId", tenantId,
		"schedulerName", config.Name,
		"configTenantId", config.TenantId,
		"maxWorkers", config.MaxWorkers,
		"queueSize", config.QueueSize,
		"defaultTimeout", config.DefaultTimeout,
		"defaultRetries", config.DefaultRetries,
		"scheduleInterval", config.ScheduleInterval)

	return config, nil
}

// createDefaultSchedulerConfig 创建默认调度器配置
// 当数据库中没有配置或加载失败时，使用此方法创建默认的调度器配置
// 默认配置适用于大多数场景，提供合理的性能和资源使用平衡
// 参数:
//
//	tenantId: 租户ID，用于生成调度器名称
//	schedulerId: 调度器ID，用于唯一标识
//
// 返回:
//
//	*timer.SchedulerConfig: 默认的调度器配置对象
func (init *BaseTaskInitializer) createDefaultSchedulerConfig(tenantId, schedulerId string) *timer.SchedulerConfig {
	config := &timer.SchedulerConfig{
		ID:               schedulerId,                                                       // 调度器唯一标识
		Name:             fmt.Sprintf("%s调度器_%s", init.factory.GetExecutorType(), tenantId), // 调度器显示名称
		TenantId:         tenantId,                                                          // 租户ID标识
		MaxWorkers:       10,                                                                // 最大工作线程数
		QueueSize:        100,                                                               // 任务队列大小
		DefaultTimeout:   30 * time.Minute,                                                  // 默认任务超时时间
		DefaultRetries:   3,                                                                 // 默认重试次数
		ScheduleInterval: 10 * time.Second,                                                  // 调度扫描间隔
		Tasks:            make(map[string]*timer.TaskConfig),                                // 初始化任务映射
	}

	logger.Info("使用默认调度器配置",
		"schedulerId", schedulerId,
		"tenantId", tenantId,
		"configTenantId", config.TenantId,
		"maxWorkers", config.MaxWorkers,
		"queueSize", config.QueueSize,
		"defaultTimeout", config.DefaultTimeout,
		"defaultRetries", config.DefaultRetries,
		"scheduleInterval", config.ScheduleInterval)

	return config
}
