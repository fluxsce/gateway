# GoHub 部署脚本使用指南

本目录包含了 GoHub 应用的部署和管理脚本，适用于 Linux/Unix 系统。

## 脚本概览

### 核心管理脚本
- `start.sh` - 启动 GoHub 应用
- `stop.sh` - 停止 GoHub 应用
- `restart.sh` - 重启 GoHub 应用
- `status.sh` - 查看应用状态
- `deploy.sh` - 自动化部署脚本

## 使用说明

### 1. 设置执行权限（Linux/Unix）

```bash
chmod +x scripts/deploy/*.sh
```

### 2. 启动应用

```bash
# 使用默认配置启动
./scripts/deploy/start.sh

# 使用自定义配置目录启动
./scripts/deploy/start.sh /path/to/configs
```

### 3. 停止应用

```bash
# 优雅停止
./scripts/deploy/stop.sh

# 强制停止
./scripts/deploy/stop.sh force

# 停止所有相关进程
./scripts/deploy/stop.sh all
```

### 4. 重启应用

```bash
# 重启（使用默认配置）
./scripts/deploy/restart.sh

# 重启（使用自定义配置）
./scripts/deploy/restart.sh /path/to/configs
```

### 5. 查看状态

```bash
# 基本状态信息
./scripts/deploy/status.sh

# 详细状态信息（包含日志）
./scripts/deploy/status.sh --verbose
```

### 6. 自动化部署

```bash
# 标准部署到 /opt/gohub
./scripts/deploy/deploy.sh

# 部署到自定义目录
./scripts/deploy/deploy.sh -t /home/app/gohub

# 跨平台编译和部署
./scripts/deploy/deploy.sh --os windows --arch amd64

# 预览部署操作（不执行）
./scripts/deploy/deploy.sh --dry-run

# 强制部署（覆盖现有）
./scripts/deploy/deploy.sh --force

# 详细输出部署过程
./scripts/deploy/deploy.sh --verbose
```

## 高级用法

### 环境变量配置

```bash
# 设置配置文件目录
export GOHUB_CONFIG_DIR="/etc/gohub/configs"

# 启动应用
./scripts/deploy/start.sh
```

### 生产环境部署完整流程

```bash
# 1. 编译和部署
./scripts/deploy/deploy.sh -t /opt/gohub --verbose

# 2. 创建系统服务（如果支持）
sudo systemctl enable gohub
sudo systemctl start gohub

# 3. 验证部署
./scripts/deploy/status.sh --verbose

# 4. 查看日志
tail -f /opt/gohub/logs/app.log
```

### 开发环境快速部署

```bash
# 开发环境部署（跳过备份）
./scripts/deploy/deploy.sh \
  -t ./deployment \
  --skip-backup \
  --force

# 启动开发环境
cd deployment && ./scripts/start.sh
```

## 脚本选项详解

### start.sh 选项
- `config_dir` - 配置文件目录路径（可选）

### stop.sh 选项
- `force` - 强制停止进程
- `all` - 停止所有相关进程
- `help` - 显示帮助信息

### status.sh 选项
- `--verbose` / `-v` - 显示详细信息
- `--help` / `-h` - 显示帮助信息

### deploy.sh 选项
- `-t, --target DIR` - 目标部署目录
- `-b, --backup DIR` - 备份目录
- `--os OS` - 目标操作系统 (linux/windows/darwin)
- `--arch ARCH` - 目标架构 (amd64/arm64)
- `--skip-build` - 跳过编译步骤
- `--skip-backup` - 跳过备份步骤
- `--force` - 强制覆盖现有部署
- `--dry-run` - 预览操作，不执行
- `-v, --verbose` - 详细输出
- `-h, --help` - 显示帮助信息

## 常见问题

### Q: 如何在 Windows 上使用这些脚本？
A: 这些脚本是为 Linux/Unix 系统设计的。在 Windows 上，建议：
1. 使用 WSL（Windows Subsystem for Linux）
2. 使用 Git Bash
3. 参考脚本逻辑编写 PowerShell 版本

### Q: 权限错误怎么办？
A: 确保脚本有执行权限：
```bash
chmod +x scripts/deploy/*.sh
```

### Q: 如何自定义部署目录？
A: 使用 `-t` 参数指定目标目录：
```bash
./scripts/deploy/deploy.sh -t /your/custom/path
```

### Q: 如何查看部署日志？
A: 使用 verbose 模式：
```bash
./scripts/deploy/deploy.sh --verbose
```

### Q: 部署失败如何回滚？
A: 脚本会自动创建备份，可以手动恢复：
```bash
# 如果有备份目录
sudo rm -rf /opt/gohub
sudo mv /opt/gohub_backup_20231201_120000 /opt/gohub
```

## 脚本维护

### 日志文件位置
- 应用日志：`logs/app.log`
- PID 文件：`logs/gohub.pid`

### 配置文件检查
脚本会检查以下配置文件：
- `configs/app.yaml`
- `configs/database.yaml`
- `configs/logger.yaml`
- `configs/web.yaml`
- `configs/gateway.yaml`

### 健康检查端点
- Web 应用：`http://localhost:8080/health`
- 网关：`http://localhost:9090/health`

## 联系支持

如果遇到脚本问题，请提供：
1. 错误日志
2. 操作系统信息
3. 执行的命令
4. 环境变量设置

---

**更新时间**: 2024年12月23日  
**版本**: v1.0 