#!/bin/bash
# ============================================
# FLUX Gateway - Kubernetes 部署脚本
# ============================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
NAMESPACE="gateway"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 配置文件目录（相对于脚本目录）
CONFIG_DIR="$SCRIPT_DIR/../../configs"

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
FLUX Gateway Kubernetes 部署脚本

用法:
    $0 [命令] [选项]

命令:
    install         安装 Gateway 到 Kubernetes
    uninstall       卸载 Gateway
    upgrade         升级 Gateway
    status          查看部署状态
    logs            查看日志
    debug           诊断 Pod 问题（查看初始化容器日志）
    restart         重启 Gateway
    help            显示此帮助信息

选项:
    -n, --namespace NS    指定命名空间 (默认: gateway)
    -c, --config-dir DIR  配置文件目录 (默认: ../../configs)
    -f, --force           强制执行操作
    -h, --help            显示此帮助信息

示例:
    # 安装 Gateway（使用默认配置目录）
    $0 install

    # 安装 Gateway（指定配置目录）
    $0 install --config-dir /path/to/configs

    # 查看状态
    $0 status

    # 查看日志
    $0 logs

    # 升级 Gateway
    $0 upgrade

    # 卸载 Gateway
    $0 uninstall

EOF
}

# 检查 kubectl 是否安装
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl 未安装，请先安装 kubectl"
        exit 1
    fi
}

# 检查集群连接
check_cluster() {
    if ! kubectl cluster-info &> /dev/null; then
        print_error "无法连接到 Kubernetes 集群"
        exit 1
    fi
    print_success "Kubernetes 集群连接正常"
}

