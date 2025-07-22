# GoHub Windows 部署指南

本文档介绍如何在Windows环境下部署GoHub应用并将其注册为Windows服务。

## 概述

GoHub提供了两种Windows服务安装方式：
1. **原生方式** - 使用Windows内置的sc命令
2. **NSSM方式** - 使用NSSM (Non-Sucking Service Manager) 工具

## 前置条件

### 1. 编译GoHub应用

首先需要编译GoHub应用程序：

```cmd
# 编译标准版本
cd /d F:\goWorkSpace\gohub
.\scripts\build\build-win10-oracle.cmd

# 或编译Windows 2008兼容版本
.\scripts\build\build-win2008-oracle.cmd
```

编译完成后，会在`dist`目录生成对应的可执行文件。

### 2. 准备部署目录

建议将编译好的程序部署到固定目录，例如：
- `C:\Program Files\GoHub`
- `D:\Apps\GoHub`

确保目录结构如下：
```
GoHub/
├── gohub-win10-oracle-amd64.exe (或 gohub-win2008-oracle-amd64.exe)
├── configs/
│   ├── app.yaml
│   ├── database.yaml
│   └── logger.yaml
├── logs/
└── scripts/
    ├── install-service.cmd
    ├── install-service-nssm.cmd
    └── uninstall-service.cmd
```

## 服务安装方式

### 方式一：原生Windows服务 (推荐用于测试)

使用Windows内置的sc命令注册服务：

```cmd
# 以管理员身份运行命令提示符
# 安装标准版本服务
.\scripts\deploy\install-service.cmd

# 安装Oracle版本服务
.\scripts\deploy\install-service.cmd oracle

# 指定自定义目录
.\scripts\deploy\install-service.cmd -d "D:\Apps\GoHub"
```

**注意：** 此方式要求GoHub程序支持`--service`参数，目前需要修改程序代码以支持Windows服务模式。

**智能检测功能：** 脚本具备以下自动检测能力：

1. **程序目录检测** - 自动检测应用程序所在目录：
   - 源码目录执行：自动检测脚本上级目录
   - 部署目录执行：自动检测程序实际位置  
   - 同目录执行：自动检测脚本当前目录

2. **文件名检测** - 自动检测以下文件名：
   - Oracle版本：`gohub-win10-oracle-amd64.exe`、`gohub-win2008-oracle-amd64.exe`、`gohub-oracle.exe`
   - 标准版本：`gohub.exe`、`gohub-win10-amd64.exe`、`gohub-win2008-amd64.exe`

### 方式二：NSSM方式 (推荐用于生产环境)

使用NSSM工具将GoHub包装为Windows服务：

#### 1. 下载NSSM

