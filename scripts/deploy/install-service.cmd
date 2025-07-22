@echo off
chcp 65001
setlocal EnableDelayedExpansion

:: 设置错误处理 - 即使有错误也不退出
set ORIGINAL_ERRORLEVEL=%errorlevel%

:: GoHub Windows 服务安装脚本
:: 用法: install-service.cmd [options]

title GoHub Windows 服务安装

:: 显示调试信息
echo.
echo ==========================================
echo  GoHub Windows 服务安装调试信息
echo ==========================================
echo.
echo [DEBUG] 脚本路径: %~dp0
echo [DEBUG] 脚本名称: %~nx0
echo [DEBUG] 当前目录: %CD%
echo [DEBUG] 传入参数: %*
echo [DEBUG] 系统时间: %date% %time%
echo.

:: 检查管理员权限
echo [INFO] 检查管理员权限...
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo.
    echo [ERROR] 此脚本需要管理员权限运行
    echo 请右键选择"以管理员身份运行"
    echo.
    echo [DEBUG] 按任意键继续以调试模式运行（功能受限）...
    pause
    set DEBUG_MODE=true
    echo [WARN] 继续运行调试模式...
) else (
    echo [INFO] ✓ 管理员权限检查通过
    set DEBUG_MODE=false
)

:: 默认配置
set SERVICE_NAME=GoHub
set SERVICE_DISPLAY_NAME=GoHub Application Service
set SERVICE_DESCRIPTION=GoHub API Gateway and Web Application Service
set ORACLE_VERSION=false

:: 检查是否为Oracle版本
if "%~1"=="oracle" (
    set ORACLE_VERSION=true
    set SERVICE_NAME=GoHub
    set SERVICE_DISPLAY_NAME=GoHub Application Service
    set SERVICE_DESCRIPTION=GoHub API Gateway and Web Application Service
    shift
)

:: 解析其他参数
:parse_args
if "%~1"=="" goto :args_done
if "%~1"=="-d" (
    set APP_DIR=%~2
    shift & shift
    goto :parse_args
)
if "%~1"=="--dir" (
    set APP_DIR=%~2
    shift & shift
    goto :parse_args
)
if "%~1"=="-c" (
    set CONFIG_DIR=%~2
    shift & shift
    goto :parse_args
)
if "%~1"=="--config" (
    set CONFIG_DIR=%~2
    shift & shift
    goto :parse_args
)
if "%~1"=="-h" goto :show_help
if "%~1"=="--help" goto :show_help
echo [ERROR] 未知参数: %~1
echo [DEBUG] 继续执行而不是退出...
goto :show_help

:args_done

:: 显示帮助
:show_help
echo.
echo GoHub Windows 服务安装脚本
echo.
echo 用法: %~nx0 [oracle] [OPTIONS]
echo.
echo 参数:
echo   oracle                  安装Oracle版本服务
echo.
echo 选项:
echo   -d, --dir DIR          应用程序目录 (默认: 自动检测)
echo   -c, --config DIR       配置文件目录 (默认: 自动检测)
echo   -h, --help             显示帮助信息
echo.
echo 示例:
echo   %~nx0                           # 安装标准版本服务
echo   %~nx0 oracle                    # 安装Oracle版本服务
echo   %~nx0 -d "C:\Program Files\GoHub"  # 指定安装目录
echo.
echo 注意:
echo   - 脚本会自动检测Oracle版本可执行文件
echo   - 所有版本都使用统一的服务名称GoHub
echo.
if "%~1"=="-h" (
    pause
    exit /b 0
)
if "%~1"=="--help" (
    pause
    exit /b 0
)
echo [DEBUG] 参数错误，但继续执行调试...
echo.

:: 智能检测应用程序目录和可执行文件
echo [INFO] 开始检测应用程序目录和可执行文件...
call :detect_app_and_exe
if errorlevel 1 (
    echo [ERROR] 应用程序检测失败
    pause
    exit /b 1
)

:: 根据可执行文件路径自动设置其他路径
:: 转换为绝对路径
for %%i in ("%EXE_FILE%") do set EXE_FILE=%%~fi
for %%i in ("%EXE_FILE%") do set EXE_DIR=%%~dpi
set EXE_DIR=%EXE_DIR:~0,-1%

:: 如果没有指定配置目录，使用可执行文件目录下的configs
if "%CONFIG_DIR%"=="" (
    set CONFIG_DIR=%EXE_DIR%\configs
    echo [INFO] 使用默认配置目录: %CONFIG_DIR%
)

:: 转换为绝对路径
for %%i in ("%CONFIG_DIR%") do set CONFIG_DIR=%%~fi

:: 设置日志目录
set LOG_DIR=%EXE_DIR%\logs

