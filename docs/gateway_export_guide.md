# 网关配置导出功能使用指南

网关提供了完整的配置导出功能，支持将当前运行的网关配置导出为YAML格式，便于备份、版本控制和配置迁移。

## 功能特性

### 1. 完整配置导出
- 导出所有配置字段（包括默认值）
- 包含完整的配置结构，便于版本回退
- 确保导出的配置可以完整还原网关状态
- 用于生产环境的配置备份和管理

### 2. 分模块导出
- 单独导出路由配置
- 单独导出服务配置
- 便于模块化管理和部分配置备份

## API 接口

### ExportToYAML() - 完整配置导出

```go
// 导出完整配置（包含所有字段和默认值）
yamlContent, err := gateway.ExportToYAML()
if err != nil {
    log.Printf("导出失败: %v", err)
    return
}
fmt.Println(yamlContent)
```

**特点:**
- 包含所有配置字段
- 显示默认值和空字段
- 用于完整备份和版本回退
- 确保配置的完整性和一致性

**示例输出:**
```yaml
gateway:
  base:
    listen: :9080
    enable_https: false
    cert_file: ""
    key_file: ""
    use_gin: true
    read_timeout: 10s
    write_timeout: 10s
    idle_timeout: 2m0s
  router:
    routes:
      - id: api-route
        name: API路由
        path: /api/v1
        match_type: ""
        methods: [GET, POST, PUT, DELETE]
        service_id: api-service
        enabled: true
        priority: 100
        strip_prefix: true
        rewrite_path: /v1
        metadata:
          description: API服务路由
        assertion_group:
          all_required: true
          assertions: []
  proxy:
    timeout: 30s
    follow_redirects: true
    keep_alive: true
    max_idle_conns: 100
    idle_conn_timeout: 1m30s
    copy_response_body: true
    services:
      - id: api-service
        name: API服务
        load_balance: round-robin
        endpoints:
          - http://localhost:8081
          - http://localhost:8082
  cors:
    enabled: true
    allow_origins: ["*"]
    allow_methods: [GET, POST, PUT, DELETE, OPTIONS]
    allow_headers: [Content-Type, Authorization, X-Requested-With]
    allow_credentials: false
    max_age: 86400
  rate_limit:
    enabled: true
    algorithm: token-bucket
    rate: 1000
    burst: 500
    reject_status_code: 429
  circuit_breaker:
    enabled: true
    minimum_requests: 100
    error_threshold_percentage: 50
    break_duration: 30s
    break_status_code: 503
  auth:
    enabled: false
    strategy: none
    jwt:
      secret: ""
      issuer: ""
      algorithm: HS256
      verify_expiration: true
      verify_issuer: true
```

### ExportToFile() - 导出到文件

```go
// 导出配置到指定文件
err := gateway.ExportToFile("backup/gateway_backup.yaml")
if err != nil {
    log.Printf("导出到文件失败: %v", err)
    return
}
```

**功能:**
- 自动创建目录（如果不存在）
- 以UTF-8编码保存文件
- 文件权限设置为0644
- 使用完整格式导出，包含所有配置项

### ExportCurrentRoutes() - 导出路由配置

```go
// 导出当前路由配置
routesYAML, err := gateway.ExportCurrentRoutes()
if err != nil {
    log.Printf("导出路由配置失败: %v", err)
    return
}
fmt.Println(routesYAML)
```

**示例输出:**
```yaml
routes:
  - id: api-route
    name: API路由
    path: /api/v1
    match_type: ""
    methods: [GET, POST, PUT, DELETE]
    service_id: api-service
    enabled: true
    priority: 100
    strip_prefix: true
    rewrite_path: /v1
    assertion_group:
      all_required: true
      assertions: []
```

### ExportCurrentServices() - 导出服务配置

```go
// 导出当前服务配置
servicesYAML, err := gateway.ExportCurrentServices()
if err != nil {
    log.Printf("导出服务配置失败: %v", err)
    return
}
fmt.Println(servicesYAML)
```

**示例输出:**
```yaml
services:
  - id: api-service
    name: API服务
    load_balance: round-robin
    endpoints:
      - http://localhost:8081
      - http://localhost:8082
```

## 使用场景

### 1. 版本回退
```go
// 导出当前配置用于版本回退
backupFile := fmt.Sprintf("backup/pre_upgrade_%s.yaml", 
    time.Now().Format("20060102_150405"))
err := gateway.ExportToFile(backupFile)

// 升级失败时可以使用此配置文件回退
// gateway.UpdateConfigFromFile(backupFile)
```

### 2. 环境迁移
```go
// 从开发环境导出完整配置
devConfig, _ := devGateway.ExportToYAML()

// 保存到文件，用于生产环境部署
os.WriteFile("prod_config.yaml", []byte(devConfig), 0644)
```

### 3. 配置备份
```go
// 定期完整备份配置
backupFile := fmt.Sprintf("backup/gateway_%s.yaml", 
    time.Now().Format("20060102_150405"))
err := gateway.ExportToFile(backupFile)
```

### 4. 配置版本控制
```go
// 导出配置用于版本控制
yamlContent, _ := gateway.ExportToYAML()
// 将yamlContent提交到Git仓库
// 确保每次变更都有完整的配置快照
```

