package bootstrap

import (
	"sync"

	"gohub/pkg/logger"
	"gohub/pkg/utils/huberrors"
)

// 网关连接池设计说明：
// 1. gatewayPool 结构体为私有，外部无法直接创建实例
// 2. newGatewayPool() 函数为私有，外部无法调用
// 3. GatewayPool 接口为公开，定义了对外的API
// 4. GetGlobalPool() 是获取连接池的唯一入口，返回接口类型
// 5. 通过 sync.Once 确保全局只创建一个连接池实例
// 这样设计确保了严格的单例模式，外部无法创建多个连接池实例

// gatewayPool 网关实例连接池（私有结构体，强制单例）
// 使用map缓存网关实例，key是InstanceID，value是Gateway
type gatewayPool struct {
	// 网关实例映射 - key: InstanceID, value: Gateway
	gateways map[string]*Gateway
	
	// 读写锁保护并发访问
	mu sync.RWMutex
}

// newGatewayPool 创建新的网关连接池（私有函数，只能内部调用）
func newGatewayPool() *gatewayPool {
	return &gatewayPool{
		gateways: make(map[string]*Gateway),
	}
}

// Add 添加网关实例到连接池
func (p *gatewayPool) Add(instanceID string, gateway *Gateway) error {
	if instanceID == "" {
		return huberrors.NewError("实例ID不能为空")
	}
	
	if gateway == nil {
		return huberrors.NewError("网关实例不能为nil")
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// 检查实例ID是否已存在
	if _, exists := p.gateways[instanceID]; exists {
		return huberrors.NewError("网关实例ID '%s' 已存在", instanceID)
	}
	
	// 添加到连接池
	p.gateways[instanceID] = gateway
	
	logger.Info("网关实例已添加到连接池", "instanceId", instanceID)
	return nil
}

// Remove 从连接池中移除网关实例
func (p *gatewayPool) Remove(instanceID string) error {
	if instanceID == "" {
		return huberrors.NewError("实例ID不能为空")
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// 检查实例是否存在
	gateway, exists := p.gateways[instanceID]
	if !exists {
		return huberrors.NewError("网关实例ID '%s' 不存在", instanceID)
	}
	
	// 如果网关正在运行，先停止它
	if gateway.IsRunning() {
		if err := gateway.Stop(); err != nil {
			logger.Error("停止网关实例失败", err, "instanceId", instanceID)
			return huberrors.WrapError(err, "停止网关实例失败")
		}
	}
	
	// 从连接池中删除
	delete(p.gateways, instanceID)
	
	logger.Info("网关实例已从连接池中移除", "instanceId", instanceID)
	return nil
}

// Get 获取指定的网关实例
func (p *gatewayPool) Get(instanceID string) (*Gateway, error) {
	if instanceID == "" {
		return nil, huberrors.NewError("实例ID不能为空")
	}
	
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	gateway, exists := p.gateways[instanceID]
	if !exists {
		return nil, huberrors.NewError("网关实例ID '%s' 不存在", instanceID)
	}
	
	return gateway, nil
}

// GetAll 获取所有网关实例
func (p *gatewayPool) GetAll() map[string]*Gateway {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	// 创建副本以避免并发修改
	result := make(map[string]*Gateway)
	for id, gateway := range p.gateways {
		result[id] = gateway
	}
	
	return result
}

// Exists 检查指定实例ID是否存在
func (p *gatewayPool) Exists(instanceID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	_, exists := p.gateways[instanceID]
	return exists
}

// Count 获取连接池中的实例总数
func (p *gatewayPool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return len(p.gateways)
}

// GetRunningGateways 获取所有正在运行的网关实例
func (p *gatewayPool) GetRunningGateways() map[string]*Gateway {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	result := make(map[string]*Gateway)
	for id, gateway := range p.gateways {
		if gateway.IsRunning() {
			result[id] = gateway
		}
	}
	
	return result
}

// StartAll 启动所有网关实例
func (p *gatewayPool) StartAll() error {
	p.mu.RLock()
	gateways := make(map[string]*Gateway)
	for id, gateway := range p.gateways {
		gateways[id] = gateway
	}
	p.mu.RUnlock()
	
	var errors []string
	for id, gateway := range gateways {
		if !gateway.IsRunning() {
			if err := gateway.Start(); err != nil {
				errors = append(errors, id+": "+err.Error())
				logger.Error("启动网关实例失败", err, "instanceId", id)
			} else {
				logger.Info("网关实例启动成功", "instanceId", id)
			}
		}
	}
	
	if len(errors) > 0 {
		return huberrors.NewError("部分网关实例启动失败: %v", errors)
	}
	
	logger.Info("所有网关实例启动成功", "count", len(gateways))
	return nil
}

// StopAll 停止所有网关实例
func (p *gatewayPool) StopAll() error {
	p.mu.RLock()
	gateways := make(map[string]*Gateway)
	for id, gateway := range p.gateways {
		gateways[id] = gateway
	}
	p.mu.RUnlock()
	
	var errors []string
	for id, gateway := range gateways {
		if gateway.IsRunning() {
			if err := gateway.Stop(); err != nil {
				errors = append(errors, id+": "+err.Error())
				logger.Error("停止网关实例失败", err, "instanceId", id)
			} else {
				logger.Info("网关实例停止成功", "instanceId", id)
			}
		}
	}
	
	if len(errors) > 0 {
		return huberrors.NewError("部分网关实例停止失败: %v", errors)
	}
	
	logger.Info("所有网关实例停止成功", "count", len(gateways))
	return nil
}

// Clear 清空连接池（会先停止所有实例）
func (p *gatewayPool) Clear() error {
	// 先停止所有实例
	if err := p.StopAll(); err != nil {
		return err
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// 清空map
	p.gateways = make(map[string]*Gateway)
	
	logger.Info("网关连接池已清空")
	return nil
}

// GetInstanceIDs 获取所有实例ID列表
func (p *gatewayPool) GetInstanceIDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	ids := make([]string, 0, len(p.gateways))
	for id := range p.gateways {
		ids = append(ids, id)
	}
	
	return ids
}

// GatewayPool 网关连接池接口（公开接口）
type GatewayPool interface {
	Add(instanceID string, gateway *Gateway) error
	Remove(instanceID string) error
	Get(instanceID string) (*Gateway, error)
	GetAll() map[string]*Gateway
	Exists(instanceID string) bool
	Count() int
	GetRunningGateways() map[string]*Gateway
	StartAll() error
	StopAll() error
	Clear() error
	GetInstanceIDs() []string
}

// 全局网关连接池实例
var (
	globalPool *gatewayPool
	poolOnce   sync.Once
)

// GetGlobalPool 获取全局网关连接池实例（单例模式）
// 这是获取连接池的唯一方式，确保全局只有一个连接池实例
func GetGlobalPool() GatewayPool {
	poolOnce.Do(func() {
		globalPool = newGatewayPool()
		logger.Info("全局网关连接池初始化完成")
	})
	return globalPool
} 