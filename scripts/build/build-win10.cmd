@echo off
setlocal EnableDelayedExpansion

:: Gateway Build Script for Windows 10/11
:: Supports optional Oracle build (default: MySQL only, use --oracle to enable Oracle)

title Gateway Build - Windows 10/11

echo ==========================================
echo  Gateway Build for Windows 10/11
echo ==========================================
echo.

:: Parse command line arguments
set INCLUDE_ORACLE=0
set ARG1=%1

if /i "!ARG1!"=="--oracle" set INCLUDE_ORACLE=1
if /i "!ARG1!"=="--all" set INCLUDE_ORACLE=1
if /i "!ARG1!"=="--no-oracle" set INCLUDE_ORACLE=0
if /i "!ARG1!"=="--mysql-only" set INCLUDE_ORACLE=0

if !INCLUDE_ORACLE! EQU 1 (
    echo [INFO] Building with Oracle support
    echo Usage: %~nx0 [--oracle^|--all^|--no-oracle^|--mysql-only]
) else (
    echo [INFO] Building without Oracle support (MySQL only, default^)
    echo Usage: %~nx0 [--oracle^|--all^|--no-oracle^|--mysql-only]
)

:: Show current Go version
for /f "tokens=3" %%v in ('go version') do set CURRENT_GO=%%v
set CURRENT_GO=!CURRENT_GO:go=!
echo Current Go version: !CURRENT_GO!
echo Using default go.mod file
echo.

:: Oracle environment setup (only if Oracle is included)
if !INCLUDE_ORACLE! EQU 1 (
    set "ORACLE_HOME=D:\sdk\instantclient\instantclient_21_18"
    set "PATH=!ORACLE_HOME!;%PATH%"
    set "CGO_CFLAGS=-I!ORACLE_HOME!\sdk\include"
    set "CGO_LDFLAGS=-L!ORACLE_HOME!\sdk\lib\msvc -loci"
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
) else (
    :: No Oracle - disable CGO
    set CGO_ENABLED=0
    echo [INFO] Oracle support disabled, building MySQL-only version
    echo CGO_ENABLED=0
    echo.
)

:: Build configuration
:: Note: Oracle support uses !no_oracle tag (default disabled, use --oracle to enable)
if !INCLUDE_ORACLE! EQU 1 (
    set BUILD_TAGS=netgo,osusergo,windows
    set VERSION_SUFFIX=win10-oracle
) else (
    set BUILD_TAGS=netgo,osusergo,no_oracle,windows
    set VERSION_SUFFIX=win10
)

:: Get to project root
cd /d "%~dp0..\.."
set "PROJECT_ROOT=%CD%"

:: Clean and create output directory
echo Cleaning dist directory...
if exist "dist" (
    rmdir /S /Q "dist" 2>nul
    if errorlevel 1 (
        echo [WARNING] Failed to clean dist directory, some files may be in use
    ) else (
        echo [OK] Dist directory cleaned
    )
)
if not exist "dist" mkdir dist
echo.

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

:: Output file - always use gateway.exe
set OUTPUT_FILE=dist\gateway.exe
set PACKAGE_DIR=dist\gateway
set VERSION_INFO=!VERSION_SUFFIX!-v3.1

:: Build flags with optimizations for modern Windows
:: Note: Use separate -X flags to avoid quote issues with spaces in BUILD_TIMESTAMP
set LDFLAGS=-s -w -X main.Version=!VERSION_INFO! -X "main.BuildTime=!BUILD_TIMESTAMP!" -X main.GitCommit=!GIT_COMMIT!

if !INCLUDE_ORACLE! EQU 1 (
    echo Building with Oracle support...
) else (
    echo Building MySQL-only version...
)
echo Output: !OUTPUT_FILE!
echo Version: !VERSION_INFO!
echo Build Tags: !BUILD_TAGS!
echo.

:: Prepare dependencies
echo Preparing dependencies...
go mod download
go mod tidy

:: Execute build with verbose output
echo.
echo Running build with verbose output...
:: Clear any existing LDFLAGS environment variable that might interfere
if defined LDFLAGS set LDFLAGS=

