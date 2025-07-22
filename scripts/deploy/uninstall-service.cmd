@echo off
chcp 65001
setlocal EnableDelayedExpansion

:: 设置错误处理 - 即使有错误也不退出
set ORIGINAL_ERRORLEVEL=%errorlevel%

:: Gateway Windows 服务卸载脚本
:: 用法: uninstall-service.cmd [options]

title Gateway Windows 服务卸载

:: 显示调试信息
echo.
echo ==========================================
echo  Gateway Windows 服务卸载调试信息
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
    echo [INFO] [OK] 管理员权限检查通过
    set DEBUG_MODE=false
)

:: 默认配置
set SERVICE_NAME=Gateway
set ORACLE_VERSION=false
set CLEAN_LOGS=false
set CLEAN_ENV=true



:: 检查是否为Oracle版本
if "%~1"=="oracle" (
    set ORACLE_VERSION=true
    set SERVICE_NAME=Gateway
    shift
)

:: 解析其他参数
:parse_args
if "%~1"=="" goto :args_done
if "%~1"=="-s" (
    set SERVICE_NAME=%~2
    shift & shift
    goto :parse_args
)
if "%~1"=="--service" (
    set SERVICE_NAME=%~2
    shift & shift
    goto :parse_args
)
if "%~1"=="--clean-logs" (
    set CLEAN_LOGS=true
    shift
    goto :parse_args
)
if "%~1"=="--keep-env" (
    set CLEAN_ENV=false
    shift
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
echo Gateway Windows 服务卸载脚本
echo.
echo 用法: %~nx0 [oracle] [OPTIONS]
echo.
echo 参数:
echo   oracle                  卸载Oracle版本服务
echo.
echo 选项:
echo   -s, --service NAME     服务名称 (默认: %SERVICE_NAME%)
echo   --clean-logs           同时删除日志文件
echo   --keep-env             保留环境变量设置

echo   -h, --help             显示帮助信息
echo.
echo 注意:
echo   - 脚本会自动检测Oracle版本服务
echo   - 所有版本都使用统一的服务名称Gateway
echo   - 兼容旧版本的Gateway-Oracle服务清理
echo   - 如果检测到多个Gateway服务，会自动卸载所有相关服务
echo   - 默认自动确认，无需手动输入
echo.
echo 示例:
echo   %~nx0                           # 卸载标准版本服务
echo   %~nx0 oracle                    # 卸载Oracle版本服务
echo   %~nx0 oracle --clean-logs       # 卸载Oracle服务并清理日志
echo   %~nx0 -s "Custom-Service"       # 卸载自定义服务名
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

echo.
echo ==========================================
echo  Gateway Windows 服务卸载
echo ==========================================
echo.
echo 服务名称: %SERVICE_NAME%
echo 清理日志: %CLEAN_LOGS%
echo 清理环境变量: %CLEAN_ENV%
echo.

:: 自动检测Oracle版本服务（保持兼容性）
if "%ORACLE_VERSION%"=="false" (
    echo [INFO] 检查是否有旧版本的Oracle服务...
    sc query "Gateway-Oracle" >nul 2>&1
    if %errorlevel% equ 0 (
        echo [INFO] 检测到旧版本Oracle服务，切换到Gateway-Oracle进行清理
        set SERVICE_NAME=Gateway-Oracle
        set ORACLE_VERSION=true
    )
)

:: 如果两个服务都存在，自动卸载所有相关服务
sc query "Gateway" >nul 2>&1
set GO_HUB_EXISTS=%errorlevel%
sc query "Gateway-Oracle" >nul 2>&1
set GO_HUB_ORACLE_EXISTS=%errorlevel%

if %GO_HUB_EXISTS% equ 0 if %GO_HUB_ORACLE_EXISTS% equ 0 (
    echo [INFO] 检测到系统中同时存在两个Gateway服务，将自动卸载所有相关服务
    echo [INFO] 首先卸载: %SERVICE_NAME%
)

:: 检查服务是否存在
echo [INFO] Checking if service exists...
sc query "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARN] 服务 '%SERVICE_NAME%' 不存在或已被删除
    echo.
    
    echo [DEBUG] 尝试列出所有相关服务...
    if "%DEBUG_MODE%"=="false" (
        sc query type= service | findstr /i "gateway" || echo [DEBUG] 未找到任何Gateway相关服务
    ) else (
        echo [DEBUG] 跳过服务列表检查（权限限制）
    )
    echo.
    
    echo [DEBUG] 检查注册表中的服务配置...
    reg query "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%" >nul 2>&1
    if %errorlevel% equ 0 (
        echo [DEBUG] 找到注册表项，但服务查询失败
        echo [DEBUG] 尝试强制删除注册表项...
        reg delete "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%" /f >nul 2>&1
        if %errorlevel% equ 0 (
            echo [INFO] 成功清理注册表项
        ) else (
            echo [WARN] 注册表项清理失败
        )
    ) else (
        echo [DEBUG] 注册表中也没有找到服务配置
    )
    echo.
    
    :: 即使服务不存在，也尝试清理残留的环境变量
    if "%CLEAN_ENV%"=="true" (
        echo [INFO] 清理可能的残留环境变量...
        call :cleanup_environment
    )
    
    echo [INFO] 卸载完成
    echo [INFO] 建议运行 debug-uninstall.cmd 进行详细诊断
    pause
    exit /b 0
)