# 安装 Gateway
install_gateway() {
    print_info "开始安装 FLUX Gateway..."
    
    # 创建命名空间
    print_info "创建命名空间: $NAMESPACE"
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        kubectl create namespace "$NAMESPACE"
        print_success "命名空间 $NAMESPACE 已创建"
    else
        print_warning "命名空间 $NAMESPACE 已存在"
    fi
    
    # 创建 ConfigMap（从本地配置文件）
    print_info "创建 ConfigMap..."
    if [[ -f "$SCRIPT_DIR/create-configmap.sh" ]]; then
        # 检查配置文件目录是否存在
        if [[ ! -d "$CONFIG_DIR" ]]; then
            print_error "配置文件目录不存在: $CONFIG_DIR"
            exit 1
        fi
        bash "$SCRIPT_DIR/create-configmap.sh" --namespace "$NAMESPACE" --config-dir "$CONFIG_DIR"
    else
        print_error "create-configmap.sh 脚本不存在"
        exit 1
    fi
    
    # 应用 Deployment
    print_info "部署应用..."
    kubectl apply -f "$SCRIPT_DIR/deployment.yaml" -n "$NAMESPACE"
    
    # 强制重启 Pod（即使镜像没有变化）
    print_info "重启 Pod 以应用最新配置..."
    kubectl rollout restart deployment/gateway -n "$NAMESPACE" 2>/dev/null || true
    
    # 应用 Service
    print_info "创建服务..."
    kubectl apply -f "$SCRIPT_DIR/service.yaml" -n "$NAMESPACE"
    
    # 应用 Ingress（可选）
    if [[ -f "$SCRIPT_DIR/ingress.yaml" ]]; then
        print_info "配置 Ingress..."
        kubectl apply -f "$SCRIPT_DIR/ingress.yaml" -n "$NAMESPACE"
    fi
    
    # 等待 Deployment 就绪
    print_info "等待 Deployment 就绪..."
    kubectl rollout status deployment/gateway -n "$NAMESPACE" --timeout=180s || {
        print_warning "Deployment 未能在 180 秒内就绪，正在诊断问题..."
        print_info ""
        print_info "Pod 状态："
        kubectl get pods -n "$NAMESPACE" -l app=gateway
        print_info ""
        
        # 获取第一个 Pod 名称
        POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l app=gateway -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
        
        if [[ -n "$POD_NAME" ]]; then
            print_info "Pod 详细信息："
            kubectl describe pod "$POD_NAME" -n "$NAMESPACE" | tail -30
            print_info ""
            
            print_info "Pod 日志："
            kubectl logs "$POD_NAME" -n "$NAMESPACE" --tail=30 2>/dev/null || print_warning "无法获取 Pod 日志"
            
            print_info ""
            print_warning "请检查以上信息，常见问题："
            print_warning "  1. 镜像是否可以正常拉取"
            print_warning "  2. ConfigMap 配置是否正确"
            print_warning "  3. 数据库连接配置是否正确"
            print_warning "  4. 资源限制是否合理"
        fi
    }
    
    print_success "=========================================="
    print_success "FLUX Gateway 安装完成！"
    print_success "=========================================="
    print_info ""
    
    # 获取访问信息
    print_info "访问信息:"
    print_info "----------------------------------------"
    
    # 获取 NodePort
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null || echo "N/A")
    HTTP_PORT=$(kubectl get svc gateway-service -n "$NAMESPACE" -o jsonpath='{.spec.ports[?(@.name=="http")].nodePort}' 2>/dev/null || echo "N/A")
    WEB_PORT=$(kubectl get svc gateway-service -n "$NAMESPACE" -o jsonpath='{.spec.ports[?(@.name=="web")].nodePort}' 2>/dev/null || echo "N/A")
    TUNNEL_PORT=$(kubectl get svc gateway-service -n "$NAMESPACE" -o jsonpath='{.spec.ports[?(@.name=="tunnel")].nodePort}' 2>/dev/null || echo "N/A")
    
    if [[ "$NODE_IP" != "N/A" ]]; then
        print_info "通过 NodePort 访问:"
        print_info "  API Gateway:  http://$NODE_IP:$HTTP_PORT"
        print_info "  Web 控制台:   http://$NODE_IP:$WEB_PORT/gatewayweb"
        print_info "  隧道控制:     $NODE_IP:$TUNNEL_PORT"
    fi
    
    # 检查 Ingress
    if kubectl get ingress -n "$NAMESPACE" &> /dev/null; then
        print_info ""
        print_info "通过 Ingress 访问:"
        print_info "  API Gateway:  http://$NODE_IP/api"
        print_info "  Web 控制台:   http://$NODE_IP/gatewayweb"
    fi
    
    print_info ""
    print_info "默认登录信息:"
    print_info "  用户名: admin"
    print_info "  密码:   123456"
    print_info ""
    print_info "查看部署状态:"
    print_info "  $0 status"
    print_info ""
    print_info "查看日志:"
    print_info "  $0 logs"
}

# 卸载 Gateway
uninstall_gateway() {
    print_warning "开始卸载 FLUX Gateway..."
    
    # 检查命名空间是否存在
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_error "命名空间 $NAMESPACE 不存在"
        return 1
    fi
    
    # 删除 Ingress
    if kubectl get ingress -n "$NAMESPACE" &> /dev/null; then
        print_info "删除 Ingress..."
        kubectl delete -f "$SCRIPT_DIR/ingress.yaml" -n "$NAMESPACE" --ignore-not-found=true
    fi
    
    # 删除 Service
    print_info "删除服务..."
    kubectl delete -f "$SCRIPT_DIR/service.yaml" -n "$NAMESPACE" --ignore-not-found=true
    
    # 删除 Deployment
    print_info "删除应用..."
    kubectl delete -f "$SCRIPT_DIR/deployment.yaml" -n "$NAMESPACE" --ignore-not-found=true
    
    # 删除 ConfigMap
    print_info "删除配置..."
    kubectl delete configmap gateway-config -n "$NAMESPACE" --ignore-not-found=true
    
    # 删除命名空间
    print_warning "是否删除命名空间 $NAMESPACE? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_info "删除命名空间..."
        kubectl delete namespace "$NAMESPACE"
        print_success "命名空间已删除"
    fi
    
    print_success "FLUX Gateway 卸载完成"
}