if !INCLUDE_ORACLE! EQU 1 (
    :: Oracle build requires external linking mode
    :: Combine LDFLAGS with -linkmode external (space-separated, not =)
    set ORACLE_LDFLAGS=!LDFLAGS! -linkmode external
    go build -v -x -tags !BUILD_TAGS! -ldflags "!ORACLE_LDFLAGS!" -o "!OUTPUT_FILE!" cmd\app\main.go
) else (
    :: MySQL-only build uses internal linking (default)
    go build -v -x -tags !BUILD_TAGS! -ldflags "!LDFLAGS!" -o "!OUTPUT_FILE!" cmd\app\main.go
)
set BUILD_RESULT=%ERRORLEVEL%

:: Check if build succeeded
if errorlevel 1 goto build_failed_section

:: Build succeeded, check output file
if not exist "!OUTPUT_FILE!" goto build_failed_section

:: Get file size
for %%F in ("!OUTPUT_FILE!") do set FILE_SIZE=%%~zF
set /a FILE_SIZE_MB=!FILE_SIZE!/1048576

:: Check if file has valid size
if !FILE_SIZE_MB! LEQ 0 goto build_failed_section

:: Build and file check passed
echo.
echo [SUCCESS] Build completed successfully!
echo Output: !OUTPUT_FILE!
echo Size: !FILE_SIZE_MB! MB
echo Go version: !CURRENT_GO!
if !INCLUDE_ORACLE! EQU 1 (
    echo Build type: With Oracle support
) else (
    echo Build type: MySQL-only
)

echo.
echo ==========================================
echo  Packaging deployment structure...
echo ==========================================

:: Create package directory structure
echo Creating directory structure...
if not exist "!PACKAGE_DIR!" mkdir "!PACKAGE_DIR!"
if not exist "!PACKAGE_DIR!\configs" mkdir "!PACKAGE_DIR!\configs"
if not exist "!PACKAGE_DIR!\web" mkdir "!PACKAGE_DIR!\web"
if not exist "!PACKAGE_DIR!\web\static" mkdir "!PACKAGE_DIR!\web\static"
if not exist "!PACKAGE_DIR!\web\frontend" mkdir "!PACKAGE_DIR!\web\frontend"
if not exist "!PACKAGE_DIR!\web\frontend\dist" mkdir "!PACKAGE_DIR!\web\frontend\dist"
if not exist "!PACKAGE_DIR!\logs" mkdir "!PACKAGE_DIR!\logs"
if not exist "!PACKAGE_DIR!\backup" mkdir "!PACKAGE_DIR!\backup"
if not exist "!PACKAGE_DIR!\scripts" mkdir "!PACKAGE_DIR!\scripts"
if not exist "!PACKAGE_DIR!\scripts\db" mkdir "!PACKAGE_DIR!\scripts\db"
if not exist "!PACKAGE_DIR!\scripts\data" mkdir "!PACKAGE_DIR!\scripts\data"
if not exist "!PACKAGE_DIR!\scripts\deploy" mkdir "!PACKAGE_DIR!\scripts\deploy"
if not exist "!PACKAGE_DIR!\pprof_analysis" mkdir "!PACKAGE_DIR!\pprof_analysis"

:: Copy executable file
echo Copying executable file...
copy /Y "!OUTPUT_FILE!" "!PACKAGE_DIR!\gateway.exe" >nul
if errorlevel 1 (
    echo [WARNING] Failed to copy executable file
) else (
    echo [OK] Executable file copied
)

:: Copy configuration files
echo Copying configuration files...
if exist "configs" (
    xcopy /Y /E /I /Q "configs\*" "!PACKAGE_DIR!\configs\" >nul 2>&1
    if errorlevel 1 (
        echo [WARNING] Failed to copy configuration files
    ) else (
        echo [OK] Configuration files copied
    )
) else (
    echo [WARNING] Configuration directory not found
)

:: Copy web static resources
echo Copying web static resources...
if exist "web\static" (
    xcopy /Y /E /I /Q "web\static\*" "!PACKAGE_DIR!\web\static\" >nul 2>&1
    if errorlevel 1 (
        echo [WARNING] Failed to copy web static resources
    ) else (
        echo [OK] Web static resources copied
    )
) else (
    echo [WARNING] Web static directory not found
)

