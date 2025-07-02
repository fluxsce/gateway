@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

:: Oracle开发环境快速设置脚本 - 支持非标准目录结构
:: 用于GoHub项目Oracle驱动编译环境配置
:: 版本: 1.1

title Oracle开发环境快速设置

cls
echo ==========================================
echo  Oracle开发环境快速设置 v1.1
echo ==========================================
echo.
echo 此脚本帮助您快速配置Oracle开发环境
echo 用于支持GoHub项目中Oracle驱动的编译
echo.
echo 功能:
echo   ✓ 自动检测现有Oracle客户端 (支持非标准结构)
echo   ✓ 下载建议和安装指导
echo   ✓ 环境变量配置
echo   ✓ 验证开发环境完整性
echo.
pause

echo.
echo [1/5] 检测现有Oracle安装...

:: 检查ORACLE_HOME环境变量
if defined ORACLE_HOME (
    echo [发现] ORACLE_HOME已设置: %ORACLE_HOME%
    
    :: 灵活检查oci.dll
    set OCI_DLL_FOUND=0
    if exist "%ORACLE_HOME%\bin\oci.dll" (
        echo [✓] Oracle客户端库存在 (标准位置: bin\oci.dll)
        set ORACLE_EXISTS=1
        set OCI_DLL_FOUND=1
    ) else if exist "%ORACLE_HOME%\oci.dll" (
        echo [✓] Oracle客户端库存在 (根目录: oci.dll)
        set ORACLE_EXISTS=1
        set OCI_DLL_FOUND=1
    ) else (
        echo [搜索] 在ORACLE_HOME中查找oci.dll...
        for /r "%ORACLE_HOME%" %%F in (oci.dll) do (
            if exist "%%F" (
                echo [✓] Oracle客户端库存在 (找到: %%F)
                set ORACLE_EXISTS=1
                set OCI_DLL_FOUND=1
                goto DLL_FOUND
            )
        )
        :DLL_FOUND
        if !OCI_DLL_FOUND! EQU 0 (
            echo [✗] Oracle客户端库不存在
            set ORACLE_EXISTS=0
        )
    )
) else (
    echo [信息] ORACLE_HOME未设置，搜索常见安装位置...
    set ORACLE_EXISTS=0
    
    :: 搜索常见Oracle安装位置 - 支持非标准结构
    set SEARCH_PATHS=C:\oracle\* C:\app\oracle\* C:\Oracle\* D:\oracle\* "%ProgramFiles%\oracle\*" "%ProgramFiles(x86)%\oracle\*"
    
    for %%P in (%SEARCH_PATHS%) do (
        for /D %%D in (%%P) do (
            echo [检查] %%D
            :: 首先检查标准的bin目录
            if exist "%%D\bin\oci.dll" (
                echo [发现] Oracle客户端: %%D (标准结构)
                set FOUND_ORACLE=%%D
                set ORACLE_EXISTS=1
                goto FOUND
            ) else (
                :: 搜索其他可能的位置
                for /r "%%D" %%F in (oci.dll) do (
                    if exist "%%F" (
                        set FOUND_DIR=%%~dpF
                        :: 去掉末尾的反斜杠，获取父目录作为ORACLE_HOME
                        for %%G in ("!FOUND_DIR!..") do set FOUND_ORACLE=%%~fG
                        echo [发现] Oracle客户端: !FOUND_ORACLE! (非标准结构，oci.dll在 !FOUND_DIR!)
                        set ORACLE_EXISTS=1
                        goto FOUND
                    )
                )
            )
        )
    )
)

:FOUND
echo.
echo [2/5] Oracle客户端状态分析...

