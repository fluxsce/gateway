# CMD 公共工具包

这个包提供了 Gateway 各个 cmd 应用程序共用的工具函数。

## 配置目录管理

### 功能特性

- 🎯 **统一配置路径管理** - 所有 cmd 应用使用统一的配置路径获取方式
- 🔧 **命令行参数支持** - 支持 `--config` 参数指定配置目录
- 🌍 **环境变量支持** - 支持 `GATEWAY_CONFIG_DIR` 环境变量
- ⚡ **优先级控制** - 命令行参数 > 环境变量 > 默认值

### 使用方法

#### 1. 导入包

```go
import "gateway/cmd/common/utils"
```

#### 2. 获取配置目录

```go
// 获取配置目录路径
configDir := utils.GetConfigDir()

// 获取特定配置文件路径
dbConfigPath := utils.GetConfigPath("database.yaml")
gatewayConfigPath := utils.GetConfigPath("gateway.yaml")
```

#### 3. 命令行参数使用

支持的命令行参数：

```bash
# 使用默认配置目录 (./configs)
./gateway-app

# 通过命令行参数指定配置目录
./gateway-app --config /path/to/config

# 通过环境变量指定配置目录
GATEWAY_CONFIG_DIR=/path/to/config ./gateway-app

# 命令行参数优先于环境变量
GATEWAY_CONFIG_DIR=/env/config ./gateway-app --config /cli/config
# 实际使用 /cli/config
```

### 优先级说明

配置目录的获取优先级：

1. **命令行参数** (`--config`)
2. **环境变量** (`GATEWAY_CONFIG_DIR`)  
3. **默认值** (`./configs`)

### API 参考

#### `GetConfigDir() string`

获取配置目录路径。

**返回值：**
- `string` - 配置目录的绝对或相对路径

#### `GetConfigPath(filename string) string`

获取指定配置文件的完整路径。

**参数：**
- `filename` - 配置文件名（如 "database.yaml"）

**返回值：**
- `string` - 配置文件的完整路径

#### `SetConfigDir(dir string)`

动态设置配置目录路径（主要用于程序内部）。

**参数：**
- `dir` - 配置目录路径

#### `ResetFlags()`

重置 flag 解析状态（主要用于测试）。

### 使用示例

#### 在 main.go 中使用

```go
package main

import (
    "gateway/cmd/common/utils"
    "gateway/pkg/config"
)

func initConfig() error {
    // 获取配置目录（自动处理命令行参数和环境变量）
    configDir := utils.GetConfigDir()
    
    // 加载配置
    return config.LoadConfig(configDir, options)
}

func initDatabase() error {
    // 获取数据库配置文件路径
    configPath := utils.GetConfigPath("database.yaml")
    
    // 加载数据库连接
    return database.LoadAllConnections(configPath)
}
```

#### 在其他 cmd 应用中使用

```go
package gatewayapp

import "gateway/cmd/common/utils"

func loadFromFile() error {
    // 获取网关配置文件路径
    configFile := utils.GetConfigPath("gateway.yaml")
    
    // 加载配置
    return loadConfig(configFile)
}
```

### 测试

运行测试：

```bash
go test ./cmd/common/utils
```

### 兼容性

- ✅ 与现有的环境变量方式完全兼容
- ✅ 与现有的配置加载逻辑完全兼容
- ✅ 不影响现有部署脚本的环境变量设置

### 部署示例

```bash
# 开发环境 - 使用默认配置目录
./gateway-app

# 生产环境 - 使用环境变量
GATEWAY_CONFIG_DIR=/opt/gateway/configs ./gateway-app

# 临时测试 - 使用命令行参数
./gateway-app --config /tmp/test-configs

# Docker 容器中
docker run -e GATEWAY_CONFIG_DIR=/app/configs gateway-app

# Windows 服务中（自动设置环境变量）
# 服务会自动使用 GATEWAY_CONFIG_DIR 环境变量
``` 