# 升级 Gateway
upgrade_gateway() {
    print_info "升级 FLUX Gateway..."
    
    # 更新 ConfigMap
    print_info "更新配置..."
    if [[ -f "$SCRIPT_DIR/create-configmap.sh" ]]; then
        # 检查配置文件目录是否存在
        if [[ ! -d "$CONFIG_DIR" ]]; then
            print_error "配置文件目录不存在: $CONFIG_DIR"
            exit 1
        fi
        # 删除旧的 ConfigMap
        kubectl delete configmap gateway-config -n "$NAMESPACE" --ignore-not-found=true
        # 创建新的 ConfigMap
        bash "$SCRIPT_DIR/create-configmap.sh" --namespace "$NAMESPACE" --config-dir "$CONFIG_DIR"
    fi
    
    # 滚动更新 Deployment
    print_info "滚动更新应用..."
    kubectl apply -f "$SCRIPT_DIR/deployment.yaml" -n "$NAMESPACE"
    
    # 更新 Service
    print_info "更新服务..."
    kubectl apply -f "$SCRIPT_DIR/service.yaml" -n "$NAMESPACE"
    
    # 更新 Ingress
    if [[ -f "$SCRIPT_DIR/ingress.yaml" ]]; then
        print_info "更新 Ingress..."
        kubectl apply -f "$SCRIPT_DIR/ingress.yaml" -n "$NAMESPACE"
    fi
    
    # 等待更新完成
    print_info "等待更新完成..."
    kubectl rollout status deployment/gateway -n "$NAMESPACE" --timeout=180s
    
    print_success "升级完成"
    
    # 显示新的 Pod 状态
    print_info ""
    print_info "新的 Pod 状态:"
    kubectl get pods -n "$NAMESPACE" -l app=gateway
}

# 查看状态
show_status() {
    print_info "FLUX Gateway 部署状态"
    print_info "=========================================="
    
    # 检查命名空间
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_error "命名空间 $NAMESPACE 不存在"
        return 1
    fi
    
    # Pods 状态
    print_info ""
    print_info "Pods 状态:"
    kubectl get pods -n "$NAMESPACE" -l app=gateway
    
    # Deployment 状态
    print_info ""
    print_info "Deployment 状态:"
    kubectl get deployment -n "$NAMESPACE"
    
    # Service 状态
    print_info ""
    print_info "Service 状态:"
    kubectl get svc -n "$NAMESPACE"
    
    # ConfigMap 状态
    print_info ""
    print_info "ConfigMap 状态:"
    kubectl get configmap -n "$NAMESPACE"
    
    # Ingress 状态
    if kubectl get ingress -n "$NAMESPACE" &> /dev/null; then
        print_info ""
        print_info "Ingress 状态:"
        kubectl get ingress -n "$NAMESPACE"
    fi
    
    # 显示访问信息
    print_info ""
    print_info "=========================================="
    print_info "访问信息:"
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null || echo "N/A")
    HTTP_PORT=$(kubectl get svc gateway-service -n "$NAMESPACE" -o jsonpath='{.spec.ports[?(@.name=="http")].nodePort}' 2>/dev/null || echo "N/A")
    WEB_PORT=$(kubectl get svc gateway-service -n "$NAMESPACE" -o jsonpath='{.spec.ports[?(@.name=="web")].nodePort}' 2>/dev/null || echo "N/A")
    TUNNEL_PORT=$(kubectl get svc gateway-service -n "$NAMESPACE" -o jsonpath='{.spec.ports[?(@.name=="tunnel")].nodePort}' 2>/dev/null || echo "N/A")
    
    if [[ "$NODE_IP" != "N/A" && "$HTTP_PORT" != "N/A" ]]; then
        print_info "  API Gateway:  http://$NODE_IP:$HTTP_PORT"
        print_info "  Web 控制台:   http://$NODE_IP:$WEB_PORT/gatewayweb"
        print_info "  隧道控制:     $NODE_IP:$TUNNEL_PORT"
    fi
}

