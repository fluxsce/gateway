#!/bin/bash

# Gateway Linux服务安装脚本
# 版本: 2.3
# 功能: 智能检测可执行文件并安装为systemd服务
# 更新: 使用检测到的应用目录作为安装位置，配置和日志使用相对路径

# 设置错误处理（即使错误也不立即退出）
set +e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量 - 将在安装过程中基于检测到的APP_DIR进行更新
SERVICE_NAME="gateway"
SERVICE_USER="gateway"
SERVICE_GROUP="gateway"
INSTALL_DIR="/opt/gateway"  # 默认值，将被检测到的应用目录覆盖
LOG_DIR=""                # 将设置为 $APP_DIR/logs
CONFIG_DIR=""             # 将设置为 $APP_DIR/configs
WORK_DIR=""               # 将设置为 $APP_DIR
ORACLE_VERSION=false
DEBUG_MODE=false
USE_SYSTEM_PATHS=false    # 是否使用系统路径 (/opt, /var/log, etc.)

# 函数：打印带颜色的消息
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

# 函数：显示帮助
show_help() {
    echo "Gateway Linux 服务安装脚本 v2.3"
    echo
    echo "用法: $0 [oracle] [OPTIONS]"
    echo
    echo "参数:"
    echo "  oracle                安装Oracle版本服务"
    echo
    echo "选项:"
    echo "  -d, --dir DIR        应用程序目录 (默认: 自动检测)"
    echo "  -c, --config DIR     配置文件源目录 (默认: 自动检测)"
    echo "  -s, --system         使用系统标准路径 (/opt/gateway, /var/log/gateway)"
    echo "  -h, --help           显示帮助信息"
    echo
    echo "示例:"
    echo "  $0                   # 安装标准版本服务，使用检测到的目录"
    echo "  $0 oracle            # 安装Oracle版本服务"
    echo "  $0 -d /opt/gateway     # 指定安装目录"
    echo "  $0 -s                # 使用系统标准目录结构"
    echo
    echo "注意:"
    echo "  - 默认将安装到检测到的应用目录中"
    echo "  - 使用 -s 选项可以安装到标准系统目录"
    echo "  - 配置文件和日志将存放在应用目录下相应子目录中"
    echo
}

# 函数：检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要root权限运行"
        echo "请使用: sudo $0"
        log_debug "按Enter键继续以调试模式运行（功能受限）..."
        read
        DEBUG_MODE=true
        log_warn "继续运行调试模式..."
    else
        log_info "✓ root权限检查通过"
        DEBUG_MODE=false
    fi
}

# 函数：检查systemd是否可用
check_systemd() {
    if ! command -v systemctl &> /dev/null; then
        log_error "systemd未找到，此脚本仅支持systemd系统"
        exit 1
    fi
}

# 函数：智能检测可执行文件
detect_executable() {
    # 如果已经通过参数指定了APP_DIR，则直接在其中查找
    if [[ ! -z "$APP_DIR" ]]; then
        log_debug "使用指定的应用程序目录: $APP_DIR"
        find_exe_in_dir "$APP_DIR"
        return $?
    fi
    
    local script_dir=$(dirname "$(readlink -f "$0")")
    local project_root=$(dirname "$(dirname "$script_dir")")
    
    log_info "正在智能检测可执行文件..."
    log_debug "脚本目录: $script_dir"
    log_debug "项目根目录: $project_root"
    
    # 方案1: 检查脚本上级目录（适用于源码目录中的脚本）
    local candidate_dir="$script_dir/.."
    log_debug "检查位置1: $candidate_dir"
    find_exe_in_dir "$candidate_dir"
    if [[ $? -eq 0 ]]; then
        APP_DIR="$candidate_dir"
        return 0
    fi
    
    # 方案2: 检查项目根目录（适用于源码目录中的脚本，scripts/deploy目录）
    log_debug "检查位置2: $project_root"
    find_exe_in_dir "$project_root"
    if [[ $? -eq 0 ]]; then
        APP_DIR="$project_root"
        return 0
    fi
    
    # 方案3: 检查脚本当前目录（适用于脚本与程序在同一目录）
    log_debug "检查位置3: $script_dir"
    find_exe_in_dir "$script_dir"
    if [[ $? -eq 0 ]]; then
        APP_DIR="$script_dir"
        return 0
    fi
    
    # 如果都没找到，提示用户
    log_error "无法自动检测应用程序目录"
    echo
    echo "脚本目录: $script_dir"
    echo
    echo "已检查以下位置："
    echo "  1. $candidate_dir"
    echo "  2. $project_root"
    echo "  3. $script_dir"
    echo
    
    log_debug "列出各检测位置的内容："
    echo
    log_debug "检查位置1: $candidate_dir"
    if [[ -d "$candidate_dir" ]]; then
        ls -la "$candidate_dir" | grep -i "gateway" || echo "  - 未找到Gateway相关文件"
    else
        echo "  - 目录不存在"
    fi
    echo
    log_debug "检查位置2: $project_root"
    if [[ -d "$project_root" ]]; then
        ls -la "$project_root" | grep -i "gateway" || echo "  - 未找到Gateway相关文件"
    else
        echo "  - 目录不存在"
    fi
    echo
    log_debug "检查位置3: $script_dir"
    if [[ -d "$script_dir" ]]; then
        ls -la "$script_dir" | grep -i "gateway" || echo "  - 未找到Gateway相关文件"
    else
        echo "  - 目录不存在"
    fi
    echo
    
    echo "请使用 -d 参数指定应用程序目录："
    echo "  $0 -d /path/to/gateway"
    echo
    log_debug "继续执行并使用脚本目录作为默认值..."
    read -p "按Enter继续..."
    return 1
}

