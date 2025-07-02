@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

:: GoHub 交叉编译批处理脚本
:: 提供简单的图形化选择界面

title GoHub 交叉编译工具

:MAIN_MENU
cls
echo ==========================================
echo          GoHub 交叉编译工具
echo ==========================================
echo.
echo 请选择操作:
echo.
echo 1. 构建 Linux AMD64 版本
echo 2. 构建 Linux AMD64 版本 (包含 Oracle)
echo 3. 构建 Linux ARM64 版本  
echo 4. 构建 macOS AMD64 版本
echo 5. 构建 macOS ARM64 版本
echo 6. 构建 Windows Server 2008 64位版本
echo 7. 构建 Windows Server 2008 兼容版本 (推荐)
echo 8. 构建所有版本
echo 9. 构建所有版本 (包含 Oracle)
echo 10. 自定义构建
echo 11. 查看帮助
echo 0. 退出
echo.
set /p choice="请输入选择 (0-11): "

if "%choice%"=="1" goto BUILD_LINUX_AMD64
if "%choice%"=="2" goto BUILD_LINUX_AMD64_ORACLE
if "%choice%"=="3" goto BUILD_LINUX_ARM64
if "%choice%"=="4" goto BUILD_DARWIN_AMD64
if "%choice%"=="5" goto BUILD_DARWIN_ARM64
if "%choice%"=="6" goto BUILD_WINDOWS_AMD64
if "%choice%"=="7" goto BUILD_WINDOWS_2008_COMPATIBLE
if "%choice%"=="8" goto BUILD_ALL
if "%choice%"=="9" goto BUILD_ALL_ORACLE
if "%choice%"=="10" goto CUSTOM_BUILD
if "%choice%"=="11" goto SHOW_HELP
if "%choice%"=="0" goto EXIT
goto INVALID_CHOICE

:BUILD_LINUX_AMD64
echo.
echo 正在构建 Linux AMD64 版本...
powershell -ExecutionPolicy Bypass -File "%~dp0crossbuild.ps1" -TargetOS linux -TargetArch amd64
goto BUILD_COMPLETE

:BUILD_LINUX_AMD64_ORACLE
echo.
echo 正在构建 Linux AMD64 版本 (包含 Oracle)...
powershell -ExecutionPolicy Bypass -File "%~dp0crossbuild.ps1" -TargetOS linux -TargetArch amd64 -EnableOracle
goto BUILD_COMPLETE

:BUILD_LINUX_ARM64
echo.
echo 正在构建 Linux ARM64 版本...
powershell -ExecutionPolicy Bypass -File "%~dp0crossbuild.ps1" -TargetOS linux -TargetArch arm64
goto BUILD_COMPLETE

:BUILD_DARWIN_AMD64
echo.
echo 正在构建 macOS AMD64 版本...
powershell -ExecutionPolicy Bypass -File "%~dp0crossbuild.ps1" -TargetOS darwin -TargetArch amd64
goto BUILD_COMPLETE

:BUILD_DARWIN_ARM64
echo.
echo 正在构建 macOS ARM64 版本...
powershell -ExecutionPolicy Bypass -File "%~dp0crossbuild.ps1" -TargetOS darwin -TargetArch arm64
goto BUILD_COMPLETE

:BUILD_WINDOWS_AMD64
echo.
echo 正在构建 Windows Server 2008 64位版本...
powershell -ExecutionPolicy Bypass -File "%~dp0crossbuild.ps1" -TargetOS windows -TargetArch amd64
goto BUILD_COMPLETE

:BUILD_WINDOWS_2008_COMPATIBLE
echo.
echo 正在构建 Windows Server 2008 兼容版本...
echo 此版本专门针对 Windows Server 2008 系统调用兼容性进行了优化
powershell -ExecutionPolicy Bypass -File "%~dp0crossbuild-win2008.ps1"
goto BUILD_COMPLETE

:BUILD_ALL
echo.
echo 正在构建所有版本...
powershell -ExecutionPolicy Bypass -File "%~dp0build-all.ps1"
goto BUILD_COMPLETE