# 查看日志
show_logs() {
    print_info "查看 Gateway 日志 (Ctrl+C 退出)..."
    kubectl logs -f -n "$NAMESPACE" -l app=gateway --tail=100
}

# 诊断 Pod 问题
debug_gateway() {
    print_info "=========================================="
    print_info "FLUX Gateway 诊断信息"
    print_info "=========================================="
    
    # 检查命名空间
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_error "命名空间 $NAMESPACE 不存在"
        return 1
    fi
    
    # 获取所有 Pod
    print_info ""
    print_info "1. Pod 列表："
    kubectl get pods -n "$NAMESPACE" -l app=gateway -o wide
    
    # 获取第一个 Pod 名称
    POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l app=gateway -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    if [[ -z "$POD_NAME" ]]; then
        print_error "未找到 Gateway Pod"
        return 1
    fi
    
    print_info ""
    print_info "2. Pod 详细状态："
    kubectl get pod "$POD_NAME" -n "$NAMESPACE" -o yaml | grep -A 20 "status:"
    
    print_info ""
    print_info "3. Pod 事件："
    kubectl describe pod "$POD_NAME" -n "$NAMESPACE" | grep -A 30 "Events:"
    
    print_info ""
    print_info "4. 应用容器日志："
    kubectl logs "$POD_NAME" -n "$NAMESPACE" --tail=50 2>/dev/null || {
        print_warning "无法获取容器日志"
    }
    
    print_info ""
    print_info "5. ConfigMap 配置："
    kubectl get configmap gateway-config -n "$NAMESPACE" -o yaml 2>/dev/null || {
        print_warning "ConfigMap 不存在"
    }
    
    print_info ""
    print_info "6. Service 状态："
    kubectl get svc gateway-service -n "$NAMESPACE" -o wide 2>/dev/null || {
        print_warning "Service 不存在"
    }
    
    print_info ""
    print_info "=========================================="
    print_info "常见问题排查："
    print_info "=========================================="
    print_warning "1. 镜像拉取失败："
    print_warning "   - 检查镜像地址是否正确"
    print_warning "   - 检查镜像仓库凭证是否配置"
    print_warning "   - 命令: kubectl describe pod $POD_NAME -n $NAMESPACE | grep -A 5 'Events'"
    print_warning ""
    print_warning "2. 应用启动失败："
    print_warning "   - 检查 ConfigMap 配置是否正确"
    print_warning "   - 检查数据库连接配置"
    print_warning "   - 查看详细日志: kubectl logs $POD_NAME -n $NAMESPACE --tail=100"
    print_warning ""
    print_warning "3. 健康检查失败："
    print_warning "   - 检查 /health 接口是否正常"
    print_warning "   - 检查资源限制是否合理"
    print_warning "   - 命令: kubectl get pod $POD_NAME -n $NAMESPACE -o yaml | grep -A 10 'conditions'"
}

# 重启 Gateway
restart_gateway() {
    print_info "重启 Gateway..."
    kubectl rollout restart deployment/gateway -n "$NAMESPACE"
    
    print_info "等待重启完成..."
    kubectl rollout status deployment/gateway -n "$NAMESPACE"
    
    print_success "重启完成"
}

# 解析命令行参数
COMMAND=""
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        install|uninstall|upgrade|status|logs|debug|restart|help)
            COMMAND="$1"
            shift
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -c|--config-dir)
            CONFIG_DIR="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
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

# 检查命令
if [[ -z "$COMMAND" ]]; then
    print_error "请指定命令"
    show_help
    exit 1
fi

# 检查环境
check_kubectl
check_cluster

# 执行命令
case $COMMAND in
    install)
        install_gateway
        ;;
    uninstall)
        uninstall_gateway
        ;;
    upgrade)
        upgrade_gateway
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    debug)
        debug_gateway
        ;;
    restart)
        restart_gateway
        ;;
    help)
        show_help
        ;;
    *)
        print_error "未知命令: $COMMAND"
        show_help
        exit 1
        ;;
esac

