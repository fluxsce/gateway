package frontend

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/registry/client"
	"gateway/internal/registry/config"
	"gateway/internal/registry/core"
	"gateway/internal/registry/manager"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// RegistryManager 注册中心管理器 - 用于前台界面集成
type RegistryManager struct {
	servers map[string]*manager.Manager
	clients map[string]*client.Client
	db      database.Database
	mutex   sync.RWMutex
}

// NewRegistryManager 创建注册中心管理器
func NewRegistryManager(db database.Database) *RegistryManager {
	return &RegistryManager{
		servers: make(map[string]*manager.Manager),
		clients: make(map[string]*client.Client),
		db:      db,
	}
}

// ================== 服务器管理 ==================

// StartServer 启动注册中心服务器
func (rm *RegistryManager) StartServer(ctx context.Context, serverId, configFile string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 检查服务器是否已存在
	if _, exists := rm.servers[serverId]; exists {
		return fmt.Errorf("服务器 %s 已经在运行", serverId)
	}

	// 使用默认配置（实际项目中应该从文件加载）
	cfg := config.DefaultConfig()
	// TODO: 实现从 configFile 加载配置的逻辑

	// 创建管理器
	server := manager.NewManager(cfg, rm.db)
	if err := server.Initialize(); err != nil {
		return fmt.Errorf("初始化管理器失败: %w", err)
	}

	// 启动服务器
	if err := server.Start(); err != nil {
		return fmt.Errorf("启动服务器失败: %w", err)
	}

	// 保存服务器实例
	rm.servers[serverId] = server

	// 创建对应的客户端
	baseURL := fmt.Sprintf("http://%s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)
	rm.clients[serverId] = client.NewClient(baseURL, "default", "system")

	logger.Info("注册中心服务器启动成功", "serverId", serverId)
	return nil
}

// StartServerWithConfig 使用配置启动服务器
func (rm *RegistryManager) StartServerWithConfig(ctx context.Context, serverId string, cfg *config.Config) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 检查服务器是否已存在
	if _, exists := rm.servers[serverId]; exists {
		return fmt.Errorf("服务器 %s 已经在运行", serverId)
	}

	// 创建管理器
	server := manager.NewManager(cfg, rm.db)
	if err := server.Initialize(); err != nil {
		return fmt.Errorf("初始化管理器失败: %w", err)
	}

	// 启动服务器
	if err := server.Start(); err != nil {
		return fmt.Errorf("启动服务器失败: %w", err)
	}

	// 保存服务器实例
	rm.servers[serverId] = server

	// 创建对应的客户端
	baseURL := fmt.Sprintf("http://%s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)
	rm.clients[serverId] = client.NewClient(baseURL, "default", "system")

	logger.Info("注册中心服务器启动成功", "serverId", serverId)
	return nil
}

// StopServer 停止注册中心服务器
func (rm *RegistryManager) StopServer(serverId string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	server, exists := rm.servers[serverId]
	if !exists {
		return fmt.Errorf("服务器 %s 不存在", serverId)
	}

	// 停止服务器
	if err := server.Stop(); err != nil {
		return fmt.Errorf("停止服务器失败: %w", err)
	}

	// 清理资源
	delete(rm.servers, serverId)
	delete(rm.clients, serverId)

	logger.Info("注册中心服务器停止成功", "serverId", serverId)
	return nil
}

// GetServerStatus 获取服务器状态
func (rm *RegistryManager) GetServerStatus(serverId string) (map[string]interface{}, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	server, exists := rm.servers[serverId]
	if !exists {
		return map[string]interface{}{
			"status":  "not_found",
			"running": false,
		}, nil
	}

	return map[string]interface{}{
		"status":       "running",
		"running":      server.IsRunning(),
		"healthStatus": server.GetHealthStatus(),
		"stats":        server.GetStats(),
	}, nil
}

// ListServers 列出所有服务器
func (rm *RegistryManager) ListServers() []string {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var serverIds []string
	for serverId := range rm.servers {
		serverIds = append(serverIds, serverId)
	}
	return serverIds
}

// ================== 客户端操作 ==================

// GetClient 获取客户端
func (rm *RegistryManager) GetClient(serverId string) (*client.Client, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	client, exists := rm.clients[serverId]
	if !exists {
		return nil, fmt.Errorf("服务器 %s 的客户端不存在", serverId)
	}

	return client, nil
}

// RegisterService 注册服务
func (rm *RegistryManager) RegisterService(serverId string, instance *ServiceInstanceRequest) error {
	client, err := rm.GetClient(serverId)
	if err != nil {
		return err
	}

	// 转换为核心类型
	coreInstance := &core.ServiceInstance{
		ServiceName:    instance.ServiceName,
		GroupName:      instance.GroupName,
		HostAddress:    instance.HostAddress,
		PortNumber:     instance.PortNumber,
		ContextPath:    instance.ContextPath,
		InstanceStatus: core.InstanceStatusUp,
		HealthStatus:   core.HealthStatusHealthy,
		WeightValue:    instance.WeightValue,
		ClientType:     core.ClientTypeService,
		RegisterTime:   time.Now(),
		AddTime:        time.Now(),
		EditTime:       time.Now(),
		CurrentVersion: 1,
		ActiveFlag:     core.FlagYes,
	}

	// 设置元数据
	if len(instance.Metadata) > 0 {
		if err := coreInstance.SetMetadata(instance.Metadata); err != nil {
			return fmt.Errorf("设置元数据失败: %w", err)
		}
	}

	// 设置标签
	if len(instance.Tags) > 0 {
		if err := coreInstance.SetTags(instance.Tags); err != nil {
			return fmt.Errorf("设置标签失败: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.RegisterInstance(ctx, coreInstance)
}

// DeregisterService 注销服务
func (rm *RegistryManager) DeregisterService(serverId, instanceId string) error {
	client, err := rm.GetClient(serverId)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.DeregisterInstance(ctx, instanceId)
}

// DiscoverServices 发现服务
func (rm *RegistryManager) DiscoverServices(serverId, serviceName, groupName string, filters map[string]string) ([]*ServiceInstanceResponse, error) {
	client, err := rm.GetClient(serverId)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	instances, err := client.DiscoverInstances(ctx, serviceName, groupName, filters)
	if err != nil {
		return nil, err
	}

	// 转换为响应类型
	var responses []*ServiceInstanceResponse
	for _, instance := range instances {
		response := &ServiceInstanceResponse{
			InstanceId:     instance.ServiceInstanceId,
			ServiceName:    instance.ServiceName,
			GroupName:      instance.GroupName,
			HostAddress:    instance.HostAddress,
			PortNumber:     instance.PortNumber,
			ContextPath:    instance.ContextPath,
			InstanceStatus: instance.InstanceStatus,
			HealthStatus:   instance.HealthStatus,
			WeightValue:    instance.WeightValue,
			RegisterTime:   instance.RegisterTime,
			Metadata:       instance.GetMetadata(),
			Tags:           instance.GetTags(),
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// SendHeartbeat 发送心跳
func (rm *RegistryManager) SendHeartbeat(serverId, instanceId string) error {
	client, err := rm.GetClient(serverId)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.Heartbeat(ctx, instanceId)
}

// GetServiceStats 获取服务统计
func (rm *RegistryManager) GetServiceStats(serverId string) (map[string]interface{}, error) {
	client, err := rm.GetClient(serverId)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.GetStats(ctx)
}

// ================== 数据类型 ==================

// ServiceInstanceRequest 服务实例注册请求
type ServiceInstanceRequest struct {
	ServiceName string            `json:"serviceName"`
	GroupName   string            `json:"groupName"`
	HostAddress string            `json:"hostAddress"`
	PortNumber  int               `json:"portNumber"`
	ContextPath string            `json:"contextPath,omitempty"`
	WeightValue int               `json:"weightValue"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// ServiceInstanceResponse 服务实例响应
type ServiceInstanceResponse struct {
	InstanceId     string            `json:"instanceId"`
	ServiceName    string            `json:"serviceName"`
	GroupName      string            `json:"groupName"`
	HostAddress    string            `json:"hostAddress"`
	PortNumber     int               `json:"portNumber"`
	ContextPath    string            `json:"contextPath,omitempty"`
	InstanceStatus string            `json:"instanceStatus"`
	HealthStatus   string            `json:"healthStatus"`
	WeightValue    int               `json:"weightValue"`
	RegisterTime   time.Time         `json:"registerTime"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	ServerId string                 `json:"serverId"`
	Status   string                 `json:"status"`
	Config   map[string]interface{} `json:"config"`
	Stats    map[string]interface{} `json:"stats"`
}

// ================== 批量操作 ==================

// BatchRegisterServices 批量注册服务
func (rm *RegistryManager) BatchRegisterServices(serverId string, instances []*ServiceInstanceRequest) []error {
	var errors []error

	for _, instance := range instances {
		if err := rm.RegisterService(serverId, instance); err != nil {
			errors = append(errors, fmt.Errorf("注册服务 %s 失败: %w", instance.ServiceName, err))
		} else {
			errors = append(errors, nil)
		}
	}

	return errors
}

// BatchDeregisterServices 批量注销服务
func (rm *RegistryManager) BatchDeregisterServices(serverId string, instanceIds []string) []error {
	var errors []error

	for _, instanceId := range instanceIds {
		if err := rm.DeregisterService(serverId, instanceId); err != nil {
			errors = append(errors, fmt.Errorf("注销服务实例 %s 失败: %w", instanceId, err))
		} else {
			errors = append(errors, nil)
		}
	}

	return errors
}

// ================== 健康检查 ==================

// CheckServerHealth 检查服务器健康状态
func (rm *RegistryManager) CheckServerHealth(serverId string) (map[string]interface{}, error) {
	client, err := rm.GetClient(serverId)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.HealthCheck(ctx)
}

// CheckAllServersHealth 检查所有服务器健康状态
func (rm *RegistryManager) CheckAllServersHealth() map[string]interface{} {
	rm.mutex.RLock()
	serverIds := make([]string, 0, len(rm.servers))
	for serverId := range rm.servers {
		serverIds = append(serverIds, serverId)
	}
	rm.mutex.RUnlock()

	results := make(map[string]interface{})
	for _, serverId := range serverIds {
		health, err := rm.CheckServerHealth(serverId)
		if err != nil {
			results[serverId] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
		} else {
			results[serverId] = health
		}
	}

	return results
}

// ================== 资源清理 ==================

// Close 关闭管理器
func (rm *RegistryManager) Close() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 停止所有服务器
	for serverId := range rm.servers {
		if err := rm.StopServer(serverId); err != nil {
			logger.Error("停止服务器失败", "serverId", serverId, "error", err)
		}
	}

	// 关闭数据库连接
	if rm.db != nil {
		return rm.db.Close()
	}

	return nil
}

// GetAllServerInfo 获取所有服务器信息
func (rm *RegistryManager) GetAllServerInfo() ([]*ServerInfo, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var serverInfos []*ServerInfo
	for serverId, server := range rm.servers {
		status := map[string]interface{}{
			"status":       "running",
			"running":      server.IsRunning(),
			"healthStatus": server.GetHealthStatus(),
			"stats":        server.GetStats(),
		}

		serverInfo := &ServerInfo{
			ServerId: serverId,
			Status:   "running",
			Config:   make(map[string]interface{}),
			Stats:    status,
		}

		// 获取配置信息
		cfg := server.GetConfig()
		if cfg != nil {
			serverInfo.Config = map[string]interface{}{
				"http_port": cfg.Server.HTTP.Port,
				"http_host": cfg.Server.HTTP.Host,
			}
		}

		serverInfos = append(serverInfos, serverInfo)
	}

	return serverInfos, nil
}