echo.
echo ==========================================
echo  GoHub Windows 服务安装
echo ==========================================
echo.
echo 服务名称: %SERVICE_NAME%
echo 显示名称: %SERVICE_DISPLAY_NAME%
echo 可执行文件: %EXE_FILE%
echo 配置目录: %CONFIG_DIR%
echo 日志目录: %LOG_DIR%
echo.

:: 检查服务是否已存在
echo [DEBUG] 检查服务是否已存在...
sc query "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% equ 0 (
    echo [WARN] 服务 '%SERVICE_NAME%' 已存在
    echo [DEBUG] 显示现有服务信息...
    sc query "%SERVICE_NAME%"
    echo.
    set /p CHOICE="是否要重新安装服务？(y/N): "
    if /i "!CHOICE!" neq "y" (
        echo 取消安装
        echo [DEBUG] 用户选择不重新安装
        pause
        exit /b 0
    )
    
    echo [INFO] 停止现有服务...
    sc stop "%SERVICE_NAME%" >nul 2>&1
    timeout /t 3 /nobreak >nul
    
    echo [INFO] 删除现有服务...
    sc delete "%SERVICE_NAME%" >nul 2>&1
    timeout /t 2 /nobreak >nul
)

:: 创建必要目录
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

:: 安装服务
echo [INFO] 正在安装服务...
echo [DEBUG] 执行服务创建命令...
echo [DEBUG] 命令: sc create "%SERVICE_NAME%" binPath= "\"%EXE_FILE%\" --config=\"%CONFIG_DIR%\" --service" DisplayName= "%SERVICE_DISPLAY_NAME%" start= auto type= own

:: 使用sc命令创建服务
sc create "%SERVICE_NAME%" ^
    binPath= "\"%EXE_FILE%\" --config=\"%CONFIG_DIR%\" --service" ^
    DisplayName= "%SERVICE_DISPLAY_NAME%" ^
    start= auto ^
    type= own

if %errorlevel% neq 0 (
    echo [ERROR] 服务创建失败
    echo [DEBUG] 错误码: %errorlevel%
    echo [DEBUG] 可能的原因：
    echo   1. 服务名称已存在
    echo   2. 权限不足
    echo   3. 可执行文件路径错误
    echo   4. 命令语法错误
    echo.
    echo [DEBUG] 继续执行后续步骤...
    pause
)

:: 设置服务描述
sc description "%SERVICE_NAME%" "%SERVICE_DESCRIPTION%"

:: 设置服务恢复选项（失败时重启）
sc failure "%SERVICE_NAME%" reset= 86400 actions= restart/5000/restart/5000/restart/5000

:: 设置服务环境变量
echo [INFO] 设置服务环境变量...
echo [DEBUG] 创建环境变量注册表项...
reg add "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%\Environment" /f >nul 2>&1
if %errorlevel% neq 0 (
    echo [DEBUG] 环境变量注册表项创建失败，错误码: %errorlevel%
) else (
    echo [DEBUG] ✓ 环境变量注册表项创建成功
)

echo [DEBUG] 设置GOHUB_CONFIG_DIR环境变量...
reg add "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%\Environment" /v GOHUB_CONFIG_DIR /t REG_SZ /d "%CONFIG_DIR%" /f >nul 2>&1
if %errorlevel% neq 0 (
    echo [DEBUG] GOHUB_CONFIG_DIR环境变量设置失败，错误码: %errorlevel%
) else (
    echo [DEBUG] ✓ GOHUB_CONFIG_DIR环境变量设置成功
)

:: 设置日志重定向环境变量
echo [DEBUG] 设置日志重定向环境变量...
reg add "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%\Environment" /v GOHUB_LOG_FILE /t REG_SZ /d "%LOG_DIR%\service.log" /f >nul 2>&1
if %errorlevel% neq 0 (
    echo [DEBUG] GOHUB_LOG_FILE环境变量设置失败，错误码: %errorlevel%
) else (
    echo [DEBUG] ✓ GOHUB_LOG_FILE环境变量设置成功
)

reg add "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%\Environment" /v GOHUB_ERROR_LOG_FILE /t REG_SZ /d "%LOG_DIR%\service-error.log" /f >nul 2>&1
if %errorlevel% neq 0 (
    echo [DEBUG] GOHUB_ERROR_LOG_FILE环境变量设置失败，错误码: %errorlevel%
) else (
    echo [DEBUG] ✓ GOHUB_ERROR_LOG_FILE环境变量设置成功
)

