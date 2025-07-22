@echo off
setlocal EnableDelayedExpansion

:: Gateway CentOS 7 Build Script for Windows 10/11
:: Optimized for cross-compilation to CentOS 7

title Gateway CentOS 7 Build - Windows 10/11

echo ==========================================
echo  Gateway CentOS 7 Build for Windows 10/11
echo ==========================================
echo.

:: Auto-detect Go version first
for /f "tokens=3" %%v in ('go version') do set CURRENT_GO=%%v
set CURRENT_GO=!CURRENT_GO:go=!

:: Extract major version (1.XX)
set GO_MAJOR=!CURRENT_GO:~0,1!
set GO_MINOR_FULL=!CURRENT_GO:~2!
for /f "delims=. tokens=1" %%a in ("!GO_MINOR_FULL!") do set GO_MINOR=%%a
set CURRENT_GO_MAJOR=!GO_MAJOR!.!GO_MINOR!

:: Go version selection
set GO_VERSION=!CURRENT_GO_MAJOR!
echo Auto-detected Go version: !CURRENT_GO! (Using !GO_VERSION! for build)

:: Check if version parameter is provided to override auto-detection
if not "%1"=="" (
    set GO_VERSION=%1
    echo Overriding with specified Go version: !GO_VERSION!
)

:: Validate Go version and set module file
if "!GO_VERSION!"=="1.19" (
    set GO_MOD_FILE=go.mod.1.19
) else if "!GO_VERSION!"=="1.20" (
    set GO_MOD_FILE=go.mod.1.20
) else if "!GO_VERSION!"=="1.21" (
    set GO_MOD_FILE=go.mod
) else if "!GO_VERSION!"=="1.22" (
    set GO_MOD_FILE=go.mod
) else if "!GO_VERSION!"=="1.23" (
    set GO_MOD_FILE=go.mod
) else if "!GO_VERSION!"=="1.24" (
    set GO_MOD_FILE=go.mod
) else (
    echo [INFO] Using default configuration for Go !GO_VERSION!
    set GO_MOD_FILE=go.mod
)

echo Using Go version: !GO_VERSION!
echo Module file: !GO_MOD_FILE!
echo.

:: Build configuration for CentOS 7
set BUILD_TAGS=netgo,osusergo,no_oracle
set OUTPUT_SUFFIX=centos7
set VERSION_SUFFIX=centos7

:: Cross-compilation settings for CentOS 7
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

:: Get to project root
set SCRIPT_DIR=%~dp0
set PROJECT_ROOT=%SCRIPT_DIR%..\..
pushd "!PROJECT_ROOT!"

:: Create output directory
if not exist "dist" mkdir dist

:: Build info
for /f "tokens=2 delims==" %%a in ('wmic os get localdatetime /value') do set datetime=%%a
set BUILD_DATE=%datetime:~0,4%-%datetime:~4,2%-%datetime:~6,2%
set BUILD_TIME=%datetime:~8,2%:%datetime:~10,2%:%datetime:~12,2%
set BUILD_TIMESTAMP=!BUILD_DATE!T!BUILD_TIME!

:: Get Git commit
set GIT_COMMIT=unknown
for /f %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i

:: Output file
set OUTPUT_FILE=dist\gateway-!VERSION_SUFFIX!-amd64
set VERSION_INFO=!VERSION_SUFFIX!-v3.1

:: Build flags optimized for CentOS 7
set BUILD_FLAGS=-ldflags="-s -w -X main.Version=!VERSION_INFO! -X main.BuildTime=!BUILD_TIMESTAMP! -X main.GitCommit=!GIT_COMMIT!"

echo Building CentOS 7 version...
echo Output: !OUTPUT_FILE!
echo Version: !VERSION_INFO!
echo Build Tags: !BUILD_TAGS!
echo Cross-compilation: GOOS=!GOOS! GOARCH=!GOARCH!
echo.

:: Use specified Go version compatible dependencies
echo Using Go !GO_VERSION! compatible dependencies...
if exist "!GO_MOD_FILE!" (
    copy go.mod go.mod.bak >nul
    copy go.sum go.sum.bak >nul
    copy "!GO_MOD_FILE!" go.mod >nul
    if exist go.sum del go.sum
    go clean -modcache
    go mod download
    go mod tidy -compat=!GO_VERSION!
) else (
    echo [WARNING] !GO_MOD_FILE! not found, using current go.mod
)

:: Execute build with verbose output
echo.
echo Running build...
go build -v -x -tags !BUILD_TAGS! !BUILD_FLAGS! -o "!OUTPUT_FILE!" cmd\app\main.go
set BUILD_RESULT=%ERRORLEVEL%

:: Check if the output file exists and has size
if !BUILD_RESULT! EQU 0 (
    if exist "!OUTPUT_FILE!" (
        for %%F in ("!OUTPUT_FILE!") do (
            set /a FILE_SIZE_MB=%%~zF/1048576
            if !FILE_SIZE_MB! GTR 0 (
                echo.
                echo [SUCCESS] Build completed successfully!
                echo Output: !OUTPUT_FILE!
                echo Size: !FILE_SIZE_MB! MB
                echo Go version: !CURRENT_GO!
                echo Module file: !GO_MOD_FILE!
                
                echo.
                echo Opening output directory...
                start "" "dist"
                
                :: Restore original go.mod and go.sum
                echo.
                echo Restoring original module files...
                copy go.mod.bak go.mod /y >nul
                if exist go.sum.bak copy go.sum.bak go.sum /y >nul
                del go.mod.bak >nul
                if exist go.sum.bak del go.sum.bak >nul
                echo Module files restored.
                
                echo.
                echo [DEPLOYMENT INSTRUCTIONS]
                echo 1. Copy !OUTPUT_FILE! to your CentOS 7 server
                echo 2. Make sure the binary has execute permissions:
                echo    chmod +x !OUTPUT_FILE!
                
                echo.
                echo [TIP] To use Go !GO_VERSION! compatible dependencies permanently:
                echo copy !GO_MOD_FILE! go.mod
                echo go mod tidy -compat=!GO_VERSION!
            ) else (
                echo [ERROR] Build output file exists but has zero size
                goto :build_failed
            )
        )
    ) else (
        echo [ERROR] Build completed but output file not found
        goto :build_failed
    )
) else (
    :build_failed
    echo.
    echo [FAILED] Build failed with error code !BUILD_RESULT!
    echo.
    echo Debug Information:
    echo -----------------
    echo Go Information:
    echo Version: !CURRENT_GO!
    echo Module file: !GO_MOD_FILE!
    go version
    echo.
    echo Environment Variables:
    echo GOOS: !GOOS!
    echo GOARCH: !GOARCH!
    echo CGO_ENABLED: !CGO_ENABLED!
    
    echo.
    echo [TIP] Common issues and solutions:
    echo 1. Make sure you have the correct Go version (!GO_VERSION!)
    echo 2. Try running 'go clean -cache' and rebuild
    echo 3. Check if your code is compatible with CentOS 7
    echo 4. Verify that all dependencies support linux/amd64
    pause
    exit /b 1
)

popd
pause 