if %ORACLE_EXISTS% EQU 1 (
    if defined FOUND_ORACLE (
        set ORACLE_HOME=%FOUND_ORACLE%
        echo [建议] 建议设置ORACLE_HOME为: %FOUND_ORACLE%
    )
    
    echo 检查开发文件完整性:
    
    :: 检查运行时库 - 支持多种位置
    set OCI_DLL_LOCATIONS="%ORACLE_HOME%\bin\oci.dll" "%ORACLE_HOME%\oci.dll" "%ORACLE_HOME%\lib\oci.dll"
    set RUNTIME_FOUND=0
    
    for %%L in (%OCI_DLL_LOCATIONS%) do (
        if exist %%L (
            echo [✓] 运行时库: %%L
            set RUNTIME_FOUND=1
            goto RUNTIME_DONE
        )
    )
    
    if !RUNTIME_FOUND! EQU 0 (
        echo [搜索] 查找oci.dll...
        for /r "%ORACLE_HOME%" %%F in (oci.dll) do (
            if exist "%%F" (
                echo [✓] 运行时库: %%F
                set RUNTIME_FOUND=1
                goto RUNTIME_DONE
            )
        )
    )
    
    :RUNTIME_DONE
    if !RUNTIME_FOUND! EQU 0 (
        echo [✗] 运行时库缺失: oci.dll
    )
    
    :: 检查开发头文件 - 支持多种位置
    set OCI_HEADER_LOCATIONS="%ORACLE_HOME%\sdk\include\oci.h" "%ORACLE_HOME%\include\oci.h" "%ORACLE_HOME%\oci.h"
    set SDK_EXISTS=0
    
    for %%L in (%OCI_HEADER_LOCATIONS%) do (
        if exist %%L (
            echo [✓] 开发头文件: %%L
            set SDK_EXISTS=1
            goto SDK_DONE
        )
    )
    
    if !SDK_EXISTS! EQU 0 (
        echo [搜索] 查找oci.h...
        for /r "%ORACLE_HOME%" %%F in (oci.h) do (
            if exist "%%F" (
                echo [✓] 开发头文件: %%F
                set SDK_EXISTS=1
                goto SDK_DONE
            )
        )
    )
    
    :SDK_DONE
    if !SDK_EXISTS! EQU 0 (
        echo [✗] 开发头文件缺失: oci.h
        echo [!] 需要下载SDK包
    )
    
    :: 检查链接库 - 支持多种位置和格式
    set OCI_LIB_LOCATIONS="%ORACLE_HOME%\lib\oci.lib" "%ORACLE_HOME%\lib\liboci.a" "%ORACLE_HOME%\bin\oci.lib" "%ORACLE_HOME%\bin\liboci.a"
    set LIB_EXISTS=0
    
    for %%L in (%OCI_LIB_LOCATIONS%) do (
        if exist %%L (
            echo [✓] 链接库: %%L
            set LIB_EXISTS=1
            goto LIB_DONE
        )
    )
    
    if !LIB_EXISTS! EQU 0 (
        echo [搜索] 查找链接库文件...
        for /r "%ORACLE_HOME%" %%F in (oci.lib liboci.a) do (
            if exist "%%F" (
                echo [✓] 链接库: %%F
                set LIB_EXISTS=1
                goto LIB_DONE
            )
        )
    )
    
    :LIB_DONE
    if !LIB_EXISTS! EQU 0 (
        echo [✗] 链接库缺失
        echo [!] 需要检查SDK包或重新安装
    )
    
) else (
    echo [信息] 未检测到Oracle客户端安装
    echo.
    echo [3/5] Oracle Instant Client下载指导...
    echo.
    echo 步骤1: 访问Oracle官网
    echo   https://www.oracle.com/database/technologies/instant-client.html
    echo.
    echo 步骤2: 选择Windows x64平台
    echo.
    echo 步骤3: 下载以下两个包:
    echo   - instantclient-basic-windows.x64-21.8.0.0.0dbru.zip (基础包)
    echo   - instantclient-sdk-windows.x64-21.8.0.0.0dbru.zip   (开发包)
    echo.
    echo 步骤4: 解压到统一目录
    echo   建议路径: C:\oracle\instantclient_21_8
    echo.
    
    set /p MANUAL_ORACLE_PATH="如果您已手动安装，请输入Oracle客户端路径 (留空跳过): "
    
    if not "!MANUAL_ORACLE_PATH!"=="" (
        :: 检查用户指定的路径
        set USER_PATH_VALID=0
        
        :: 检查标准位置
        if exist "!MANUAL_ORACLE_PATH!\bin\oci.dll" (
            echo [✓] 找到Oracle客户端: !MANUAL_ORACLE_PATH! (标准结构)
            set ORACLE_HOME=!MANUAL_ORACLE_PATH!
            set ORACLE_EXISTS=1
            set USER_PATH_VALID=1
        ) else (
            :: 搜索用户指定目录
            for /r "!MANUAL_ORACLE_PATH!" %%F in (oci.dll) do (
                if exist "%%F" (
                    echo [✓] 找到Oracle客户端: !MANUAL_ORACLE_PATH! (oci.dll在 %%~dpF)
                    set ORACLE_HOME=!MANUAL_ORACLE_PATH!
                    set ORACLE_EXISTS=1
                    set USER_PATH_VALID=1
                    goto USER_PATH_FOUND
                )
            )
            :USER_PATH_FOUND
        )
        
        if !USER_PATH_VALID! EQU 0 (
            echo [✗] 指定路径无效: !MANUAL_ORACLE_PATH!
            echo [!] 未找到oci.dll文件
        )
    )
)

echo.
echo [4/5] 环境变量配置...

