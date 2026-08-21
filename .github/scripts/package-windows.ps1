# CI-owned Windows amd64 packaging. Not used by local scripts/build.
# Env:
#   VERSION     required, e.g. 3.2.8
#   ORACLE      0 (MySQL/SQLite/ClickHouse) or 1 (also Oracle)
#   ORACLE_HOME required when ORACLE=1
$ErrorActionPreference = "Stop"

if (-not $env:VERSION) {
    throw "VERSION is required"
}

$oracle = if ($env:ORACLE) { $env:ORACLE } else { "0" }
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = (Resolve-Path (Join-Path $scriptDir "..\..")).Path
Set-Location $root

if (-not (Test-Path "go.mod")) {
    throw "go.mod not found in $root"
}
if (-not (Test-Path "web\frontend\dist")) {
    throw "web\frontend\dist missing; frontend job must run first"
}

if ($oracle -eq "1") {
    if (-not $env:ORACLE_HOME) {
        throw "ORACLE_HOME is required when ORACLE=1"
    }
    $ociH = Join-Path $env:ORACLE_HOME "sdk\include\oci.h"
    if (-not (Test-Path $ociH)) {
        throw "Oracle SDK header not found: $ociH"
    }
    $buildTags = "netgo,osusergo,windows"
    $variant = "windows-amd64-oracle"
    if (-not $env:CGO_CFLAGS) {
        $env:CGO_CFLAGS = "-I$($env:ORACLE_HOME)\sdk\include"
    }
    if (-not $env:CGO_LDFLAGS) {
        $env:CGO_LDFLAGS = "-L$($env:ORACLE_HOME)\sdk\lib\msvc -loci"
    }
    $env:Path = "$($env:ORACLE_HOME);$env:Path"
    Write-Host "[INFO] Building with Oracle support"
} else {
    $buildTags = "netgo,osusergo,no_oracle,windows"
    $variant = "windows-amd64"
    Write-Host "[INFO] Building without Oracle (MySQL/SQLite/ClickHouse)"
}

$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

$gitCommit = "unknown"
try { $gitCommit = (git rev-parse --short HEAD 2>$null).Trim() } catch { }
if (-not $gitCommit) { $gitCommit = "unknown" }
$buildTime = [DateTime]::UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ")
$ldflags = "-s -w -X main.Version=$($env:VERSION) -X main.BuildTime=$buildTime -X main.GitCommit=$gitCommit"
if ($oracle -eq "1") {
    $ldflags = "$ldflags -linkmode external"
}

$packageDir = Join-Path $root "dist\gateway"
$archiveName = "gateway-$variant-$($env:VERSION).zip"
if (Test-Path "dist") {
    Remove-Item -Recurse -Force "dist"
}
New-Item -ItemType Directory -Force -Path $packageDir | Out-Null

Write-Host "[INFO] go build gateway tags=$buildTags"
& go build -tags "$buildTags" -ldflags "$ldflags" -o (Join-Path $packageDir "gateway.exe") cmd\app\main.go
if ($LASTEXITCODE -ne 0) { throw "gateway build failed" }

Write-Host "[INFO] go build password_plugin"
& go build -tags "$buildTags" -ldflags "-s -w" -o (Join-Path $packageDir "password_plugin.exe") cmd\plugins\password_plugin\main.go
if ($LASTEXITCODE -ne 0) { throw "password_plugin build failed" }

function Copy-Tree([string]$src, [string]$dest) {
    New-Item -ItemType Directory -Force -Path $dest | Out-Null
    if (Test-Path $src) {
        Copy-Item -Path (Join-Path $src "*") -Destination $dest -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Copy-Tree "configs" (Join-Path $packageDir "configs")
Copy-Tree "web\static" (Join-Path $packageDir "web\static")
Copy-Tree "web\frontend\dist" (Join-Path $packageDir "web\frontend\dist")
Copy-Tree "scripts\db" (Join-Path $packageDir "scripts\db")
Copy-Tree "scripts\deploy" (Join-Path $packageDir "scripts\deploy")
Copy-Tree "scripts\k8s" (Join-Path $packageDir "scripts\k8s")
Copy-Tree "scripts\test" (Join-Path $packageDir "scripts\test")
New-Item -ItemType Directory -Force -Path (Join-Path $packageDir "logs") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $packageDir "backup") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $packageDir "scripts\data") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $packageDir "pprof_analysis") | Out-Null

if (Test-Path "scripts\docker") {
    Copy-Tree "scripts\docker" (Join-Path $packageDir "scripts\docker")
    $pushSh = Join-Path $packageDir "scripts\docker\push.sh"
    if (Test-Path $pushSh) {
        Remove-Item -Force $pushSh
    }
}

if ($oracle -eq "1") {
    Write-Host "[INFO] Oracle Instant Client is not bundled (OTN license). Install it on the target host."
}

$archivePath = Join-Path $root "dist\$archiveName"
Push-Location (Join-Path $root "dist")
try {
    & tar -a -c -f $archiveName gateway
    if ($LASTEXITCODE -ne 0) { throw "failed to create $archiveName" }
} finally {
    Pop-Location
}

Write-Host "[OK] $archivePath"
Get-Item $archivePath | Format-List Name, Length
