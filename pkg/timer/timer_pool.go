package timer

import (
	"errors"
	"fmt"
	"sync"
)

// TimerPool 全局定时器池，用于管理多个调度器实例
type TimerPool struct {
	mu         sync.RWMutex
	schedulers map[string]TaskScheduler // 调度器ID -> 调度器实例映射
	configs    map[string]*SchedulerConfig // 调度器ID -> 配置映射
}

// globalTimerPool 全局定时器池实例
var globalTimerPool *TimerPool
var poolOnce sync.Once

// GetTimerPool 获取全局定时器池实例（单例模式）
// 返回:
//   *TimerPool: 全局定时器池实例
func GetTimerPool() *TimerPool {
	poolOnce.Do(func() {
		globalTimerPool = &TimerPool{
			schedulers: make(map[string]TaskScheduler),
			configs:    make(map[string]*SchedulerConfig),
		}
	})
	return globalTimerPool
}

// CreateScheduler 创建新的调度器实例并添加到池中
// 参数:
//   config: 调度器配置，包含ID、名称等信息
// 返回:
//   TaskScheduler: 创建的调度器实例
//   error: 创建失败时返回错误信息
func (p *TimerPool) CreateScheduler(config *SchedulerConfig) (TaskScheduler, error) {
	if config == nil {
		return nil, errors.New("scheduler config cannot be nil")
	}
	
	if config.ID == "" {
		return nil, errors.New("scheduler ID cannot be empty")
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// 检查调度器ID是否已存在
	if _, exists := p.schedulers[config.ID]; exists {
		return nil, fmt.Errorf("scheduler with ID %s already exists", config.ID)
	}
	
	// 创建新的调度器实例
	scheduler := NewStandardScheduler(config)
	
	// 添加到池中
	p.schedulers[config.ID] = scheduler
	p.configs[config.ID] = config
	
	return scheduler, nil
}

// GetScheduler 根据ID获取调度器实例
// 参数:
//   schedulerID: 调度器ID
// 返回:
//   TaskScheduler: 调度器实例
//   error: 获取失败时返回错误信息
func (p *TimerPool) GetScheduler(schedulerID string) (TaskScheduler, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	scheduler, exists := p.schedulers[schedulerID]
	if !exists {
		return nil, fmt.Errorf("scheduler with ID %s not found", schedulerID)
	}
	
	return scheduler, nil
}

// GetSchedulerConfig 根据ID获取调度器配置
// 参数:
//   schedulerID: 调度器ID
// 返回:
//   *SchedulerConfig: 调度器配置
//   error: 获取失败时返回错误信息
func (p *TimerPool) GetSchedulerConfig(schedulerID string) (*SchedulerConfig, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	config, exists := p.configs[schedulerID]
	if !exists {
		return nil, fmt.Errorf("scheduler config with ID %s not found", schedulerID)
	}
	
	return config, nil
}

// RemoveScheduler 从池中移除指定的调度器
// 参数:
//   schedulerID: 要移除的调度器ID
// 返回:
//   error: 移除失败时返回错误信息
func (p *TimerPool) RemoveScheduler(schedulerID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	scheduler, exists := p.schedulers[schedulerID]
	if !exists {
		return fmt.Errorf("scheduler with ID %s not found", schedulerID)
	}
	
	// 如果调度器正在运行，先停止它
	if scheduler.IsRunning() {
		if err := scheduler.Stop(); err != nil {
			return fmt.Errorf("failed to stop scheduler %s: %w", schedulerID, err)
		}
	}
	
	// 从池中删除
	delete(p.schedulers, schedulerID)
	delete(p.configs, schedulerID)
	
	return nil
}

// ListSchedulers 获取所有调度器的ID列表
// 返回:
//   []string: 所有调度器ID的切片
func (p *TimerPool) ListSchedulers() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	ids := make([]string, 0, len(p.schedulers))
	for id := range p.schedulers {
		ids = append(ids, id)
	}
	
	return ids
}

