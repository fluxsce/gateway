# Gateway 交叉编译构建脚本

本目录包含用于 Gateway 项目的交叉编译构建脚本，支持在 Windows 环境下构建多个目标平台的二进制文件。

## 支持的目标平台

- **Linux**
  - AMD64 (x86_64)
  - ARM64
  - Oracle 数据库支持（可选）

- **macOS (Darwin)**
  - AMD64 (x86_64)
  - ARM64 (Apple Silicon)

- **Windows**
  - AMD64 (x86_64) - 兼容 Windows Server 2008 及以上版本

## 构建脚本

### build.cmd
Windows 批处理脚本，提供图形化菜单界面：

```cmd
.\scripts\build\build.cmd
```

**菜单选项：**
1. 构建 Linux AMD64 版本
2. 构建 Linux AMD64 版本 (包含 Oracle)
3. 构建 Linux ARM64 版本
4. 构建 macOS AMD64 版本
5. 构建 macOS ARM64 版本
6. **构建 Windows Server 2008 64位版本** ⭐ 新增
7. 构建所有版本
8. 构建所有版本 (包含 Oracle)
9. 自定义构建
10. 查看帮助

### build-centos7.sh ⭐ 新增
Linux Shell 脚本，用于在 Linux 环境下构建 CentOS 7 兼容版本：

```bash
# 使用当前 Go 版本构建
./scripts/build/build-centos7.sh

# 指定 Go 版本构建
./scripts/build/build-centos7.sh 1.20
```

**特性：**
- 自动检测 Go 版本
- 支持指定 Go 版本构建（1.19, 1.20, 1.21+）
- 自动选择对应的 go.mod 文件
- 纯 Go 编译（CGO_ENABLED=0）
- 优化的 CentOS 7 兼容性
- 自动恢复原始 go.mod

### build-win10-centos7.cmd
Windows 批处理脚本，用于在 Windows 环境下交叉编译 CentOS 7 版本：

```cmd
.\scripts\build\build-win10-centos7.cmd

REM 指定 Go 版本
.\scripts\build\build-win10-centos7.cmd 1.20
```

**特性：**
- 与 build-centos7.sh 功能对应
- 在 Windows 上交叉编译 Linux 二进制
- 自动检测和切换 Go 版本

### crossbuild.ps1
PowerShell 交叉编译脚本：

```powershell
# 构建 Windows 版本
.\scripts\build\crossbuild.ps1 -TargetOS windows -TargetArch amd64

# 构建 Linux 版本（包含 Oracle）
.\scripts\build\crossbuild.ps1 -TargetOS linux -TargetArch amd64 -EnableOracle

# 构建 macOS ARM64 版本
.\scripts\build\crossbuild.ps1 -TargetOS darwin -TargetArch arm64
```

### build-all.ps1
批量构建所有目标平台：

```powershell
# 构建所有平台（包括 Windows）
.\scripts\build\build-all.ps1

# 构建所有平台（包含 Oracle，仅限 Linux）
.\scripts\build\build-all.ps1 -EnableOracle

# 构建指定目标
.\scripts\build\build-all.ps1 -Targets @("linux-amd64", "windows-amd64")
```

## Windows 构建特性

### 兼容性
- **目标系统**: Windows Server 2008 R2 及以上版本
- **架构**: 64位 (AMD64)
- **运行时**: 无需额外依赖，纯 Go 编译

### 构建特点
- 使用 `CGO_ENABLED=0` 进行纯 Go 编译
- 生成 `.exe` 可执行文件
- 不支持 Oracle 数据库驱动（Oracle 驱动需要 CGO）
- 文件名格式：`gateway-windows-amd64.exe`

### 系统要求
- Windows 10/11 开发环境
- Docker Desktop（用于交叉编译）
- PowerShell 5.0+

## CentOS 7 专用构建 ⭐ 新增

针对 CentOS 7 系统的优化构建，确保在旧版 Linux 系统上的兼容性。

### ⚠️ 重要说明：SQLite 支持

**SQLite 需要 CGO 支持**，构建时有两个选项：

#### 选项 1：纯 Go 构建（推荐）
- ✅ **优点**：完全静态编译，无外部依赖，跨平台部署简单
- ❌ **缺点**：不支持 SQLite
- 🗄️ **支持的数据库**：MySQL, ClickHouse
- 🔧 **构建标签**：`netgo,osusergo,no_oracle,no_sqlite`
- ⚙️ **编译配置**：`CGO_ENABLED=0`

#### 选项 2：包含 SQLite 支持
- ✅ **优点**：支持 SQLite 数据库
- ❌ **缺点**：需要 GCC 编译器，二进制文件有外部依赖
- 🗄️ **支持的数据库**：MySQL, SQLite, ClickHouse
- 🔧 **构建标签**：`netgo,no_oracle`
- ⚙️ **编译配置**：`CGO_ENABLED=1`

### 构建标签说明
- `netgo`: 使用纯 Go 网络实现
- `osusergo`: 使用纯 Go 用户/组查询
- `no_oracle`: 禁用 Oracle 驱动（避免 CGO 依赖）
- `no_sqlite`: 禁用 SQLite 驱动（避免 CGO 依赖）

