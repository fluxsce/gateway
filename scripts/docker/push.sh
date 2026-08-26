#!/bin/bash
# ============================================
# FLUX Gateway - Docker 镜像推送脚本
# ============================================
#
# 说明:
#   - 支持推送到 Docker Hub 和阿里云镜像仓库
#   - 默认推送到阿里云仓库
#   - 阿里云凭证只从环境变量读取，不要写进本文件
#
# ============================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
DEFAULT_IMAGE_NAME="datahub-images/gateway"
VERSION="3.3.0"

# 阿里云仓库地址可覆盖；账号密码必须来自环境变量
ALIYUN_REGISTRY="${ALIYUN_REGISTRY:-crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com}"
ALIYUN_NAMESPACE="${ALIYUN_NAMESPACE:-datahub-images}"

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
FLUX Gateway Docker 镜像推送脚本

用法:
    $0 [选项]

选项:
    -t, --type TYPE       镜像类型: oracle (默认) 或 standard
    -r, --registry REG    目标仓库: aliyun (默认), dockerhub, both
    -n, --name NAME       本地镜像名称 (默认: $DEFAULT_IMAGE_NAME)
    -l, --latest          同时推送 latest（标准版）或 latest-oracle（Oracle 版）
    -h, --help            显示此帮助信息

环境变量（推送阿里云时必填）:
    ALIYUN_USERNAME       ACR 登录用户名
    ALIYUN_PASSWORD       ACR 固定密码（不要写进脚本或提交到 Git）
    ALIYUN_REGISTRY       可选，默认 $ALIYUN_REGISTRY
    ALIYUN_NAMESPACE      可选，默认 $ALIYUN_NAMESPACE

标签与 build.sh 一致: 标准版 :{version}，Oracle 版 :{version}-oracle

示例:
    export ALIYUN_USERNAME='your-account'
    export ALIYUN_PASSWORD='your-registry-password'

    # 推送 Oracle 版到阿里云（默认），远程标签 :${VERSION}-oracle
    $0

    # 推送标准版到阿里云，远程标签 :${VERSION}
    $0 --type standard

    # 推送到 Docker Hub
    $0 --registry dockerhub

    # 推送到阿里云和 Docker Hub
    $0 --registry both

    # 推送并标记 latest / latest-oracle
    $0 --latest
    $0 --type standard --latest

EOF
}

# 解析命令行参数
IMAGE_TYPE="oracle"
TARGET_REGISTRY="aliyun"
IMAGE_NAME="$DEFAULT_IMAGE_NAME"
PUSH_LATEST=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            IMAGE_TYPE="$2"
            shift 2
            ;;
        -r|--registry)
            TARGET_REGISTRY="$2"
            shift 2
            ;;
        -n|--name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        -l|--latest)
            PUSH_LATEST=true
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

# 验证镜像类型
if [[ "$IMAGE_TYPE" != "standard" && "$IMAGE_TYPE" != "oracle" ]]; then
    print_error "无效的构建类型: $IMAGE_TYPE (必须是 standard 或 oracle)"
    exit 1
fi

# 验证镜像仓库
if [[ "$TARGET_REGISTRY" != "dockerhub" && "$TARGET_REGISTRY" != "aliyun" && "$TARGET_REGISTRY" != "both" ]]; then
    print_error "无效的镜像仓库: $TARGET_REGISTRY (必须是 dockerhub、aliyun 或 both)"
    exit 1
fi

if [[ "$TARGET_REGISTRY" == "aliyun" || "$TARGET_REGISTRY" == "both" ]]; then
    if [[ -z "${ALIYUN_USERNAME:-}" || -z "${ALIYUN_PASSWORD:-}" ]]; then
        print_error "推送阿里云需要环境变量 ALIYUN_USERNAME 和 ALIYUN_PASSWORD"
        print_info "示例: export ALIYUN_USERNAME='...' ; export ALIYUN_PASSWORD='...'"
        exit 1
    fi
fi

# 标准版 :{version}；Oracle 版 :{version}-oracle
if [[ "$IMAGE_TYPE" == "oracle" ]]; then
    VERSION_SUFFIX="-oracle"
else
    VERSION_SUFFIX=""
fi

# 本地镜像标签（与 build.sh 一致）
LOCAL_IMAGE="${IMAGE_NAME}:${VERSION}${VERSION_SUFFIX}"
LOCAL_LATEST="${IMAGE_NAME}:latest${VERSION_SUFFIX}"

print_info "=========================================="
print_info "Docker 镜像推送"
print_info "=========================================="
print_info "镜像类型: $IMAGE_TYPE"
print_info "镜像版本: $VERSION"
print_info "目标仓库: $TARGET_REGISTRY"
print_info "本地镜像: $LOCAL_IMAGE"

# 检查本地镜像是否存在
if ! docker image inspect "$LOCAL_IMAGE" &> /dev/null; then
    print_error "本地镜像不存在: $LOCAL_IMAGE"
    print_info "请先运行构建脚本: ./build.sh --type $IMAGE_TYPE"
    exit 1
fi

