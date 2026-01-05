#!/bin/bash

# Gateway CentOS 7 Build Script for Linux
# Optimized for cross-compilation to CentOS 7

set -e

echo "=========================================="
echo " Gateway CentOS 7 Build for Linux"
echo "=========================================="
echo ""

# Show current Go version
CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
echo "Current Go version: $CURRENT_GO"
echo "Using default go.mod file"
echo ""

# Build configuration for CentOS 7
# Default: Build with MySQL/SQLite only (no Oracle support)
# Use --oracle or --all to enable Oracle support
BUILD_TAGS="netgo,no_oracle"
export CGO_ENABLED=1

# Parse command line arguments
ORACLE_ENABLED=false
if [[ "$1" == "--oracle" ]] || [[ "$1" == "--all" ]]; then
    ORACLE_ENABLED=true
    # Remove no_oracle tag to enable Oracle support
    BUILD_TAGS="netgo"
fi

# Oracle environment check
ORACLE_SUPPORT_ENABLED=false
ORACLE_CHECK_FAILED=false

if [ "$ORACLE_ENABLED" = true ]; then
    echo "[INFO] Oracle support is enabled"
    
    if [ -n "$ORACLE_HOME" ]; then
        echo "[INFO] Oracle environment detected"
        echo "  ORACLE_HOME: $ORACLE_HOME"
        
        # Check for Oracle library
        ORACLE_LIB_FOUND=false
        if [ -f "${ORACLE_HOME}/libclntsh.so" ]; then
            ORACLE_LIB_FOUND=true
            echo "  [OK] Found libclntsh.so"
        elif [ -n "$(ls ${ORACLE_HOME}/libclntsh.so.* 2>/dev/null)" ]; then
            ORACLE_LIB_FOUND=true
            echo "  [OK] Found libclntsh.so.*"
        else
            echo "  [ERROR] Oracle library (libclntsh.so) not found in $ORACLE_HOME"
            ORACLE_CHECK_FAILED=true
        fi
        
        # Check for Oracle header
        if [ -f "${ORACLE_HOME}/sdk/include/oci.h" ]; then
            echo "  [OK] Found oci.h"
        else
            echo "  [ERROR] Oracle header (oci.h) not found in ${ORACLE_HOME}/sdk/include"
            ORACLE_CHECK_FAILED=true
        fi
        
        if [ "$ORACLE_CHECK_FAILED" = false ]; then
            # Set Oracle CGO flags for Linux
            export CGO_CFLAGS="-I${ORACLE_HOME}/sdk/include"
            export CGO_LDFLAGS="-L${ORACLE_HOME} -lclntsh"
            
            # Add Oracle library path to LD_LIBRARY_PATH if not already present
            if [[ ":$LD_LIBRARY_PATH:" != *":$ORACLE_HOME:"* ]]; then
                export LD_LIBRARY_PATH="${ORACLE_HOME}:${LD_LIBRARY_PATH}"
            fi
            
            echo "  CGO_CFLAGS: $CGO_CFLAGS"
            echo "  CGO_LDFLAGS: $CGO_LDFLAGS"
            ORACLE_SUPPORT_ENABLED=true
            echo ""
        else
            echo ""
            echo "=========================================="
            echo " [ERROR] Oracle environment check failed!"
            echo "=========================================="
            echo ""
            echo "Oracle support is enabled in build tags, but required files are missing:"
            echo ""
            echo "Required files:"
            echo "  1. ${ORACLE_HOME}/libclntsh.so (or libclntsh.so.*)"
            echo "  2. ${ORACLE_HOME}/sdk/include/oci.h"
            echo ""
            echo "Please install Oracle Instant Client:"
            echo "  1. Download Oracle Instant Client from:"
            echo "     https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html"
            echo "  2. Extract to a directory (e.g., /usr/lib/oracle/21/client64)"
            echo "  3. Set ORACLE_HOME environment variable:"
            echo "     export ORACLE_HOME=/usr/lib/oracle/21/client64"
            echo ""
            echo "Or build without Oracle support by adding 'no_oracle' to BUILD_TAGS"
            echo ""
            exit 1
        fi
    else
        echo ""
        echo "=========================================="
        echo " [ERROR] Oracle environment not configured!"
        echo "=========================================="
        echo ""
        echo "Oracle support is enabled in build tags, but ORACLE_HOME is not set."
        echo ""
        echo "To enable Oracle support:"
        echo "  1. Install Oracle Instant Client"
        echo "  2. Set ORACLE_HOME environment variable:"
        echo "     export ORACLE_HOME=/usr/lib/oracle/21/client64"
        echo ""
        echo "Or build without Oracle support by adding 'no_oracle' to BUILD_TAGS"
        echo ""
        exit 1
    fi
else
    echo "[INFO] Oracle support is disabled in build tags"
    echo ""
fi

# Display build configuration
if [ "$ORACLE_SUPPORT_ENABLED" = true ]; then
    echo "[INFO] Building with all features (CGO_ENABLED=1)"
    echo "  - Supports: MySQL, SQLite, ClickHouse, Oracle"
    echo "  - Build tags: $BUILD_TAGS"
