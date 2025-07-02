#!/bin/bash

# GoHub 状态检查脚本
# 用法: ./status.sh [--verbose]

set -e

# 默认配置
APP_NAME="gohub"
APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="${APP_DIR}/logs"
PID_FILE="${LOG_DIR}/${APP_NAME}.pid"
LOG_FILE="${LOG_DIR}/app.log"
VERBOSE="${1:-false}"

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
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查进程状态
check_process_status() {
    local status="STOPPED"
    local pid=""
    local uptime=""
    
    if [ -f "$PID_FILE" ]; then
        pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            status="RUNNING"
            # 获取进程启动时间
            if command -v ps >/dev/null 2>&1; then
                uptime=$(ps -o etime= -p "$pid" 2>/dev/null | tr -d ' ' || echo "unknown")
            fi
        else
            status="DEAD"
        fi
    fi
    
    echo "Process Status: $status"
    if [ -n "$pid" ]; then
        echo "PID: $pid"
    fi
    if [ -n "$uptime" ]; then
        echo "Uptime: $uptime"
    fi
}

# 检查端口状态
check_port_status() {
    local config_file="${APP_DIR}/configs/app.yaml"
    local web_config_file="${APP_DIR}/configs/web.yaml"
    
    echo "Port Status:"
    
    # 检查Web端口
    if [ -f "$web_config_file" ]; then
        local web_port=$(grep -E "^\s*port\s*:" "$web_config_file" | head -1 | sed 's/.*:\s*//' | tr -d ' ')
        if [ -n "$web_port" ]; then
            if netstat -tlun 2>/dev/null | grep -q ":$web_port "; then
                echo "  Web Port $web_port: LISTENING"
            else
                echo "  Web Port $web_port: NOT LISTENING"
            fi
        fi
    fi
    
    # 检查网关端口
    local gateway_config_file="${APP_DIR}/configs/gateway.yaml"
    if [ -f "$gateway_config_file" ]; then
        local gateway_ports=$(grep -E "^\s*port\s*:" "$gateway_config_file" | sed 's/.*:\s*//' | tr -d ' ')
        for port in $gateway_ports; do
            if [ -n "$port" ]; then
                if netstat -tlun 2>/dev/null | grep -q ":$port "; then
                    echo "  Gateway Port $port: LISTENING"
                else
                    echo "  Gateway Port $port: NOT LISTENING"
                fi
            fi
        done
    fi
}

# 检查日志文件
check_log_status() {
    echo "Log Status:"
    
    if [ -f "$LOG_FILE" ]; then
        local log_size=$(du -h "$LOG_FILE" 2>/dev/null | cut -f1)
        local log_lines=$(wc -l < "$LOG_FILE" 2>/dev/null)
        local last_modified=$(stat -c %y "$LOG_FILE" 2>/dev/null | cut -d'.' -f1)
        
        echo "  Log File: $LOG_FILE"
        echo "  Size: $log_size"
        echo "  Lines: $log_lines"
        echo "  Last Modified: $last_modified"
        
        # 检查最近的错误
        local error_count=$(grep -i "error\|fail\|panic" "$LOG_FILE" 2>/dev/null | wc -l)
        echo "  Error Count: $error_count"
        
        if [ "$VERBOSE" = "--verbose" ] || [ "$VERBOSE" = "-v" ]; then
            echo "  Recent Errors:"
            grep -i "error\|fail\|panic" "$LOG_FILE" 2>/dev/null | tail -5 | sed 's/^/    /' || echo "    No recent errors"
        fi
    else
        echo "  Log File: NOT FOUND"
    fi
}

# 检查配置文件
check_config_status() {
    echo "Configuration Status:"
    
    local config_dir="${APP_DIR}/configs"
    if [ -d "$config_dir" ]; then
        echo "  Config Directory: $config_dir"
        
        local configs=("app.yaml" "database.yaml" "logger.yaml" "web.yaml" "gateway.yaml")
        for config in "${configs[@]}"; do
            if [ -f "$config_dir/$config" ]; then
                echo "  $config: FOUND"
            else
                echo "  $config: MISSING"
            fi
        done
    else
        echo "  Config Directory: NOT FOUND"
    fi
}