:: 显示服务状态
echo [INFO] Current service status:
if "%DEBUG_MODE%"=="false" (
    sc query "%SERVICE_NAME%"
) else (
    echo [DEBUG] 跳过服务状态显示（权限限制）
)
echo.

:: 自动确认卸载
echo [INFO] Auto confirming uninstall for service '%SERVICE_NAME%'
echo.

:: 停止服务
echo [INFO] Stopping service...
sc stop "%SERVICE_NAME%" >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARN] Service stop failed, but continue uninstall...
) else (
    echo [INFO] Service stopped
)

:: 等待服务完全停止
echo [INFO] Waiting for service to stop completely...
timeout /t 5 /nobreak >nul

:: 删除服务
echo [INFO] Deleting service...
sc delete "%SERVICE_NAME%"
set DELETE_RESULT=%errorlevel%
if %DELETE_RESULT% neq 0 (
    echo [ERROR] Service delete failed
    echo [DEBUG] Error code: %DELETE_RESULT%
    echo.
    echo Possible reasons:
    echo 1. Service is still running
    echo 2. Service is occupied by other program
    echo 3. Insufficient permissions
    echo 4. Service configuration corrupted
    echo.
    echo [DEBUG] Trying force delete...
    sc delete "%SERVICE_NAME%" /f >nul 2>&1
    if %errorlevel% equ 0 (
        echo [INFO] Force delete service successful!
    ) else (
        echo [WARN] Force delete also failed, error code: %errorlevel%
        echo [DEBUG] Trying manual registry cleanup...
        reg delete "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%" /f >nul 2>&1
        if %errorlevel% equ 0 (
            echo [INFO] Manual registry cleanup successful!
        ) else (
            echo [WARN] Manual registry cleanup also failed, but continue cleanup steps
        )
    )
) else (
    echo [INFO] Service delete successful!
)

:: 清理环境变量
if "%CLEAN_ENV%"=="true" (
    echo [INFO] Cleaning environment variables...
    call :cleanup_environment
)

:: 清理日志文件
if "%CLEAN_LOGS%"=="true" (
    echo [INFO] Cleaning log files...
    call :cleanup_logs
)

:: 如果还有另一个服务存在，自动卸载
if %GO_HUB_EXISTS% equ 0 if %GO_HUB_ORACLE_EXISTS% equ 0 (
    echo.
    echo [INFO] Checking if other Gateway services need to be uninstalled...
    if "%SERVICE_NAME%"=="Gateway" (
        sc query "Gateway-Oracle" >nul 2>&1
        if %errorlevel% equ 0 (
            echo [INFO] Auto uninstalling Gateway-Oracle service...
            call :uninstall_single_service "Gateway-Oracle"
        )
    ) else (
        sc query "Gateway" >nul 2>&1
        if %errorlevel% equ 0 (
            echo [INFO] Auto uninstalling Gateway service...
            call :uninstall_single_service "Gateway"
        )
    )
)

echo.
echo ==========================================
echo Uninstall process completed!
echo ==========================================
echo.
echo [DEBUG] 最终状态总结：
echo   服务名称: %SERVICE_NAME%
echo   清理日志: %CLEAN_LOGS%
echo   清理环境变量: %CLEAN_ENV%
echo   强制模式: %FORCE_REMOVE%
echo   调试模式: %DEBUG_MODE%
echo.
echo [DEBUG] 验证服务状态...
if "%DEBUG_MODE%"=="false" (
    sc query "%SERVICE_NAME%" >nul 2>&1
    if %errorlevel% equ 0 (
        echo [DEBUG] [WARN] 服务仍然存在
        sc query "%SERVICE_NAME%" | findstr STATE
    ) else (
        echo [DEBUG] [OK] 服务已被成功删除
    )
) else (
    echo [DEBUG] 跳过服务状态检查（权限限制）
)
echo.
echo [DEBUG] 验证环境变量...
if "%DEBUG_MODE%"=="false" (
    reg query "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%\Environment" >nul 2>&1
    if %errorlevel% equ 0 (
        echo [DEBUG] [WARN] 环境变量仍然存在
    ) else (
        echo [DEBUG] [OK] 环境变量已被清理
    )
) else (
    echo [DEBUG] 跳过环境变量检查（权限限制）
)
echo.
echo [INFO] 卸载完成！
echo [INFO] 如需重新安装，请运行 install-service.cmd
echo.
echo [INFO] 如果卸载过程中出现问题，请检查：
echo   - Windows事件日志（Windows Logs -> Application）
echo   - 服务管理器 (services.msc) 中是否还有残留服务
echo   - 注册表中是否有残留配置
echo.
echo [INFO] 按任意键退出...
pause
exit /b 0