### 使用场景
- 需要在 CentOS 7 上部署
- 需要最大化兼容性
- 不需要 Oracle 数据库支持
- 根据是否需要 SQLite 选择构建模式

### 构建方式

**在 Linux 上构建：**
```bash
./scripts/build/build-centos7.sh
# 脚本会提示选择：
# 1) 包含 SQLite 支持（需要 GCC）
# 2) 纯 Go 构建（推荐，无 SQLite）
```

**在 Windows 上交叉编译：**
```cmd
.\scripts\build\build-win10-centos7.cmd
# 脚本会提示选择：
# 1) 纯 Go 构建（推荐，无 SQLite）
# 2) 包含 SQLite 支持（需要 MinGW/GCC）
```

## 构建产物

构建完成的二进制文件将保存在 `./dist/` 目录下：

```
dist/
├── gateway-linux-amd64           # Linux 64位版本
├── gateway-linux-amd64-oracle    # Linux 64位版本（Oracle支持）
├── gateway-linux-arm64           # Linux ARM64版本
├── gateway-centos7-amd64         # CentOS 7 版本 ⭐
├── gateway-darwin-amd64          # macOS Intel版本
├── gateway-darwin-arm64          # macOS ARM版本
└── gateway-windows-amd64.exe     # Windows 64位版本
```

**注意：**
- `gateway-centos7-amd64` 文件名固定，不随构建模式改变（用于应用自动注册）
- 构建时可选择是否包含 SQLite 支持，但输出文件名保持一致

## 使用示例

### 快速构建 CentOS 7 版本

**在 Linux 环境：**
```bash
# 赋予执行权限
chmod +x ./scripts/build/build-centos7.sh

# 使用当前 Go 版本构建
./scripts/build/build-centos7.sh

# 使用指定 Go 版本构建
./scripts/build/build-centos7.sh 1.20
```

**在 Windows 环境：**
```cmd
# 使用当前 Go 版本构建
.\scripts\build\build-win10-centos7.cmd

# 使用指定 Go 版本构建
.\scripts\build\build-win10-centos7.cmd 1.20
```

### 在 CentOS 7 上部署
1. 将构建的 `gateway-centos7-amd64` 复制到目标服务器
2. 准备配置文件（`configs/` 目录）
3. 赋予执行权限并运行：
   ```bash
   chmod +x gateway-centos7-amd64
   ./gateway-centos7-amd64
   ```

### 快速构建 Windows 版本
```cmd
# 使用批处理脚本
.\scripts\build\build.cmd
# 选择选项 6

# 或直接使用 PowerShell
powershell -ExecutionPolicy Bypass -File ".\scripts\build\crossbuild.ps1" -TargetOS windows -TargetArch amd64
```

### 在 Windows Server 2008 上部署
1. 将构建的 `gateway-windows-amd64.exe` 复制到目标服务器
2. 准备配置文件（`configs/` 目录）
3. 直接运行：
   ```cmd
   gateway-windows-amd64.exe
   ```

## 故障排查

### 常见问题

**1. Docker 相关错误**
- 确保 Docker Desktop 正在运行
- 检查 Docker 版本兼容性

**2. 网络连接问题**
- 构建过程需要下载 Go 模块和依赖
- 确保网络连接正常

**3. 磁盘空间不足**
- 构建过程需要至少 2GB 可用空间
- Docker 镜像会占用额外空间

**4. Windows 版本特定问题**
- Windows 版本不支持 Oracle 驱动
- 如果需要 Oracle 支持，请使用 Linux 版本

### 调试模式
```powershell
# 启用详细输出
.\scripts\build\crossbuild.ps1 -TargetOS windows -TargetArch amd64 -Verbose
```

## 更新日志

### v1.2.0 (2025-10-24)
- ✅ 新增 CentOS 7 专用构建脚本（Linux 和 Windows 版本）
- ✅ 支持自动检测和切换 Go 版本
- ✅ 优化构建标签以提高 CentOS 7 兼容性
- ✅ 改进模块文件管理和恢复机制
- ✅ 增强错误处理和诊断信息

### v1.1.0
- ✅ 新增 Windows Server 2008 64位构建支持
- ✅ 更新所有构建脚本以支持 Windows 目标
- ✅ 优化文件命名规则（Windows 版本添加 .exe 扩展名）
- ✅ 更新帮助文档和错误提示

### v1.0.0
- ✅ 支持 Linux AMD64/ARM64 构建
- ✅ 支持 macOS AMD64/ARM64 构建
- ✅ 支持 Oracle 数据库驱动（仅限 Linux）
- ✅ Docker 容器化构建环境 

# Gateway Windows Server 2008 Oracle兼容构建脚本

## 📋 概述

本目录包含了专门为Windows Server 2008系统设计的Gateway构建脚本，支持Oracle数据库驱动和纯Go两种构建模式。

## 🎯 主要特性

### ✅ 双模式构建支持
- **Oracle驱动版本**: 完整的Oracle数据库支持，包含CGO编译
- **纯Go版本**: 轻量级构建，支持MySQL/SQLite，无需额外依赖

