# 注册中心模块 API 使用指南

## 概述

注册中心模块已经按照 `register.go` 和 `gateway.go` 的设计模式进行了重构，提供了公共的初始化方法和前台界面调用接口，无需独立的 main 函数。

## 核心API

### 1. 基础服务器管理

#### 启动注册中心服务器

```go
import (
    "context"
    "github.com/gateway/internal/registry"
    "github.com/gateway/pkg/database"
)

// 方法1: 使用配置文件启动
func startWithConfigFile() {
    db, _ := database.NewDatabase("mysql", "root:password@tcp(localhost:3306)/registry")
    defer db.Close()
    
    ctx := context.Background()
    server, err := registry.StartRegistryServer(ctx, db, "configs/registry.yaml")
    if err != nil {
        // 处理错误
        return
    }
    
    // 服务器已启动，可以使用
    defer registry.StopRegistryServer(server)
}

// 方法2: 使用自定义配置启动
func startWithCustomConfig() {
    db, _ := database.NewDatabase("mysql", "root:password@tcp(localhost:3306)/registry")
    defer db.Close()
    
    cfg := config.DefaultConfig()
    cfg.Server.HTTP.Port = 8081
    
    ctx := context.Background()
    server, err := registry.StartRegistryServerWithConfig(ctx, db, cfg)
    if err != nil {
        // 处理错误
        return
    }
    
    defer registry.StopRegistryServer(server)
}
```

#### 服务器状态管理

```go
// 获取服务器状态
status, err := registry.GetRegistryServerStatus(server)
if err != nil {
    // 处理错误
}

// 检查服务器是否运行
if server.IsRunning() {
    // 服务器正在运行
}

// 获取健康状态
health := server.GetHealthStatus()

// 获取统计信息
stats := server.GetStats()
```

### 2. 前台界面集成

#### 使用 RegistryManager

```go
import "github.com/gateway/internal/registry/frontend"

// 创建管理器
manager := frontend.NewRegistryManager(db)
defer manager.Close()

// 启动服务器
ctx := context.Background()
err := manager.StartServer(ctx, "server1", "configs/registry.yaml")
if err != nil {
    // 处理错误
}

// 注册服务
serviceReq := &frontend.ServiceInstanceRequest{
    ServiceName: "user-service",
    GroupName:   "business",
    HostAddress: "192.168.1.100",
    PortNumber:  8080,
    WeightValue: 100,
    Metadata: map[string]string{
        "version": "1.0.0",
        "protocol": "http",
    },
    Tags: []string{"production", "stable"},
}

err = manager.RegisterService("server1", serviceReq)
if err != nil {
    // 处理错误
}

// 发现服务
instances, err := manager.DiscoverServices("server1", "user-service", "business", nil)
if err != nil {
    // 处理错误
}

// 发送心跳
err = manager.SendHeartbeat("server1", instanceId)

// 停止服务器
err = manager.StopServer("server1")
```

### 3. 客户端API

#### 使用 Client

```go
import "github.com/gateway/internal/registry/client"

// 创建客户端
client := client.NewClient("http://localhost:8080", "default", "system")

// 注册服务实例
instance := &core.ServiceInstance{
    ServiceName:   "user-service",
    GroupName:     "business",
    HostAddress:   "192.168.1.100",
    PortNumber:    8080,
    InstanceStatus: core.InstanceStatusUp,
    HealthStatus:  core.HealthStatusHealthy,
    WeightValue:   100,
}

ctx := context.Background()
err := client.RegisterInstance(ctx, instance)

// 发现服务实例
instances, err := client.DiscoverInstances(ctx, "user-service", "business", nil)

// 发送心跳
err = client.Heartbeat(ctx, instanceId)

// 获取统计信息
stats, err := client.GetStats(ctx)

// 健康检查
health, err := client.HealthCheck(ctx)
```

## 前台界面调用示例

### Web界面集成

```go
// 在Web控制器中使用
type RegistryController struct {
    manager *frontend.RegistryManager
}

func NewRegistryController(db database.Database) *RegistryController {
    return &RegistryController{
        manager: frontend.NewRegistryManager(db),
    }
}

// 启动服务器接口
func (rc *RegistryController) StartServer(w http.ResponseWriter, r *http.Request) {
    serverId := r.FormValue("serverId")
    configFile := r.FormValue("configFile")
    
    ctx := r.Context()
    err := rc.manager.StartServer(ctx, serverId, configFile)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "message": "服务器启动成功",
    })
}

// 获取服务器状态接口
func (rc *RegistryController) GetServerStatus(w http.ResponseWriter, r *http.Request) {
    serverId := r.FormValue("serverId")
    
    status, err := rc.manager.GetServerStatus(serverId)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(status)
}

// 注册服务接口
func (rc *RegistryController) RegisterService(w http.ResponseWriter, r *http.Request) {
    var req frontend.ServiceInstanceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    serverId := r.FormValue("serverId")
    err := rc.manager.RegisterService(serverId, &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "message": "服务注册成功",
    })
}
```