# 函数：在指定目录中查找可执行文件
find_exe_in_dir() {
    local check_dir="$1"
    check_dir=$(readlink -f "$check_dir")
    
    log_debug "在 $check_dir 中查找可执行文件..."
    
    # 首先检查Oracle版本的文件（如果指定了oracle参数）
    if [[ "$ORACLE_VERSION" == true ]]; then
        if [[ -f "$check_dir/gateway-oracle" && -x "$check_dir/gateway-oracle" ]]; then
            EXECUTABLE_PATH="$check_dir/gateway-oracle"
            log_info "找到Oracle版本可执行文件: $EXECUTABLE_PATH"
            return 0
        elif [[ -f "$check_dir/gateway-linux-oracle-amd64" && -x "$check_dir/gateway-linux-oracle-amd64" ]]; then
            EXECUTABLE_PATH="$check_dir/gateway-linux-oracle-amd64"
            log_info "找到Oracle版本可执行文件: $EXECUTABLE_PATH"
            return 0
        elif [[ -f "$check_dir/gateway-centos7-oracle-amd64" && -x "$check_dir/gateway-centos7-oracle-amd64" ]]; then
            EXECUTABLE_PATH="$check_dir/gateway-centos7-oracle-amd64"
            log_info "找到Oracle版本可执行文件: $EXECUTABLE_PATH"
            return 0
        fi
    fi
    
    # 如果没有找到Oracle版本文件，检查标准版本文件
    if [[ -f "$check_dir/gateway" && -x "$check_dir/gateway" ]]; then
        EXECUTABLE_PATH="$check_dir/gateway"
        log_info "找到标准版本可执行文件: $EXECUTABLE_PATH"
        return 0
    elif [[ -f "$check_dir/gateway-linux-amd64" && -x "$check_dir/gateway-linux-amd64" ]]; then
        EXECUTABLE_PATH="$check_dir/gateway-linux-amd64"
        log_info "找到标准版本可执行文件: $EXECUTABLE_PATH"
        return 0
    elif [[ -f "$check_dir/gateway-centos7-amd64" && -x "$check_dir/gateway-centos7-amd64" ]]; then
        EXECUTABLE_PATH="$check_dir/gateway-centos7-amd64"
        log_info "找到标准版本可执行文件: $EXECUTABLE_PATH"
        return 0
    fi
    
    # 如果还是没有找到，尝试自动检测Oracle版本
    if [[ -f "$check_dir/gateway-linux-oracle-amd64" && -x "$check_dir/gateway-linux-oracle-amd64" ]]; then
        EXECUTABLE_PATH="$check_dir/gateway-linux-oracle-amd64"
        ORACLE_VERSION=true
        SERVICE_NAME="gateway"
        log_info "自动检测到Oracle版本可执行文件"
        return 0
    elif [[ -f "$check_dir/gateway-oracle" && -x "$check_dir/gateway-oracle" ]]; then
        EXECUTABLE_PATH="$check_dir/gateway-oracle"
        ORACLE_VERSION=true
        SERVICE_NAME="gateway"
        log_info "自动检测到Oracle版本可执行文件"
        return 0
    elif [[ -f "$check_dir/gateway-centos7-oracle-amd64" && -x "$check_dir/gateway-centos7-oracle-amd64" ]]; then
        EXECUTABLE_PATH="$check_dir/gateway-centos7-oracle-amd64"
        ORACLE_VERSION=true
        SERVICE_NAME="gateway"
        log_info "自动检测到Oracle版本可执行文件"
        return 0
    fi
    
    # 搜索模式：如果上面没有找到，则使用原来的模式继续搜索
    local patterns=(
        "gateway"
        "gateway-*"
        "Gateway"
        "Gateway-*"
        "main"
    )
    
    # 搜索目录
    local search_dirs=(
        "$check_dir"
        "$check_dir/dist"
        "$check_dir/build"
        "$check_dir/bin"
    )
    
    # 搜索可执行文件
    for dir in "${search_dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            log_debug "搜索目录: $dir"
            for pattern in "${patterns[@]}"; do
                for file in "$dir"/$pattern; do
                    if [[ -f "$file" && -x "$file" ]]; then
                        # 排除脚本文件
                        if [[ "$file" != *.sh && "$file" != *.bat && "$file" != *.cmd ]]; then
                            EXECUTABLE_PATH="$file"
                            log_info "检测到可执行文件: $EXECUTABLE_PATH"
                            return 0
                        fi
                    fi
                done
            done
        fi
    done
    
    # 没找到相关文件
    return 1
}