:: Copy frontend dist resources
echo Copying frontend dist resources...
if exist "web\frontend\dist" (
    xcopy /Y /E /I /Q "web\frontend\dist\*" "!PACKAGE_DIR!\web\frontend\dist\" >nul 2>&1
    if errorlevel 1 (
        echo [WARNING] Failed to copy frontend dist resources
    ) else (
        echo [OK] Frontend dist resources copied
    )
) else (
    echo [WARNING] Frontend dist directory not found
    echo [INFO] Please build frontend first: cd web\frontend ^&^& npm run build
)

:: Copy scripts directories
echo Copying scripts directories...

:: Copy db scripts
if exist "scripts\db" (
    xcopy /Y /E /I /Q "scripts\db\*" "!PACKAGE_DIR!\scripts\db\" >nul 2>&1
    if errorlevel 1 (
        echo [WARNING] Failed to copy db scripts
    ) else (
        echo [OK] Database scripts copied
    )
) else (
    echo [WARNING] Database scripts directory not found
)

:: Copy deploy scripts
if exist "scripts\deploy" (
    xcopy /Y /E /I /Q "scripts\deploy\*" "!PACKAGE_DIR!\scripts\deploy\" >nul 2>&1
    if errorlevel 1 (
        echo [WARNING] Failed to copy deploy scripts
    ) else (
        echo [OK] Deploy scripts copied
    )
) else (
    echo [WARNING] Deploy scripts directory not found
)

:: Copy docker scripts
if exist "scripts\docker" (
    xcopy /Y /E /I /Q "scripts\docker\*" "!PACKAGE_DIR!\scripts\docker\" >nul 2>&1
    if errorlevel 1 (
        echo [WARNING] Failed to copy docker scripts
    ) else (
        echo [OK] Docker scripts copied
    )
) else (
    echo [WARNING] Docker scripts directory not found
)

:: Copy k8s scripts
if exist "scripts\k8s" (
    xcopy /Y /E /I /Q "scripts\k8s\*" "!PACKAGE_DIR!\scripts\k8s\" >nul 2>&1
    if errorlevel 1 (
        echo [WARNING] Failed to copy k8s scripts
    ) else (
        echo [OK] K8s scripts copied
    )
) else (
    echo [WARNING] K8s scripts directory not found
)

:: Copy test scripts
if exist "scripts\test" (
    xcopy /Y /E /I /Q "scripts\test\*" "!PACKAGE_DIR!\scripts\test\" >nul 2>&1
    if errorlevel 1 (
        echo [WARNING] Failed to copy test scripts
    ) else (
        echo [OK] Test scripts copied
    )
) else (
    echo [WARNING] Test scripts directory not found
)

:: Note: scripts/data directory is created empty (not copied from source)

echo.
echo ==========================================
echo  Package structure created successfully!
echo ==========================================
echo Package directory: !PACKAGE_DIR!
echo.
echo Directory structure:
dir /B /AD "!PACKAGE_DIR!" 2>nul
echo.

echo Opening output directory...
start "" "dist"
goto build_success_end

:build_failed_section
    echo.
    echo [FAILED] Build failed with error code !BUILD_RESULT!
    echo.
    echo Debug Information:
    echo -----------------
    echo Go Information:
    echo Version: !CURRENT_GO!
    go version
    echo.
    if !INCLUDE_ORACLE! EQU 1 (
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
    ) else (
        echo CGO_ENABLED: !CGO_ENABLED!
    )
    
    echo.
    echo [TIP] Common issues and solutions:
    if !INCLUDE_ORACLE! EQU 1 (
        echo 1. Make sure Oracle Instant Client is properly installed
        echo 2. Verify GCC or Visual Studio Build Tools is installed
        echo 3. Check if PATH includes Oracle and GCC directories
    )
    echo 4. Try running 'go clean -cache' and rebuild
    echo 5. Make sure you're using a compatible Go version
    pause
    exit /b 1

:build_success_end
pause

