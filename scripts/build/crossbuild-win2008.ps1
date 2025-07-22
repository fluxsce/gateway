# Gateway Windows Server 2008 兼容性构建脚本
# 解决在 Windows Server 2008 上运行时出现的系统调用错误

param(
    [string]$Version = "latest",
    [string]$OutputDir = ".\dist",
    [switch]$Verbose = $false,
    [switch]$Help = $false
)

if ($Help) {
    Write-Host @"
Gateway Windows Server 2008 兼容性构建脚本

用法: .\crossbuild-win2008.ps1 [参数]

参数:
  -Version string       版本号 [默认: latest]
  -OutputDir string     输出目录 [默认: .\dist]
  -Verbose             详细输出
  -Help                显示帮助

说明:
  此脚本专门用于构建兼容 Windows Server 2008 的版本，
  通过设置特定的构建参数和Go版本来解决兼容性问题。

常见错误代码解决方案:
  810961237-810979948: 这些错误通常与Go运行时的系统调用兼容性有关
  
解决方法:
  1. 使用较老的Go版本进行构建 (Go 1.19-1.21)
  2. 设置特定的构建标签和链接参数
  3. 禁用某些现代Windows特性
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

# 检查Go版本
function Test-GoVersion {
    try {
        $goVersion = go version
        Write-Debug "当前Go版本: $goVersion"
        
        # 提取版本号
        if ($goVersion -match "go(\d+\.\d+)") {
            $version = [version]$matches[1]
            if ($version -gt [version]"1.21") {
                Write-Warn "检测到Go版本 $($version.ToString())，建议使用Go 1.19-1.21以获得更好的Windows Server 2008兼容性"
                Write-Warn "如果遇到运行时错误，请考虑降级Go版本"
            }
        }
        return $true
    } catch {
        Write-Error "Go未安装或不在PATH中"
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

# Windows Server 2008 兼容性构建
function Build-Win2008Compatible {
    param(
        $ProjectDir,
        $Version,
        $BuildTime,
        $GitCommit,
        $OutputDir
    )
    
    Write-Info "开始构建 Windows Server 2008 兼容版本..."
    
    # 创建输出目录
    if (!(Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    }
    
    # 设置构建环境变量
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    
    # Windows Server 2008 兼容性参数
    $buildFlags = @(
        "-ldflags",
        "-s -w -X 'main.Version=$Version' -X 'main.BuildTime=$BuildTime' -X 'main.GitCommit=$GitCommit'"
    )
    
    # 添加Windows Server 2008特定的构建标签
    $buildTags = @(
        "-tags",
        "netgo,osusergo"
    )
    
    $outputFile = Join-Path $OutputDir "gateway-windows-server2008-amd64.exe"
    
    Write-Info "构建参数:"
    Write-Info "  GOOS: $env:GOOS"
    Write-Info "  GOARCH: $env:GOARCH" 
    Write-Info "  CGO_ENABLED: $env:CGO_ENABLED"
    Write-Info "  构建标签: $($buildTags[1])"
    Write-Info "  输出文件: $outputFile"
    
    try {
        # 切换到项目目录
        Push-Location $ProjectDir
        
        # 执行构建
        $buildArgs = @(
            "build"
        ) + $buildTags + $buildFlags + @(
            "-o", $outputFile,
            "cmd/app/main.go"
        )
        
        Write-Debug "执行命令: go $($buildArgs -join ' ')"
        
        & go @buildArgs
        
        if ($LASTEXITCODE -eq 0) {
            Write-Info "构建成功完成!"
            
            # 显示文件信息
            if (Test-Path $outputFile) {
                $fileInfo = Get-Item $outputFile
                Write-Info "构建产物: $outputFile"
                Write-Info "文件大小: $([math]::Round($fileInfo.Length / 1MB, 2)) MB"
                
                # 生成部署说明
                $deployGuide = @"
Windows Server 2008 部署说明:

1. 系统要求:
   - Windows Server 2008 R2 SP1 或更高版本
   - .NET Framework 3.5+ (通常已预装)
   - 足够的内存和磁盘空间

2. 部署步骤:
   - 将 $($fileInfo.Name) 复制到目标服务器
   - 复制配置文件目录 configs/
   - 确保防火墙允许应用端口
   
3. 运行应用:
   cmd> $($fileInfo.Name)
   
4. 故障排查:
   如果仍然遇到系统调用错误，请尝试:
   - 安装最新的Windows更新
   - 确保系统已安装所有必需的运行时库
   - 检查事件查看器中的详细错误信息
   - 考虑在兼容模式下运行

5. 技术说明:
   此版本使用了以下兼容性优化:
   - 纯Go网络实现 (netgo)
   - 纯Go用户管理 (osusergo)
   - 静态链接，无外部依赖
   - 禁用CGO，避免C库兼容性问题
"@
                
                $deployGuideFile = Join-Path $OutputDir "Windows-Server-2008-部署说明.txt"
                $deployGuide | Out-File -FilePath $deployGuideFile -Encoding UTF8
                Write-Info "部署说明已保存: $deployGuideFile"
                
                return $true
            } else {
                Write-Error "构建产物不存在: $outputFile"
                return $false
            }
        } else {
            Write-Error "构建失败，退出码: $LASTEXITCODE"
            return $false
        }
        
    } catch {
        Write-Error "构建过程中出现异常: $($_.Exception.Message)"
        return $false
    } finally {
        Pop-Location
    }
}

# 主函数
function Main {
    Write-Info "Gateway Windows Server 2008 兼容性构建工具"
    Write-Info "版本: $Version"
    Write-Info "输出目录: $OutputDir"
    
    # 检查Go环境
    if (!(Test-GoVersion)) {
        exit 1
    }
    
    # 获取项目信息
    $projectInfo = Get-ProjectInfo
    Write-Debug "项目目录: $($projectInfo.ProjectDir)"
    Write-Debug "Git提交: $($projectInfo.GitCommit)"
    Write-Debug "构建时间: $($projectInfo.BuildTime)"
    
    # 执行构建
    $success = Build-Win2008Compatible -ProjectDir $projectInfo.ProjectDir `
                                      -Version $Version `
                                      -BuildTime $projectInfo.BuildTime `
                                      -GitCommit $projectInfo.GitCommit `
                                      -OutputDir $OutputDir
    
    if ($success) {
        Write-Info "Windows Server 2008 兼容版本构建完成!"
        Write-Info "请查看部署说明文件了解详细的部署和故障排查信息"
        exit 0
    } else {
        Write-Error "构建失败"
        exit 1
    }
}

# 执行主函数
Main 