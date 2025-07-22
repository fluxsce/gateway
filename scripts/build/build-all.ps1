# Gateway 批量交叉编译脚本
# 一次性构建多个目标平台版本

param(
    [string]$Version = "latest",
    [string]$OutputDir = ".\dist",
    [switch]$EnableOracle = $false,
    [switch]$Verbose = $false,
    [switch]$Help = $false,
    [array]$Targets = @()
)

if ($Help) {
    Write-Host @"
Gateway 批量交叉编译脚本

用法: .\build-all.ps1 [参数]

参数:
  -Version string       版本号 [默认: latest]
  -OutputDir string     输出目录 [默认: .\dist]
  -EnableOracle         启用 Oracle 数据库支持
  -Verbose             详细输出
  -Targets array       指定构建目标 [默认: 所有支持的目标]
  -Help                显示帮助

可用目标:
  linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64

示例:
  .\build-all.ps1                                    # 构建所有目标
  .\build-all.ps1 -EnableOracle                      # 构建所有目标 (包含 Oracle)
  .\build-all.ps1 -Targets @("linux-amd64", "linux-arm64")  # 仅构建指定目标
  .\build-all.ps1 -Version "v1.0.0" -Verbose         # 指定版本并显示详细输出
"@
    exit 0
}

# 默认构建目标
$defaultTargets = @(
    @{OS="linux"; Arch="amd64"},
    @{OS="linux"; Arch="arm64"},
    @{OS="darwin"; Arch="amd64"},
    @{OS="darwin"; Arch="arm64"},
    @{OS="windows"; Arch="amd64"}
)

# 解析目标参数
$buildTargets = @()
if ($Targets.Count -eq 0) {
    $buildTargets = $defaultTargets
} else {
    foreach ($target in $Targets) {
        $parts = $target.Split('-')
        if ($parts.Count -eq 2) {
            $buildTargets += @{OS=$parts[0]; Arch=$parts[1]}
        } else {
            Write-Host "[ERROR] 无效的目标格式: $target" -ForegroundColor Red
            Write-Host "正确格式: os-arch (如: linux-amd64)" -ForegroundColor Yellow
            exit 1
        }
    }
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

# 获取脚本目录
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$crossbuildScript = Join-Path $scriptDir "crossbuild.ps1"

# 检查交叉编译脚本
if (!(Test-Path $crossbuildScript)) {
    Write-Error "交叉编译脚本不存在: $crossbuildScript"
    exit 1
}

# 构建统计
$buildResults = @()
$successCount = 0
$failCount = 0
$startTime = Get-Date

Write-Info "开始批量交叉编译..."
Write-Info "版本: $Version"
Write-Info "输出目录: $OutputDir"
Write-Info "Oracle 支持: $EnableOracle"
Write-Info "构建目标数量: $($buildTargets.Count)"

# 遍历构建目标
foreach ($target in $buildTargets) {
    $targetName = "$($target.OS)-$($target.Arch)"
    Write-Info "----------------------------------------"
    Write-Info "构建目标: $targetName"
    
    $buildStartTime = Get-Date
    
    try {
        # 构建参数
        $buildArgs = @(
            "-TargetOS", $target.OS,
            "-TargetArch", $target.Arch,
            "-Version", $Version,
            "-OutputDir", $OutputDir
        )
        
        if ($EnableOracle -and $target.OS -eq "linux") {
            $buildArgs += "-EnableOracle"
        }
        
        if ($Verbose) {
            $buildArgs += "-Verbose"
        }
        
        # 执行构建
        & $crossbuildScript @buildArgs
        
        if ($LASTEXITCODE -eq 0) {
            $buildTime = (Get-Date) - $buildStartTime
            Write-Info "✓ $targetName 构建成功 (耗时: $($buildTime.TotalSeconds.ToString('F1'))秒)"
            
            $buildResults += @{
                Target = $targetName
                Status = "SUCCESS"
                Time = $buildTime.TotalSeconds
                Error = $null
            }
            $successCount++
        } else {
            throw "构建失败 (退出码: $LASTEXITCODE)"
        }
        
    } catch {
        $buildTime = (Get-Date) - $buildStartTime
        Write-Error "✗ $targetName 构建失败: $($_.Exception.Message)"
        
        $buildResults += @{
            Target = $targetName
            Status = "FAILED"
            Time = $buildTime.TotalSeconds
            Error = $_.Exception.Message
        }
        $failCount++
    }
}

# 显示构建总结
$totalTime = (Get-Date) - $startTime
Write-Info "========================================"
Write-Info "批量构建完成"
Write-Info "========================================"
Write-Info "总耗时: $($totalTime.TotalMinutes.ToString('F1')) 分钟"
Write-Info "成功: $successCount"
Write-Info "失败: $failCount"
Write-Info "总计: $($buildTargets.Count)"

if ($successCount -gt 0) {
    Write-Info ""
    Write-Info "成功构建的目标:"
    foreach ($result in $buildResults) {
        if ($result.Status -eq "SUCCESS") {
            $suffix = if ($EnableOracle -and $result.Target.StartsWith("linux")) { "-oracle" } else { "" }
            $extension = if ($result.Target.StartsWith("windows")) { ".exe" } else { "" }
            $fileName = "gateway-$($result.Target)$suffix$extension"
            $filePath = Join-Path $OutputDir $fileName
            if (Test-Path $filePath) {
                $fileSize = [math]::Round((Get-Item $filePath).Length / 1MB, 2)
                Write-Info "  ✓ $($result.Target) -> $fileName ($fileSize MB, $($result.Time.ToString('F1'))s)"
            }
        }
    }
}

if ($failCount -gt 0) {
    Write-Error ""
    Write-Error "失败的目标:"
    foreach ($result in $buildResults) {
        if ($result.Status -eq "FAILED") {
            Write-Error "  ✗ $($result.Target): $($result.Error)"
        }
    }
}

# 生成构建报告
$reportPath = Join-Path $OutputDir "build-report.json"
$report = @{
    Version = $Version
    BuildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    TotalTime = $totalTime.TotalSeconds
    EnableOracle = $EnableOracle
    Results = $buildResults
    Summary = @{
        Total = $buildTargets.Count
        Success = $successCount
        Failed = $failCount
    }
}

$report | ConvertTo-Json -Depth 10 | Out-File -FilePath $reportPath -Encoding UTF8
Write-Info "构建报告已保存: $reportPath"

# 设置退出码
if ($failCount -gt 0) {
    exit 1
} else {
    exit 0
} 