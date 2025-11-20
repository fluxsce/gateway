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
VERSION="2.0.3"
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
    -t, --type TYPE       构建类型: standard (默认) 或 oracle
    -v, --version VER     镜像版本 (默认: $VERSION)
    -n, --name NAME       镜像名称 (默认: $DEFAULT_IMAGE_NAME)
    -l, --latest          同时标记为 latest
    -h, --help            显示此帮助信息

示例:
    # 构建标准版本
    $0

    # 构建 Oracle 版本
    $0 --type oracle

    # 构建并标记为 latest
    $0 --latest

    # 构建特定版本
    $0 --version 2.1.0

    # 构建后推送镜像
    $0 && ./push.sh

EOF
}

# 解析命令行参数
BUILD_TYPE="standard"
TAG_LATEST=false
IMAGE_NAME="$DEFAULT_IMAGE_NAME"

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            BUILD_TYPE="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -n|--name)
            IMAGE_NAME="$2"
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

# 选择 Dockerfile 和构建标签
if [[ "$BUILD_TYPE" == "oracle" ]]; then
    DOCKERFILE="scripts/docker/Dockerfile.oracle"
    VERSION_SUFFIX="-oracle"
else
    DOCKERFILE="scripts/docker/Dockerfile"
    VERSION_SUFFIX=""
fi

# 构建镜像标签
IMAGE_TAG="${IMAGE_NAME}:${VERSION}${VERSION_SUFFIX}"
LATEST_TAG="${IMAGE_NAME}:latest${VERSION_SUFFIX}"

print_info "使用 Dockerfile: $DOCKERFILE"
print_info "镜像标签: $IMAGE_TAG"
if [[ "$TAG_LATEST" == true ]]; then
    print_info "Latest 标签: $LATEST_TAG"
fi

# 检查 Dockerfile 是否存在
if [[ ! -f "$DOCKERFILE" ]]; then
    print_error "Dockerfile 不存在: $DOCKERFILE"
    exit 1
fi

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

# 构建镜像
print_info "开始构建 Docker 镜像..."

docker build \
    -f "$DOCKERFILE" \
    -t "$IMAGE_TAG" \
    --build-arg VERSION="$VERSION" \
    --build-arg BUILD_DATE="$BUILD_DATE" \
    --build-arg GIT_COMMIT="$GIT_COMMIT" \
    --progress=plain \
    .

if [[ $? -eq 0 ]]; then
    print_success "镜像构建成功: $IMAGE_TAG"
else
    print_error "镜像构建失败"
    exit 1
fi

# 标记为 latest
if [[ "$TAG_LATEST" == true ]]; then
    print_info "标记为 latest: $LATEST_TAG"
    docker tag "$IMAGE_TAG" "$LATEST_TAG"
    print_success "标记成功: $LATEST_TAG"
fi

# 显示镜像信息
print_info "镜像信息:"
docker images | grep "$IMAGE_NAME" | grep "$VERSION"

# 完成
print_success "=========================================="
print_success "Docker 镜像构建完成！"
print_success "=========================================="
print_info "构建的镜像:"
print_info "  - $IMAGE_TAG"
if [[ "$TAG_LATEST" == true ]]; then
    print_info "  - $LATEST_TAG"
fi
print_info ""
print_info "运行镜像:"
print_info "  docker run -d -p 8080:8080 -p 12003:12003 -p 7000:7000 $IMAGE_TAG"
print_info ""
print_info "使用 Docker Compose:"
print_info "  cd scripts/docker && docker-compose up -d"
print_info ""
print_info "推送镜像:"
print_info "  ./scripts/docker/push.sh"