else
    echo "[INFO] Building without Oracle support (CGO_ENABLED=1)"
    echo "  - Supports: MySQL, SQLite, ClickHouse"
    echo "  - Build tags: $BUILD_TAGS"
fi
echo ""

# Output configuration
VERSION_SUFFIX="centos7"

# Cross-compilation settings for CentOS 7
export GOOS=linux
export GOARCH=amd64

# Get to project root (script is in scripts/build/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Clean and create output directory
echo "Cleaning dist directory..."
if [ -d "dist" ]; then
    rm -rf dist
    echo "[OK] Dist directory cleaned"
fi
mkdir -p dist

# Build info
BUILD_DATE=$(date +%Y-%m-%d)
BUILD_TIME=$(date +%H:%M:%S)
BUILD_TIMESTAMP="${BUILD_DATE}T${BUILD_TIME}"

# Get Git commit
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Output file - always use gateway
OUTPUT_FILE="dist/gateway"
PACKAGE_DIR="dist/gateway"
VERSION_INFO="${VERSION_SUFFIX}-v3.1"

# Build flags optimized for CentOS 7
LDFLAGS="-s -w -X main.Version=${VERSION_INFO} -X main.BuildTime=${BUILD_TIMESTAMP} -X main.GitCommit=${GIT_COMMIT}"

echo "Building CentOS 7 version..."
echo "Output: $OUTPUT_FILE"
echo "Version: $VERSION_INFO"
echo "Build Tags: $BUILD_TAGS"
echo "Cross-compilation: GOOS=$GOOS GOARCH=$GOARCH"
echo ""

