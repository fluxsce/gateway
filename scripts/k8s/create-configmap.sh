#!/bin/bash
# ============================================
# FLUX Gateway - 从本地配置文件创建 ConfigMap
# ============================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
FLUX Gateway - 从本地配置文件创建 Kubernetes ConfigMap

用法:
    $0 [选项]

选项:
    -n, --namespace NS    Kubernetes 命名空间 (默认: gateway)
    -c, --config-dir DIR  配置文件目录 (默认: ../../configs)
    -d, --dry-run         仅生成 YAML，不应用到集群
    -o, --output FILE     输出 YAML 到文件
    -h, --help            显示此帮助信息

示例:
    # 从默认配置目录创建 ConfigMap
    $0

    # 指定配置目录
    $0 --config-dir /path/to/configs

    # 仅生成 YAML 不应用
    $0 --dry-run

    # 生成 YAML 并保存到文件
    $0 --output configmap-generated.yaml

    # 指定命名空间
    $0 --namespace production

EOF
}

# 默认参数
NAMESPACE="gateway"
CONFIG_DIR=""
DRY_RUN=false
OUTPUT_FILE=""

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -c|--config-dir)
            CONFIG_DIR="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 确定配置目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -z "$CONFIG_DIR" ]]; then
    CONFIG_DIR="$(cd "$SCRIPT_DIR/../../configs" && pwd)"
fi

print_info "=========================================="
print_info "创建 Gateway ConfigMap"
print_info "=========================================="
print_info "命名空间: $NAMESPACE"
print_info "配置目录: $CONFIG_DIR"

# 检查配置目录是否存在
if [[ ! -d "$CONFIG_DIR" ]]; then
    print_error "配置目录不存在: $CONFIG_DIR"
    exit 1
fi

# 检查是否有配置文件
CONFIG_FILES=("$CONFIG_DIR"/*.yaml)
if [[ ! -e "${CONFIG_FILES[0]}" ]]; then
    print_error "配置目录中没有找到 .yaml 文件: $CONFIG_DIR"
    exit 1
fi

print_info "找到配置文件:"
for file in "$CONFIG_DIR"/*.yaml; do
    if [[ -f "$file" ]]; then
        print_info "  - $(basename "$file")"
    fi
done

# 检查 kubectl 是否可用
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl 未安装或不在 PATH 中"
    exit 1
fi

# 检查命名空间是否存在
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    print_warning "命名空间 '$NAMESPACE' 不存在"
    read -p "是否创建命名空间? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl create namespace "$NAMESPACE"
        print_success "命名空间 '$NAMESPACE' 已创建"
    else
        print_error "取消操作"
        exit 1
    fi
fi

# 删除旧的 ConfigMap（如果存在）
if kubectl get configmap gateway-config -n "$NAMESPACE" &> /dev/null; then
    print_warning "ConfigMap 'gateway-config' 已存在，将被删除并重新创建"
    kubectl delete configmap gateway-config -n "$NAMESPACE"
fi

# 创建 ConfigMap
print_info "创建 ConfigMap..."

if [[ "$DRY_RUN" == true ]] || [[ -n "$OUTPUT_FILE" ]]; then
    # 生成 YAML
    YAML_OUTPUT=$(kubectl create configmap gateway-config \
        --namespace="$NAMESPACE" \
        --from-file="$CONFIG_DIR" \
        --dry-run=client \
        -o yaml)
    
    # 添加标签
    YAML_OUTPUT=$(echo "$YAML_OUTPUT" | sed '/metadata:/a\
  labels:\
    app: gateway\
    component: config')
    
    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "$YAML_OUTPUT" > "$OUTPUT_FILE"
        print_success "ConfigMap YAML 已保存到: $OUTPUT_FILE"
    else
        echo "$YAML_OUTPUT"
    fi
    
    if [[ "$DRY_RUN" == true ]]; then
        print_info "Dry-run 模式，未应用到集群"
        exit 0
    fi
fi

# 应用到集群
kubectl create configmap gateway-config \
    --namespace="$NAMESPACE" \
    --from-file="$CONFIG_DIR"

# 添加标签
kubectl label configmap gateway-config \
    --namespace="$NAMESPACE" \
    app=gateway \
    component=config \
    --overwrite

if [[ $? -eq 0 ]]; then
    print_success "ConfigMap 创建成功！"
else
    print_error "ConfigMap 创建失败"
    exit 1
fi

# 显示 ConfigMap 信息
print_info ""
print_info "ConfigMap 详情:"
kubectl describe configmap gateway-config -n "$NAMESPACE"

print_info ""
print_success "=========================================="
print_success "ConfigMap 创建完成！"
print_success "=========================================="
print_info "命名空间: $NAMESPACE"
print_info "ConfigMap 名称: gateway-config"
print_info ""
print_info "查看 ConfigMap:"
print_info "  kubectl get configmap gateway-config -n $NAMESPACE"
print_info ""
print_info "查看 ConfigMap 内容:"
print_info "  kubectl describe configmap gateway-config -n $NAMESPACE"
print_info ""
print_info "编辑 ConfigMap:"
print_info "  kubectl edit configmap gateway-config -n $NAMESPACE"
print_info ""
print_info "删除 ConfigMap:"
print_info "  kubectl delete configmap gateway-config -n $NAMESPACE"