:BUILD_ALL_ORACLE
echo.
echo 正在构建所有版本 (包含 Oracle)...
powershell -ExecutionPolicy Bypass -File "%~dp0build-all.ps1" -EnableOracle
goto BUILD_COMPLETE

:CUSTOM_BUILD
cls
echo ==========================================
echo            自定义构建选项
echo ==========================================
echo.
echo 操作系统:
echo 1. Linux
echo 2. macOS (Darwin)
echo 3. Windows
echo.
set /p os_choice="请选择操作系统 (1-3): "

if "%os_choice%"=="1" set target_os=linux
if "%os_choice%"=="2" set target_os=darwin
if "%os_choice%"=="3" set target_os=windows
if "%target_os%"=="" goto INVALID_CHOICE

echo.
echo 架构:
echo 1. AMD64 (x86_64)
echo 2. ARM64
echo.
set /p arch_choice="请选择架构 (1-2): "

if "%arch_choice%"=="1" set target_arch=amd64
if "%arch_choice%"=="2" set target_arch=arm64
if "%target_arch%"=="" goto INVALID_CHOICE

echo.
echo 是否启用 Oracle 支持? (仅限 Linux)
echo 1. 是
echo 2. 否
echo.
set /p oracle_choice="请选择 (1-2): "

set oracle_flag=
if "%oracle_choice%"=="1" if "%target_os%"=="linux" set oracle_flag=-EnableOracle

echo.
set /p version="请输入版本号 (默认: latest): "
if "%version%"=="" set version=latest

echo.
echo 构建配置:
echo   操作系统: %target_os%
echo   架构: %target_arch%
echo   Oracle 支持: %oracle_choice%
echo   版本: %version%
echo.
pause

echo 正在构建 %target_os%/%target_arch% 版本...
powershell -ExecutionPolicy Bypass -File "%~dp0crossbuild.ps1" -TargetOS %target_os% -TargetArch %target_arch% -Version %version% %oracle_flag%
goto BUILD_COMPLETE

:SHOW_HELP
cls
echo ==========================================
echo              帮助信息
echo ==========================================
echo.
echo 系统要求:
echo   - Windows 10/11
echo   - Docker Desktop
echo   - PowerShell 5.0+
echo.
echo 构建产物位置:
echo   .\dist\ 目录下
echo.
echo 支持的目标平台:
echo   - Linux AMD64/ARM64
echo   - macOS AMD64/ARM64
echo   - Windows AMD64 (Server 2008+)
echo.
echo Oracle 支持说明:
echo   - 仅支持 Linux 平台
echo   - 需要 Docker 容器内下载 Oracle 客户端
echo   - 构建时间会更长
echo.
echo 故障排查:
echo   1. 确保 Docker Desktop 正在运行
echo   2. 检查网络连接 (需要下载依赖)
echo   3. 确保有足够的磁盘空间 (2GB+)
echo.
echo 如果遇到问题，请查看详细日志或使用 PowerShell 脚本获取更多信息。
echo.
pause
goto MAIN_MENU

:BUILD_COMPLETE
echo.
if %ERRORLEVEL% EQU 0 (
    echo [成功] 构建完成!
    echo 构建产物位置: %~dp0..\..\dist
    echo.
    echo 是否打开输出目录?
    set /p open_dir="(Y/N): "
    if /I "!open_dir!"=="Y" start "" "%~dp0..\..\dist"
) else (
    echo [失败] 构建失败! 错误代码: %ERRORLEVEL%
    echo.
    echo 可能的原因:
    echo   - Docker 未运行
    echo   - 网络连接问题
    echo   - 磁盘空间不足
    echo.
    echo 建议:
    echo   1. 检查 Docker Desktop 状态
    echo   2. 使用 PowerShell 脚本获取详细错误信息
    echo   3. 查看构建日志
)
echo.
pause
goto MAIN_MENU

:INVALID_CHOICE
echo.
echo [错误] 无效选择，请重新输入
echo.
pause
goto MAIN_MENU

:EXIT
echo.
echo 感谢使用 GoHub 交叉编译工具!
pause
exit /b 0 