# 函数：创建系统用户
create_user() {
    if ! id "$SERVICE_USER" &>/dev/null; then
        log_info "创建系统用户: $SERVICE_USER"
        useradd --system --no-create-home --shell /bin/false "$SERVICE_USER"
        log_info "系统用户 $SERVICE_USER 创建成功"
    else
        log_info "系统用户 $SERVICE_USER 已存在"
    fi
}

# 函数：创建目录结构
create_directories() {
    log_info "创建目录结构..."
    
    # 创建安装目录
    mkdir -p "$INSTALL_DIR"
    
    # 创建配置目录 (在安装目录下)
    mkdir -p "$CONFIG_DIR"
    
    # 创建日志目录
    mkdir -p "$LOG_DIR"
    
    # 创建工作目录
    mkdir -p "$WORK_DIR"
    
    # 设置目录权限
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$LOG_DIR"
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$CONFIG_DIR"
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$WORK_DIR"
    
    chmod 755 "$INSTALL_DIR"
    chmod 755 "$LOG_DIR"
    chmod 755 "$CONFIG_DIR"
    chmod 755 "$WORK_DIR"
    
    log_info "目录结构创建完成"
}

# 函数：安装可执行文件
install_executable() {
    log_info "安装可执行文件..."
    
    local exe_name=$(basename "$EXECUTABLE_PATH")
    local target_path="$INSTALL_DIR/$exe_name"
    
    # 如果可执行文件已经在目标位置，不需要复制
    if [[ "$EXECUTABLE_PATH" == "$target_path" ]]; then
        log_info "可执行文件已在正确位置，无需复制"
        INSTALLED_EXE="$EXECUTABLE_PATH"
        return
    fi
    
    cp "$EXECUTABLE_PATH" "$target_path"
    chmod 755 "$target_path"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$target_path"
    
    INSTALLED_EXE="$target_path"
    log_info "可执行文件安装完成: $INSTALLED_EXE"
}