# Prepare dependencies
echo "Preparing dependencies..."
go mod download
go mod tidy

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
            if [ "$ORACLE_SUPPORT_ENABLED" = true ]; then
                echo "Build type: With all features (MySQL, SQLite, ClickHouse, Oracle support)"
            else
                echo "Build type: With features (MySQL, SQLite, ClickHouse support, Oracle disabled)"
            fi
            
            echo ""
            echo "=========================================="
            echo " Packaging deployment structure..."
            echo "=========================================="
            
            # Handle executable file if it exists (it's a file, not a directory)
            TEMP_EXECUTABLE=""
            if [ -f "$OUTPUT_FILE" ]; then
                # Backup executable file temporarily
                TEMP_EXECUTABLE="${OUTPUT_FILE}.tmp"
                if mv "$OUTPUT_FILE" "$TEMP_EXECUTABLE" 2>/dev/null; then
                    echo "[INFO] Executable file moved temporarily for directory creation"
                else
                    echo "[WARNING] Failed to move executable file, may cause directory creation issues"
                fi
            fi
            
            # Create package directory structure
            echo "Creating directory structure..."
            if [ -d "$PACKAGE_DIR" ]; then
                # If directory already exists, remove it first
                rm -rf "$PACKAGE_DIR"
                echo "[INFO] Removed existing package directory"
            fi
            mkdir -p "$PACKAGE_DIR"
            mkdir -p "$PACKAGE_DIR/configs"
            mkdir -p "$PACKAGE_DIR/web"
            mkdir -p "$PACKAGE_DIR/web/static"
            mkdir -p "$PACKAGE_DIR/web/frontend"
            mkdir -p "$PACKAGE_DIR/web/frontend/dist"
            mkdir -p "$PACKAGE_DIR/logs"
            mkdir -p "$PACKAGE_DIR/backup"
            mkdir -p "$PACKAGE_DIR/scripts"
            mkdir -p "$PACKAGE_DIR/scripts/db"
            mkdir -p "$PACKAGE_DIR/scripts/data"
            mkdir -p "$PACKAGE_DIR/scripts/deploy"
            mkdir -p "$PACKAGE_DIR/pprof_analysis"
            
            # Copy executable file
            echo "Copying executable file..."
            if [ -n "$TEMP_EXECUTABLE" ] && [ -f "$TEMP_EXECUTABLE" ]; then
                # Copy from temporary location
                if cp "$TEMP_EXECUTABLE" "$PACKAGE_DIR/gateway"; then
                    chmod +x "$PACKAGE_DIR/gateway"
                    rm -f "$TEMP_EXECUTABLE"
                    echo "[OK] Executable file copied"
                else
                    echo "[WARNING] Failed to copy executable file"
                    # Restore original file if copy failed
                    if [ -f "$TEMP_EXECUTABLE" ]; then
                        mv "$TEMP_EXECUTABLE" "$OUTPUT_FILE"
                    fi
                fi
            elif [ -f "$OUTPUT_FILE" ]; then
                # Fallback: if temp move didn't work, try direct copy
                if cp "$OUTPUT_FILE" "$PACKAGE_DIR/gateway"; then
                    chmod +x "$PACKAGE_DIR/gateway"
                    echo "[OK] Executable file copied"
                else
                    echo "[WARNING] Failed to copy executable file"
                fi
            else
                echo "[WARNING] Executable file not found: $OUTPUT_FILE"
            fi
            
            # Copy configuration files
            echo "Copying configuration files..."
            if [ -d "configs" ]; then
                if cp -r configs/* "$PACKAGE_DIR/configs/" 2>/dev/null; then
                    echo "[OK] Configuration files copied"
                else
                    echo "[WARNING] Failed to copy configuration files"
                fi
            else
                echo "[WARNING] Configuration directory not found"
            fi
            
            # Copy web static resources
            echo "Copying web static resources..."
            if [ -d "web/static" ]; then
                if cp -r web/static/* "$PACKAGE_DIR/web/static/" 2>/dev/null; then
                    echo "[OK] Web static resources copied"
                else
                    echo "[WARNING] Failed to copy web static resources"
                fi
            else
                echo "[WARNING] Web static directory not found"
            fi
            
            # Copy frontend dist resources
            echo "Copying frontend dist resources..."
            if [ -d "web/frontend/dist" ]; then
                if cp -r web/frontend/dist/* "$PACKAGE_DIR/web/frontend/dist/" 2>/dev/null; then
                    echo "[OK] Frontend dist resources copied"
                else
                    echo "[WARNING] Failed to copy frontend dist resources"
                fi
            else
                echo "[WARNING] Frontend dist directory not found"
                echo "[INFO] Please build frontend first: cd web/frontend && npm run build"
            fi
            
            # Copy scripts directories
            echo "Copying scripts directories..."
            
            # Copy db scripts
            if [ -d "scripts/db" ]; then
                if cp -r scripts/db/* "$PACKAGE_DIR/scripts/db/" 2>/dev/null; then
                    echo "[OK] Database scripts copied"
                else
                    echo "[WARNING] Failed to copy db scripts"
                fi
            else
                echo "[WARNING] Database scripts directory not found"
            fi
            
            # Copy deploy scripts
            if [ -d "scripts/deploy" ]; then
                if cp -r scripts/deploy/* "$PACKAGE_DIR/scripts/deploy/" 2>/dev/null; then
                    echo "[OK] Deploy scripts copied"
                else
                    echo "[WARNING] Failed to copy deploy scripts"
                fi
            else
                echo "[WARNING] Deploy scripts directory not found"
            fi
            
            # Copy docker scripts
            if [ -d "scripts/docker" ]; then
                if cp -r scripts/docker/* "$PACKAGE_DIR/scripts/docker/" 2>/dev/null; then
                    echo "[OK] Docker scripts copied"
                else
                    echo "[WARNING] Failed to copy docker scripts"
                fi
            else
                echo "[WARNING] Docker scripts directory not found"
            fi
            
            # Copy k8s scripts
            if [ -d "scripts/k8s" ]; then
                if cp -r scripts/k8s/* "$PACKAGE_DIR/scripts/k8s/" 2>/dev/null; then
                    echo "[OK] K8s scripts copied"
                else
                    echo "[WARNING] Failed to copy k8s scripts"
                fi
            else
                echo "[WARNING] K8s scripts directory not found"
            fi
            
            # Copy test scripts
            if [ -d "scripts/test" ]; then
                if cp -r scripts/test/* "$PACKAGE_DIR/scripts/test/" 2>/dev/null; then
                    echo "[OK] Test scripts copied"
                else
                    echo "[WARNING] Failed to copy test scripts"
                fi
            else
                echo "[WARNING] Test scripts directory not found"
            fi
            
            # Note: scripts/data directory is created empty (not copied from source)
            
            echo ""
            echo "=========================================="
            echo " Package structure created successfully!"
            echo "=========================================="
            echo "Package directory: $PACKAGE_DIR"
            echo ""
            echo "Directory structure:"
            ls -d "$PACKAGE_DIR"/*/ 2>/dev/null | sed 's|/$||' | xargs -n1 basename
            echo ""
            
            echo "Build artifacts location:"
            ls -lh "$OUTPUT_FILE"
            echo ""
            echo "[DEPLOYMENT INSTRUCTIONS]"
            echo "1. Copy $PACKAGE_DIR directory to your CentOS 7 server"
            echo "2. Make sure the binary has execute permissions:"
            echo "   chmod +x $PACKAGE_DIR/gateway"
            echo "3. Run the binary: $PACKAGE_DIR/gateway"
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
    go version
    echo ""
    echo "Environment Variables:"
    echo "GOOS: $GOOS"
    echo "GOARCH: $GOARCH"
    echo "CGO_ENABLED: $CGO_ENABLED"
    
    echo ""
    echo "[TIP] Common issues and solutions:"
    echo "1. Make sure GCC compiler is installed (required for CGO)"
    echo "2. For Oracle support, ensure ORACLE_HOME is set and Oracle Instant Client is installed"
    echo "3. Try running 'go clean -cache' and rebuild"
    echo "4. Check if your code is compatible with CentOS 7"
    echo "5. Verify that all dependencies support linux/amd64"
    if [ -n "$ORACLE_HOME" ]; then
        echo ""
        echo "Oracle environment:"
        echo "  ORACLE_HOME: $ORACLE_HOME"
        echo "  CGO_CFLAGS: $CGO_CFLAGS"
        echo "  CGO_LDFLAGS: $CGO_LDFLAGS"
    fi
    exit 1
fi