echo [INFO] 服务安装成功！
echo.
echo 服务信息：
echo   服务名称: %SERVICE_NAME%
echo   显示名称: %SERVICE_DISPLAY_NAME%
echo   可执行文件: %EXE_FILE%
echo   配置目录: %CONFIG_DIR%
echo   日志目录: %LOG_DIR%
echo   日志文件: %LOG_DIR%\service.log
echo   错误日志: %LOG_DIR%\service-error.log
echo   环境变量: GOHUB_CONFIG_DIR=%CONFIG_DIR%
echo   日志环境变量: GOHUB_LOG_FILE=%LOG_DIR%\service.log
echo   Oracle版本: %ORACLE_VERSION%
echo.

:: 询问是否立即启动服务
set /p START_NOW="是否立即启动服务？(Y/n): "
if /i "%START_NOW%"=="n" goto :end

echo [INFO] 正在启动服务...
echo [DEBUG] 执行服务启动命令: sc start "%SERVICE_NAME%"
sc start "%SERVICE_NAME%"
set START_RESULT=%errorlevel%

if %START_RESULT% equ 0 (
    echo [INFO] 服务启动成功！
    echo [DEBUG] 检查服务状态...
    sc query "%SERVICE_NAME%"
    echo.
    echo 服务管理命令：
    echo   启动服务: sc start "%SERVICE_NAME%"
    echo   停止服务: sc stop "%SERVICE_NAME%"
    echo   查看状态: sc query "%SERVICE_NAME%"
    echo   删除服务: sc delete "%SERVICE_NAME%"
    echo   查看日志: type "%LOG_DIR%\service.log"
    echo   查看错误: type "%LOG_DIR%\service-error.log"
    echo.
    echo 或者使用 Windows 服务管理器 (services.msc) 进行图形化管理
) else (
    echo [ERROR] 服务启动失败
    echo [DEBUG] 服务启动错误码: %START_RESULT%
    echo [DEBUG] 检查服务状态...
    sc query "%SERVICE_NAME%"
    echo.
    echo 请检查配置文件和日志文件
    echo 日志目录: %LOG_DIR%
    echo.
    echo [DEBUG] 可能的启动失败原因：
    echo   1. 可执行文件不存在或无法访问
    echo   2. 配置文件有误
    echo   3. 端口被占用
    echo   4. 数据库连接失败
    echo   5. 权限不足
)

:end
echo.
echo ==========================================
echo  安装过程完成！
echo ==========================================
echo.
echo [DEBUG] 最终状态总结：
echo   服务名称: %SERVICE_NAME%
echo   显示名称: %SERVICE_DISPLAY_NAME%
echo   可执行文件: %EXE_FILE%
echo   配置目录: %CONFIG_DIR%
echo   日志目录: %LOG_DIR%
echo   调试模式: %DEBUG_MODE%
echo.
echo [DEBUG] 验证服务状态...
if "%DEBUG_MODE%"=="false" (
    sc query "%SERVICE_NAME%" 2>nul | findstr STATE
) else (
    echo [DEBUG] 跳过服务状态检查（权限限制）
)
echo.
echo [DEBUG] 验证环境变量...
if "%DEBUG_MODE%"=="false" (
    reg query "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%\Environment" /v GOHUB_CONFIG_DIR 2>nul | findstr "GOHUB_CONFIG_DIR" || echo [DEBUG] 环境变量未设置
) else (
    echo [DEBUG] 跳过环境变量检查（权限限制）
)
echo.
echo [INFO] 如需查看详细日志，请检查：
echo   - Windows事件日志（Windows Logs -> Application）
echo   - 应用程序日志目录：%LOG_DIR%
echo.
echo [INFO] 按任意键退出...
pause
exit /b 0

:: 智能检测应用程序目录和可执行文件
:detect_app_and_exe
:: 如果已经通过参数指定了APP_DIR，则直接使用
if not "%APP_DIR%"=="" (
    call :find_exe_in_dir "%APP_DIR%"
    exit /b %errorlevel%
)

:: 尝试检测应用程序目录
set SCRIPT_DIR=%~dp0
if "%SCRIPT_DIR:~-1%"=="\" set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

:: 方案1: 检查脚本上级目录（适用于源码目录中的脚本）
set CANDIDATE_DIR=%SCRIPT_DIR%\..
call :find_exe_in_dir "%CANDIDATE_DIR%"
if not errorlevel 1 (
    set APP_DIR=%CANDIDATE_DIR%
    exit /b 0
)

:: 方案2: 检查项目根目录（适用于源码目录中的脚本，scripts\deploy目录）
set PROJECT_DIR=%SCRIPT_DIR%\..\..
call :find_exe_in_dir "%PROJECT_DIR%"
if not errorlevel 1 (
    set APP_DIR=%PROJECT_DIR%
    exit /b 0
)

:: 方案3: 检查脚本当前目录（适用于脚本与程序在同一目录）
call :find_exe_in_dir "%SCRIPT_DIR%"
if not errorlevel 1 (
    set APP_DIR=%SCRIPT_DIR%
    exit /b 0
)