### ✅ 智能Oracle环境检测
- 自动检测多种Oracle安装结构（标准/非标准）
- 支持oci.dll在任意子目录的安装
- 智能搜索头文件和链接库
- 自动配置CGO编译参数

### ✅ 增强的错误处理
- 详细的构建失败分析
- 针对性的解决方案建议
- 完整的Oracle环境诊断工具

## 📁 脚本文件说明

| 文件名 | 功能 | 使用场景 |
|--------|------|----------|
| `build-win2008-oracle.cmd` | 主构建脚本 | 日常构建使用 |
| `setup-oracle-env.cmd` | Oracle环境设置 | 首次配置Oracle环境 |
| `Oracle环境检查.cmd` | 环境诊断工具 | 构建过程中自动生成 |

## 🚀 快速开始

### 1. 首次使用
```bash
# 运行Oracle环境设置脚本（可选）
scripts\build\setup-oracle-env.cmd

# 运行主构建脚本
scripts\build\build-win2008-oracle.cmd
```

### 2. 选择构建模式
运行主脚本后，系统会提示选择：
- **选项1**: Oracle驱动版本 - 需要Oracle客户端和GCC编译器
- **选项2**: 纯Go版本 - 无需额外依赖，构建速度快

### 3. 构建输出
成功构建后，在`dist/`目录会生成：
- 可执行文件
- 部署说明文档
- 启动脚本
- 环境检查工具（Oracle版本）

## 🔧 Oracle环境要求

### 必需组件
- **Oracle Instant Client Basic**: 运行时库
- **Oracle Instant Client SDK**: 开发头文件和链接库
- **MinGW/GCC编译器**: CGO编译支持

### 支持的Oracle目录结构

#### 标准结构
```
C:\oracle\instantclient_21_8\
├── bin\
│   └── oci.dll
├── sdk\
│   └── include\
│       └── oci.h
└── lib\
    └── oci.lib
```

#### 非标准结构（自动检测）
```
C:\oracle\custom_install\
├── oci.dll          # 根目录
├── include\
│   └── oci.h         # 直接在include目录
└── oci.lib           # 与DLL同目录
```

## 📝 使用指南

### Oracle版本构建
1. 确保Oracle客户端已安装
2. 运行构建脚本选择选项1
3. 脚本会自动：
   - 检测Oracle安装位置
   - 验证必需文件存在性
   - 配置CGO编译参数
   - 执行构建

### 纯Go版本构建
1. 运行构建脚本选择选项2
2. 无需额外配置
3. 支持MySQL和SQLite数据库

## 🛠 故障排查

### 常见问题

#### 1. Oracle环境问题
**错误**: `cannot find -loci`
**解决**: 
- 运行`setup-oracle-env.cmd`配置环境
- 确保下载了SDK包
- 检查ORACLE_HOME环境变量

#### 2. 编译器问题
**错误**: `gcc not found`
**解决**: 
- 安装MinGW或TDM-GCC
- 确保GCC在PATH中
- 推荐使用64位版本

#### 3. 非标准Oracle安装
**问题**: 脚本无法找到Oracle文件
**解决**: 
- 脚本会自动搜索常见位置
- 支持任意子目录结构
- 手动设置ORACLE_HOME环境变量

### 诊断工具
```bash
# 运行Oracle环境检查（构建后生成）
dist\Oracle环境检查.cmd

# 手动验证Oracle环境
where oci.dll
echo %ORACLE_HOME%
gcc --version
```

## 📋 输出文件说明

### Oracle版本构建输出
- `gateway-windows-server2008-oracle-amd64.exe`: 主程序
- `Windows-Server-2008-Oracle驱动版本-部署说明.txt`: 详细部署文档
- `启动Gateway-Oracle驱动版本.cmd`: 启动脚本
- `Oracle环境检查.cmd`: 环境诊断工具

### 纯Go版本构建输出
- `gateway-windows-server2008-pure-amd64.exe`: 主程序
- `Windows-Server-2008-纯Go版本-部署说明.txt`: 部署文档
- `启动Gateway-纯Go版本.cmd`: 启动脚本

## 🔄 版本历史

### v1.1 (最新)
- ✅ 新增双模式构建选择
- ✅ 智能Oracle环境检测
- ✅ 支持非标准Oracle目录结构
- ✅ 增强错误处理和诊断
- ✅ 独立的Oracle环境设置工具

### v1.0
- ✅ 基础Oracle构建支持
- ✅ Windows Server 2008兼容性
- ✅ 基本的环境检查

## 📞 技术支持

如果遇到问题：
1. 运行Oracle环境检查工具
2. 查看详细的错误信息和解决建议
3. 参考生成的部署说明文档
4. 考虑使用纯Go版本作为备选方案

## 🎯 最佳实践

1. **首次使用**: 先运行`setup-oracle-env.cmd`配置环境
2. **开发环境**: 推荐使用Oracle版本获得完整数据库支持
3. **生产部署**: 如果不需要Oracle，使用纯Go版本更简单
4. **环境迁移**: 使用生成的部署说明文档确保一致性 