访问 [https://nssm.cc/download](https://nssm.cc/download) 下载NSSM：
- 下载对应架构的版本 (win32/win64)
- 解压到 `scripts\deploy\tools\` 目录
- 确保 `scripts\deploy\tools\nssm.exe` 存在

#### 2. 安装服务

```cmd
# 以管理员身份运行命令提示符
# 安装标准版本服务
.\scripts\deploy\install-service-nssm.cmd

# 安装Oracle版本服务
.\scripts\deploy\install-service-nssm.cmd oracle

# 指定自定义目录
.\scripts\deploy\install-service-nssm.cmd -d "D:\Apps\GoHub"
```

## 服务管理

### 1. 启动/停止服务

使用Windows服务管理器：
```cmd
# 打开服务管理器
services.msc
```

或使用命令行：
```cmd
# 启动服务
sc start GoHub
# 或
net start GoHub

# 停止服务
sc stop GoHub
# 或
net stop GoHub

# 查看服务状态
sc query GoHub
```

### 2. 使用NSSM管理服务

```cmd
# 启动服务
nssm start GoHub

# 停止服务
nssm stop GoHub

# 重启服务
nssm restart GoHub

# 查看服务状态
nssm status GoHub

# 编辑服务配置
nssm edit GoHub
```

### 3. 卸载服务

**基本卸载：**
```cmd
# 卸载标准版本服务
.\scripts\deploy\uninstall-service.cmd

# 卸载Oracle版本服务
.\scripts\deploy\uninstall-service.cmd oracle
```

**高级卸载选项：**
```cmd
# 强制卸载（跳过确认）
.\scripts\deploy\uninstall-service.cmd oracle --force

# 卸载并清理日志文件
.\scripts\deploy\uninstall-service.cmd oracle --clean-logs

# 卸载但保留环境变量
.\scripts\deploy\uninstall-service.cmd oracle --keep-env

# 卸载自定义服务名
.\scripts\deploy\uninstall-service.cmd -s "Custom-GoHub-Service"

# 完整清理（推荐）
.\scripts\deploy\uninstall-service.cmd oracle --clean-logs --force
```

**卸载脚本特性：**
- 🔍 **智能检测** - 自动识别NSSM服务和原生Windows服务
- 🗑️ **完整清理** - 可选择清理日志文件和环境变量
- ⚡ **强制模式** - 支持无人值守批量卸载
- 🛡️ **安全检查** - 等待服务完全停止后再删除
- 📝 **详细反馈** - 显示每个步骤的执行状态

## 配置说明

### 1. 环境变量

服务会自动设置以下环境变量：
- `GOHUB_CONFIG_DIR` - 配置文件目录路径

### 2. 日志文件

服务日志位置：
- **原生服务**：`logs\app.log`
- **NSSM服务**：
  - 标准输出：`logs\service.log`
  - 错误输出：`logs\service-error.log`

### 3. 服务配置

服务默认配置：
- **服务名称**：GoHub (或 GoHub-Oracle)
- **显示名称**：GoHub Application Service
- **启动类型**：自动启动
- **恢复策略**：失败时自动重启

## 完整部署流程

### 1. 开发环境快速部署

```cmd
# 1. 编译程序
.\scripts\build\build-win10-oracle.cmd

# 2. 创建部署目录
mkdir C:\GoHub-Dev
xcopy /E /I configs C:\GoHub-Dev\configs
copy dist\gohub-win10-oracle-amd64.exe C:\GoHub-Dev\gohub-win10-oracle-amd64.exe
xcopy /E /I scripts C:\GoHub-Dev\scripts

# 3. 安装服务 (使用NSSM)
cd C:\GoHub-Dev
.\scripts\deploy\install-service-nssm.cmd oracle -d "C:\GoHub-Dev"
```

### 2. 生产环境部署

```cmd
# 1. 编译程序
.\scripts\build\build-win2008-oracle.cmd

# 2. 创建部署目录
mkdir "C:\Program Files\GoHub"
xcopy /E /I configs "C:\Program Files\GoHub\configs"
copy dist\gohub-win2008-oracle-amd64.exe "C:\Program Files\GoHub\gohub-win2008-oracle-amd64.exe"
xcopy /E /I scripts "C:\Program Files\GoHub\scripts"

# 3. 安装服务
cd "C:\Program Files\GoHub"
.\scripts\deploy\install-service-nssm.cmd oracle

# 4. 验证服务
sc query GoHub-Oracle
```

## 故障排除

### 1. 服务启动失败

检查事项：
- 可执行文件是否存在
- 配置文件是否正确
- 日志文件权限是否足够
- 依赖的数据库是否可访问

查看日志：
```cmd
# 查看服务日志
type "C:\Program Files\GoHub\logs\service.log"
type "C:\Program Files\GoHub\logs\service-error.log"
```

### 2. 权限问题

确保以管理员身份运行安装脚本：
```cmd
# 右键点击"命令提示符"，选择"以管理员身份运行"
```

### 3. Oracle版本问题

如果使用Oracle版本，确保：
- Oracle Instant Client已正确安装
- 相关DLL文件在PATH中
- 数据库连接配置正确

### 4. 端口冲突

检查配置文件中的端口设置：
- Gateway端口 (默认8080)
- Web端口 (默认8081)
- API端口 (默认8082)

## 高级配置

### 1. 自定义服务参数

编辑NSSM服务配置：
```cmd
nssm edit GoHub
```

可以配置：
- 启动参数
- 环境变量
- 用户账户
- 依赖服务

### 2. 服务监控

可以配置服务监控和告警：
- Windows事件日志
- 性能计数器
- 第三方监控工具

### 3. 多实例部署

可以部署多个GoHub实例：
```cmd
# 部署第二个实例
.\scripts\deploy\install-service-nssm.cmd oracle -d "C:\GoHub-Instance2"
```

## 安全建议

1. **最小权限原则**
   - 为服务创建专用用户账户
   - 限制文件和目录访问权限

2. **网络安全**
   - 配置防火墙规则
   - 使用HTTPS和安全认证

3. **日志管理**
   - 定期清理日志文件
   - 设置日志轮转策略

4. **更新维护**
   - 定期更新程序版本
   - 监控服务运行状态

## 🆕 v2.0 智能检测功能

### 新增特性

- **🔍 智能程序目录检测**：自动检测应用程序所在目录，无需手动指定路径
- **📁 多场景执行支持**：脚本可在源码目录、部署目录或任意位置执行
- **🎯 精确文件识别**：自动识别正确的可执行文件版本
- **⚡ 动态路径解析**：根据脚本执行位置智能推导程序位置
- **🛡️ 增强容错处理**：检测失败时提供详细的诊断信息
- **⚙️ 配置目录统一化**：修复配置文件路径不一致问题，统一使用GOHUB_CONFIG_DIR环境变量

### 支持的执行场景

| 执行场景 | 脚本位置 | 程序位置 | 检测方式 |
|----------|----------|----------|----------|
| **源码开发** | `F:\gohub\scripts\deploy\` | `F:\gohub\` | 检测脚本上级目录 |
| **部署环境** | `C:\Program Files\GoHub\scripts\` | `C:\Program Files\GoHub\` | 检测脚本父目录 |
| **便携模式** | `C:\GoHub\` | `C:\GoHub\` | 检测脚本当前目录 |
| **自定义路径** | 任意位置 | 通过 `-d` 参数指定 | 使用指定目录 |

### 文件检测优先级

脚本按以下优先级自动检测可执行文件：

**Oracle版本** (使用 `oracle` 参数时)：
1. `gohub-win10-oracle-amd64.exe` - Windows 10/11版本
2. `gohub-win2008-oracle-amd64.exe` - Windows Server 2008版本  
3. `gohub-oracle.exe` - 通用Oracle版本

**标准版本** (不使用 `oracle` 参数时)：
1. `gohub.exe` - 标准版本
2. `gohub-win10-amd64.exe` - Windows 10/11版本
3. `gohub-win2008-amd64.exe` - Windows Server 2008版本

### 使用示例

```cmd
# 在源码目录执行 - 自动检测到项目根目录
F:\gohub\scripts\deploy> install-service-nssm.cmd oracle

# 在部署目录执行 - 自动检测到应用目录  
C:\Program Files\GoHub\scripts> install-service-nssm.cmd oracle

# 在任意位置执行 - 手动指定目录
D:\Tools> install-service-nssm.cmd oracle -d "C:\Program Files\GoHub"
```

## 配置目录统一化修复

**修复的问题：**
- 原先数据库配置、缓存配置使用硬编码路径 `configs/database.yaml`
- 主应用配置支持 `GOHUB_CONFIG_DIR` 环境变量，但其他组件不支持
- 导致配置文件路径不一致，部署时可能出现配置文件找不到的问题

**修复后的效果：**
- 所有组件统一使用 `GOHUB_CONFIG_DIR` 环境变量
- 配置文件路径完全一致，支持灵活的部署场景
- 服务安装脚本自动设置环境变量，确保服务启动时使用正确的配置目录

**配置目录优先级：**
1. `GOHUB_CONFIG_DIR` 环境变量指定的目录
2. `./configs` （相对于程序启动目录）
3. `.` （程序启动目录）

**验证配置目录：**
```cmd
# 运行配置目录测试脚本
.\scripts\deploy\test-config-dir.cmd
```

## 联系支持

如有问题，请参考：
- 项目文档：`docs/`
- 配置示例：`configs/`
- 故障排除：`docs/DEBUGGING.md` 