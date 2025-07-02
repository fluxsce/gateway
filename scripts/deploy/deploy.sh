#!/bin/bash

# GoHub 自动部署脚本
# 用法: ./deploy.sh [options]

set -e

# 默认配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
TARGET_DIR="/opt/gohub"
APP_NAME="gohub"
BUILD_TARGET="linux"
BUILD_ARCH="amd64"
BACKUP_DIR=""
SKIP_BUILD=false
SKIP_BACKUP=false
FORCE_INSTALL=false
DRY_RUN=false
VERBOSE=false

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# 显示帮助
show_help() {
    cat << EOF
GoHub 自动部署脚本

用法: $0 [OPTIONS]

选项:
  -t, --target DIR        目标部署目录 (默认: $TARGET_DIR)
  -b, --backup DIR        备份目录 (默认: 自动生成)
  --os OS                 目标操作系统 (默认: $BUILD_TARGET)
  --arch ARCH             目标架构 (默认: $BUILD_ARCH)
  --skip-build            跳过编译步骤
  --skip-backup           跳过备份步骤
  --force                 强制安装，覆盖现有部署
  --dry-run               预览操作，不执行实际部署
  -v, --verbose           详细输出
  -h, --help              显示此帮助信息

示例:
  $0                                    # 标准部署
  $0 -t /home/app/gohub                # 部署到自定义目录
  $0 --os windows --arch amd64         # 为Windows编译
  $0 --skip-build --dry-run            # 预览部署（跳过编译）
  $0 --force -v                        # 强制部署并显示详细信息

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--target)
                TARGET_DIR="$2"
                shift 2
                ;;
            -b|--backup)
                BACKUP_DIR="$2"
                shift 2
                ;;
            --os)
                BUILD_TARGET="$2"
                shift 2
                ;;
            --arch)
                BUILD_ARCH="$2"
                shift 2
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-backup)
                SKIP_BACKUP=true
                shift
                ;;
            --force)
                FORCE_INSTALL=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查前置条件
check_prerequisites() {
    log_info "检查前置条件..."
    
    # 检查Go环境
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_debug "Go版本: $go_version"
    
    # 检查项目目录
    if [ ! -f "$PROJECT_DIR/go.mod" ]; then
        log_error "项目目录无效: $PROJECT_DIR"
        exit 1
    fi
    
    # 检查目标目录权限
    if [ ! -d "$TARGET_DIR" ]; then
        log_debug "目标目录不存在，将创建: $TARGET_DIR"
    else
        if [ ! -w "$TARGET_DIR" ]; then
            log_error "目标目录无写权限: $TARGET_DIR"
            exit 1
        fi
    fi
}

# 创建备份
create_backup() {
    if [ "$SKIP_BACKUP" = true ]; then
        log_info "跳过备份步骤"
        return 0
    fi
    
    if [ ! -d "$TARGET_DIR" ]; then
        log_info "目标目录不存在，跳过备份"
        return 0
    fi
    
    if [ -z "$BACKUP_DIR" ]; then
        BACKUP_DIR="${TARGET_DIR}_backup_$(date +%Y%m%d_%H%M%S)"
    fi
    
    log_info "创建备份: $TARGET_DIR -> $BACKUP_DIR"
    
    if [ "$DRY_RUN" = false ]; then
        cp -r "$TARGET_DIR" "$BACKUP_DIR"
        log_info "备份创建成功: $BACKUP_DIR"
    else
        log_info "[DRY RUN] 将创建备份: $BACKUP_DIR"
    fi
}

