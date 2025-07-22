# Gateway 交叉编译脚本 (Windows PowerShell)
# 使用 Docker 在 Windows 上交叉编译 Linux/macOS/Windows 版本

param(
    [string]$TargetOS = "linux",
    [string]$TargetArch = "amd64", 
    [switch]$EnableOracle = $false,
    [string]$Version = "latest",
    [string]$OutputDir = ".\dist",
    [switch]$BuildOnly = $false,
    [switch]$Verbose = $false,
    [switch]$Help = $false
)

# 显示帮助
if ($Help) {
    Write-Host @"
Gateway 交叉编译脚本

用法: .\crossbuild.ps1 [参数]

参数:
  -TargetOS string      目标操作系统 (linux, darwin, windows) [默认: linux]
  -TargetArch string    目标架构 (amd64, arm64) [默认: amd64]  
  -EnableOracle         启用 Oracle 数据库支持
  -Version string       版本号 [默认: latest]
  -OutputDir string     输出目录 [默认: .\dist]
  -BuildOnly           仅构建，不复制到输出目录
  -Verbose             详细输出
  -Help                显示帮助

示例:
  .\crossbuild.ps1                                    # 构建 Linux amd64 版本
  .\crossbuild.ps1 -EnableOracle                      # 构建带 Oracle 支持的版本
  .\crossbuild.ps1 -TargetOS darwin -TargetArch arm64 # 构建 macOS ARM64 版本
  .\crossbuild.ps1 -TargetOS linux -TargetArch arm64 -EnableOracle # 构建 Linux ARM64 Oracle 版本
"@
    exit 0
}

