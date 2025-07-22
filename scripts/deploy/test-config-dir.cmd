@echo off
chcp 65001 > nul
setlocal enabledelayedexpansion

title GoHub 配置目录测试

echo.
echo ==========================================
echo  GoHub 配置目录测试
echo ==========================================
echo.

:: 获取脚本目录
set SCRIPT_DIR=%~dp0
if "%SCRIPT_DIR:~-1%"=="\" set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

:: 获取项目根目录
set PROJECT_DIR=%SCRIPT_DIR%\..\..
if "%PROJECT_DIR:~-1%"=="\" set PROJECT_DIR=%PROJECT_DIR:~0,-1%

echo 脚本目录: %SCRIPT_DIR%
echo 项目目录: %PROJECT_DIR%
echo.

:: 测试场景1：使用默认配置目录
echo [测试1] 使用默认配置目录
set GOHUB_CONFIG_DIR=
set TEST_DIR=%PROJECT_DIR%\configs
echo 期望目录: %TEST_DIR%
if exist "%TEST_DIR%" (
    echo ✓ 默认配置目录存在
    dir /b "%TEST_DIR%\*.yaml" 2>nul | find /c ".yaml" > temp_count.txt
    set /p FILE_COUNT=<temp_count.txt
    del temp_count.txt
    echo   发现 !FILE_COUNT! 个YAML配置文件
) else (
    echo ✗ 默认配置目录不存在
)
echo.

:: 测试场景2：使用环境变量指定配置目录
echo [测试2] 使用环境变量指定配置目录
set GOHUB_CONFIG_DIR=%PROJECT_DIR%\configs
echo 设置环境变量: GOHUB_CONFIG_DIR=%GOHUB_CONFIG_DIR%
echo 期望目录: %GOHUB_CONFIG_DIR%
if exist "%GOHUB_CONFIG_DIR%" (
    echo ✓ 环境变量配置目录存在
    dir /b "%GOHUB_CONFIG_DIR%\*.yaml" 2>nul | find /c ".yaml" > temp_count.txt
    set /p FILE_COUNT=<temp_count.txt
    del temp_count.txt
    echo   发现 !FILE_COUNT! 个YAML配置文件
) else (
    echo ✗ 环境变量配置目录不存在
)
echo.

:: 测试场景3：使用自定义配置目录
echo [测试3] 使用自定义配置目录（模拟生产环境）
set GOHUB_CONFIG_DIR=C:\Program Files\GoHub\configs
echo 设置环境变量: GOHUB_CONFIG_DIR=%GOHUB_CONFIG_DIR%
echo 期望目录: %GOHUB_CONFIG_DIR%
if exist "%GOHUB_CONFIG_DIR%" (
    echo ✓ 自定义配置目录存在
    dir /b "%GOHUB_CONFIG_DIR%\*.yaml" 2>nul | find /c ".yaml" > temp_count.txt
    set /p FILE_COUNT=<temp_count.txt
    del temp_count.txt
    echo   发现 !FILE_COUNT! 个YAML配置文件
) else (
    echo ✗ 自定义配置目录不存在（这是正常的，除非您已经部署到生产环境）
)
echo.

:: 检查关键配置文件
echo [检查] 关键配置文件
set GOHUB_CONFIG_DIR=%PROJECT_DIR%\configs
set CONFIG_FILES=app.yaml database.yaml logger.yaml web.yaml gateway.yaml
for %%f in (%CONFIG_FILES%) do (
    if exist "%GOHUB_CONFIG_DIR%\%%f" (
        echo ✓ %%f 存在
    ) else (
        echo ✗ %%f 不存在
    )
)
echo.

:: 配置目录优先级说明
echo [说明] 配置目录优先级
echo 1. GOHUB_CONFIG_DIR 环境变量指定的目录
echo 2. ./configs （相对于程序启动目录）
echo 3. . （程序启动目录）
echo.

:: 修复建议
echo [建议] 配置目录使用最佳实践
echo 1. 开发环境：使用默认的 ./configs 目录
echo 2. 生产环境：设置 GOHUB_CONFIG_DIR 环境变量
echo 3. 容器部署：在容器中设置环境变量
echo 4. Windows服务：在服务安装脚本中设置环境变量
echo.

echo ==========================================
echo 测试完成！
echo ==========================================
echo.
pause 