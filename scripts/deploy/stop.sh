#!/bin/bash

# GoHub 停止脚本
# 用法: ./stop.sh [force]

set -e

# 默认配置
APP_NAME="gohub"
APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="${APP_DIR}/logs"
PID_FILE="${LOG_DIR}/${APP_NAME}.pid"
FORCE_STOP="${1:-false}"

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

# 检查进程是否运行
check_process() {
    if [ ! -f "$PID_FILE" ]; then
        log_warn "PID file not found: $PID_FILE"
        return 1
    fi
    
    local pid=$(cat "$PID_FILE")
    if ! kill -0 "$pid" 2>/dev/null; then
        log_warn "Process not running (PID: $pid)"
        rm -f "$PID_FILE"
        return 1
    fi
    
    return 0
}

# 优雅停止
graceful_stop() {
    local pid=$(cat "$PID_FILE")
    log_info "Sending SIGTERM to $APP_NAME (PID: $pid)..."
    
    kill -TERM "$pid"
    
    # 等待进程停止
    local count=0
    local max_wait=30
    
    while kill -0 "$pid" 2>/dev/null && [ $count -lt $max_wait ]; do
        echo -n "."
        sleep 1
        count=$((count + 1))
    done
    echo
    
    if kill -0 "$pid" 2>/dev/null; then
        log_warn "Process still running after ${max_wait}s"
        return 1
    else
        log_info "Process stopped gracefully"
        rm -f "$PID_FILE"
        return 0
    fi
}

# 强制停止
force_stop() {
    local pid=$(cat "$PID_FILE")
    log_warn "Force stopping $APP_NAME (PID: $pid)..."
    
    kill -KILL "$pid" 2>/dev/null || true
    sleep 1
    
    if kill -0 "$pid" 2>/dev/null; then
        log_error "Failed to force stop process"
        return 1
    else
        log_info "Process force stopped"
        rm -f "$PID_FILE"
        return 0
    fi
}

# 停止所有相关进程
stop_all_processes() {
    log_info "Searching for all $APP_NAME processes..."
    
    local pids=$(pgrep -f "$APP_NAME" || true)
    if [ -z "$pids" ]; then
        log_info "No $APP_NAME processes found"
        return 0
    fi
    
    log_info "Found processes: $pids"
    
    for pid in $pids; do
        log_info "Stopping process $pid..."
        kill -TERM "$pid" 2>/dev/null || true
    done
    
    # 等待所有进程停止
    sleep 3
    
    # 检查是否还有进程运行
    local remaining=$(pgrep -f "$APP_NAME" || true)
    if [ -n "$remaining" ]; then
        log_warn "Some processes still running, force stopping: $remaining"
        for pid in $remaining; do
            kill -KILL "$pid" 2>/dev/null || true
        done
    fi
    
    rm -f "$PID_FILE"
    log_info "All processes stopped"
}

# 显示状态
show_status() {
    echo "----------------------------------------"
    echo "Service Status:"
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo "Status: Running (PID: $pid)"
        else
            echo "Status: Stopped (stale PID file removed)"
            rm -f "$PID_FILE"
        fi
    else
        echo "Status: Stopped"
    fi
    echo "----------------------------------------"
}

# 主函数
main() {
    log_info "Stopping $APP_NAME..."
    
    if ! check_process; then
        log_info "$APP_NAME is not running"
        show_status
        return 0
    fi
    
    case "$FORCE_STOP" in
        "force"|"--force"|"-f")
            if ! force_stop; then
                log_error "Failed to force stop $APP_NAME"
                exit 1
            fi
            ;;
        "all"|"--all"|"-a")
            stop_all_processes
            ;;
        *)
            if ! graceful_stop; then
                log_warn "Graceful stop failed, trying force stop..."
                if ! force_stop; then
                    log_error "Failed to stop $APP_NAME"
                    exit 1
                fi
            fi
            ;;
    esac
    
    show_status
    log_info "$APP_NAME stopped successfully"
}

# 显示帮助
show_help() {
    echo "Usage: $0 [option]"
    echo "Options:"
    echo "  (none)    Graceful stop with fallback to force stop"
    echo "  force     Force stop immediately"
    echo "  all       Stop all gohub processes"
    echo "  help      Show this help message"
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