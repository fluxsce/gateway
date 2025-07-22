package timer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gateway/pkg/logger"
	"gateway/pkg/timer/cron"
	"gateway/pkg/timer/logwrite"
)

// StandardScheduler 标准任务调度器实现（使用独立调度协程）
// 采用定期扫描的方式调度任务，支持多种调度类型和并发执行
type StandardScheduler struct {
	config    *SchedulerConfig        // 调度器配置，包含任务存储和全局设置
	executors map[string]TaskExecutor // 任务ID到执行器的映射表

	// 调度器状态管理
	mu      sync.RWMutex       // 读写锁，保护调度器状态的并发安全
	running bool               // 调度器运行状态标志
	ctx     context.Context    // 调度器上下文，用于控制生命周期
	cancel  context.CancelFunc // 取消函数，用于停止调度器

	// 工作线程池管理
	taskQueue chan *taskJob  // 任务执行队列，缓冲待执行的任务
	wg        sync.WaitGroup // 等待组，确保所有goroutine正确退出

	// Cron表达式解析器
	cronParser cron.CronParser // 用于解析和计算cron表达式的执行时间

	// 调度控制机制
	schedulerTicker    *time.Ticker       // 调度定时器，控制任务扫描频率
	scheduleIntervalCh chan time.Duration // 调度间隔调整通道，支持动态修改扫描间隔
}

// taskJob 任务作业
// 封装了一次任务执行所需的所有信息
type taskJob struct {
	taskID   string       // 任务唯一标识符
	params   interface{}  // 任务执行参数
	executor TaskExecutor // 任务执行器，定义具体执行逻辑
	config   *TaskConfig  // 任务配置，包含调度规则和状态信息
}

// NewStandardScheduler 创建标准调度器实例
// 初始化调度器的所有组件，包括任务队列、执行器映射、cron解析器等
// 调度器创建后需要调用Start()方法才能开始工作
// 参数:
//
//	config: 调度器配置，如果为nil则使用默认配置
//
// 返回:
//
//	*StandardScheduler: 初始化完成的调度器实例
func NewStandardScheduler(config *SchedulerConfig) *StandardScheduler {
	// 如果没有提供配置，使用默认配置
	if config == nil {
		config = DefaultSchedulerConfig()
	} else {
		// 如果提供了配置但Tasks字段为nil，需要初始化它
		if config.Tasks == nil {
			config.Tasks = make(map[string]*TaskConfig)
		}
	}

	return &StandardScheduler{
		config:             config,                                // 调度器配置
		executors:          make(map[string]TaskExecutor),         // 初始化执行器映射表
		taskQueue:          make(chan *taskJob, config.QueueSize), // 创建任务队列，大小由配置决定
		cronParser:         cron.NewStandardCronParser(),          // 初始化cron表达式解析器
		scheduleIntervalCh: make(chan time.Duration, 1),           // 创建调度间隔调整通道
	}
}

// AddTask 添加新任务到调度器
// 将任务配置和执行器注册到调度器中，任务添加后会根据其调度类型自动计算下次执行时间
// 如果任务ID已存在，会先删除旧任务再添加新任务，确保任务配置是最新的
// 参数:
//
//	config: 任务配置，包含调度规则、重试策略等信息
//	executor: 任务执行器，定义具体的执行逻辑
//
// 返回:
//
//	error: 添加失败时返回错误信息
func (s *StandardScheduler) AddTask(config *TaskConfig, executor TaskExecutor) error {
	// 验证任务配置的有效性
	if err := ValidateTaskConfig(config); err != nil {
		return fmt.Errorf("invalid task config: %w", err)
	}

	// 验证执行器不能为空
	if executor == nil {
		return errors.New("executor cannot be nil")
	}

	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查任务ID是否已存在
	if existingExecutor, exists := s.executors[config.ID]; exists {
		// 任务已存在，先删除旧任务（在当前锁范围内直接处理，避免死锁）
		logger.Info("Task already exists, removing old task before adding new one", "taskID", config.ID)

		// 释放旧执行器资源
		if err := existingExecutor.Close(); err != nil {
			logger.Warn("关闭已存在的执行器失败", "taskID", config.ID, "error", err)
		} else {
			logger.Info("成功关闭已存在的执行器", "taskID", config.ID)
		}

		// 从内存中删除旧的执行器映射（注意：这里不能调用RemoveTask方法，会导致死锁）
		delete(s.executors, config.ID)

		// 从调度器配置中删除旧的任务配置
		s.config.RemoveTask(config.ID)

		logger.Info("Old task removed successfully", "taskID", config.ID)
	}

	// 初始化任务的基本信息
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now() // 设置创建时间
	}
	if config.Status == 0 {
		config.UpdateStatus(TaskStatusPending) // 设置初始状态为待执行
	}

	// 根据任务配置计算并设置下次执行时间
	nextRunTime := s.calculateNextRunTime(config)
	if !nextRunTime.IsZero() {
		config.SetNextRunTime(&nextRunTime) // 设置下次执行时间
	}

	// 将任务添加到调度器配置中
	s.config.AddTask(config)

	// 在内存中注册执行器，建立任务ID到执行器的映射
	s.executors[config.ID] = executor

	return nil
}

