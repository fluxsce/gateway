package timer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	
	"gohub/pkg/timer/cron"
)

// StandardScheduler 标准任务调度器实现
type StandardScheduler struct {
	config    *SchedulerConfig
	storage   TaskStorage
	executors map[string]TaskExecutor // 任务ID -> 执行器映射
	
	// 调度器状态
	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	
	// 工作线程池
	taskQueue chan *taskJob
	wg        sync.WaitGroup
	
	// 任务调度
	timers map[string]*time.Timer // 任务ID -> 定时器映射
	
	// Cron解析器
	cronParser cron.CronParser
}

// taskJob 任务作业
type taskJob struct {
	taskID   string
	params   interface{}
	executor TaskExecutor
	config   *TaskConfig
}

// NewStandardScheduler 创建标准调度器实例
// 初始化调度器的各个组件，包括存储、执行器映射、任务队列等
// 参数:
//   config: 调度器配置，如果为nil则使用默认配置
//   storage: 任务存储接口实现，用于持久化任务数据
// 返回:
//   *StandardScheduler: 初始化完成的调度器实例
func NewStandardScheduler(config *SchedulerConfig, storage TaskStorage) *StandardScheduler {
	// 使用默认配置（如果未提供配置）
	if config == nil {
		config = DefaultSchedulerConfig()
	}
	
	return &StandardScheduler{
		config:     config,                                    // 调度器配置
		storage:    storage,                                   // 存储接口
		executors:  make(map[string]TaskExecutor),            // 任务ID到执行器的映射
		timers:     make(map[string]*time.Timer),             // 任务ID到定时器的映射
		taskQueue:  make(chan *taskJob, config.QueueSize),    // 任务执行队列
		cronParser: cron.NewStandardCronParser(),             // Cron表达式解析器
	}
}

// AddTask 添加新任务到调度器
// 将任务配置和执行器注册到调度器中，如果调度器正在运行且任务已启用，会立即开始调度
// 参数:
//   config: 任务配置，包含调度规则、重试策略等信息
//   executor: 任务执行器，定义具体的执行逻辑
// 返回:
//   error: 添加失败时返回错误信息
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
	
	// 检查任务ID是否已存在（任务ID必须唯一）
	if _, exists := s.executors[config.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", config.ID)
	}
	
	// 将任务配置持久化到存储中
	if err := s.storage.SaveTaskConfig(config); err != nil {
		return fmt.Errorf("failed to save task config: %w", err)
	}
	
	// 创建并保存任务运行时信息
	taskInfo := NewTaskInfo(config)
	if err := s.storage.SaveTaskInfo(taskInfo); err != nil {
		return fmt.Errorf("failed to save task info: %w", err)
	}
	
	// 在内存中注册执行器
	s.executors[config.ID] = executor
	
	// 如果调度器正在运行且任务已启用，立即开始调度
	if s.running && config.Enabled && config.ShouldStart() {
		s.scheduleTask(config)
	}
	
	return nil
}

// RemoveTask 从调度器中移除指定任务
// 停止任务的调度，清理相关资源，并从存储中删除任务数据
// 参数:
//   taskID: 要移除的任务ID
// 返回:
//   error: 移除失败时返回错误信息
func (s *StandardScheduler) RemoveTask(taskID string) error {
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 停止任务调度定时器
	if timer, exists := s.timers[taskID]; exists {
		timer.Stop()                    // 停止定时器
		delete(s.timers, taskID)        // 从定时器映射中删除
	}
	
	// 从内存中删除执行器
	delete(s.executors, taskID)
	
	// 从存储中删除任务配置数据
	return s.storage.DeleteTaskConfig(taskID)
}

// GetTask 获取指定任务的详细信息
// 从存储中加载任务的运行时信息，包含配置、状态、执行历史等
// 参数:
//   taskID: 任务ID
// 返回:
//   *TaskInfo: 任务信息对象，包含完整的任务状态
//   error: 获取失败时返回错误信息
func (s *StandardScheduler) GetTask(taskID string) (*TaskInfo, error) {
	return s.storage.LoadTaskInfo(taskID)
}

