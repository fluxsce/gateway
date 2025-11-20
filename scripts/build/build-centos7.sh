#!/bin/bash

# Gateway CentOS 7 Build Script for Linux
# Optimized for cross-compilation to CentOS 7

set -e

echo "=========================================="
echo " Gateway CentOS 7 Build for Linux"
echo "=========================================="
echo ""

# Auto-detect Go version
CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo "$CURRENT_GO" | cut -d. -f1)
GO_MINOR=$(echo "$CURRENT_GO" | cut -d. -f2)
CURRENT_GO_MAJOR="${GO_MAJOR}.${GO_MINOR}"

# Go version selection
GO_VERSION="$CURRENT_GO_MAJOR"
echo "Auto-detected Go version: $CURRENT_GO (Using $GO_VERSION for build)"

# Check if version parameter is provided to override auto-detection
if [ ! -z "$1" ]; then
    GO_VERSION="$1"
    echo "Overriding with specified Go version: $GO_VERSION"
fi

# Validate Go version and set module file
case "$GO_VERSION" in
    1.19)
        GO_MOD_FILE="go.mod.1.19"
        ;;
    1.20)
        GO_MOD_FILE="go.mod.1.20"
        ;;
    1.21|1.22|1.23|1.24)
        GO_MOD_FILE="go.mod"
        ;;
    *)
        echo "[INFO] Using default configuration for Go $GO_VERSION"
        GO_MOD_FILE="go.mod"
        ;;
esac

echo "Using Go version: $GO_VERSION"
echo "Module file: $GO_MOD_FILE"
echo ""

# Build configuration for CentOS 7
# Note: SQLite requires CGO (C compiler), pure Go build will exclude SQLite
echo ""
echo "========================================"
echo "Choose build mode:"
echo "========================================"
echo "1) With SQLite support (CGO_ENABLED=1)"
echo "   - Supports: MySQL, SQLite, ClickHouse"
echo "   - Requires: GCC compiler on build machine"
echo "   - Binary has external dependencies"
echo ""
echo "2) Pure Go build (CGO_ENABLED=0) [Recommended]"
echo "   - Supports: MySQL, ClickHouse"
echo "   - No SQLite support"
echo "   - Fully static binary, no external dependencies"
echo "   - Better for cross-platform deployment"
echo "========================================"
read -p "Enter your choice [1-2] (default: 2): " BUILD_MODE
BUILD_MODE=${BUILD_MODE:-2}

if [ "$BUILD_MODE" = "1" ]; then
    BUILD_TAGS="netgo,no_oracle"
    export CGO_ENABLED=1
    echo ""
    echo "✓ Building with SQLite support (CGO_ENABLED=1)"
    echo "  Build tags: $BUILD_TAGS"
else
    BUILD_TAGS="netgo,osusergo,no_oracle,no_sqlite"
    export CGO_ENABLED=0
    echo ""
    echo "✓ Building pure Go version (CGO_ENABLED=0)"
    echo "  Build tags: $BUILD_TAGS"
    echo "  Note: SQLite is disabled in this build"
fi

# Output configuration (fixed name for auto-registration)
OUTPUT_SUFFIX="centos7"
VERSION_SUFFIX="centos7"

# Cross-compilation settings for CentOS 7
export GOOS=linux
export GOARCH=amd64

# Get to project root (script is in scripts/build/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Create output directory
mkdir -p dist

# Build info
BUILD_DATE=$(date +%Y-%m-%d)
BUILD_TIME=$(date +%H:%M:%S)
BUILD_TIMESTAMP="${BUILD_DATE}T${BUILD_TIME}"

# Get Git commit
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Output file
OUTPUT_FILE="dist/gateway-${VERSION_SUFFIX}-amd64"
VERSION_INFO="${VERSION_SUFFIX}-v3.1"

# Build flags optimized for CentOS 7
LDFLAGS="-s -w -X main.Version=${VERSION_INFO} -X main.BuildTime=${BUILD_TIMESTAMP} -X main.GitCommit=${GIT_COMMIT}"