# 编译应用
build_application() {
    if [ "$SKIP_BUILD" = true ]; then
        log_info "跳过编译步骤"
        return 0
    fi
    
    log_info "编译应用..."
    log_debug "项目目录: $PROJECT_DIR"
    log_debug "目标系统: $BUILD_TARGET/$BUILD_ARCH"
    
    cd "$PROJECT_DIR"
    
    # 设置构建文件名
    local binary_name="$APP_NAME"
    if [ "$BUILD_TARGET" = "windows" ]; then
        binary_name="${APP_NAME}.exe"
    fi
    
    local build_output="$PROJECT_DIR/bin/${binary_name}"
    
    if [ "$DRY_RUN" = false ]; then
        # 创建bin目录
        mkdir -p "$PROJECT_DIR/bin"
        
        # 编译
        log_info "正在编译 ${BUILD_TARGET}/${BUILD_ARCH}..."
        GOOS="$BUILD_TARGET" GOARCH="$BUILD_ARCH" go build \
            -ldflags="-s -w" \
            -o "$build_output" \
            cmd/app/main.go
        
        if [ -f "$build_output" ]; then
            log_info "编译成功: $build_output"
            log_debug "文件大小: $(du -h "$build_output" | cut -f1)"
        else
            log_error "编译失败"
            exit 1
        fi
    else
        log_info "[DRY RUN] 将编译为: $build_output"
    fi
}

# 部署应用
deploy_application() {
    log_info "部署应用..."
    
    # 检查是否已存在部署
    if [ -d "$TARGET_DIR" ] && [ "$FORCE_INSTALL" = false ]; then
        log_error "目标目录已存在: $TARGET_DIR"
        log_error "使用 --force 参数强制覆盖"
        exit 1
    fi
    
    if [ "$DRY_RUN" = false ]; then
        # 创建目标目录结构
        log_info "创建目录结构..."
        mkdir -p "$TARGET_DIR"/{configs,logs,backup,scripts}
        
        # 复制可执行文件
        local binary_name="$APP_NAME"
        if [ "$BUILD_TARGET" = "windows" ]; then
            binary_name="${APP_NAME}.exe"
        fi
        
        local source_binary="$PROJECT_DIR/bin/$binary_name"
        local target_binary="$TARGET_DIR/$binary_name"
        
        if [ -f "$source_binary" ]; then
            log_info "复制可执行文件: $source_binary -> $target_binary"
            cp "$source_binary" "$target_binary"
            chmod +x "$target_binary"
        else
            log_error "可执行文件不存在: $source_binary"
            exit 1
        fi
        
        # 复制配置文件
        log_info "复制配置文件..."
        if [ -d "$PROJECT_DIR/configs" ]; then
            cp -r "$PROJECT_DIR/configs/"* "$TARGET_DIR/configs/"
            log_debug "配置文件复制完成"
        else
            log_warn "配置文件目录不存在: $PROJECT_DIR/configs"
        fi
        
        # 复制脚本文件
        log_info "复制管理脚本..."
        local script_files=("start.sh" "stop.sh" "restart.sh" "status.sh")
        for script in "${script_files[@]}"; do
            if [ -f "$SCRIPT_DIR/$script" ]; then
                cp "$SCRIPT_DIR/$script" "$TARGET_DIR/scripts/"
                chmod +x "$TARGET_DIR/scripts/$script"
                log_debug "复制脚本: $script"
            fi
        done
        
        # 设置权限
        log_info "设置权限..."
        chown -R $(whoami):$(whoami) "$TARGET_DIR" 2>/dev/null || true
        chmod -R 644 "$TARGET_DIR/configs/"* 2>/dev/null || true
        
        log_info "部署完成: $TARGET_DIR"
    else
        log_info "[DRY RUN] 将部署到: $TARGET_DIR"
        log_info "[DRY RUN] 将复制可执行文件、配置文件和脚本"
    fi
}

