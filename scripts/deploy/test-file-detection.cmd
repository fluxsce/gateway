@echo off
chcp 65001
setlocal EnableDelayedExpansion

:: GoHub 文件名检测测试脚本
:: 用于验证服务安装脚本的文件名检测功能

title GoHub 文件名检测测试

set APP_DIR=
set ORACLE_VERSION=false

:: 检查是否为Oracle版本
if "%~1"=="oracle" (
    set ORACLE_VERSION=true
    shift
)

if "%~1"=="" (
    echo 用法: %~nx0 [oracle]
    echo.
    echo 示例:
    echo   %~nx0         # 测试标准版本文件检测
    echo   %~nx0 oracle  # 测试Oracle版本文件检测
    echo.
    echo 默认测试标准版本...
    echo.
)

:: 智能检测应用程序目录
call :detect_app_dir

echo.
echo ==========================================
echo  GoHub 文件名检测测试
echo ==========================================
echo.
echo 测试目录: %APP_DIR%
echo 版本类型: %ORACLE_VERSION%
echo.

:: 文件检测逻辑（与服务安装脚本相同）
set EXE_FILE=
if "%ORACLE_VERSION%"=="true" (
    echo 检测Oracle版本可执行文件...
    echo.
    if exist "%APP_DIR%\gohub-win10-oracle-amd64.exe" (
        set EXE_FILE=%APP_DIR%\gohub-win10-oracle-amd64.exe
        echo ✓ 找到: gohub-win10-oracle-amd64.exe
    ) else (
        echo ✗ 未找到: gohub-win10-oracle-amd64.exe
    )
    
    if exist "%APP_DIR%\gohub-win2008-oracle-amd64.exe" (
        if "%EXE_FILE%"=="" set EXE_FILE=%APP_DIR%\gohub-win2008-oracle-amd64.exe
        echo ✓ 找到: gohub-win2008-oracle-amd64.exe
    ) else (
        echo ✗ 未找到: gohub-win2008-oracle-amd64.exe
    )
    
    if exist "%APP_DIR%\gohub-oracle.exe" (
        if "%EXE_FILE%"=="" set EXE_FILE=%APP_DIR%\gohub-oracle.exe
        echo ✓ 找到: gohub-oracle.exe
    ) else (
        echo ✗ 未找到: gohub-oracle.exe
    )
) else (
    echo 检测标准版本可执行文件...
    echo.
    if exist "%APP_DIR%\gohub.exe" (
        set EXE_FILE=%APP_DIR%\gohub.exe
        echo ✓ 找到: gohub.exe
    ) else (
        echo ✗ 未找到: gohub.exe
    )
    
    if exist "%APP_DIR%\gohub-win10-amd64.exe" (
        if "%EXE_FILE%"=="" set EXE_FILE=%APP_DIR%\gohub-win10-amd64.exe
        echo ✓ 找到: gohub-win10-amd64.exe
    ) else (
        echo ✗ 未找到: gohub-win10-amd64.exe
    )
    
    if exist "%APP_DIR%\gohub-win2008-amd64.exe" (
        if "%EXE_FILE%"=="" set EXE_FILE=%APP_DIR%\gohub-win2008-amd64.exe
        echo ✓ 找到: gohub-win2008-amd64.exe
    ) else (
        echo ✗ 未找到: gohub-win2008-amd64.exe
    )
)

echo.
echo ==========================================
echo  检测结果
echo ==========================================
echo.

if "%EXE_FILE%"=="" (
    echo ❌ 未找到可用的可执行文件
    echo.
    if "%ORACLE_VERSION%"=="true" (
        echo 期望的Oracle版本文件名：
        echo   - gohub-win10-oracle-amd64.exe
        echo   - gohub-win2008-oracle-amd64.exe
        echo   - gohub-oracle.exe
    ) else (
        echo 期望的标准版本文件名：
        echo   - gohub.exe
        echo   - gohub-win10-amd64.exe
        echo   - gohub-win2008-amd64.exe
    )
    echo.
    echo 请运行构建脚本生成可执行文件：
    if "%ORACLE_VERSION%"=="true" (
        echo   .\scripts\build\build-win10-oracle.cmd
        echo   .\scripts\build\build-win2008-oracle.cmd
    ) else (
        echo   .\scripts\build\build-win10.cmd
        echo   .\scripts\build\build-win2008.cmd
    )
) else (
    echo ✅ 检测到可执行文件: %EXE_FILE%
    echo.
    echo 文件信息：
    for %%F in ("%EXE_FILE%") do (
        echo   文件名: %%~nxF
        echo   大小: %%~zF 字节
        echo   修改时间: %%~tF
    )
    echo.
    echo 🎉 文件检测成功！可以使用服务安装脚本了
)

echo.
pause
exit /b 0

:: 智能检测应用程序目录
:detect_app_dir
:: 尝试检测应用程序目录
set SCRIPT_DIR=%~dp0
if "%SCRIPT_DIR:~-1%"=="\" set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

:: 方案1: 检查脚本上级目录（适用于源码目录中的脚本，scripts目录）
set CANDIDATE_DIR=%SCRIPT_DIR%\..
call :check_app_files "%CANDIDATE_DIR%"
if not errorlevel 1 (
    set APP_DIR=%CANDIDATE_DIR%
    exit /b 0
)

:: 方案2: 检查项目根目录（适用于源码目录中的脚本，scripts\deploy目录）
set PROJECT_DIR=%SCRIPT_DIR%\..\..
call :check_app_files "%PROJECT_DIR%"
if not errorlevel 1 (
    set APP_DIR=%PROJECT_DIR%
    exit /b 0
)

:: 方案3: 检查脚本当前目录（适用于脚本与程序在同一目录）
call :check_app_files "%SCRIPT_DIR%"
if not errorlevel 1 (
    set APP_DIR=%SCRIPT_DIR%
    exit /b 0
)

:: 如果都没找到，使用默认值进行测试
echo [WARN] 无法检测到应用程序目录，使用项目根目录进行测试
set APP_DIR=%SCRIPT_DIR%\..\..
exit /b 0

:: 检查目录中是否包含GoHub可执行文件
:check_app_files
set CHECK_DIR=%~1
if "%CHECK_DIR:~-1%"=="\" set CHECK_DIR=%CHECK_DIR:~0,-1%

:: 检查是否存在任何GoHub可执行文件
if exist "%CHECK_DIR%\gohub*.exe" exit /b 0
if exist "%CHECK_DIR%\*.exe" (
    :: 进一步检查是否是GoHub相关的exe文件
    for %%f in ("%CHECK_DIR%\*.exe") do (
        set filename=%%~nxf
        echo !filename! | findstr /i "gohub" >nul && exit /b 0
    )
)

:: 如果没找到可执行文件，检查是否有configs目录（可能是正确的应用目录）
if exist "%CHECK_DIR%\configs\app.yaml" exit /b 0
if exist "%CHECK_DIR%\configs\database.yaml" exit /b 0

:: 没找到相关文件
exit /b 1 