// GetSchedulerCount 获取池中调度器的数量
// 返回:
//   int: 调度器数量
func (p *TimerPool) GetSchedulerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return len(p.schedulers)
}

// StartAllSchedulers 启动池中所有的调度器
// 返回:
//   map[string]error: 调度器ID -> 启动错误的映射，成功启动的调度器不在映射中
func (p *TimerPool) StartAllSchedulers() map[string]error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	errors := make(map[string]error)
	
	for id, scheduler := range p.schedulers {
		if !scheduler.IsRunning() {
			if err := scheduler.Start(); err != nil {
				errors[id] = err
			}
		}
	}
	
	return errors
}

// StopAllSchedulers 停止池中所有的调度器
// 返回:
//   map[string]error: 调度器ID -> 停止错误的映射，成功停止的调度器不在映射中
func (p *TimerPool) StopAllSchedulers() map[string]error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	errors := make(map[string]error)
	
	for id, scheduler := range p.schedulers {
		if scheduler.IsRunning() {
			if err := scheduler.Stop(); err != nil {
				errors[id] = err
			}
		}
	}
	
	return errors
}

// GetRunningSchedulers 获取正在运行的调度器ID列表
// 返回:
//   []string: 正在运行的调度器ID切片
func (p *TimerPool) GetRunningSchedulers() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var runningIds []string
	for id, scheduler := range p.schedulers {
		if scheduler.IsRunning() {
			runningIds = append(runningIds, id)
		}
	}
	
	return runningIds
}

// GetStoppedSchedulers 获取已停止的调度器ID列表
// 返回:
//   []string: 已停止的调度器ID切片
func (p *TimerPool) GetStoppedSchedulers() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var stoppedIds []string
	for id, scheduler := range p.schedulers {
		if !scheduler.IsRunning() {
			stoppedIds = append(stoppedIds, id)
		}
	}
	
	return stoppedIds
}

// SchedulerInfo 调度器信息结构
type SchedulerInfo struct {
	ID        string           `json:"id"`        // 调度器ID
	Config    *SchedulerConfig `json:"config"`    // 调度器配置
	IsRunning bool             `json:"isRunning"` // 是否正在运行
	TaskCount int              `json:"taskCount"` // 任务数量
}

// GetSchedulerInfo 获取指定调度器的详细信息
// 参数:
//   schedulerID: 调度器ID
// 返回:
//   *SchedulerInfo: 调度器信息
//   error: 获取失败时返回错误信息
func (p *TimerPool) GetSchedulerInfo(schedulerID string) (*SchedulerInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	scheduler, exists := p.schedulers[schedulerID]
	if !exists {
		return nil, fmt.Errorf("scheduler with ID %s not found", schedulerID)
	}
	
	config, exists := p.configs[schedulerID]
	if !exists {
		return nil, fmt.Errorf("scheduler config with ID %s not found", schedulerID)
	}
	
	// 获取任务数量
	tasks, _ := scheduler.ListTasks()
	taskCount := len(tasks)
	
	return &SchedulerInfo{
		ID:        schedulerID,
		Config:    config,
		IsRunning: scheduler.IsRunning(),
		TaskCount: taskCount,
	}, nil
}

// GetAllSchedulerInfo 获取所有调度器的详细信息
// 返回:
//   []*SchedulerInfo: 所有调度器信息的切片
func (p *TimerPool) GetAllSchedulerInfo() []*SchedulerInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	infos := make([]*SchedulerInfo, 0, len(p.schedulers))
	
	for id, scheduler := range p.schedulers {
		config := p.configs[id]
		
		// 获取任务数量
		tasks, _ := scheduler.ListTasks()
		taskCount := len(tasks)
		
		infos = append(infos, &SchedulerInfo{
			ID:        id,
			Config:    config,
			IsRunning: scheduler.IsRunning(),
			TaskCount: taskCount,
		})
	}
	
	return infos
}