# 创建系统服务
create_systemd_service() {
    if [ "$BUILD_TARGET" != "linux" ]; then
        log_debug "非Linux系统，跳过systemd服务创建"
        return 0
    fi
    
    log_info "创建systemd服务..."
    
    local service_file="/etc/systemd/system/$APP_NAME.service"
    local service_content="[Unit]
Description=GoHub Application
Documentation=https://github.com/your-org/gohub
After=network.target

[Service]
Type=simple
User=$(whoami)
Group=$(whoami)
WorkingDirectory=$TARGET_DIR
ExecStart=$TARGET_DIR/$APP_NAME
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30

Environment=GOHUB_CONFIG_DIR=$TARGET_DIR/configs

Restart=always
RestartSec=5
StartLimitInterval=60s
StartLimitBurst=3

LimitNOFILE=65536

[Install]
WantedBy=multi-user.target"

    if [ "$DRY_RUN" = false ]; then
        if command -v systemctl >/dev/null 2>&1; then
            echo "$service_content" | sudo tee "$service_file" >/dev/null
            sudo systemctl daemon-reload
            log_info "systemd服务已创建: $service_file"
            log_info "使用以下命令管理服务:"
            log_info "  sudo systemctl enable $APP_NAME"
            log_info "  sudo systemctl start $APP_NAME"
        else
            log_warn "systemctl 不可用，跳过服务安装"
        fi
    else
        log_info "[DRY RUN] 将创建systemd服务: $service_file"
    fi
}

# 验证部署
verify_deployment() {
    log_info "验证部署..."
    
    local binary_path="$TARGET_DIR/$APP_NAME"
    if [ "$BUILD_TARGET" = "windows" ]; then
        binary_path="${binary_path}.exe"
    fi
    
    # 检查文件存在性
    local checks=0
    local failures=0
    
    # 检查可执行文件
    checks=$((checks + 1))
    if [ -f "$binary_path" ] && [ -x "$binary_path" ]; then
        log_debug "✓ 可执行文件: $binary_path"
    else
        log_error "✗ 可执行文件: $binary_path"
        failures=$((failures + 1))
    fi
    
    # 检查配置目录
    checks=$((checks + 1))
    if [ -d "$TARGET_DIR/configs" ]; then
        log_debug "✓ 配置目录: $TARGET_DIR/configs"
    else
        log_error "✗ 配置目录: $TARGET_DIR/configs"
        failures=$((failures + 1))
    fi
    
    # 检查管理脚本
    local script_files=("start.sh" "stop.sh" "restart.sh" "status.sh")
    for script in "${script_files[@]}"; do
        checks=$((checks + 1))
        if [ -f "$TARGET_DIR/scripts/$script" ] && [ -x "$TARGET_DIR/scripts/$script" ]; then
            log_debug "✓ 管理脚本: $script"
        else
            log_warn "✗ 管理脚本: $script"
            failures=$((failures + 1))
        fi
    done
    
    log_info "验证结果: $((checks - failures))/$checks 通过"
    
    if [ $failures -gt 0 ]; then
        log_warn "部署验证发现问题，请检查"
        return 1
    else
        log_info "部署验证通过"
        return 0
    fi
}

# 显示部署信息
show_deployment_info() {
    echo "========================================"
    echo "GoHub 部署完成"
    echo "========================================"
    echo "部署目录: $TARGET_DIR"
    echo "可执行文件: $TARGET_DIR/$APP_NAME"
    echo "配置目录: $TARGET_DIR/configs"
    echo "日志目录: $TARGET_DIR/logs"
    echo "管理脚本: $TARGET_DIR/scripts/"
    
    if [ -n "$BACKUP_DIR" ] && [ -d "$BACKUP_DIR" ]; then
        echo "备份目录: $BACKUP_DIR"
    fi
    
    echo ""
    echo "管理命令:"
    echo "  启动: $TARGET_DIR/scripts/start.sh"
    echo "  停止: $TARGET_DIR/scripts/stop.sh"
    echo "  重启: $TARGET_DIR/scripts/restart.sh"
    echo "  状态: $TARGET_DIR/scripts/status.sh"
    echo ""
    echo "快速启动:"
    echo "  cd $TARGET_DIR && ./scripts/start.sh"
    echo "========================================"
}

# 主函数
main() {
    log_info "GoHub 自动部署开始..."
    
    parse_args "$@"
    
    if [ "$DRY_RUN" = true ]; then
        log_warn "预览模式 - 不会执行实际操作"
    fi
    
    check_prerequisites
    create_backup
    build_application
    deploy_application
    create_systemd_service
    
    if verify_deployment; then
        show_deployment_info
        log_info "部署成功完成!"
    else
        log_error "部署存在问题，请检查"
        exit 1
    fi
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 