# 颜色输出函数
function Write-Info {
    param($Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param($Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param($Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-Debug {
    param($Message)
    if ($Verbose) {
        Write-Host "[DEBUG] $Message" -ForegroundColor Blue
    }
}

# 检查 Docker 是否安装
function Test-Docker {
    try {
        $dockerVersion = docker --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Debug "Docker 版本: $dockerVersion"
            return $true
        }
    } catch {
        return $false
    }
    return $false
}

# 检查 Docker 是否运行
function Test-DockerRunning {
    try {
        docker info 2>$null | Out-Null
        return $LASTEXITCODE -eq 0
    } catch {
        return $false
    }
}

# 获取项目信息
function Get-ProjectInfo {
    $projectDir = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
    $gitCommit = ""
    $buildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    
    try {
        if (Test-Path "$projectDir\.git") {
            $gitCommit = git rev-parse --short HEAD 2>$null
            if ($LASTEXITCODE -ne 0) { $gitCommit = "unknown" }
        } else {
            $gitCommit = "unknown"
        }
    } catch {
        $gitCommit = "unknown"
    }
    
    return @{
        ProjectDir = $projectDir
        GitCommit = $gitCommit
        BuildTime = $buildTime
    }
}

# 构建 Docker 镜像
function Build-DockerImage {
    param(
        $ProjectDir,
        $TargetOS,
        $TargetArch,
        $EnableOracle,
        $Version,
        $BuildTime,
        $GitCommit
    )
    
    $imageName = "gateway-crossbuild"
    $dockerfile = "$ProjectDir\scripts\build\Dockerfile.crossbuild"
    
    Write-Info "构建 Docker 镜像..."
    Write-Debug "Dockerfile: $dockerfile"
    Write-Debug "项目目录: $ProjectDir"
    
    $buildArgs = @(
        "--build-arg", "TARGET_OS=$TargetOS",
        "--build-arg", "TARGET_ARCH=$TargetArch", 
        "--build-arg", "ENABLE_ORACLE=$($EnableOracle.ToString().ToLower())",
        "--build-arg", "VERSION=$Version",
        "--build-arg", "BUILD_TIME=$BuildTime",
        "--build-arg", "GIT_COMMIT=$GitCommit",
        "--target", "builder",
        "-f", $dockerfile,
        "-t", $imageName,
        $ProjectDir
    )
    
    Write-Debug "Docker build 命令: docker build $($buildArgs -join ' ')"
    
    $buildResult = & docker @buildArgs
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Docker 镜像构建失败"
        exit 1
    }
    
    Write-Info "Docker 镜像构建成功: $imageName"
    return $imageName
}

# 从容器中复制构建产物
function Copy-BuildArtifacts {
    param(
        $ImageName,
        $OutputDir,
        $TargetOS,
        $TargetArch,
        $EnableOracle
    )
    
    # 创建输出目录
    if (!(Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
        Write-Debug "创建输出目录: $OutputDir"
    }
    
    Write-Info "从容器复制构建产物..."
    
    # 创建临时容器
    $containerName = "gateway-crossbuild-temp-$(Get-Random)"
    
    try {
        # 创建容器（不启动）
        $createResult = docker create --name $containerName $ImageName
        if ($LASTEXITCODE -ne 0) {
            Write-Error "创建容器失败"
            return $false
        }
        
        # 确定二进制文件名
        $suffix = if ($EnableOracle) { "-oracle" } else { "" }
        $extension = if ($TargetOS -eq "windows") { ".exe" } else { "" }
        $binaryName = "gateway-$TargetOS-$TargetArch$suffix$extension"
        
        # 复制二进制文件
        $sourcePath = "$containerName"+":/build/$binaryName"
        $destPath = "$OutputDir\$binaryName"
        
        Write-Debug "复制路径: $sourcePath -> $destPath"
        
        $copyResult = docker cp $sourcePath $destPath
        if ($LASTEXITCODE -eq 0) {
            Write-Info "构建产物已复制到: $destPath"
            
            # 显示文件信息
            $fileInfo = Get-Item $destPath
            Write-Info "文件大小: $([math]::Round($fileInfo.Length / 1MB, 2)) MB"
            
            return $true
        } else {
            Write-Error "复制构建产物失败"
            return $false
        }
        
    } finally {
        # 清理临时容器
        docker rm $containerName 2>$null | Out-Null
    }
}

# 清理 Docker 资源
function Clear-DockerResources {
    param($ImageName)
    
    Write-Info "清理 Docker 资源..."
    
    # 删除构建镜像
    docker rmi $ImageName 2>$null | Out-Null
    
    # 清理悬挂镜像
    $danglingImages = docker images -f "dangling=true" -q
    if ($danglingImages) {
        docker rmi $danglingImages 2>$null | Out-Null
    }
    
    Write-Debug "Docker 资源清理完成"
}

# 验证构建产物
function Test-BuildArtifact {
    param($FilePath)
    
    if (!(Test-Path $FilePath)) {
        Write-Error "构建产物不存在: $FilePath"
        return $false
    }
    
    $fileInfo = Get-Item $FilePath
    if ($fileInfo.Length -eq 0) {
        Write-Error "构建产物为空文件: $FilePath"
        return $false
    }
    
    Write-Info "构建产物验证通过: $FilePath"
    return $true
}

# 主函数
function Main {
    Write-Info "开始 Gateway 交叉编译..."
    Write-Info "目标平台: $TargetOS/$TargetArch"
    Write-Info "Oracle 支持: $EnableOracle"
    Write-Info "版本: $Version"
    
    # 检查 Docker
    if (!(Test-Docker)) {
        Write-Error "Docker 未安装或不在 PATH 中"
        Write-Error "请先安装 Docker Desktop: https://www.docker.com/products/docker-desktop"
        exit 1
    }
    
    if (!(Test-DockerRunning)) {
        Write-Error "Docker 未运行，请启动 Docker Desktop"
        exit 1
    }
    
    # 验证参数
    $supportedOS = @("linux", "darwin", "windows")
    $supportedArch = @("amd64", "arm64")
    
    if ($TargetOS -notin $supportedOS) {
        Write-Error "不支持的操作系统: $TargetOS"
        Write-Error "支持的操作系统: $($supportedOS -join ', ')"
        exit 1
    }
    
    if ($TargetArch -notin $supportedArch) {
        Write-Error "不支持的架构: $TargetArch"
        Write-Error "支持的架构: $($supportedArch -join ', ')"
        exit 1
    }
    
    # 暂时不支持 macOS 和 Windows 的 Oracle 交叉编译
    if (($TargetOS -eq "darwin" -or $TargetOS -eq "windows") -and $EnableOracle) {
        Write-Warn "暂时不支持 $TargetOS 的 Oracle 交叉编译，将禁用 Oracle 支持"
        $EnableOracle = $false
    }
    
    # 获取项目信息
    $projectInfo = Get-ProjectInfo
    Write-Debug "项目目录: $($projectInfo.ProjectDir)"
    Write-Debug "Git 提交: $($projectInfo.GitCommit)"
    Write-Debug "构建时间: $($projectInfo.BuildTime)"
    
    try {
        # 构建 Docker 镜像
        $imageName = Build-DockerImage -ProjectDir $projectInfo.ProjectDir `
                                      -TargetOS $TargetOS `
                                      -TargetArch $TargetArch `
                                      -EnableOracle $EnableOracle `
                                      -Version $Version `
                                      -BuildTime $projectInfo.BuildTime `
                                      -GitCommit $projectInfo.GitCommit
        
        if (!$BuildOnly) {
            # 复制构建产物
            $success = Copy-BuildArtifacts -ImageName $imageName `
                                         -OutputDir $OutputDir `
                                         -TargetOS $TargetOS `
                                         -TargetArch $TargetArch `
                                         -EnableOracle $EnableOracle
            
            if ($success) {
                $suffix = if ($EnableOracle) { "-oracle" } else { "" }
                $extension = if ($TargetOS -eq "windows") { ".exe" } else { "" }
                $artifactPath = "$OutputDir\gateway-$TargetOS-$TargetArch$suffix$extension"
                
                if (Test-BuildArtifact $artifactPath) {
                    Write-Info "交叉编译成功完成!"
                    Write-Info "构建产物: $artifactPath"
                } else {
                    exit 1
                }
            } else {
                exit 1
            }
        } else {
            Write-Info "仅构建模式 - 构建完成，产物保留在容器中"
        }
        
    } finally {
        # 清理 Docker 资源
        if ($imageName) {
            Clear-DockerResources -ImageName $imageName
        }
    }
}

# 执行主函数
Main 