// ListTasks 获取所有任务的列表
// 从存储中加载所有任务的配置和运行时信息
// 返回:
//   []*TaskInfo: 所有任务信息的切片，按任务ID排序
//   error: 获取失败时返回错误信息
func (s *StandardScheduler) ListTasks() ([]*TaskInfo, error) {
	// 从存储中获取所有任务配置
	configs, err := s.storage.ListTaskConfigs()
	if err != nil {
		return nil, err
	}
	
	var tasks []*TaskInfo
	// 为每个任务配置获取对应的运行时信息
	for _, config := range configs {
		taskInfo, err := s.storage.LoadTaskInfo(config.ID)
		if err != nil {
			// 如果任务信息不存在，基于配置创建一个基本的任务信息
			taskInfo = NewTaskInfo(config)
		}
		tasks = append(tasks, taskInfo)
	}
	
	return tasks, nil
}

// StartTask 启动指定的任务调度
// 将任务标记为启用状态，如果调度器正在运行则立即开始调度该任务
// 参数:
//   taskID: 要启动的任务ID
// 返回:
//   error: 启动失败时返回错误信息
func (s *StandardScheduler) StartTask(taskID string) error {
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 从存储中加载任务配置
	config, err := s.storage.LoadTaskConfig(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	
	// 如果任务未启用，将其设置为启用状态
	if !config.Enabled {
		config.Enabled = true
		if err := s.storage.SaveTaskConfig(config); err != nil {
			return fmt.Errorf("failed to update task config: %w", err)
		}
	}
	
	// 如果调度器正在运行且任务满足启动条件，立即开始调度
	if s.running && config.ShouldStart() {
		s.scheduleTask(config)
	}
	
	return nil
}

// StopTask 停止指定的任务调度
// 停止任务的定时器，将任务标记为禁用状态，但不删除任务配置
// 参数:
//   taskID: 要停止的任务ID
// 返回:
//   error: 停止失败时返回错误信息
func (s *StandardScheduler) StopTask(taskID string) error {
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 停止任务的定时器
	if timer, exists := s.timers[taskID]; exists {
		timer.Stop()                    // 停止定时器
		delete(s.timers, taskID)        // 从定时器映射中删除
	}
	
	// 从存储中加载任务配置并更新状态
	config, err := s.storage.LoadTaskConfig(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	
	// 将任务标记为禁用状态并保存
	config.Enabled = false
	return s.storage.SaveTaskConfig(config)
}

// TriggerTask 手动触发任务执行
// 立即将任务加入执行队列，不影响任务的正常调度计划
// 参数:
//   taskID: 要触发的任务ID
//   params: 执行参数，会覆盖任务配置中的参数，如果为nil则使用配置中的参数
// 返回:
//   error: 触发失败时返回错误信息
func (s *StandardScheduler) TriggerTask(taskID string, params interface{}) error {
	// 获取任务执行器（使用读锁提高并发性能）
	s.mu.RLock()
	executor, exists := s.executors[taskID]
	s.mu.RUnlock()
	
	// 检查任务是否存在
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}
	
	// 从存储中加载任务配置
	config, err := s.storage.LoadTaskConfig(taskID)
	if err != nil {
		return fmt.Errorf("failed to load task config: %w", err)
	}
	
	// 确定执行参数：优先使用提供的参数，否则使用配置中的参数
	taskParams := params
	if taskParams == nil {
		taskParams = config.Params
	}
	
	// 将任务提交到执行队列（非阻塞方式）
	select {
	case s.taskQueue <- &taskJob{
		taskID:   taskID,
		params:   taskParams,
		executor: executor,
		config:   config,
	}:
		return nil
	default:
		// 如果队列已满，返回错误而不是阻塞
		return errors.New("task queue is full")
	}
}

// Start 启动调度器
// 初始化工作线程池，加载已有任务并开始调度
// 这是调度器的主要启动入口，必须在添加任务后调用
// 返回:
//   error: 启动失败时返回错误信息
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
	s.running = true
	
	// 启动工作线程池处理任务执行
	for i := 0; i < s.config.MaxWorkers; i++ {
		s.wg.Add(1)
		go s.worker() // 启动工作协程
	}
	
	// 从存储中加载所有任务配置
	configs, err := s.storage.ListTaskConfigs()
	if err != nil {
		return fmt.Errorf("failed to load task configs: %w", err)
	}
	
	// 为所有已启用且满足启动条件的任务开始调度
	for _, config := range configs {
		if config.Enabled && config.ShouldStart() {
			s.scheduleTask(config) // 计算下次执行时间并设置定时器
		}
	}
	
	log.Printf("Scheduler %s started with %d workers", s.config.Name, s.config.MaxWorkers)
	return nil
}