### 5. 灾难恢复
```go
// 定期导出配置用于灾难恢复
ticker := time.NewTicker(12 * time.Hour)
go func() {
    for range ticker.C {
        config, _ := gateway.ExportToYAML()
        // 将配置上传到备份存储
        uploadToBackupStorage(config)
    }
}()
```

## 最佳实践

### 1. 定期完整备份
```go
// 设置定期完整备份任务
func setupPeriodicBackup(gateway *Gateway) {
    ticker := time.NewTicker(24 * time.Hour) // 每天备份一次
    go func() {
        for range ticker.C {
            backupFile := fmt.Sprintf("backup/daily/gateway_%s.yaml", 
                time.Now().Format("20060102"))
            if err := gateway.ExportToFile(backupFile); err != nil {
                log.Printf("定期备份失败: %v", err)
            } else {
                log.Printf("定期备份成功: %s", backupFile)
            }
        }
    }()
}
```

### 2. 变更前备份
```go
// 在进行配置变更前备份当前配置
func backupBeforeChange(gateway *Gateway, changeDesc string) (string, error) {
    timestamp := time.Now().Format("20060102_150405")
    backupFile := fmt.Sprintf("backup/pre_change_%s_%s.yaml", changeDesc, timestamp)
    return backupFile, gateway.ExportToFile(backupFile)
}

// 使用示例
backupFile, err := backupBeforeChange(gateway, "add_new_route")
if err != nil {
    log.Printf("变更前备份失败: %v", err)
    return
}
log.Printf("变更前配置已备份到: %s", backupFile)
```

### 3. 文件命名规范
```go
// 推荐的文件命名格式
type BackupType string

const (
    DailyBackup   BackupType = "daily"
    PreChange     BackupType = "pre_change" 
    PostChange    BackupType = "post_change"
    Emergency     BackupType = "emergency"
)

func generateBackupFilename(backupType BackupType, description string) string {
    timestamp := time.Now().Format("20060102_150405")
    if description != "" {
        return fmt.Sprintf("backup/%s/gateway_%s_%s.yaml", 
            backupType, description, timestamp)
    }
    return fmt.Sprintf("backup/%s/gateway_%s.yaml", 
        backupType, timestamp)
}
```

### 4. 完整的错误处理
```go
// 带重试的导出功能
func exportWithRetry(gateway *Gateway, filename string, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        if err := gateway.ExportToFile(filename); err != nil {
            log.Printf("导出失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
            if i < maxRetries-1 {
                time.Sleep(time.Second * time.Duration(i+1))
            }
            continue
        }
        
        // 验证导出的文件
        if err := validateBackupFile(filename); err != nil {
            log.Printf("备份文件验证失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
            if i < maxRetries-1 {
                time.Sleep(time.Second * time.Duration(i+1))
            }
            continue
        }
        
        return nil
    }
    return fmt.Errorf("导出失败，已重试 %d 次", maxRetries)
}

// 验证备份文件
func validateBackupFile(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("读取备份文件失败: %w", err)
    }
    
    // 验证YAML格式
    var config map[string]interface{}
    if err := yaml.Unmarshal(data, &config); err != nil {
        return fmt.Errorf("备份文件YAML格式无效: %w", err)
    }
    
    // 检查是否包含gateway配置
    if _, exists := config["gateway"]; !exists {
        return fmt.Errorf("备份文件缺少gateway配置节点")
    }
    
    return nil
}
```

### 5. 备份文件管理
```go
// 清理过期备份文件
func cleanupOldBackups(backupDir string, retentionDays int) error {
    cutoff := time.Now().AddDate(0, 0, -retentionDays)
    
    return filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
            if info.ModTime().Before(cutoff) {
                log.Printf("删除过期备份文件: %s", path)
                return os.Remove(path)
            }
        }
        
        return nil
    })
}
```

## 注意事项

1. **敏感信息**: 导出的配置包含所有字段，可能包含敏感信息（如JWT密钥），请妥善保管
2. **文件大小**: 完整配置导出的文件较大，注意磁盘空间管理
3. **版本兼容性**: 导出的配置反映当前版本的结构，确保回退时版本兼容
4. **备份验证**: 定期验证备份文件的完整性和可用性
5. **存储安全**: 备份文件应存储在安全的位置，并设置适当的访问权限

## 完整示例

参考 `examples/gateway_export/main.go` 文件，了解如何使用配置导出功能：

```go
package main

import (
    "fmt"
    "log"
    "time"
    "gohub/internal/gateway"
)

func main() {
    // 创建网关实例
    gw, err := gateway.New("configs/gateway.yaml")
    if err != nil {
        log.Fatalf("创建网关失败: %v", err)
    }

    // 1. 完整配置导出
    yamlContent, _ := gw.ExportToYAML()
    fmt.Println("完整配置:", yamlContent)

    // 2. 导出到文件
    backupFile := fmt.Sprintf("backup/gateway_%s.yaml", 
        time.Now().Format("20060102_150405"))
    gw.ExportToFile(backupFile)

    // 3. 分模块导出
    routes, _ := gw.ExportCurrentRoutes()
    services, _ := gw.ExportCurrentServices()
    
    fmt.Println("路由配置:", routes)
    fmt.Println("服务配置:", services)
}
``` 