@echo off
REM Proto Code Generator and Cleaner (Windows)

:menu
cls
echo ========================================
echo  Proto Code Manager
echo ========================================
echo.
echo  1. Generate gRPC code
echo  2. Clean generated files
echo  3. Clean and regenerate (Recommended)
echo  4. Exit
echo.
echo ========================================
set /p choice="Please select an option (1-4): "

if "%choice%"=="1" goto generate
if "%choice%"=="2" goto clean
if "%choice%"=="3" goto clean_and_generate
if "%choice%"=="4" goto exit
echo [ERROR] Invalid option, please try again.
timeout /t 2 >nul
goto menu

:generate
echo.
echo ========================================
echo  Generating gRPC Code
echo ========================================
echo.

REM Check if protoc is installed
where protoc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] protoc not found!
    echo Please install Protocol Buffers compiler from:
    echo https://github.com/protocolbuffers/protobuf/releases
    echo.
    pause
    goto menu
)

REM Show protoc version
echo [INFO] Checking protoc version...
protoc --version
echo.

REM Add GOPATH\bin to PATH
echo [INFO] Configuring Go environment...
for /f "tokens=*" %%i in ('go env GOPATH') do set GOPATH=%%i
if not "%GOPATH%"=="" (
    set "PATH=%PATH%;%GOPATH%\bin"
    echo [OK] Added %GOPATH%\bin to PATH
) else (
    echo [WARNING] Could not determine GOPATH
)
echo.

REM Install protoc plugins
echo [1/3] Installing protoc plugins...
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to install protoc-gen-go
    pause
    goto menu
)

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to install protoc-gen-go-grpc
    pause
    goto menu
)

echo [OK] Plugins installed successfully!
echo.

REM Count proto files
set proto_count=0
for %%f in (*.proto) do set /a proto_count+=1
echo [INFO] Found %proto_count% proto files
echo.

REM Generate gRPC code
echo [2/3] Generating gRPC code from proto files...
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to generate gRPC code
    pause
    goto menu
)

echo [OK] Code generation completed!
echo.

REM List generated files
echo [3/3] Generated files:
dir /B *.pb.go 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] No .pb.go files found
) else (
    echo.
    echo [OK] All files generated successfully!
)
echo.

pause
goto menu

:clean
echo.
echo ========================================
echo  Cleaning Generated Files
echo ========================================
echo.

REM Check if there are files to clean
set file_count=0
for %%f in (*.pb.go) do set /a file_count+=1

if %file_count% EQU 0 (
    echo [INFO] No generated files to clean
) else (
    echo [INFO] Found %file_count% files to clean
    del /Q *.pb.go 2>nul
    del /Q *_grpc.pb.go 2>nul
    echo [OK] Clean completed!
)
echo.

pause
goto menu

:clean_and_generate
echo.
echo ========================================
echo  Clean and Regenerate (Recommended)
echo ========================================
echo.

REM Step 1: Clean
echo [Step 1/2] Cleaning old files...
set file_count=0
for %%f in (*.pb.go) do set /a file_count+=1

if %file_count% EQU 0 (
    echo [INFO] No old files to clean
) else (
    echo [INFO] Cleaning %file_count% files...
    del /Q *.pb.go 2>nul
    del /Q *_grpc.pb.go 2>nul
    echo [OK] Old files cleaned!
)
echo.

REM Step 2: Generate
echo [Step 2/2] Generating new code...
echo.

REM Check if protoc is installed
where protoc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] protoc not found!
    echo Please install Protocol Buffers compiler from:
    echo https://github.com/protocolbuffers/protobuf/releases
    echo.
    pause
    goto menu
)

REM Show protoc version
echo [INFO] protoc version:
protoc --version
echo.

REM Add GOPATH\bin to PATH
echo [INFO] Configuring Go environment...
for /f "tokens=*" %%i in ('go env GOPATH') do set GOPATH=%%i
if not "%GOPATH%"=="" (
    set "PATH=%PATH%;%GOPATH%\bin"
    echo [OK] Added %GOPATH%\bin to PATH
) else (
    echo [WARNING] Could not determine GOPATH
)
echo.

REM Install protoc plugins
echo [1/3] Installing protoc plugins...
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to install protoc-gen-go
    pause
    goto menu
)

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to install protoc-gen-go-grpc
    pause
    goto menu
)

echo [OK] Plugins installed!
echo.

REM Count proto files
set proto_count=0
for %%f in (*.proto) do set /a proto_count+=1
echo [INFO] Found %proto_count% proto files
echo.

REM Generate gRPC code
echo [2/3] Generating gRPC code...
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to generate gRPC code
    pause
    goto menu
)

echo [OK] Code generation completed!
echo.

REM List generated files
echo [3/3] Generated files:
dir /B *.pb.go 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] No .pb.go files found
) else (
    echo.
    echo [SUCCESS] Clean and regenerate completed!
)
echo.

pause
goto menu

:exit
echo.
echo Goodbye!
exit /b 0

