#!/bin/bash
# ============================================
# FLUX Gateway - Docker 镜像构建脚本
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
VERSION="3.2.2"
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 镜像仓库配置
DOCKERHUB_REGISTRY="docker.io"
ALIYUN_REGISTRY="crpi-25xt72cd1prwdj5s.cn-hangzhou.personal.cr.aliyuncs.com"

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
FLUX Gateway Docker 镜像构建脚本

用法:
    $0 [选项]

选项:
    -t, --type TYPE       构建类型: oracle (默认) 或 standard
    -n, --name NAME       镜像名称 (默认: $DEFAULT_IMAGE_NAME)
    -m, --mirror HOST     Debian 镜像源主机（解决 apt 拉取超时），如 mirrors.aliyun.com
    -l, --latest          同时打 latest（标准版）或 latest-oracle（Oracle 版）
    -h, --help            显示此帮助信息

    也可通过环境变量 DEBIAN_MIRROR 指定镜像（与 -m 二选一）。

标签:
    standard  -> {name}:{version}           以及可选 :latest
    oracle    -> {name}:{version}-oracle    以及可选 :latest-oracle

示例:
    # 构建 Oracle 版（默认），标签为 :${VERSION}-oracle
    $0

    # 国内/内网环境 apt 超时时，使用镜像源构建
    $0 --mirror mirrors.aliyun.com
    DEBIAN_MIRROR=mirrors.tuna.tsinghua.edu.cn $0

    # 构建标准版，标签为 :${VERSION}
    $0 --type standard

    # 构建并标记 latest / latest-oracle
    $0 --latest
    $0 --type standard --latest

    # 构建后推送镜像
    $0 && ./push.sh

EOF
}

# 解析命令行参数
BUILD_TYPE="oracle"
TAG_LATEST=false
IMAGE_NAME="$DEFAULT_IMAGE_NAME"
# 从环境变量读取，命令行 -m 可覆盖
DEBIAN_MIRROR="${DEBIAN_MIRROR:-}"

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            BUILD_TYPE="$2"
            shift 2
            ;;
        -n|--name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        -m|--mirror)
            DEBIAN_MIRROR="$2"
            shift 2
            ;;
        -l|--latest)
            TAG_LATEST=true
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

# 验证构建类型
if [[ "$BUILD_TYPE" != "standard" && "$BUILD_TYPE" != "oracle" ]]; then
    print_error "无效的构建类型: $BUILD_TYPE (必须是 standard 或 oracle)"
    exit 1
fi

# 切换到项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

print_info "项目根目录: $PROJECT_ROOT"
print_info "构建类型: $BUILD_TYPE"
print_info "镜像版本: $VERSION"
print_info "Git Commit: $GIT_COMMIT"

# 检查必要的文件
print_info "检查必要文件..."
REQUIRED_DIRS=("configs" "web" "cmd")
for dir in "${REQUIRED_DIRS[@]}"; do
    if [[ ! -d "$dir" ]]; then
        print_error "目录不存在: $dir"
        exit 1
    fi
done

if [[ ! -f "go.mod" ]]; then
    print_error "go.mod 文件不存在"
    exit 1
fi

print_success "文件检查通过"

# 标准版 :{version}；Oracle 版 :{version}-oracle
tag_suffix() {
    if [[ "$1" == "oracle" ]]; then
        echo "-oracle"
    else
        echo ""
    fi
}

# 定义构建函数
build_image() {
    local build_type=$1
    local dockerfile=""
    local version_suffix=""
    
    if [[ "$build_type" == "oracle" ]]; then
        dockerfile="scripts/docker/Dockerfile.oracle"
        version_suffix="$(tag_suffix oracle)"
    else
        dockerfile="scripts/docker/Dockerfile"
        version_suffix="$(tag_suffix standard)"
    fi
    
    # 检查 Dockerfile 是否存在
    if [[ ! -f "$dockerfile" ]]; then
        print_error "Dockerfile 不存在: $dockerfile"
        return 1
    fi
    
    # 构建镜像标签
    local image_tag="${IMAGE_NAME}:${VERSION}${version_suffix}"
    local latest_tag="${IMAGE_NAME}:latest${version_suffix}"
    
    print_info ""
    print_info "=========================================="
    print_info "构建 $build_type 版本"
    print_info "=========================================="
    print_info "使用 Dockerfile: $dockerfile"
    print_info "镜像标签: $image_tag"
    if [[ "$TAG_LATEST" == true ]]; then
        print_info "Latest 标签: $latest_tag"
    fi
    if [[ -n "$DEBIAN_MIRROR" ]]; then
        print_info "使用 Debian 镜像源: $DEBIAN_MIRROR"
    fi
    
    # 构建参数（oracle 类型可传 Debian 镜像以应对网络超时）
    local build_args=(
        -f "$dockerfile"
        -t "$image_tag"
        --build-arg "VERSION=$VERSION"
        --build-arg "BUILD_DATE=$BUILD_DATE"
        --build-arg "GIT_COMMIT=$GIT_COMMIT"
        --progress=plain
    )
    if [[ -n "$DEBIAN_MIRROR" && "$build_type" == "oracle" ]]; then
        build_args+=(--build-arg "DEBIAN_MIRROR=$DEBIAN_MIRROR")
    fi
    
    # 构建镜像
    print_info "开始构建 Docker 镜像..."
    
    docker build "${build_args[@]}" .
    
    if [[ $? -eq 0 ]]; then
        print_success "镜像构建成功: $image_tag"
        
        # 标记为 latest
        if [[ "$TAG_LATEST" == true ]]; then
            print_info "标记为 latest: $latest_tag"
            docker tag "$image_tag" "$latest_tag"
            print_success "标记成功: $latest_tag"
        fi
        
        return 0
    else
        print_error "镜像构建失败: $image_tag"
        return 1
    fi
}

# 根据构建类型执行构建
BUILD_FAILED=0
BUILT_IMAGES=()

# 构建指定版本
IMAGE_TAG="${IMAGE_NAME}:${VERSION}$(tag_suffix "$BUILD_TYPE")"
LATEST_TAG="${IMAGE_NAME}:latest$(tag_suffix "$BUILD_TYPE")"
if build_image "$BUILD_TYPE"; then
    BUILT_IMAGES+=("$IMAGE_TAG")
    if [[ "$TAG_LATEST" == true ]]; then
        BUILT_IMAGES+=("$LATEST_TAG")
    fi
else
    BUILD_FAILED=1
fi

# 显示镜像信息
print_info ""
print_info "=========================================="
print_info "构建的镜像列表:"
print_info "=========================================="
for img in "${BUILT_IMAGES[@]}"; do
    print_info "  - $img"
done

print_info ""
print_info "镜像信息:"
docker images | grep "$IMAGE_NAME" | grep -E "$VERSION|latest" || true

# 完成
print_info ""
if [[ $BUILD_FAILED -eq 0 ]]; then
    print_success "=========================================="
    print_success "Docker 镜像构建完成！"
    print_success "=========================================="
else
    print_warning "=========================================="
    print_warning "部分镜像构建失败，请检查错误信息"
    print_warning "=========================================="
fi

print_info ""
print_info "运行镜像示例:"
print_info "  docker run -d -p 8080:8080 -p 12003:12003 -p 7000:7000 ${IMAGE_TAG}"
print_info ""
print_info "使用 Docker Compose:"
print_info "  cd scripts/docker && docker-compose up -d"
print_info ""
print_info "推送镜像:"
print_info "  ./scripts/docker/push.sh --type ${BUILD_TYPE}"

if [[ $BUILD_FAILED -ne 0 ]]; then
    exit 1
fi