# 推送到 Docker Hub
if [[ "$TARGET_REGISTRY" == "dockerhub" || "$TARGET_REGISTRY" == "both" ]]; then
    print_info ""
    print_info "=========================================="
    print_info "推送到 Docker Hub"
    print_info "=========================================="

    # 检查是否登录
    if ! docker info 2>/dev/null | grep -q "Username"; then
        print_warning "未登录 Docker Hub，请先登录"
        docker login
    fi

    # 推送版本标签
    print_info "推送: $LOCAL_IMAGE"
    docker push "$LOCAL_IMAGE"

    if [[ $? -eq 0 ]]; then
        print_success "推送成功: $LOCAL_IMAGE"
    else
        print_error "推送失败: $LOCAL_IMAGE"
        exit 1
    fi

    # 推送 latest 标签
    if [[ "$PUSH_LATEST" == true ]]; then
        if docker image inspect "$LOCAL_LATEST" &> /dev/null; then
            print_info "推送: $LOCAL_LATEST"
            docker push "$LOCAL_LATEST"

            if [[ $? -eq 0 ]]; then
                print_success "推送成功: $LOCAL_LATEST"
            else
                print_error "推送失败: $LOCAL_LATEST"
                exit 1
            fi
        else
            print_warning "本地不存在 latest 标签，跳过推送"
            print_info "请使用 --latest 参数构建镜像"
        fi
    fi
fi

# 推送到阿里云
if [[ "$TARGET_REGISTRY" == "aliyun" || "$TARGET_REGISTRY" == "both" ]]; then
    print_info ""
    print_info "=========================================="
    print_info "推送到阿里云镜像仓库"
    print_info "=========================================="

    # 阿里云镜像标签
    ALIYUN_IMAGE="${ALIYUN_REGISTRY}/${ALIYUN_NAMESPACE}/gateway:${VERSION}${VERSION_SUFFIX}"
    ALIYUN_LATEST="${ALIYUN_REGISTRY}/${ALIYUN_NAMESPACE}/gateway:latest${VERSION_SUFFIX}"

    print_info "目标镜像: $ALIYUN_IMAGE"

    # 登录阿里云镜像仓库（密码走 stdin，避免出现在进程列表）
    print_info "登录阿里云镜像仓库: $ALIYUN_REGISTRY （用户 ${ALIYUN_USERNAME}）"
    echo "$ALIYUN_PASSWORD" | docker login "$ALIYUN_REGISTRY" \
        --username "$ALIYUN_USERNAME" \
        --password-stdin

    if [[ $? -eq 0 ]]; then
        print_success "登录成功"
    else
        print_error "登录失败，请检查 ALIYUN_USERNAME / ALIYUN_PASSWORD"
        exit 1
    fi

    # 标记镜像
    print_info "标记镜像: $LOCAL_IMAGE -> $ALIYUN_IMAGE"
    docker tag "$LOCAL_IMAGE" "$ALIYUN_IMAGE"

    # 推送版本标签
    print_info "推送: $ALIYUN_IMAGE"
    docker push "$ALIYUN_IMAGE"

    if [[ $? -eq 0 ]]; then
        print_success "推送成功: $ALIYUN_IMAGE"
    else
        print_error "推送失败: $ALIYUN_IMAGE"
        exit 1
    fi

    # 推送 latest 标签
    if [[ "$PUSH_LATEST" == true ]]; then
        if docker image inspect "$LOCAL_LATEST" &> /dev/null; then
            print_info "标记镜像: $LOCAL_LATEST -> $ALIYUN_LATEST"
            docker tag "$LOCAL_LATEST" "$ALIYUN_LATEST"

            print_info "推送: $ALIYUN_LATEST"
            docker push "$ALIYUN_LATEST"

            if [[ $? -eq 0 ]]; then
                print_success "推送成功: $ALIYUN_LATEST"
            else
                print_error "推送失败: $ALIYUN_LATEST"
                exit 1
            fi
        else
            print_warning "本地不存在 latest 标签，跳过推送"
            print_info "请使用 --latest 参数构建镜像"
        fi
    fi
fi

# 完成
print_info ""
print_success "=========================================="
print_success "镜像推送完成！"
print_success "=========================================="

if [[ "$TARGET_REGISTRY" == "dockerhub" || "$TARGET_REGISTRY" == "both" ]]; then
    print_info "Docker Hub:"
    print_info "  - $LOCAL_IMAGE"
    if [[ "$PUSH_LATEST" == true ]]; then
        print_info "  - $LOCAL_LATEST"
    fi
fi

if [[ "$TARGET_REGISTRY" == "aliyun" || "$TARGET_REGISTRY" == "both" ]]; then
    print_info "阿里云镜像仓库:"
    print_info "  - $ALIYUN_IMAGE"
    if [[ "$PUSH_LATEST" == true ]]; then
        print_info "  - $ALIYUN_LATEST"
    fi
fi

print_info ""
print_info "拉取镜像:"
if [[ "$TARGET_REGISTRY" == "dockerhub" || "$TARGET_REGISTRY" == "both" ]]; then
    print_info "  docker pull $LOCAL_IMAGE"
fi
if [[ "$TARGET_REGISTRY" == "aliyun" || "$TARGET_REGISTRY" == "both" ]]; then
    print_info "  docker pull $ALIYUN_IMAGE"
fi
