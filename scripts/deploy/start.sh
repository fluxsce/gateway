#!/bin/bash

# GoHub 启动脚本
# 用法: ./start.sh [config_dir]

set -e

# 默认配置
APP_NAME="gohub"
APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CONFIG_DIR="${1:-${APP_DIR}/configs}"
LOG_DIR="${APP_DIR}/logs"
PID_FILE="${LOG_DIR}/${APP_NAME}.pid"
LOG_FILE="${LOG_DIR}/app.log"

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

# 检查是否已经运行
check_running() {
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_error "$APP_NAME is already running with PID $pid"
            exit 1
        else
            log_warn "Removing stale PID file"
            rm -f "$PID_FILE"
        fi
    fi
}

# 创建必要的目录
create_dirs() {
    mkdir -p "$LOG_DIR"
    mkdir -p "${APP_DIR}/backup"
}

# 检查可执行文件
check_executable() {
    local exe_file="${APP_DIR}/${APP_NAME}"
    if [ ! -f "$exe_file" ]; then
        log_error "Executable file not found: $exe_file"
        exit 1
    fi
    
    if [ ! -x "$exe_file" ]; then
        log_error "File is not executable: $exe_file"
        exit 1
    fi
}

# 检查配置文件
check_config() {
    if [ ! -d "$CONFIG_DIR" ]; then
        log_error "Config directory not found: $CONFIG_DIR"
        exit 1
    fi
    
    local required_configs=("app.yaml" "database.yaml" "logger.yaml")
    for config in "${required_configs[@]}"; do
        if [ ! -f "$CONFIG_DIR/$config" ]; then
            log_warn "Config file not found: $CONFIG_DIR/$config"
        fi
    done
}

# 启动应用
start_app() {
    log_info "Starting $APP_NAME..."
    log_info "App directory: $APP_DIR"
    log_info "Config directory: $CONFIG_DIR"
    log_info "Log file: $LOG_FILE"
    
    cd "$APP_DIR"
    
    # 设置环境变量
    export GOHUB_CONFIG_DIR="$CONFIG_DIR"
    
    # 启动应用
    nohup "./${APP_NAME}" > "$LOG_FILE" 2>&1 &
    local pid=$!
    
    # 保存 PID
    echo "$pid" > "$PID_FILE"
    
    # 检查启动是否成功
    sleep 2
    if kill -0 "$pid" 2>/dev/null; then
        log_info "$APP_NAME started successfully with PID $pid"
        log_info "Log file: $LOG_FILE"
        log_info "PID file: $PID_FILE"
    else
        log_error "$APP_NAME failed to start"
        rm -f "$PID_FILE"
        exit 1
    fi
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
            echo "Status: Stopped (stale PID file)"
        fi
    else
        echo "Status: Stopped"
    fi
    echo "----------------------------------------"
}

# 主函数
main() {
    log_info "GoHub deployment starting..."
    
    check_running
    create_dirs
    check_executable
    check_config
    start_app
    show_status
    
    log_info "Use './stop.sh' to stop the service"
    log_info "Use 'tail -f $LOG_FILE' to view logs"
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 