#!/bin/bash

# GoHub 重启脚本
# 用法: ./restart.sh [config_dir]

set -e

# 默认配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
START_SCRIPT="${SCRIPT_DIR}/start.sh"
STOP_SCRIPT="${SCRIPT_DIR}/stop.sh"
CONFIG_DIR="${1:-}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# 检查脚本是否存在
check_scripts() {
    if [ ! -f "$START_SCRIPT" ]; then
        log_error "Start script not found: $START_SCRIPT"
        exit 1
    fi
    
    if [ ! -f "$STOP_SCRIPT" ]; then
        log_error "Stop script not found: $STOP_SCRIPT"
        exit 1
    fi
    
    if [ ! -x "$START_SCRIPT" ]; then
        log_warn "Making start script executable: $START_SCRIPT"
        chmod +x "$START_SCRIPT"
    fi
    
    if [ ! -x "$STOP_SCRIPT" ]; then
        log_warn "Making stop script executable: $STOP_SCRIPT"
        chmod +x "$STOP_SCRIPT"
    fi
}

# 重启应用
restart_app() {
    log_info "Restarting GoHub..."
    
    # 停止应用
    log_info "Step 1: Stopping GoHub..."
    if ! "$STOP_SCRIPT"; then
        log_error "Failed to stop GoHub"
        exit 1
    fi
    
    # 等待一段时间确保完全停止
    log_info "Waiting for complete shutdown..."
    sleep 3
    
    # 启动应用
    log_info "Step 2: Starting GoHub..."
    if [ -n "$CONFIG_DIR" ]; then
        if ! "$START_SCRIPT" "$CONFIG_DIR"; then
            log_error "Failed to start GoHub"
            exit 1
        fi
    else
        if ! "$START_SCRIPT"; then
            log_error "Failed to start GoHub"
            exit 1
        fi
    fi
    
    log_info "GoHub restarted successfully"
}

# 显示帮助
show_help() {
    echo "Usage: $0 [config_dir]"
    echo "Parameters:"
    echo "  config_dir    Optional. Configuration directory path"
    echo ""
    echo "Examples:"
    echo "  $0                        # Use default config directory"
    echo "  $0 /etc/gohub/configs     # Use custom config directory"
}

# 主函数
main() {
    check_scripts
    restart_app
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    case "${1:-}" in
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            main "$@"
            ;;
    esac
fi 