// Stop 停止整个调度器
// 停止所有任务的调度，关闭工作线程池，等待正在执行的任务完成
// 这是调度器的优雅关闭方法，会等待所有任务执行完成后才返回
// 返回:
//   error: 停止失败时返回错误信息
func (s *StandardScheduler) Stop() error {
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 检查调度器是否正在运行
	if !s.running {
		return errors.New("scheduler is not running")
	}
	
	// 取消上下文，通知所有工作协程停止
	s.cancel()
	
	// 停止所有任务的定时器
	for _, timer := range s.timers {
		timer.Stop()
	}
	s.timers = make(map[string]*time.Timer)  // 重置定时器映射
	
	// 关闭任务队列，不再接受新任务
	close(s.taskQueue)
	
	// 等待所有工作线程完成当前任务并退出
	s.wg.Wait()
	
	// 更新调度器状态
	s.running = false
	log.Printf("Scheduler %s stopped", s.config.Name)
	return nil
}

// IsRunning 检查调度器是否正在运行
// 线程安全地获取调度器的运行状态
// 返回:
//   bool: true表示调度器正在运行，false表示调度器已停止
func (s *StandardScheduler) IsRunning() bool {
	// 使用读锁，允许并发读取状态
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetTaskHistory 获取指定任务的执行历史记录
// 从存储中加载任务的历史执行结果，按时间倒序排列
// 参数:
//   taskID: 任务ID
//   limit: 返回记录的最大数量，0表示不限制
// 返回:
//   []*TaskResult: 执行结果列表，按时间倒序排列
//   error: 获取失败时返回错误信息
func (s *StandardScheduler) GetTaskHistory(taskID string, limit int) ([]*TaskResult, error) {
	return s.storage.LoadTaskResults(taskID, limit)
}

// GetRunningTasks 获取正在运行的任务
func (s *StandardScheduler) GetRunningTasks() ([]*TaskInfo, error) {
	tasks, err := s.ListTasks()
	if err != nil {
		return nil, err
	}
	
	var runningTasks []*TaskInfo
	for _, task := range tasks {
		if task.Status == TaskStatusRunning {
			runningTasks = append(runningTasks, task)
		}
	}
	
	return runningTasks, nil
}

// scheduleTask 调度任务（内部方法，调用时需要持有锁）
func (s *StandardScheduler) scheduleTask(config *TaskConfig) {
	// 停止现有的定时器
	if timer, exists := s.timers[config.ID]; exists {
		timer.Stop()
	}
	
	// 计算下次执行时间
	nextTime := s.calculateNextRunTime(config)
	if nextTime.IsZero() {
		log.Printf("Task %s: no next run time calculated", config.ID)
		return
	}
	
	// 更新任务信息中的下次执行时间
	taskInfo, _ := s.storage.LoadTaskInfo(config.ID)
	if taskInfo != nil {
		taskInfo.NextRunTime = &nextTime
		s.storage.SaveTaskInfo(taskInfo)
	}
	
	// 创建定时器
	duration := time.Until(nextTime)
	timer := time.AfterFunc(duration, func() {
		s.executeTask(config.ID)
	})
	
	s.timers[config.ID] = timer
	log.Printf("Task %s scheduled to run at %s", config.ID, nextTime.Format(time.RFC3339))
}

// calculateNextRunTime 计算下次执行时间
func (s *StandardScheduler) calculateNextRunTime(config *TaskConfig) time.Time {
	now := time.Now()
	
	switch config.ScheduleType {
	case ScheduleTypeOnce:
		if config.StartTime != nil {
			return *config.StartTime
		}
		return now.Add(config.Delay)
		
	case ScheduleTypeDelay:
		return now.Add(config.Delay)
		
	case ScheduleTypeInterval:
		if config.StartTime != nil && config.StartTime.After(now) {
			return *config.StartTime
		}
		return now.Add(config.Interval)
		
	case ScheduleTypeCron:
		schedule, err := s.cronParser.Parse(config.CronExpr)
		if err != nil {
			log.Printf("Failed to parse cron expression %s: %v", config.CronExpr, err)
			return time.Time{}
		}
		return schedule.Next(now)
		
	default:
		return time.Time{}
	}
}

// executeTask 执行任务
func (s *StandardScheduler) executeTask(taskID string) {
	s.mu.RLock()
	executor, exists := s.executors[taskID]
	s.mu.RUnlock()
	
	if !exists {
		log.Printf("Task %s: executor not found", taskID)
		return
	}
	
	config, err := s.storage.LoadTaskConfig(taskID)
	if err != nil {
		log.Printf("Task %s: failed to load config: %v", taskID, err)
		return
	}
	
	// 检查任务是否仍然应该执行
	if !config.Enabled || !config.ShouldStart() {
		return
	}
	
	// 提交任务到队列
	select {
	case s.taskQueue <- &taskJob{
		taskID:   taskID,
		params:   config.Params,
		executor: executor,
		config:   config,
	}:
		// 任务已提交
	default:
		log.Printf("Task %s: queue is full, skipping execution", taskID)
	}
	
	// 为重复任务重新调度
	if config.ScheduleType == ScheduleTypeInterval || config.ScheduleType == ScheduleTypeCron {
		s.mu.Lock()
		s.scheduleTask(config)
		s.mu.Unlock()
	}
}

// worker 工作线程
func (s *StandardScheduler) worker() {
	defer s.wg.Done()
	
	for job := range s.taskQueue {
		s.runTask(job)
	}
}

// runTask 运行任务
func (s *StandardScheduler) runTask(job *taskJob) {
	// 创建任务结果
	result := NewTaskResult(job.taskID)
	
	// 更新任务状态为运行中
	taskInfo, _ := s.storage.LoadTaskInfo(job.taskID)
	if taskInfo != nil {
		taskInfo.Status = TaskStatusRunning
		taskInfo.LastRunTime = &result.StartTime
		taskInfo.UpdatedAt = time.Now()
		s.storage.SaveTaskInfo(taskInfo)
	}
	
	// 设置超时
	timeout := job.config.Timeout
	if timeout <= 0 {
		timeout = s.config.DefaultTimeout
	}
	
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()
	
	// 执行任务
	err := s.executeWithRetry(ctx, job, result)
	
	// 更新任务状态和结果
	if taskInfo != nil {
		taskInfo.RunCount++
		if err != nil {
			taskInfo.Status = TaskStatusFailed
			taskInfo.FailureCount++
			result.Fail(err)
		} else {
			taskInfo.Status = TaskStatusCompleted
			result.Complete()
		}
		taskInfo.LastResult = result
		taskInfo.UpdatedAt = time.Now()
		s.storage.SaveTaskInfo(taskInfo)
	}
	
	// 保存执行结果
	s.storage.SaveTaskResult(result)
	
	if err != nil {
		log.Printf("Task %s failed: %v", job.taskID, err)
	} else {
		log.Printf("Task %s completed successfully in %v", job.taskID, result.Duration)
	}
}

// executeWithRetry 带重试的执行任务
func (s *StandardScheduler) executeWithRetry(ctx context.Context, job *taskJob, result *TaskResult) error {
	maxRetries := job.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = s.config.DefaultRetries
	}
	
	var lastErr error
	
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			result.RetryCount = i
			// 重试间隔
			retryInterval := job.config.RetryInterval
			if retryInterval <= 0 {
				retryInterval = time.Second * time.Duration(i) // 递增间隔
			}
			
			select {
			case <-time.After(retryInterval):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		
		// 执行任务
		err := job.executor.Execute(ctx, job.params)
		if err == nil {
			return nil // 成功
		}
		
		lastErr = err
		log.Printf("Task %s attempt %d failed: %v", job.taskID, i+1, err)
	}
	
	return fmt.Errorf("task failed after %d attempts: %w", maxRetries+1, lastErr)
} 