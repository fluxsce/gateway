# Gateway 构建脚本

本目录包含用于 Gateway 项目的构建脚本，支持在 Windows 和 Linux 环境下构建二进制文件。

## 支持的目标平台

- **Windows**
  - AMD64 (x86_64) - Windows 10/11 和 Windows Server 2008 及以上版本
  - Oracle 数据库支持（可选）

- **Linux**
  - AMD64 (x86_64) - CentOS 7 兼容版本
  - 支持 MySQL、SQLite、ClickHouse 数据库

## 构建脚本

### build-win10.cmd ⭐
Windows 批处理脚本，用于在 Windows 10/11 环境下构建：

```cmd
# 默认构建（包含 Oracle 支持）
.\scripts\build\build-win10.cmd

# 不包含 Oracle（仅 MySQL）
.\scripts\build\build-win10.cmd --no-oracle
```

**特性：**
- 默认包含 Oracle 支持（可选禁用）
- 使用默认 go.mod 文件
- 输出文件：`dist/gateway.exe`
- 支持 Oracle 和 MySQL 数据库

### build-centos7.sh ⭐
Linux Shell 脚本，用于在 Linux 环境下构建 CentOS 7 兼容版本：

```bash
# 默认构建所有功能（包含 SQLite 支持）
./scripts/build/build-centos7.sh
```

**特性：**
- 默认构建所有功能（CGO_ENABLED=1，包含 SQLite）
- 使用默认 go.mod 文件
- 优化的 CentOS 7 兼容性
- 输出文件：`dist/gateway`
- 支持 MySQL、SQLite、ClickHouse 数据库

### build-win2008-oracle.cmd
Windows 批处理脚本，用于构建 Windows Server 2008 兼容版本（支持 Oracle）：

```cmd
.\scripts\build\build-win2008-oracle.cmd
```

**特性：**
- 支持 Oracle 和纯 Go 两种构建模式
- 智能 Oracle 环境检测
- 输出文件：`dist/gateway-windows-server2008-*.exe`
