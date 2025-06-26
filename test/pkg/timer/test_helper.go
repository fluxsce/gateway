package timer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gohub/pkg/timer"
)

// TestTaskExecutor 测试用的任务执行器
// 实现了timer.TaskExecutor接口，用于单元测试
type TestTaskExecutor struct {
	name          string                                                    // 执行器名称
	executeFunc   func(ctx context.Context, params interface{}) error      // 自定义执行函数
	executeCalls  []ExecuteCall                                             // 执行调用记录
	mu            sync.RWMutex                                              // 读写锁保护并发访问
}

// ExecuteCall 执行调用记录
// 记录每次Execute方法的调用参数和结果，用于测试验证
type ExecuteCall struct {
	Params    interface{} // 传入的参数
	Error     error       // 返回的错误
	Timestamp time.Time   // 调用时间戳
}

// NewTestTaskExecutor 创建测试任务执行器
// 参数:
//   name: 执行器名称，用于标识和日志
//   executeFunc: 自定义执行函数，如果为nil则默认成功执行
// 返回:
//   *TestTaskExecutor: 初始化完成的测试执行器
func NewTestTaskExecutor(name string, executeFunc func(ctx context.Context, params interface{}) error) *TestTaskExecutor {
	return &TestTaskExecutor{
		name:        name,
		executeFunc: executeFunc,
		executeCalls: make([]ExecuteCall, 0),
	}
}

// Execute 实现TaskExecutor接口的Execute方法
// 执行任务逻辑，记录调用信息，支持自定义执行函数
func (e *TestTaskExecutor) Execute(ctx context.Context, params interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	var err error
	if e.executeFunc != nil {
		err = e.executeFunc(ctx, params)
	}
	
	// 记录执行调用
	e.executeCalls = append(e.executeCalls, ExecuteCall{
		Params:    params,
		Error:     err,
		Timestamp: time.Now(),
	})
	
	return err
}

// GetName 实现TaskExecutor接口的GetName方法
// 返回执行器的名称标识
func (e *TestTaskExecutor) GetName() string {
	return e.name
}

// GetExecuteCalls 获取所有执行调用记录
// 返回执行调用的副本，线程安全
func (e *TestTaskExecutor) GetExecuteCalls() []ExecuteCall {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	calls := make([]ExecuteCall, len(e.executeCalls))
	copy(calls, e.executeCalls)
	return calls
}

// GetExecuteCount 获取执行次数
// 返回Execute方法被调用的总次数
func (e *TestTaskExecutor) GetExecuteCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.executeCalls)
}

// Reset 重置执行记录
// 清空所有执行调用记录，用于测试用例之间的清理
func (e *TestTaskExecutor) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.executeCalls = e.executeCalls[:0]
}

// MemoryTaskStorage 内存存储实现
// 实现了timer.TaskStorage接口，用于测试环境的数据存储
type MemoryTaskStorage struct {
	configs map[string]*timer.TaskConfig // 任务配置存储
	infos   map[string]*timer.TaskInfo   // 任务信息存储
	results map[string][]*timer.TaskResult // 任务结果存储，按taskID分组
	mu      sync.RWMutex                 // 读写锁保护并发访问
}

// NewMemoryTaskStorage 创建内存存储实例
// 返回初始化完成的内存存储对象，用于测试环境
func NewMemoryTaskStorage() *MemoryTaskStorage {
	return &MemoryTaskStorage{
		configs: make(map[string]*timer.TaskConfig),
		infos:   make(map[string]*timer.TaskInfo),
		results: make(map[string][]*timer.TaskResult),
	}
}

// SaveTaskConfig 保存任务配置
// 实现TaskStorage接口，将任务配置保存到内存中
func (s *MemoryTaskStorage) SaveTaskConfig(config *timer.TaskConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.ID == "" {
		return fmt.Errorf("config ID cannot be empty")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 深拷贝配置以避免外部修改影响存储
	configCopy := *config
	s.configs[config.ID] = &configCopy
	return nil
}

// LoadTaskConfig 加载任务配置
// 实现TaskStorage接口，从内存中加载指定任务的配置
func (s *MemoryTaskStorage) LoadTaskConfig(taskID string) (*timer.TaskConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	config, exists := s.configs[taskID]
	if !exists {
		return nil, fmt.Errorf("task config not found: %s", taskID)
	}
	
	// 返回配置的深拷贝
	configCopy := *config
	return &configCopy, nil
}

// DeleteTaskConfig 删除任务配置
// 实现TaskStorage接口，从内存中删除指定任务的配置
func (s *MemoryTaskStorage) DeleteTaskConfig(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.configs, taskID)
	delete(s.infos, taskID)
	delete(s.results, taskID)
	return nil
}

// ListTaskConfigs 列出所有任务配置
// 实现TaskStorage接口，返回所有任务配置的副本
func (s *MemoryTaskStorage) ListTaskConfigs() ([]*timer.TaskConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	configs := make([]*timer.TaskConfig, 0, len(s.configs))
	for _, config := range s.configs {
		configCopy := *config
		configs = append(configs, &configCopy)
	}
	return configs, nil
}

