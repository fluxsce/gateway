#!/bin/bash

# GoHub 配置目录测试脚本 (Linux/Unix版本)
# 验证GOHUB_CONFIG_DIR环境变量是否正确工作

set -e

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

log_header() {
    echo -e "${BLUE}[测试]${NC} $1"
}

echo
echo "=========================================="
echo "  GoHub 配置目录测试"
echo "=========================================="
echo

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "脚本目录: ${SCRIPT_DIR}"
echo "项目目录: ${PROJECT_DIR}"
echo

# 测试场景1：使用默认配置目录
log_header "使用默认配置目录"
unset GOHUB_CONFIG_DIR
TEST_DIR="${PROJECT_DIR}/configs"
echo "期望目录: ${TEST_DIR}"
if [ -d "${TEST_DIR}" ]; then
    log_info "✓ 默认配置目录存在"
    FILE_COUNT=$(find "${TEST_DIR}" -name "*.yaml" -type f | wc -l)
    echo "  发现 ${FILE_COUNT} 个YAML配置文件"
else
    log_error "✗ 默认配置目录不存在"
fi
echo

# 测试场景2：使用环境变量指定配置目录
log_header "使用环境变量指定配置目录"
export GOHUB_CONFIG_DIR="${PROJECT_DIR}/configs"
echo "设置环境变量: GOHUB_CONFIG_DIR=${GOHUB_CONFIG_DIR}"
echo "期望目录: ${GOHUB_CONFIG_DIR}"
if [ -d "${GOHUB_CONFIG_DIR}" ]; then
    log_info "✓ 环境变量配置目录存在"
    FILE_COUNT=$(find "${GOHUB_CONFIG_DIR}" -name "*.yaml" -type f | wc -l)
    echo "  发现 ${FILE_COUNT} 个YAML配置文件"
else
    log_error "✗ 环境变量配置目录不存在"
fi
echo

# 测试场景3：使用自定义配置目录
log_header "使用自定义配置目录（模拟生产环境）"
export GOHUB_CONFIG_DIR="/opt/gohub/configs"
echo "设置环境变量: GOHUB_CONFIG_DIR=${GOHUB_CONFIG_DIR}"
echo "期望目录: ${GOHUB_CONFIG_DIR}"
if [ -d "${GOHUB_CONFIG_DIR}" ]; then
    log_info "✓ 自定义配置目录存在"
    FILE_COUNT=$(find "${GOHUB_CONFIG_DIR}" -name "*.yaml" -type f | wc -l)
    echo "  发现 ${FILE_COUNT} 个YAML配置文件"
else
    log_warn "✗ 自定义配置目录不存在（这是正常的，除非您已经部署到生产环境）"
fi
echo

# 检查关键配置文件
log_header "检查关键配置文件"
export GOHUB_CONFIG_DIR="${PROJECT_DIR}/configs"
CONFIG_FILES=("app.yaml" "database.yaml" "logger.yaml" "web.yaml" "gateway.yaml")
for config_file in "${CONFIG_FILES[@]}"; do
    if [ -f "${GOHUB_CONFIG_DIR}/${config_file}" ]; then
        log_info "✓ ${config_file} 存在"
    else
        log_warn "✗ ${config_file} 不存在"
    fi
done
echo

# 配置目录优先级说明
log_header "配置目录优先级"
echo "1. GOHUB_CONFIG_DIR 环境变量指定的目录"
echo "2. ./configs （相对于程序启动目录）"
echo "3. . （程序启动目录）"
echo

# 修复建议
log_header "配置目录使用最佳实践"
echo "1. 开发环境：使用默认的 ./configs 目录"
echo "2. 生产环境：设置 GOHUB_CONFIG_DIR 环境变量"
echo "3. 容器部署：在容器中设置环境变量"
echo "4. Systemd服务：在服务配置文件中设置环境变量"
echo

# 示例部署命令
log_header "示例部署命令"
echo "# 开发环境启动"
echo "./gohub"
echo
echo "# 生产环境启动"
echo "export GOHUB_CONFIG_DIR=/opt/gohub/configs"
echo "./gohub"
echo
echo "# Systemd服务配置"
echo "Environment=GOHUB_CONFIG_DIR=/opt/gohub/configs"
echo

echo "=========================================="
echo "测试完成！"
echo "=========================================="
echo 