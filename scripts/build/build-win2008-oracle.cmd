@echo off
setlocal EnableDelayedExpansion

:: GoHub Oracle Build Script for Windows Server 2008
:: Optimized for legacy Windows environments

title GoHub Oracle Build - Windows Server 2008

echo ==========================================
echo  GoHub Oracle Build for Windows Server 2008
echo ==========================================
echo.

:: Auto-detect Go version first
for /f "tokens=3" %%v in ('go version') do set CURRENT_GO=%%v
set CURRENT_GO=!CURRENT_GO:go=!

:: Extract major version (1.XX) - Windows Server 2008 compatible approach
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
    set GODROR_VERSION=v0.33.0
) else if "!GO_VERSION!"=="1.20" (
    set GO_MOD_FILE=go.mod.1.20
    set GODROR_VERSION=v0.33.0
) else if "!GO_VERSION!"=="1.21" (
    set GO_MOD_FILE=go.mod
    set GODROR_VERSION=v0.40.3
) else if "!GO_VERSION!"=="1.22" (
    set GO_MOD_FILE=go.mod
    set GODROR_VERSION=v0.42.0
) else if "!GO_VERSION!"=="1.23" (
    set GO_MOD_FILE=go.mod
    set GODROR_VERSION=v0.42.0
) else if "!GO_VERSION!"=="1.24" (
    set GO_MOD_FILE=go.mod
    set GODROR_VERSION=v0.42.0
) else (
    echo [INFO] Using default configuration for Go !GO_VERSION!
    set GO_MOD_FILE=go.mod
    
    :: Fix for Windows Server 2008 compatibility - avoid nested variable expansion
    set GO_VER_PREFIX=!GO_VERSION:~0,3!
    set GO_VER_DIGIT=!GO_VERSION:~2,1!
    
    if "!GO_VER_PREFIX!"=="1.2" (
        set /a GO_MINOR=!GO_VER_DIGIT! 2>nul
        if !GO_MINOR! GEQ 2 (
            set GODROR_VERSION=v0.42.0
        ) else (
            set GODROR_VERSION=v0.33.0
        )
    ) else (
        set GODROR_VERSION=v0.33.0
    )
)

echo Using Go version: !GO_VERSION!
echo Module file: !GO_MOD_FILE!
echo Godror version: !GODROR_VERSION!
echo.

:: Oracle environment setup
set ORACLE_HOME=D:\SDK\instantclient_21_18
set PATH=!ORACLE_HOME!;%PATH%
set CGO_CFLAGS=-I!ORACLE_HOME!\sdk\include
set CGO_LDFLAGS=-L!ORACLE_HOME!\sdk\lib\msvc -loci
set CGO_ENABLED=1

:: Environment verification with size checks
echo Checking Oracle environment...
echo ORACLE_HOME: !ORACLE_HOME!
echo CGO_CFLAGS: !CGO_CFLAGS!
echo CGO_LDFLAGS: !CGO_LDFLAGS!
    echo.

echo Checking Oracle files...
set MISSING_FILES=0

:: Check oci.dll with size verification
if exist "!ORACLE_HOME!\oci.dll" (
    for %%F in ("!ORACLE_HOME!\oci.dll") do (
        if %%~zF LSS 800000 (
            echo [WARNING] oci.dll found but size is suspicious (%%~zF bytes^)
            echo Expected size: ~817,152 bytes
            set /a MISSING_FILES+=1
        ) else (
            echo [OK] Found oci.dll (%%~zF bytes^)
        )
    )
        ) else (
    echo [ERROR] oci.dll not found
    set /a MISSING_FILES+=1
)

:: Check oci.h
if exist "!ORACLE_HOME!\sdk\include\oci.h" (
    for %%F in ("!ORACLE_HOME!\sdk\include\oci.h") do (
        if %%~zF LSS 200000 (
            echo [WARNING] oci.h found but size is suspicious (%%~zF bytes^)
            echo Expected size: ~236,440 bytes
            set /a MISSING_FILES+=1
        ) else (
            echo [OK] Found oci.h (%%~zF bytes^)
        )
    )
    ) else (
    echo [ERROR] oci.h not found
    set /a MISSING_FILES+=1
)