// RemoveTask 从调度器中移除指定任务
// 停止任务的调度，清理相关资源
// 参数:
//
//	taskID: 要移除的任务ID
//
// 返回:
//
//	error: 移除失败时返回错误信息
func (s *StandardScheduler) RemoveTask(taskID string) error {
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取要移除的执行器
	executor, exists := s.executors[taskID]
	if exists {
		// 释放执行器资源
		if err := executor.Close(); err != nil {
			logger.Warn("移除任务时关闭执行器失败", "taskID", taskID, "error", err)
		} else {
			logger.Info("移除任务时成功关闭执行器", "taskID", taskID)
		}

		// 从内存中删除执行器映射
		delete(s.executors, taskID)
	}

	// 从调度器配置中删除任务配置
	s.config.RemoveTask(taskID)

	return nil
}

// GetTask 获取指定任务的详细信息
// 从调度器配置中获取任务配置信息
// 参数:
//
//	taskID: 任务ID
//
// 返回:
//
//	*TaskConfig: 任务配置对象，包含完整的任务状态
//	error: 获取失败时返回错误信息
func (s *StandardScheduler) GetTask(taskID string) (*TaskConfig, error) {
	// 从调度器配置中查找任务
	config, exists := s.config.GetTask(taskID)
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}
	return config, nil
}

// ListTasks 获取所有任务的列表
// 从调度器配置中获取所有任务配置
// 返回:
//
//	[]*TaskConfig: 所有任务配置的切片
//	error: 获取失败时返回错误信息
func (s *StandardScheduler) ListTasks() ([]*TaskConfig, error) {
	// 直接返回调度器配置中的所有任务
	return s.config.ListTasks(), nil
}

// StartTask 启动指定的任务调度
// 将任务标记为启用状态，并重新计算下次执行时间
// 参数:
//
//	taskID: 要启动的任务ID
//
// 返回:
//
//	error: 启动失败时返回错误信息
func (s *StandardScheduler) StartTask(taskID string) error {
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()

	// 从调度器配置中获取任务
	config, exists := s.config.GetTask(taskID)
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// 如果任务未启用，将其设置为启用状态
	if !config.Enabled {
		config.Enabled = true // 启用任务
		// 重新计算下次执行时间，确保任务能够被调度
		nextRunTime := s.calculateNextRunTime(config)
		if !nextRunTime.IsZero() {
			config.SetNextRunTime(&nextRunTime) // 更新下次执行时间
		}
	}

	return nil
}

// StopTask 停止指定的任务调度
// 将任务标记为禁用状态，并清空下次执行时间
// 参数:
//
//	taskID: 要停止的任务ID
//
// 返回:
//
//	error: 停止失败时返回错误信息
func (s *StandardScheduler) StopTask(taskID string) error {
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()

	// 从调度器配置中获取任务
	config, exists := s.config.GetTask(taskID)
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// 如果任务已启用，将其设置为禁用状态
	if config.Enabled {
		config.Enabled = false // 禁用任务
		// 清空下次执行时间，防止任务继续被调度
		config.SetNextRunTime(nil)
	}

	return nil
}

// TriggerTask 手动触发任务执行
// 立即执行指定的任务，不影响其正常的调度计划
// 参数:
//
//	taskID: 要触发的任务ID
//	params: 执行参数，会覆盖任务配置中的参数
//
// 返回:
//
//	error: 触发失败时返回错误信息
func (s *StandardScheduler) TriggerTask(taskID string, params interface{}) error {
	// 获取任务配置和执行器（使用读锁，允许并发读取）
	s.mu.RLock()
	executor, exists := s.executors[taskID]
	config, configExists := s.config.GetTask(taskID)
	s.mu.RUnlock()

	// 检查任务和执行器是否存在
	if !exists || !configExists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// 创建任务作业并提交到队列
	job := &taskJob{
		taskID:   taskID,
		params:   params, // 使用传入的参数，覆盖配置中的参数
		executor: executor,
		config:   config,
	}

	// 尝试将任务加入执行队列
	select {
	case s.taskQueue <- job:
		return nil // 任务成功加入队列
	default:
		return errors.New("task queue is full") // 队列已满，无法加入
	}
}

