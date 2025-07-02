# Oracle 数据库支持安装指南

## 概述

本项目已经实现了完整的 Oracle 数据库支持，但由于 Oracle 驱动需要 Oracle 客户端库，默认情况下不会编译 Oracle 支持以避免编译问题。本指南将说明如何启用 Oracle 数据库支持。

## 前置条件

### 1. 安装 Oracle 客户端

#### Windows

1. 下载 Oracle Instant Client：
   - 访问 [Oracle Instant Client 下载页面](https://www.oracle.com/database/technologies/instant-client/downloads.html)
   - 下载适合您系统的版本（推荐 19c 或 21c）
   - 下载 `instantclient-basic-windows.x64-xx.x.x.x.x.zip`

2. 解压并配置环境变量：
   ```powershell
   # 解压到指定目录，例如 C:\oracle\instantclient_19_1
   # 将 Oracle 客户端目录添加到 PATH 环境变量
   setx PATH "%PATH%;C:\oracle\instantclient_19_1"
   ```

#### Linux

1. 安装 Oracle Instant Client：
   ```bash
   # CentOS/RHEL
   sudo yum install oracle-instantclient-basic
   
   # Ubuntu/Debian
   # 需要先下载 .deb 包
   wget https://download.oracle.com/otn_software/linux/instantclient/211000/oracle-instantclient-basic-21.1.0.0.0-1.x86_64.rpm
   sudo alien -i oracle-instantclient-basic-21.1.0.0.0-1.x86_64.rpm
   
   # 或者手动安装
   wget https://download.oracle.com/otn_software/linux/instantclient/211000/instantclient-basic-linux.x64-21.1.0.0.0.zip
   unzip instantclient-basic-linux.x64-21.1.0.0.0.zip
   sudo mv instantclient_21_1 /opt/oracle/
   export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_1:$LD_LIBRARY_PATH
   ```

2. 配置环境变量：
   ```bash
   echo 'export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_1:$LD_LIBRARY_PATH' >> ~/.bashrc
   source ~/.bashrc
   ```

#### macOS

1. 安装 Oracle Instant Client：
   ```bash
   # 使用 Homebrew
   brew tap InstantClientTap/instantclient
   brew install instantclient-basic
   
   # 或者手动安装
   wget https://download.oracle.com/otn_software/mac/instantclient/211000/instantclient-basic-macos.x64-21.1.0.0.0.zip
   unzip instantclient-basic-macos.x64-21.1.0.0.0.zip
   sudo mv instantclient_21_1 /usr/local/lib/
   export DYLD_LIBRARY_PATH=/usr/local/lib/instantclient_21_1:$DYLD_LIBRARY_PATH
   ```

### 2. 验证安装

验证 Oracle 客户端是否正确安装：

```bash
# Linux/macOS
ldd /path/to/instantclient/libclntsh.so

# Windows
# 检查环境变量 PATH 是否包含 Oracle 客户端目录
echo $env:PATH
```

## 启用 Oracle 支持

### 方法 1: 修改 alldriver 包 (推荐)

1. 编辑 `pkg/database/alldriver/drivers.go` 文件：
   ```go
   import (
       // 导入MySQL驱动包，确保其init()函数被调用
       _ "gohub/pkg/database/mysql"
       
       // 启用Oracle驱动 - 取消注释以下行
       _ "gohub/pkg/database/oracle"
       
       // 未来可能添加的其他驱动
       // _ "gohub/pkg/database/postgres"
       // _ "gohub/pkg/database/sqlite"
   )
   ```

2. 使用 build tag 编译：
   ```bash
   go build -tags oracle ./...
   ```

### 方法 2: 在应用代码中直接导入

在需要使用 Oracle 的 Go 文件中添加导入：

```go
package main

import (
    "gohub/pkg/database"
    _ "gohub/pkg/database/mysql"        // MySQL 支持
    _ "gohub/pkg/database/oracle"       // Oracle 支持
)

func main() {
    // 使用 Oracle 数据库
    config := &database.DbConfig{
        Driver: "oracle",
        DSN:    "oracle://username:password@host:port/service_name",
    }
    
    db, err := database.Open(config)
    if err != nil {
        panic(err)
    }
    defer db.Close()
}
```

然后使用 build tag 编译：
```bash
go build -tags oracle main.go
```

### 方法 3: 设置环境变量

设置 CGO 环境变量以确保正确链接：

```bash
# Linux/macOS
export CGO_ENABLED=1
export PKG_CONFIG_PATH=/path/to/instantclient:$PKG_CONFIG_PATH

# Windows (PowerShell)
$env:CGO_ENABLED = "1"
```

## 验证 Oracle 支持

### 1. 编译测试

```bash
# 使用 Oracle build tag 编译
go build -tags oracle ./pkg/database/oracle

# 编译整个项目
go build -tags oracle ./...
```

### 2. 运行测试

```bash
# 运行 Oracle 驱动注册测试（无需实际数据库）
go test -tags oracle -v ./test/database -run TestOracleDriverRegistration

# 运行 Oracle 连接测试（需要 Oracle 数据库）
go test -tags oracle -v ./test/database -run TestOracleConnection
```

### 3. 代码示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "gohub/pkg/database"
    _ "gohub/pkg/database/oracle"
)

func main() {
    config := &database.DbConfig{
        Name:   "oracle_test",
        Driver: "oracle",
        DSN:    "oracle://scott:tiger@localhost:1521/XEPDB1",
        Pool: database.PoolConfig{
            MaxOpenConns:    10,
            MaxIdleConns:    5,
            ConnMaxLifetime: 3600,
            ConnMaxIdleTime: 1800,
        },
        Log: database.LogConfig{
            Enable:        true,
            SlowThreshold: 200,
        },
    }

    db, err := database.Open(config)
    if err != nil {
        log.Fatal("连接Oracle失败:", err)
    }
    defer db.Close()

    // 测试连接
    ctx := context.Background()
    err = db.Ping(ctx)
    if err != nil {
        log.Fatal("Ping Oracle失败:", err)
    }

    fmt.Println("Oracle连接成功!")
}
```

编译并运行：
```bash
go build -tags oracle main.go
./main
```

## 常见问题

### 1. 编译错误

**错误**: `undefined: VersionInfo`
**解决**: 确保 Oracle 客户端已正确安装，环境变量已设置

**错误**: `cannot find package "github.com/godror/godror"`
**解决**: 运行 `go mod tidy` 确保依赖已下载

### 2. 运行时错误

**错误**: `ORA-12154: TNS:could not resolve the connect identifier specified`
**解决**: 检查 DSN 格式和网络连接

**错误**: `ORA-12505: TNS:listener does not currently know of SID given in connect descriptor`
**解决**: 确认 SERVICE_NAME 或 SID 名称正确

### 3. 性能问题

- 调整连接池参数
- 启用 SQL 日志分析慢查询
- 检查 Oracle 数据库统计信息

## 配置文件示例

参考 `configs/oracle_example.yaml` 获取完整的配置示例。

## 支持的 Oracle 版本

- Oracle 11g (11.2.0.4+)
- Oracle 12c (12.1.0.2+)
- Oracle 18c
- Oracle 19c (推荐)
- Oracle 21c

## 更多资源

- [Oracle 官方文档](https://docs.oracle.com/en/database/)
- [godror 驱动文档](https://github.com/godror/godror)
- [Oracle Instant Client 下载](https://www.oracle.com/database/technologies/instant-client.html) 