:: Check oci.lib
if exist "!ORACLE_HOME!\sdk\lib\msvc\oci.lib" (
    for %%F in ("!ORACLE_HOME!\sdk\lib\msvc\oci.lib") do (
        if %%~zF LSS 800000 (
            echo [WARNING] oci.lib found but size is suspicious (%%~zF bytes^)
            echo Expected size: ~811,436 bytes
            set /a MISSING_FILES+=1
        ) else (
            echo [OK] Found oci.lib (%%~zF bytes^)
        )
    )
) else (
    echo [ERROR] oci.lib not found
    set /a MISSING_FILES+=1
)

if !MISSING_FILES! GTR 0 (
echo.
    echo [ERROR] Some Oracle files are missing or have incorrect sizes
    echo Please check the Oracle Instant Client installation
    echo Download from: https://www.oracle.com/database/technologies/instant-client/winx64-64-downloads.html
    pause
    exit /b 1
)

:: Build configuration
set BUILD_TAGS=netgo,osusergo,ora
set OUTPUT_SUFFIX=oracle
set VERSION_SUFFIX=win2008-oracle

:: Get to project root
set SCRIPT_DIR=%~dp0
set PROJECT_ROOT=%SCRIPT_DIR%..\..
pushd "!PROJECT_ROOT!"

:: Create output directory
if not exist "dist" mkdir dist

:: Build info
set BUILD_DATE=%DATE%
set BUILD_TIME=%TIME%
set BUILD_TIMESTAMP=!BUILD_DATE! !BUILD_TIME!

:: Get Git commit
set GIT_COMMIT=unknown
for /f %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i

:: Build configuration
set GOOS=windows
set GOARCH=amd64

:: Output file
set OUTPUT_FILE=dist\gohub-!VERSION_SUFFIX!-amd64.exe
set VERSION_INFO=!VERSION_SUFFIX!-v3.1

:: Build flags with Windows Server 2008 optimizations
set BUILD_FLAGS=-ldflags="-s -w -X main.Version=!VERSION_INFO! -X main.BuildTime=!BUILD_TIMESTAMP! -X main.GitCommit=!GIT_COMMIT!"

echo Building Oracle version...
echo Output: !OUTPUT_FILE!
echo Version: !VERSION_INFO!
echo Build Tags: !BUILD_TAGS!
echo.

:: Use specified Go version compatible dependencies
echo Using Go !GO_VERSION! compatible dependencies...
if exist "!GO_MOD_FILE!" (
    copy go.mod go.mod.bak >nul
    copy go.sum go.sum.bak >nul
    copy "!GO_MOD_FILE!" go.mod >nul
    if exist go.sum del go.sum
    go mod edit -require=github.com/godror/godror@!GODROR_VERSION!
    go clean -modcache
    go mod download
    go mod tidy -compat=!GO_VERSION!
) else (
    echo [WARNING] !GO_MOD_FILE! not found, using current go.mod
)

:: Execute build with verbose output
echo.
echo Running build with verbose output...
go build -v -x -tags !BUILD_TAGS! !BUILD_FLAGS! -ldflags="-linkmode external" -o "!OUTPUT_FILE!" cmd\app\main.go
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
    echo GCC Version:
    gcc --version
    echo.
    echo Environment Variables:
    echo ORACLE_HOME: !ORACLE_HOME!
    echo CGO_ENABLED: !CGO_ENABLED!
    echo CGO_CFLAGS: !CGO_CFLAGS!
    echo CGO_LDFLAGS: !CGO_LDFLAGS!
    echo.
    echo Oracle Files:
    dir "!ORACLE_HOME!\oci.*"
    dir "!ORACLE_HOME!\sdk\include\oci.h"
    dir "!ORACLE_HOME!\sdk\lib\msvc\oci.lib"
    
    echo.
    echo [TIP] Common issues and solutions:
    echo 1. Make sure Oracle Instant Client is properly installed
    echo 2. Verify GCC or Visual Studio Build Tools is installed
    echo 3. Check if PATH includes Oracle and GCC directories
    echo 4. Try running 'go clean -cache' and rebuild
    echo 5. Make sure you're using the correct Go version (!GO_VERSION!)
    echo 6. For Windows Server 2008, ensure you have all required Windows updates
    pause
    exit /b 1
)

popd
pause