// Start 启动整个调度器
// 启动工作线程池和调度协程，开始任务调度
// 返回:
//
//	error: 启动失败时返回错误信息
func (s *StandardScheduler) Start() error {
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查调度器是否已经在运行
	if s.running {
		return errors.New("scheduler is already running")
	}

	// 创建上下文用于控制调度器生命周期
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// 启动工作线程池，处理任务执行
	for i := 0; i < s.config.MaxWorkers; i++ {
		s.wg.Add(1)
		go s.worker() // 启动工作线程
	}

	// 启动调度协程，负责扫描和调度任务
	s.wg.Add(1)
	go s.scheduler()

	// 标记调度器为运行状态
	s.running = true
	logger.Info("Task scheduler started")

	return nil
}

// Stop 停止整个调度器
// 停止所有任务调度，清理资源，等待正在执行的任务完成
// 返回:
//
//	error: 停止失败时返回错误信息
func (s *StandardScheduler) Stop() error {
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查调度器是否正在运行
	if !s.running {
		return errors.New("scheduler is not running")
	}

	// 取消上下文，通知所有goroutine停止
	if s.cancel != nil {
		s.cancel() // 发送停止信号
	}

	// 关闭任务队列，不再接收新任务（避免重复关闭）
	select {
	case <-s.taskQueue:
		// 队列已关闭
	default:
		close(s.taskQueue)
	}

	// 等待所有工作线程和调度协程结束
	s.wg.Wait()

	// 释放所有执行器资源
	s.closeAllExecutors()

	// 标记调度器为停止状态
	s.running = false
	logger.Info("Task scheduler stopped")

	return nil
}