// SaveTaskInfo 保存任务信息
// 实现TaskStorage接口，将任务运行时信息保存到内存中
func (s *MemoryTaskStorage) SaveTaskInfo(info *timer.TaskInfo) error {
	if info == nil {
		return fmt.Errorf("info cannot be nil")
	}
	if info.Config == nil || info.Config.ID == "" {
		return fmt.Errorf("info config ID cannot be empty")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 深拷贝信息以避免外部修改影响存储
	infoCopy := *info
	if info.Config != nil {
		configCopy := *info.Config
		infoCopy.Config = &configCopy
	}
	if info.LastResult != nil {
		resultCopy := *info.LastResult
		infoCopy.LastResult = &resultCopy
	}
	
	s.infos[info.Config.ID] = &infoCopy
	return nil
}

// LoadTaskInfo 加载任务信息
// 实现TaskStorage接口，从内存中加载指定任务的运行时信息
func (s *MemoryTaskStorage) LoadTaskInfo(taskID string) (*timer.TaskInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	info, exists := s.infos[taskID]
	if !exists {
		return nil, fmt.Errorf("task info not found: %s", taskID)
	}
	
	// 返回信息的深拷贝
	infoCopy := *info
	if info.Config != nil {
		configCopy := *info.Config
		infoCopy.Config = &configCopy
	}
	if info.LastResult != nil {
		resultCopy := *info.LastResult
		infoCopy.LastResult = &resultCopy
	}
	
	return &infoCopy, nil
}

// SaveTaskResult 保存任务结果
// 实现TaskStorage接口，将任务执行结果保存到内存中
func (s *MemoryTaskStorage) SaveTaskResult(result *timer.TaskResult) error {
	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}
	if result.TaskID == "" {
		return fmt.Errorf("result TaskID cannot be empty")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 深拷贝结果以避免外部修改影响存储
	resultCopy := *result
	
	// 将结果添加到对应任务的结果列表中
	s.results[result.TaskID] = append(s.results[result.TaskID], &resultCopy)
	return nil
}

// LoadTaskResults 加载任务结果
// 实现TaskStorage接口，从内存中加载指定任务的执行历史记录
func (s *MemoryTaskStorage) LoadTaskResults(taskID string, limit int) ([]*timer.TaskResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	results, exists := s.results[taskID]
	if !exists {
		return []*timer.TaskResult{}, nil
	}
	
	// 按时间倒序排列（最新的在前面）
	sortedResults := make([]*timer.TaskResult, len(results))
	for i, result := range results {
		resultCopy := *result
		sortedResults[len(results)-1-i] = &resultCopy
	}
	
	// 应用限制
	if limit > 0 && len(sortedResults) > limit {
		sortedResults = sortedResults[:limit]
	}
	
	return sortedResults, nil
}

// Clear 清空所有存储数据
// 用于测试用例之间的数据清理
func (s *MemoryTaskStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.configs = make(map[string]*timer.TaskConfig)
	s.infos = make(map[string]*timer.TaskInfo)
	s.results = make(map[string][]*timer.TaskResult)
}

// GetTaskCount 获取任务数量
// 返回当前存储中的任务配置数量，用于测试验证
func (s *MemoryTaskStorage) GetTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.configs)
}

// GetResultCount 获取指定任务的结果数量
// 返回指定任务的执行结果数量，用于测试验证
func (s *MemoryTaskStorage) GetResultCount(taskID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.results[taskID])
}

// CreateTestTaskConfig 创建测试用的任务配置
// 提供便捷的方法创建用于测试的任务配置对象
func CreateTestTaskConfig(id, name string, scheduleType timer.ScheduleType) *timer.TaskConfig {
	config := &timer.TaskConfig{
		ID:           id,
		Name:         name,
		Description:  fmt.Sprintf("Test task: %s", name),
		Priority:     timer.TaskPriorityNormal,
		ScheduleType: scheduleType,
		MaxRetries:   3,
		Timeout:      time.Minute * 5,
		Enabled:      true,
	}
	
	// 根据调度类型设置相应的调度参数
	switch scheduleType {
	case timer.ScheduleTypeOnce:
		startTime := time.Now().Add(time.Second)
		config.StartTime = &startTime
	case timer.ScheduleTypeInterval:
		config.Interval = time.Second * 2
	case timer.ScheduleTypeCron:
		config.CronExpr = "*/5 * * * * *" // 每5秒执行一次
	case timer.ScheduleTypeDelay:
		config.Delay = time.Second
	}
	
	return config
}

// WaitForCondition 等待条件满足
// 在指定超时时间内等待条件函数返回true，用于异步测试
func WaitForCondition(condition func() bool, timeout time.Duration, checkInterval time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-timer.C:
			return false // 超时
		case <-ticker.C:
			if condition() {
				return true // 条件满足
			}
		}
	}
} 