### 批量操作

```go
// 批量启动多个服务器
func startMultipleServers(manager *frontend.RegistryManager) {
    servers := []struct {
        id   string
        port int
    }{
        {"server1", 8080},
        {"server2", 8081},
        {"server3", 8082},
    }
    
    ctx := context.Background()
    for _, s := range servers {
        cfg := config.DefaultConfig()
        cfg.Server.HTTP.Port = s.port
        
        err := manager.StartServerWithConfig(ctx, s.id, cfg)
        if err != nil {
            logger.Error("启动服务器失败", "serverId", s.id, "error", err)
        } else {
            logger.Info("服务器启动成功", "serverId", s.id, "port", s.port)
        }
    }
}

// 批量注册服务
func batchRegisterServices(manager *frontend.RegistryManager) {
    services := []*frontend.ServiceInstanceRequest{
        {
            ServiceName: "user-service",
            GroupName:   "business",
            HostAddress: "192.168.1.100",
            PortNumber:  8080,
            WeightValue: 100,
        },
        {
            ServiceName: "order-service",
            GroupName:   "business",
            HostAddress: "192.168.1.101",
            PortNumber:  8080,
            WeightValue: 100,
        },
    }
    
    errors := manager.BatchRegisterServices("server1", services)
    for i, err := range errors {
        if err != nil {
            logger.Error("注册服务失败", "service", services[i].ServiceName, "error", err)
        }
    }
}
```

## 配置管理

### 动态配置

```go
// 重载配置
err := server.Reload("new-config.yaml")
if err != nil {
    // 处理错误
}

// 获取当前配置
cfg := server.GetConfig()
```

### 环境变量配置

```go
// 可以通过环境变量覆盖配置
os.Setenv("REGISTRY_HTTP_PORT", "8081")
os.Setenv("REGISTRY_DB_DSN", "root:password@tcp(localhost:3306)/registry")
```

## 监控和健康检查

### 健康检查

```go
// 检查单个服务器健康状态
health, err := manager.CheckServerHealth("server1")

// 检查所有服务器健康状态
allHealth := manager.CheckAllServersHealth()
```

### 获取统计信息

```go
// 获取服务统计
stats, err := manager.GetServiceStats("server1")

// 获取所有服务器信息
serverInfos, err := manager.GetAllServerInfo()
```

## 错误处理

```go
// 统一错误处理
func handleRegistryError(err error) {
    switch {
    case errors.Is(err, core.ErrInstanceNotFound):
        // 处理实例不存在错误
    case errors.Is(err, core.ErrServiceNotFound):
        // 处理服务不存在错误
    case errors.Is(err, core.ErrRegistryNotRunning):
        // 处理注册中心未运行错误
    default:
        // 处理其他错误
    }
}
```

## 最佳实践

### 1. 资源管理

```go
// 始终确保资源正确释放
defer func() {
    if err := registry.StopRegistryServer(server); err != nil {
        logger.Error("停止服务器失败", "error", err)
    }
}()

// 或使用管理器的Close方法
defer manager.Close()
```

### 2. 错误处理

```go
// 使用上下文控制超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

server, err := registry.StartRegistryServer(ctx, db, configFile)
if err != nil {
    return fmt.Errorf("启动注册中心失败: %w", err)
}
```

### 3. 并发安全

```go
// RegistryManager 是并发安全的
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        
        serverId := fmt.Sprintf("server%d", id)
        err := manager.StartServer(ctx, serverId, "config.yaml")
        if err != nil {
            logger.Error("启动服务器失败", "serverId", serverId, "error", err)
        }
    }(i)
}
wg.Wait()
```

## 总结

注册中心模块现在提供了完整的模块化API，支持：

1. **服务器管理**: 启动、停止、状态查询
2. **服务注册发现**: 注册、注销、发现、心跳
3. **前台界面集成**: 统一的管理器接口
4. **客户端API**: 完整的HTTP客户端
5. **批量操作**: 支持批量服务管理
6. **健康检查**: 服务器和服务健康状态监控
7. **配置管理**: 动态配置重载
8. **错误处理**: 完善的错误类型和处理机制

所有API都是线程安全的，支持并发调用，适合在前台界面和后台服务中使用。