// IsRunning 检查调度器是否正在运行
// 返回:
//
//	bool: true表示正在运行，false表示已停止
func (s *StandardScheduler) IsRunning() bool {
	// 使用读锁检查运行状态
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetRunningTasks 获取当前正在运行的任务列表
// 返回:
//
//	[]*TaskConfig: 正在运行的任务配置列表
//	error: 获取失败时返回错误信息
func (s *StandardScheduler) GetRunningTasks() ([]*TaskConfig, error) {
	// 获取所有任务配置
	tasks := s.config.ListTasks()
	runningTasks := make([]*TaskConfig, 0)

	// 筛选出状态为运行中的任务
	for _, task := range tasks {
		if task.GetStatus() == TaskStatusRunning {
			runningTasks = append(runningTasks, task)
		}
	}

	return runningTasks, nil
}

// SetScheduleInterval 动态设置调度检查间隔
// 参数:
//
//	interval: 新的调度间隔，必须大于0
//
// 返回:
//
//	error: 设置失败时返回错误信息
func (s *StandardScheduler) SetScheduleInterval(interval time.Duration) error {
	if interval <= 0 {
		return errors.New("schedule interval must be greater than 0")
	}

	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	if !running {
		// 调度器未运行，直接更新配置
		s.config.ScheduleInterval = interval
		return nil
	}

	// 调度器正在运行，通过通道发送新间隔
	select {
	case s.scheduleIntervalCh <- interval:
		return nil
	default:
		return errors.New("failed to update schedule interval: channel is full")
	}
}

// GetScheduleInterval 获取当前调度检查间隔
// 返回:
//
//	time.Duration: 当前的调度间隔
func (s *StandardScheduler) GetScheduleInterval() time.Duration {
	return s.config.ScheduleInterval
}

// closeAllExecutors 关闭所有任务执行器并释放资源
// 遍历所有注册的执行器，调用其Close方法释放资源
// 该方法在调度器停止时被调用，确保所有资源得到正确释放
func (s *StandardScheduler) closeAllExecutors() {
	// 遍历所有执行器
	for taskID, executor := range s.executors {
		// 调用执行器的Close方法释放资源
		if err := executor.Close(); err != nil {
			logger.Warn("关闭执行器失败", "taskID", taskID, "error", err)
		} else {
			logger.Info("成功关闭执行器", "taskID", taskID)
		}
	}

	// 清空执行器映射表
	s.executors = make(map[string]TaskExecutor)
	logger.Info("All task executors have been closed and cleaned up")
}

// scheduler 调度协程，定期扫描任务并将需要执行的任务放入队列
// 这是调度器的核心协程，负责周期性地检查所有任务并调度需要执行的任务
func (s *StandardScheduler) scheduler() {
	defer s.wg.Done() // 确保在退出时通知WaitGroup

	// 获取调度间隔配置
	interval := s.config.ScheduleInterval
	if interval <= 0 {
		interval = time.Second // 默认1秒扫描一次
	}

	// 创建定时器，用于定期触发任务扫描
	s.schedulerTicker = time.NewTicker(interval)
	defer s.schedulerTicker.Stop() // 确保在退出时停止定时器

	// 调度器主循环
	for {
		select {
		case <-s.schedulerTicker.C:
			// 定时器触发，执行任务扫描和调度
			s.checkAndScheduleTasks()
		case newInterval := <-s.scheduleIntervalCh:
			// 接收到动态调整调度间隔的请求
			s.schedulerTicker.Stop()                        // 停止当前定时器
			s.schedulerTicker = time.NewTicker(newInterval) // 创建新的定时器
			s.config.ScheduleInterval = newInterval         // 更新配置
			logger.Info("调度间隔已更新", "interval", newInterval)
		case <-s.ctx.Done():
			// 调度器停止信号，退出主循环
			return
		}
	}
}

// checkAndScheduleTasks 检查并调度需要执行的任务
// 扫描所有任务配置，找出需要执行的任务并将其放入执行队列
func (s *StandardScheduler) checkAndScheduleTasks() {
	now := time.Now()             // 获取当前时间作为调度基准
	tasks := s.config.ListTasks() // 获取所有任务配置

	// 遍历所有任务，检查是否需要执行
	for _, config := range tasks {
		// 检查任务是否应该在当前时间执行
		if !s.shouldExecuteNow(config, now) {
			continue // 跳过不需要执行的任务
		}

		// 获取任务对应的执行器
		s.mu.RLock()
		executor, exists := s.executors[config.ID]
		s.mu.RUnlock()

		// 如果执行器不存在，跳过该任务
		if !exists {
			continue
		}

		// 创建任务作业对象
		job := &taskJob{
			taskID:   config.ID,
			params:   config.Params, // 使用任务配置中的参数
			executor: executor,
			config:   config,
		}

		// 尝试将任务放入执行队列
		select {
		case s.taskQueue <- job:
			// 任务提交成功，清空下次执行时间（防止重复调度）
			// 实际的下次执行时间将在任务完成后更新
			config.SetNextRunTime(nil)
		case <-s.ctx.Done():
			// 调度器已停止，退出调度
			return
		default:
			// 队列已满，跳过此次调度并记录警告
			logger.Warn("任务队列已满，跳过任务执行", "taskID", config.ID)
		}
	}
}

// shouldExecuteNow 检查任务是否应该现在执行
// 综合检查任务的各种条件，判断是否应该在当前时间执行
// 参数:
//
//	config: 任务配置
//	now: 当前时间
//
// 返回:
//
//	bool: true表示应该执行，false表示不应该执行
func (s *StandardScheduler) shouldExecuteNow(config *TaskConfig, now time.Time) bool {
	// 检查任务基本条件：是否启用且满足启动条件
	if !config.Enabled || !config.ShouldStart() {
		return false
	}

	// 检查任务是否正在运行（避免重复执行）
	if config.GetStatus() == TaskStatusRunning {
		return false
	}

	// 根据调度类型进行不同的判断
	switch config.ScheduleType {
	case ScheduleTypeOnce:
		// 一次性任务：检查是否已执行过
		if config.GetRunCount() > 0 {
			return false // 已执行过，不再调度
		}

		// 检查是否设置了开始时间且到了执行时间
		if config.StartTime != nil {
			return now.After(*config.StartTime) || now.Equal(*config.StartTime)
		}

		// 没有设置开始时间，检查是否设置了下次执行时间
		nextRunTime := config.GetNextRunTime()
		return nextRunTime != nil && (now.After(*nextRunTime) || now.Equal(*nextRunTime))

	case ScheduleTypeDelay:
		// 延迟执行任务：检查是否已执行过
		if config.GetRunCount() > 0 {
			return false // 已执行过，不再调度
		}

		// 检查是否到了下次执行时间
		nextRunTime := config.GetNextRunTime()
		return nextRunTime != nil && (now.After(*nextRunTime) || now.Equal(*nextRunTime))

	case ScheduleTypeInterval:
		// 固定间隔任务：检查是否到了下次执行时间
		nextRunTime := config.GetNextRunTime()
		if nextRunTime == nil {
			return false
		}

		// 检查当前时间是否已到达或超过计划执行时间
		// 注意：now.Sub(nextRunTime)为负值表示还未到执行时间
		return now.After(*nextRunTime) || now.Equal(*nextRunTime)

	case ScheduleTypeCron:
		// Cron表达式任务：检查是否到了下次执行时间
		nextRunTime := config.GetNextRunTime()
		if nextRunTime == nil {
			return false
		}

		// 检查当前时间是否已到达或超过计划执行时间
		return now.After(*nextRunTime) || now.Equal(*nextRunTime)

	default:
		// 未知的调度类型，不执行
		logger.Warn("未知的调度类型", "taskID", config.ID, "scheduleType", config.ScheduleType)
		return false
	}
}

// updateNextRunTime 更新任务的下次执行时间
// 根据任务类型和配置重新计算下次执行时间
// 参数:
//
//	config: 任务配置
func (s *StandardScheduler) updateNextRunTime(config *TaskConfig) {
	// 重新计算下次执行时间
	nextRunTime := s.calculateNextRunTime(config)
	if nextRunTime.IsZero() {
		// 如果无法计算下次执行时间（如一次性任务已完成），清空执行时间
		logger.Warn("无法计算下次执行时间", "taskID", config.ID)
		config.SetNextRunTime(nil)
	} else {
		// 设置新的下次执行时间
		config.SetNextRunTime(&nextRunTime)
	}
}

// calculateNextRunTime 根据任务配置计算下次执行时间
// 根据不同的调度类型计算任务的下次执行时间
// 参数:
//
//	config: 任务配置
//
// 返回:
//
//	time.Time: 下次执行时间，如果返回零值表示不再调度
func (s *StandardScheduler) calculateNextRunTime(config *TaskConfig) time.Time {
	now := time.Now() // 获取当前时间作为计算基准

	// 根据调度类型计算下次执行时间
	switch config.ScheduleType {
	case ScheduleTypeOnce:
		// 一次性任务：只执行一次
		if config.GetRunCount() > 0 {
			return time.Time{} // 已执行过，不再调度
		}
		// 使用指定的开始时间，如果没有则立即执行
		if config.StartTime != nil {
			return *config.StartTime
		}
		return now

	case ScheduleTypeDelay:
		// 延迟执行任务：延迟指定时间后执行一次
		if config.GetRunCount() > 0 {
			return time.Time{} // 已执行过，不再调度
		}
		return now.Add(config.Delay) // 从当前时间延迟执行

	case ScheduleTypeInterval:
		// 固定间隔任务：按固定间隔重复执行
		lastRunTime := config.GetLastRunTime()
		if lastRunTime != nil {
			// 基于上次执行时间计算下次执行时间
			return lastRunTime.Add(config.Interval)
		}

		// 首次执行：使用开始时间或当前时间
		if config.StartTime != nil {
			return *config.StartTime
		}
		return now

	case ScheduleTypeCron:
		// Cron表达式任务：使用cron表达式计算执行时间
		if config.CronExpr == "" {
			return time.Time{} // cron表达式为空，无法调度
		}

		// 解析cron表达式
		schedule, err := s.cronParser.Parse(config.CronExpr)
		if err != nil {
			logger.Warn("解析cron表达式失败", "taskID", config.ID, "cronExpr", config.CronExpr, "error", err)
			return time.Time{} // 解析失败，无法调度
		}

		// 计算基于当前时间的下次执行时间
		return schedule.Next(now)
	}

	// 未知的调度类型，返回零值
	return time.Time{}
}

// worker 工作线程，处理任务队列中的任务
// 从任务队列中获取任务并执行，直到调度器停止
func (s *StandardScheduler) worker() {
	defer s.wg.Done() // 确保在退出时通知WaitGroup

	for {
		select {
		case job, ok := <-s.taskQueue:
			if !ok {
				return // 队列已关闭，退出工作线程
			}
			s.runTask(job) // 执行任务
		case <-s.ctx.Done():
			return // 调度器已停止，退出工作线程
		}
	}
}

// runTask 执行单个任务
// 处理任务的完整生命周期，包括状态更新、执行、重试和结果记录
// 参数:
//
//	job: 要执行的任务作业
func (s *StandardScheduler) runTask(job *taskJob) {
	// 更新任务状态为运行中
	job.config.UpdateStatus(TaskStatusRunning)

	// 创建任务执行上下文，设置超时时间
	timeout := job.config.Timeout
	if timeout <= 0 {
		timeout = s.config.DefaultTimeout // 使用默认超时时间
	}

	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel() // 确保释放上下文资源

	// 创建任务结果对象，记录执行过程
	result := NewTaskResult(job.taskID)

	// 执行任务（包含重试逻辑）
	err := s.executeWithRetry(ctx, job, result)
	if err != nil {
		result.Fail(err) // 标记任务执行失败
		logger.Error("任务执行失败", "taskID", job.taskID, "error", err, "duration", result.Duration)
	} else {
		result.Complete() // 标记任务执行成功
	}

	// 更新任务的运行信息（执行次数、最后执行时间等）
	job.config.UpdateRunInfo(result)

	// 任务完成后，为重复执行的任务更新下次执行时间
	s.updateNextRunTime(job.config)

	// 写入任务执行日志到数据库
	if err := s.writeExecutionLog(ctx, job, result); err != nil {
		logger.Error("日志写入失败", "taskID", job.taskID, "error", err)
	}

	// 记录任务执行完成的日志
	logger.Info("任务执行完成", "taskID", job.taskID, "status", result.Status.String(), "duration", result.Duration, "retryCount", result.RetryCount)
}

// executeWithRetry 带重试的任务执行
// 执行任务并处理重试逻辑，支持技术失败和业务失败的重试
// 参数:
//
//	ctx: 执行上下文，包含超时控制
//	job: 任务作业对象
//	result: 任务结果对象，用于记录执行过程
//
// 返回:
//
//	error: 最终执行失败时返回错误信息
func (s *StandardScheduler) executeWithRetry(ctx context.Context, job *taskJob, result *TaskResult) error {
	// 获取最大执行次数（包括首次执行）
	maxAttempts := job.config.MaxRetries
	if maxAttempts <= 0 {
		maxAttempts = s.config.DefaultRetries // 使用默认重试次数
	}
	// 确保至少执行一次
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	// 获取重试间隔
	retryInterval := job.config.RetryInterval
	if retryInterval <= 0 {
		retryInterval = time.Second * 5 // 默认重试间隔5秒
	}

	var lastErr error // 记录最后一次错误

	// 执行重试循环（maxAttempts次总执行次数）
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			result.RetryCount++ // 增加重试计数
			// 等待重试间隔，支持上下文取消
			select {
			case <-time.After(retryInterval):
			case <-ctx.Done():
				return ctx.Err() // 上下文已取消，返回取消错误
			}
		}

		// 调用执行器执行任务
		executeResult, err := job.executor.Execute(ctx, job.params)
		if err != nil {
			lastErr = err
			logger.Warn("任务重试失败", "taskID", job.taskID, "attempt", attempt+1, "maxAttempts", maxAttempts, "error", err)
			continue // 技术执行失败，继续重试
		}

		// 保存执行结果
		result.Result = executeResult

		// 检查业务执行是否成功
		if executeResult != nil && !executeResult.Success {
			lastErr = fmt.Errorf("task business logic failed: %s", executeResult.Message)
			logger.Warn("业务逻辑失败", "taskID", job.taskID, "attempt", attempt+1, "maxAttempts", maxAttempts, "message", executeResult.Message)
			continue // 业务逻辑失败，继续重试
		}

		// 执行成功，返回nil
		return nil
	}

	// 所有重试都失败，返回最后一次错误
	return lastErr
}

// writeExecutionLog 写入任务执行日志到数据库
// 简单调用静态方法，不处理具体的日志逻辑
// 参数:
//
//	ctx: 上下文
//	job: 任务作业对象
//	result: 任务执行结果
//
// 返回:
//
//	error: 写入失败时返回错误信息
func (s *StandardScheduler) writeExecutionLog(ctx context.Context, job *taskJob, result *TaskResult) error {
	// 直接调用静态方法写入日志，所有逻辑由logwrite包处理
	return logwrite.WriteTaskExecutionLog(
		ctx,
		job.config,            // 任务配置
		result,                // 任务执行结果
		job.config.MaxRetries, // 最大重试次数
		s.config.TenantId,     // 租户ID，由logwrite包设置默认值
		s.config.ID,           // 调度器ID，可选
	)
}