:: 清理环境变量
:cleanup_environment
echo [INFO] Cleaning registry environment variables...
echo [DEBUG] Cleaning service: %SERVICE_NAME%

if "%DEBUG_MODE%"=="true" (
    echo [DEBUG] Skip environment variables cleanup (permission limited)
    exit /b 0
)

:: 删除服务环境变量
reg delete "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%\Environment" /f >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] [OK] Service environment variables cleaned
) else (
    echo [INFO] [-] Service environment variables not found (may have been cleaned)
)

:: 尝试清理其他可能的注册表项
reg delete "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%SERVICE_NAME%" /f >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] [OK] Service registry entries cleaned
) else (
    echo [DEBUG] Service registry entries cleanup failed or not exist
)

exit /b 0

:: 卸载单个服务
:uninstall_single_service
set TARGET_SERVICE=%~1
echo [INFO] Starting to uninstall service: %TARGET_SERVICE%

:: 停止服务
echo [INFO] Stopping service %TARGET_SERVICE%...
sc stop "%TARGET_SERVICE%" >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARN] Service %TARGET_SERVICE% stop failed, but continue uninstall...
) else (
    echo [INFO] Service %TARGET_SERVICE% stopped
)

:: 等待服务完全停止
echo [INFO] Waiting for service %TARGET_SERVICE% to stop completely...
timeout /t 3 /nobreak >nul

:: 删除服务
echo [INFO] Deleting service %TARGET_SERVICE%...
sc delete "%TARGET_SERVICE%"
if %errorlevel% neq 0 (
    echo [ERROR] Service %TARGET_SERVICE% delete failed
    echo [DEBUG] Trying force delete...
    sc delete "%TARGET_SERVICE%" /f >nul 2>&1
    if %errorlevel% equ 0 (
        echo [INFO] Force delete service %TARGET_SERVICE% successful!
    ) else (
        echo [WARN] Force delete service %TARGET_SERVICE% also failed
    )
) else (
    echo [INFO] Service %TARGET_SERVICE% delete successful!
)

:: 清理环境变量
if "%CLEAN_ENV%"=="true" (
    echo [INFO] Cleaning environment variables for service %TARGET_SERVICE%...
    reg delete "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\%TARGET_SERVICE%\Environment" /f >nul 2>&1
    if %errorlevel% equ 0 (
        echo [INFO] Service %TARGET_SERVICE% environment variables cleaned
    ) else (
        echo [DEBUG] Service %TARGET_SERVICE% environment variables cleanup failed or not exist
    )
)

echo [INFO] Service %TARGET_SERVICE% uninstall completed
exit /b 0

:: 清理日志文件
:cleanup_logs
echo [DEBUG] Starting to clean log files...
:: 尝试从常见位置清理日志文件
set LOG_DIRS=logs .\logs ..\logs ..\..\logs
echo [DEBUG] Checking log directory list: %LOG_DIRS%

for %%d in (%LOG_DIRS%) do (
    echo [DEBUG] Checking directory: %%d
    if exist "%%d" (
        echo [INFO] Checking log directory: %%d
        
        if exist "%%d\service.log" (
            del "%%d\service.log" >nul 2>&1
            if %errorlevel% equ 0 (
                echo [INFO] [OK] Deleted %%d\service.log
            ) else (
                echo [DEBUG] Delete failed: %%d\service.log
            )
        )
        
        if exist "%%d\service-error.log" (
            del "%%d\service-error.log" >nul 2>&1
            if %errorlevel% equ 0 (
                echo [INFO] [OK] Deleted %%d\service-error.log
            ) else (
                echo [DEBUG] Delete failed: %%d\service-error.log
            )
        )
        
        if exist "%%d\app.log" (
            del "%%d\app.log" >nul 2>&1
            if %errorlevel% equ 0 (
                echo [INFO] [OK] Deleted %%d\app.log
            ) else (
                echo [DEBUG] Delete failed: %%d\app.log
            )
        )
        
        :: 删除Gateway相关的其他日志文件
        for %%f in ("%%d\gateway*.log" "%%d\*-gateway*.log") do (
            if exist "%%f" (
                del "%%f" >nul 2>&1
                if %errorlevel% equ 0 (
                    echo [INFO] [OK] Deleted %%f
                ) else (
                    echo [DEBUG] Delete failed: %%f
                )
            )
        )
    ) else (
        echo [DEBUG] Directory not exist: %%d
    )
)

echo [INFO] Log files cleanup completed
exit /b 0 