echo "Building CentOS 7 version..."
echo "Output: $OUTPUT_FILE"
echo "Version: $VERSION_INFO"
echo "Build Tags: $BUILD_TAGS"
echo "Cross-compilation: GOOS=$GOOS GOARCH=$GOARCH"
echo ""

# Use specified Go version compatible dependencies
echo "Using Go $GO_VERSION compatible dependencies..."
if [ -f "$GO_MOD_FILE" ] && [ "$GO_MOD_FILE" != "go.mod" ]; then
    echo "Switching to $GO_MOD_FILE..."
    cp go.mod go.mod.bak
    [ -f go.sum ] && cp go.sum go.sum.bak
    cp "$GO_MOD_FILE" go.mod
    [ -f go.sum ] && rm go.sum
    go clean -modcache
    go mod download
    go mod tidy -compat="$GO_VERSION"
elif [ "$GO_MOD_FILE" = "go.mod" ]; then
    echo "Using current go.mod (already compatible with Go $GO_VERSION)"
    go mod download
else
    echo "[WARNING] $GO_MOD_FILE not found, using current go.mod"
fi

# Execute build with verbose output
echo ""
echo "Running build..."
go build -v -x -tags "$BUILD_TAGS" -ldflags "$LDFLAGS" -o "$OUTPUT_FILE" cmd/app/main.go
BUILD_RESULT=$?

# Check if the output file exists and has size
if [ $BUILD_RESULT -eq 0 ]; then
    if [ -f "$OUTPUT_FILE" ]; then
        FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
        FILE_SIZE_BYTES=$(stat -f%z "$OUTPUT_FILE" 2>/dev/null || stat -c%s "$OUTPUT_FILE" 2>/dev/null)
        
        if [ "$FILE_SIZE_BYTES" -gt 0 ]; then
            echo ""
            echo "[SUCCESS] Build completed successfully!"
            echo "Output: $OUTPUT_FILE"
            echo "Size: $FILE_SIZE"
            echo "Go version: $CURRENT_GO"
            echo "Module file: $GO_MOD_FILE"
            
            echo ""
            echo "Build artifacts location:"
            ls -lh "$OUTPUT_FILE"
            
            # Restore original go.mod and go.sum if they were backed up
            if [ -f go.mod.bak ]; then
                echo ""
                echo "Restoring original module files..."
                mv go.mod.bak go.mod
                [ -f go.sum.bak ] && mv go.sum.bak go.sum
                echo "Module files restored."
            fi
            
            echo ""
            echo "[DEPLOYMENT INSTRUCTIONS]"
            echo "1. Copy $OUTPUT_FILE to your CentOS 7 server"
            echo "2. Make sure the binary has execute permissions:"
            echo "   chmod +x $OUTPUT_FILE"
            echo "3. Run the binary: ./$OUTPUT_FILE"
            
            echo ""
            echo "[TIP] To use Go $GO_VERSION compatible dependencies permanently:"
            echo "cp $GO_MOD_FILE go.mod"
            echo "go mod tidy -compat=$GO_VERSION"
        else
            echo "[ERROR] Build output file exists but has zero size"
            exit 1
        fi
    else
        echo "[ERROR] Build completed but output file not found"
        exit 1
    fi
else
    echo ""
    echo "[FAILED] Build failed with error code $BUILD_RESULT"
    echo ""
    echo "Debug Information:"
    echo "-----------------"
    echo "Go Information:"
    echo "Version: $CURRENT_GO"
    echo "Module file: $GO_MOD_FILE"
    go version
    echo ""
    echo "Environment Variables:"
    echo "GOOS: $GOOS"
    echo "GOARCH: $GOARCH"
    echo "CGO_ENABLED: $CGO_ENABLED"
    
    echo ""
    echo "[TIP] Common issues and solutions:"
    echo "1. Make sure you have the correct Go version ($GO_VERSION)"
    echo "2. Try running 'go clean -cache' and rebuild"
    echo "3. Check if your code is compatible with CentOS 7"
    echo "4. Verify that all dependencies support linux/amd64"
    exit 1
fi