:: 如果都没找到，提示用户
echo.
echo [ERROR] 无法自动检测应用程序目录
echo.
echo 脚本目录: %SCRIPT_DIR%
echo.
echo 已检查以下位置：
echo   1. %CANDIDATE_DIR%
echo   2. %PROJECT_DIR%
echo   3. %SCRIPT_DIR%
echo.
echo [DEBUG] 列出各检测位置的内容：
echo.
echo [DEBUG] 检查位置1: %CANDIDATE_DIR%
if exist "%CANDIDATE_DIR%" (
    dir /b "%CANDIDATE_DIR%\*.exe" 2>nul | findstr /i "gohub" || echo   - 未找到GoHub相关exe文件
) else (
    echo   - 目录不存在
)
echo.
echo [DEBUG] 检查位置2: %PROJECT_DIR%
if exist "%PROJECT_DIR%" (
    dir /b "%PROJECT_DIR%\*.exe" 2>nul | findstr /i "gohub" || echo   - 未找到GoHub相关exe文件
) else (
    echo   - 目录不存在
)
echo.
echo [DEBUG] 检查位置3: %SCRIPT_DIR%
if exist "%SCRIPT_DIR%" (
    dir /b "%SCRIPT_DIR%\*.exe" 2>nul | findstr /i "gohub" || echo   - 未找到GoHub相关exe文件
) else (
    echo   - 目录不存在
)
echo.
echo 请使用 -d 参数指定应用程序目录：
echo   %~nx0 -d "C:\Program Files\GoHub"
echo.
echo [DEBUG] 继续执行并使用脚本目录作为默认值...
pause
exit /b 1

:: 在指定目录中查找可执行文件
:find_exe_in_dir
set CHECK_DIR=%~1
if "%CHECK_DIR:~-1%"=="\" set CHECK_DIR=%CHECK_DIR:~0,-1%

:: 转换为绝对路径
for %%i in ("%CHECK_DIR%") do set CHECK_DIR=%%~fi

:: 首先检查Oracle版本的文件（如果指定了oracle参数）
if "%ORACLE_VERSION%"=="true" (
    if exist "%CHECK_DIR%\gohub-win10-oracle-amd64.exe" (
        set EXE_FILE=%CHECK_DIR%\gohub-win10-oracle-amd64.exe
        exit /b 0
    ) else if exist "%CHECK_DIR%\gohub-win2008-oracle-amd64.exe" (
        set EXE_FILE=%CHECK_DIR%\gohub-win2008-oracle-amd64.exe
        exit /b 0
    ) else if exist "%CHECK_DIR%\gohub-oracle.exe" (
        set EXE_FILE=%CHECK_DIR%\gohub-oracle.exe
        exit /b 0
    )
)

:: 如果没有找到Oracle版本文件，检查标准版本文件
if exist "%CHECK_DIR%\gohub.exe" (
    set EXE_FILE=%CHECK_DIR%\gohub.exe
    exit /b 0
) else if exist "%CHECK_DIR%\gohub-win10-amd64.exe" (
    set EXE_FILE=%CHECK_DIR%\gohub-win10-amd64.exe
    exit /b 0
) else if exist "%CHECK_DIR%\gohub-win2008-amd64.exe" (
    set EXE_FILE=%CHECK_DIR%\gohub-win2008-amd64.exe
    exit /b 0
)

:: 如果还是没有找到，尝试自动检测Oracle版本
if exist "%CHECK_DIR%\gohub-win2008-oracle-amd64.exe" (
    set EXE_FILE=%CHECK_DIR%\gohub-win2008-oracle-amd64.exe
    set ORACLE_VERSION=true
    set SERVICE_NAME=GoHub
    set SERVICE_DISPLAY_NAME=GoHub Application Service
    echo [INFO] 自动检测到Oracle版本可执行文件
    exit /b 0
) else if exist "%CHECK_DIR%\gohub-win10-oracle-amd64.exe" (
    set EXE_FILE=%CHECK_DIR%\gohub-win10-oracle-amd64.exe
    set ORACLE_VERSION=true
    set SERVICE_NAME=GoHub
    set SERVICE_DISPLAY_NAME=GoHub Application Service
    echo [INFO] 自动检测到Oracle版本可执行文件
    exit /b 0
) else if exist "%CHECK_DIR%\gohub-oracle.exe" (
    set EXE_FILE=%CHECK_DIR%\gohub-oracle.exe
    set ORACLE_VERSION=true
    set SERVICE_NAME=GoHub
    set SERVICE_DISPLAY_NAME=GoHub Application Service
    echo [INFO] 自动检测到Oracle版本可执行文件
    exit /b 0
)

:: 没找到相关文件
exit /b 1 