if %ORACLE_EXISTS% EQU 1 (
    echo.
    echo 当前Oracle路径: %ORACLE_HOME%
    echo.
    
    :: 检查环境变量是否已正确设置
    if "%ORACLE_HOME%"=="%ORACLE_HOME%" (
        echo ORACLE_HOME变量状态检查:
        
        reg query "HKEY_CURRENT_USER\Environment" /v ORACLE_HOME >nul 2>&1
        if %ERRORLEVEL% EQU 0 (
            echo [✓] ORACLE_HOME已在用户环境变量中设置
        ) else (
            echo [!] ORACLE_HOME未在持久环境变量中设置
            
            set /p SET_ORACLE_HOME="是否设置ORACLE_HOME环境变量? (Y/N): "
            if /I "!SET_ORACLE_HOME!"=="Y" (
                echo 设置ORACLE_HOME环境变量...
                setx ORACLE_HOME "%ORACLE_HOME%" >nul
                echo [✓] ORACLE_HOME已设置为: %ORACLE_HOME%
            )
        )
        
        :: 检查PATH变量 - 需要添加实际包含oci.dll的目录
        set ORACLE_BIN_PATH=
        if exist "%ORACLE_HOME%\bin\oci.dll" (
            set ORACLE_BIN_PATH=%ORACLE_HOME%\bin
        ) else if exist "%ORACLE_HOME%\oci.dll" (
            set ORACLE_BIN_PATH=%ORACLE_HOME%
        ) else (
            :: 搜索oci.dll所在目录
            for /r "%ORACLE_HOME%" %%F in (oci.dll) do (
                if exist "%%F" (
                    set ORACLE_BIN_PATH=%%~dpF
                    goto BIN_PATH_FOUND
                )
            )
            :BIN_PATH_FOUND
        )
        
        if defined ORACLE_BIN_PATH (
            echo %PATH% | findstr /i "!ORACLE_BIN_PATH!" >nul
            if %ERRORLEVEL% EQU 0 (
                echo [✓] Oracle路径已在PATH中: !ORACLE_BIN_PATH!
            ) else (
                echo [!] Oracle路径未在PATH中
                
                set /p ADD_PATH="是否将Oracle目录添加到PATH? (!ORACLE_BIN_PATH!) (Y/N): "
                if /I "!ADD_PATH!"=="Y" (
                    echo 添加Oracle路径到PATH...
                    setx PATH "%PATH%;!ORACLE_BIN_PATH!" >nul
                    echo [✓] PATH已更新
                    echo [!] 请重启命令行窗口使PATH生效
                )
            )
        ) else (
            echo [!] 无法确定Oracle bin目录位置
        )
    )
) else (
    echo [跳过] 未检测到Oracle客户端，无法配置环境变量
)

echo.
echo [5/5] 编译环境验证...

:: 检查GCC编译器
gcc --version >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [✓] GCC编译器可用
    gcc --version | findstr "gcc"
) else (
    echo [✗] GCC编译器不可用
    echo.
    echo MinGW/GCC安装建议:
    echo   选项1: TDM-GCC (推荐)
    echo     下载: https://jmeubank.github.io/tdm-gcc/
    echo     选择64位版本安装
    echo.
    echo   选项2: MSYS2
    echo     下载: https://www.msys2.org/
    echo     安装后执行: pacman -S mingw-w64-x86_64-gcc
    echo.
    echo   选项3: MinGW-w64
    echo     下载: https://www.mingw-w64.org/downloads/
    echo.
)

:: 检查Go环境
go version >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [✓] Go编程环境可用
    go version
) else (
    echo [✗] Go编程环境不可用
    echo     请先安装Go: https://golang.org/dl/
)

echo.
echo ==========================================
echo 环境配置总结
echo ==========================================
echo.

if %ORACLE_EXISTS% EQU 1 (
    if defined SDK_EXISTS (
        if %SDK_EXISTS% EQU 1 (
            if defined LIB_EXISTS (
                if %LIB_EXISTS% EQU 1 (
                    echo [✓] Oracle开发环境配置完整
                    echo     可以构建包含Oracle驱动的版本
                    echo.
                    echo 下一步: 运行GoHub构建脚本选择Oracle版本
                    echo     scripts\build\build-win2008-oracle.cmd
                ) else (
                    echo [!] Oracle环境不完整 - 缺少链接库
                    echo     建议重新下载SDK包
                )
            )
        ) else (
            echo [!] Oracle环境不完整 - 缺少SDK
            echo     请下载instantclient-sdk包
        )
    )
) else (
    echo [!] Oracle客户端未安装
    echo     建议:
    echo     1. 按照上述指导安装Oracle Instant Client
    echo     2. 或者使用纯Go版本 (不支持Oracle数据库)
)

echo.
echo ==========================================
echo 快速测试命令
echo ==========================================
echo.
echo 验证Oracle环境是否正确配置:
echo   where oci.dll
echo   echo %%ORACLE_HOME%%
echo   dir "%%ORACLE_HOME%%\sdk\include\oci.h" 2^>nul
echo   dir "%%ORACLE_HOME%%\include\oci.h" 2^>nul
echo.
echo 测试编译环境:
echo   gcc --version
echo   go version
echo.
echo 如果一切正常，可以运行GoHub构建脚本:
echo   scripts\build\build-win2008-oracle.cmd
echo.

pause

echo.
echo 配置完成！
echo 如需重新配置，请再次运行此脚本。
echo.
pause 