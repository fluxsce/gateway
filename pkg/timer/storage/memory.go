package storage

import (
	"errors"
	"sync"
	
	"gohub/pkg/timer"
)

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	mu      sync.RWMutex
	configs map[string]*timer.TaskConfig // 任务配置
	infos   map[string]*timer.TaskInfo   // 任务信息
	results map[string][]*timer.TaskResult // 任务结果历史
	maxResults int // 每个任务保存的最大结果数
}

// NewMemoryStorage 创建内存存储实例
// 提供基于内存的任务数据存储，适用于开发、测试和小规模部署
// 注意：重启后数据会丢失，生产环境建议使用持久化存储
// 返回:
//   *MemoryStorage: 初始化的内存存储实例，默认保存100条执行结果
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		configs:    make(map[string]*timer.TaskConfig),    // 任务配置映射
		infos:      make(map[string]*timer.TaskInfo),      // 任务信息映射
		results:    make(map[string][]*timer.TaskResult),  // 任务结果历史映射
		maxResults: 100,                                   // 默认保存100条结果
	}
}

// SetMaxResults 设置每个任务保存的最大结果数
// 用于控制内存使用，避免执行结果历史记录过多导致内存溢出
// 参数:
//   max: 每个任务保存的最大结果数，建议设置为100-1000之间
func (s *MemoryStorage) SetMaxResults(max int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxResults = max
}

// SaveTaskConfig 保存任务配置到内存存储
// 使用深拷贝确保数据独立性，防止外部修改影响存储的数据
// 参数:
//   config: 要保存的任务配置对象
// 返回:
//   error: 保存失败时返回错误信息
func (s *MemoryStorage) SaveTaskConfig(config *timer.TaskConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}
	
	// 深拷贝配置，确保数据独立性
	configCopy, err := timer.CopyTaskConfig(config)
	if err != nil {
		return err
	}
	
	// 线程安全地保存配置
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs[config.ID] = configCopy
	return nil
}

// LoadTaskConfig 从内存存储加载任务配置
// 返回配置的深拷贝，确保调用者修改不会影响存储的原始数据
// 参数:
//   taskID: 任务ID
// 返回:
//   *timer.TaskConfig: 任务配置对象的深拷贝
//   error: 任务不存在或拷贝失败时返回错误
func (s *MemoryStorage) LoadTaskConfig(taskID string) (*timer.TaskConfig, error) {
	// 使用读锁提高并发性能
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	config, exists := s.configs[taskID]
	if !exists {
		return nil, errors.New("task config not found")
	}
	
	// 返回深拷贝，保护原始数据
	return timer.CopyTaskConfig(config)
}

// DeleteTaskConfig 删除任务配置
func (s *MemoryStorage) DeleteTaskConfig(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.configs, taskID)
	delete(s.infos, taskID)
	delete(s.results, taskID)
	return nil
}

// ListTaskConfigs 列出所有任务配置
func (s *MemoryStorage) ListTaskConfigs() ([]*timer.TaskConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	configs := make([]*timer.TaskConfig, 0, len(s.configs))
	for _, config := range s.configs {
		configCopy, err := timer.CopyTaskConfig(config)
		if err != nil {
			return nil, err
		}
		configs = append(configs, configCopy)
	}
	
	return configs, nil
}

// SaveTaskInfo 保存任务信息
func (s *MemoryStorage) SaveTaskInfo(info *timer.TaskInfo) error {
	if info == nil {
		return errors.New("info cannot be nil")
	}
	
	// 深拷贝信息
	infoCopy, err := timer.CopyTaskInfo(info)
	if err != nil {
		return err
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	s.infos[info.Config.ID] = infoCopy
	return nil
}

// LoadTaskInfo 加载任务信息
func (s *MemoryStorage) LoadTaskInfo(taskID string) (*timer.TaskInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	info, exists := s.infos[taskID]
	if !exists {
		return nil, errors.New("task info not found")
	}
	
	// 返回深拷贝
	return timer.CopyTaskInfo(info)
}

// SaveTaskResult 保存任务结果
func (s *MemoryStorage) SaveTaskResult(result *timer.TaskResult) error {
	if result == nil {
		return errors.New("result cannot be nil")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 添加结果到历史记录
	results := s.results[result.TaskID]
	results = append(results, result)
	
	// 限制历史记录数量
	if len(results) > s.maxResults {
		results = results[len(results)-s.maxResults:]
	}
	
	s.results[result.TaskID] = results
	return nil
}

// LoadTaskResults 加载任务结果
func (s *MemoryStorage) LoadTaskResults(taskID string, limit int) ([]*timer.TaskResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	results, exists := s.results[taskID]
	if !exists {
		return []*timer.TaskResult{}, nil
	}
	
	// 返回最新的limit条记录
	if limit <= 0 || limit > len(results) {
		limit = len(results)
	}
	
	start := len(results) - limit
	return results[start:], nil
} 