# 检查系统资源
check_system_resources() {
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            echo "Resource Usage:"
            
            # CPU和内存使用率
            if command -v ps >/dev/null 2>&1; then
                local cpu_mem=$(ps -o pid,pcpu,pmem,rss,vsz -p "$pid" 2>/dev/null | tail -1)
                if [ -n "$cpu_mem" ]; then
                    echo "  CPU: $(echo $cpu_mem | awk '{print $2}')%"
                    echo "  Memory: $(echo $cpu_mem | awk '{print $3}')%"
                    echo "  RSS: $(echo $cpu_mem | awk '{print $4}')KB"
                    echo "  VSZ: $(echo $cpu_mem | awk '{print $5}')KB"
                fi
            fi
            
            # 文件句柄
            if command -v lsof >/dev/null 2>&1; then
                local fd_count=$(lsof -p "$pid" 2>/dev/null | wc -l)
                echo "  File Descriptors: $fd_count"
            fi
            
            # 网络连接
            if command -v netstat >/dev/null 2>&1; then
                local conn_count=$(netstat -pan 2>/dev/null | grep "$pid/" | wc -l)
                echo "  Network Connections: $conn_count"
            fi
        fi
    fi
}

# 执行健康检查
check_health() {
    echo "Health Check:"
    
    # 检查HTTP端点
    local web_config_file="${APP_DIR}/configs/web.yaml"
    if [ -f "$web_config_file" ]; then
        local web_port=$(grep -E "^\s*port\s*:" "$web_config_file" | head -1 | sed 's/.*:\s*//' | tr -d ' ')
        if [ -n "$web_port" ]; then
            if command -v curl >/dev/null 2>&1; then
                if curl -f -s "http://localhost:$web_port/health" >/dev/null 2>&1; then
                    echo "  Web Health: OK"
                else
                    echo "  Web Health: FAILED"
                fi
            else
                echo "  Web Health: SKIP (curl not available)"
            fi
        fi
    fi
    
    # 检查数据库连接（通过日志）
    if [ -f "$LOG_FILE" ]; then
        if grep -q "数据库连接成功" "$LOG_FILE" 2>/dev/null; then
            echo "  Database: CONNECTED"
        else
            echo "  Database: UNKNOWN"
        fi
    fi
}

# 显示系统信息
show_system_info() {
    if [ "$VERBOSE" = "--verbose" ] || [ "$VERBOSE" = "-v" ]; then
        echo "System Information:"
        echo "  OS: $(uname -s -r)"
        echo "  Hostname: $(hostname)"
        echo "  Load Average: $(uptime | awk -F'load average:' '{print $2}')"
        echo "  Memory: $(free -h 2>/dev/null | grep Mem: | awk '{print $3"/"$2}' || echo 'N/A')"
        echo "  Disk: $(df -h "$APP_DIR" 2>/dev/null | tail -1 | awk '{print $4" available"}' || echo 'N/A')"
    fi
}

# 显示最近的日志
show_recent_logs() {
    if [ "$VERBOSE" = "--verbose" ] || [ "$VERBOSE" = "-v" ]; then
        echo "Recent Logs (last 10 lines):"
        if [ -f "$LOG_FILE" ]; then
            tail -10 "$LOG_FILE" | sed 's/^/  /'
        else
            echo "  No log file found"
        fi
    fi
}

# 主函数
main() {
    echo "========================================"
    echo "GoHub Status Report"
    echo "========================================"
    echo "Time: $(date)"
    echo "Directory: $APP_DIR"
    echo ""
    
    check_process_status
    echo ""
    
    check_port_status
    echo ""
    
    check_log_status
    echo ""
    
    check_config_status
    echo ""
    
    check_system_resources
    echo ""
    
    check_health
    echo ""
    
    show_system_info
    
    show_recent_logs
    
    echo "========================================"
}

# 显示帮助
show_help() {
    echo "Usage: $0 [option]"
    echo "Options:"
    echo "  (none)      Show basic status"
    echo "  --verbose   Show detailed status with logs"
    echo "  -v          Same as --verbose"
    echo "  --help      Show this help message"
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