# 函数：复制配置文件
copy_configs() {
    log_info "复制配置文件..."
    
    # 如果通过参数指定了配置目录，使用指定的配置目录
    if [[ ! -z "$SOURCE_CONFIG_DIR" ]]; then
        if [[ -d "$SOURCE_CONFIG_DIR" ]]; then
            cp -r "$SOURCE_CONFIG_DIR"/* "$CONFIG_DIR/"
            chown -R "$SERVICE_USER:$SERVICE_GROUP" "$CONFIG_DIR"
            log_info "从指定目录复制配置文件完成"
            return
        else
            log_warn "指定的配置目录 $SOURCE_CONFIG_DIR 不存在，尝试查找默认位置"
        fi
    fi
    
    local script_dir=$(dirname "$(readlink -f "$0")")
    local project_root=$(dirname "$(dirname "$script_dir")")
    local config_source="$project_root/configs"
    
    if [[ -d "$config_source" ]]; then
        cp -r "$config_source"/* "$CONFIG_DIR/"
        chown -R "$SERVICE_USER:$SERVICE_GROUP" "$CONFIG_DIR"
        log_info "配置文件复制完成"
    elif [[ -d "$APP_DIR/configs" ]]; then
        cp -r "$APP_DIR/configs"/* "$CONFIG_DIR/"
        chown -R "$SERVICE_USER:$SERVICE_GROUP" "$CONFIG_DIR"
        log_info "配置文件复制完成"
    else
        log_warn "配置文件目录不存在，跳过复制"
        log_debug "尝试查找的位置:"
        log_debug "  - $config_source"
        log_debug "  - $APP_DIR/configs"
    fi
}

# 函数：创建systemd服务文件
create_systemd_service() {
    log_info "创建systemd服务文件..."
    
    cat > "/etc/systemd/system/$SERVICE_NAME.service" << EOF
[Unit]
Description=Gateway Gateway and Management Service
Documentation=https://github.com/your-org/gateway
After=network.target remote-fs.target nss-lookup.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
WorkingDirectory=$WORK_DIR
ExecStart=$INSTALLED_EXE --service --config=$CONFIG_DIR
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=10
StartLimitInterval=60
StartLimitBurst=3

# 环境变量
Environment=GATEWAY_CONFIG_DIR=$CONFIG_DIR
Environment=GATEWAY_LOG_DIR=$LOG_DIR

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$LOG_DIR $WORK_DIR $CONFIG_DIR

# 日志设置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

# 资源限制
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

    log_info "systemd服务文件创建完成"
}

# 函数：启用并启动服务
enable_and_start_service() {
    log_info "重新加载systemd配置..."
    systemctl daemon-reload
    
    log_info "启用服务..."
    systemctl enable "$SERVICE_NAME"
    
    # 询问是否立即启动服务
    read -p "是否立即启动服务？(Y/n): " START_NOW
    if [[ "$START_NOW" == "n" || "$START_NOW" == "N" ]]; then
        log_info "服务已安装但未启动"
        return
    fi
    
    log_info "启动服务..."
    systemctl start "$SERVICE_NAME"
    
    # 等待服务启动
    sleep 3
    
    # 检查服务状态
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_info "服务启动成功"
    else
        log_error "服务启动失败"
        log_error "查看详细日志: sudo journalctl -u $SERVICE_NAME -f"
        return 1
    fi
}

# 函数：显示服务状态
show_service_status() {
    log_info "服务状态:"
    systemctl status "$SERVICE_NAME" --no-pager
    
    echo ""
    log_info "服务管理命令:"
    echo "  启动服务: sudo systemctl start $SERVICE_NAME"
    echo "  停止服务: sudo systemctl stop $SERVICE_NAME"
    echo "  重启服务: sudo systemctl restart $SERVICE_NAME"
    echo "  重新加载配置: sudo systemctl reload $SERVICE_NAME"
    echo "  查看状态: sudo systemctl status $SERVICE_NAME"
    echo "  查看日志: sudo journalctl -u $SERVICE_NAME -f"
    echo "  禁用服务: sudo systemctl disable $SERVICE_NAME"
    
    echo ""
    log_info "配置文件位置: $CONFIG_DIR"
    log_info "日志文件位置: $LOG_DIR"
    log_info "工作目录位置: $WORK_DIR"
    if [[ "$ORACLE_VERSION" == true ]]; then
        log_info "服务类型: Oracle版本"
    else
        log_info "服务类型: 标准版本"
    fi
}

# 函数：清理旧安装
cleanup_old_installation() {
    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        log_info "停止现有服务..."
        systemctl stop "$SERVICE_NAME"
    fi
    
    if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
        log_info "禁用现有服务..."
        systemctl disable "$SERVICE_NAME"
    fi
}

# 主函数
main() {
    echo "Gateway Linux服务安装脚本 v2.3"
    echo "================================="
    echo ""
    
    # 解析命令行参数
    APP_DIR=""
    SOURCE_CONFIG_DIR=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            oracle)
                ORACLE_VERSION=true
                shift
                ;;
            -d|--dir)
                APP_DIR="$2"
                shift 2
                ;;
            -c|--config)
                SOURCE_CONFIG_DIR="$2"
                shift 2
                ;;
            -s|--system)
                USE_SYSTEM_PATHS=true
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
    
    log_debug "配置信息:"
    log_debug "  Oracle版本: $ORACLE_VERSION"
    log_debug "  指定应用目录: $APP_DIR"
    log_debug "  指定配置目录: $SOURCE_CONFIG_DIR"
    log_debug "  使用系统路径: $USE_SYSTEM_PATHS"
    
    # 检查权限
    check_root
    
    # 检查systemd
    check_systemd
    
    # 检测可执行文件
    detect_executable
    if [[ $? -ne 0 && -z "$EXECUTABLE_PATH" ]]; then
        log_error "可执行文件检测失败，无法继续安装"
        exit 1
    fi
    
    # 设置安装目录和其他路径
    if [[ "$USE_SYSTEM_PATHS" == true ]]; then
        # 使用系统标准路径
        INSTALL_DIR="/opt/gateway"
        LOG_DIR="/var/log/gateway"
        CONFIG_DIR="$INSTALL_DIR/configs"
        WORK_DIR="/var/lib/gateway"
    else
        # 使用检测到的应用目录作为基础
        INSTALL_DIR="$(dirname "$EXECUTABLE_PATH")"
        APP_DIR="$INSTALL_DIR"
        CONFIG_DIR="$APP_DIR/configs"
        LOG_DIR="$APP_DIR/logs"
        WORK_DIR="$APP_DIR"
    fi
    
    log_debug "使用配置目录: $CONFIG_DIR"
    log_debug "使用日志目录: $LOG_DIR"
    
    # 如果没有指定配置目录，使用可执行文件目录下的configs
    if [[ -z "$SOURCE_CONFIG_DIR" ]]; then
        SOURCE_CONFIG_DIR="$(dirname "$EXECUTABLE_PATH")/configs"
        if [[ ! -d "$SOURCE_CONFIG_DIR" ]]; then
            log_debug "默认配置目录不存在: $SOURCE_CONFIG_DIR"
            SOURCE_CONFIG_DIR=""
        fi
    fi
    
    echo ""
    echo "============================================"
    echo "  Gateway Linux 服务安装配置"
    echo "============================================"
    echo ""
    echo "服务名称: $SERVICE_NAME"
    echo "可执行文件: $EXECUTABLE_PATH"
    echo "安装目录: $INSTALL_DIR"
    echo "配置源目录: ${SOURCE_CONFIG_DIR:-自动检测}"
    echo "配置目标目录: $CONFIG_DIR"
    echo "日志目录: $LOG_DIR"
    echo "工作目录: $WORK_DIR"
    echo "Oracle版本: $ORACLE_VERSION"
    echo "使用系统路径: $USE_SYSTEM_PATHS"
    echo ""
    
    # 询问是否继续
    read -p "是否继续安装？(Y/n): " CONTINUE
    if [[ "$CONTINUE" == "n" || "$CONTINUE" == "N" ]]; then
        log_info "安装已取消"
        exit 0
    fi
    
    # 清理旧安装
    cleanup_old_installation
    
    # 创建系统用户
    create_user
    
    # 创建目录结构
    create_directories
    
    # 安装可执行文件
    install_executable
    
    # 复制配置文件
    copy_configs
    
    # 创建systemd服务
    create_systemd_service
    
    # 启用并启动服务
    enable_and_start_service
    
    # 显示服务状态
    show_service_status
    
    echo ""
    log_info "Gateway服务安装完成！"
    echo "============================================"
    echo "  安装过程完成！"
    echo "============================================"
    echo ""
    log_debug "最终状态总结："
    log_debug "  服务名称: $SERVICE_NAME"
    log_debug "  可执行文件: $INSTALLED_EXE"
    log_debug "  配置目录: $CONFIG_DIR"
    log_debug "  日志目录: $LOG_DIR"
    log_debug "  调试模式: $DEBUG_MODE"
    log_debug "  Oracle版本: $ORACLE_VERSION"
    echo ""
    log_debug "验证服务状态..."
    if [[ "$DEBUG_MODE" == false ]]; then
        systemctl status "$SERVICE_NAME" | grep "Active:"
    else
        log_debug "跳过服务状态检查（权限限制）"
    fi
    echo ""
    log_info "如需查看详细日志，请运行："
    echo "  sudo journalctl -u $SERVICE_NAME -f"
}

# 如果脚本被直接执行，则运行主函数
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 