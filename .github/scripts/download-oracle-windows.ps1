# Download Oracle Instant Client Basic + SDK for Windows amd64 (compile-time only).
# Does not bundle client libraries into release archives (OTN license).
$ErrorActionPreference = "Stop"

$icVersion = if ($env:ORACLE_IC_VERSION) { $env:ORACLE_IC_VERSION } else { "21.18.0.0.0dbru" }
$icPath = if ($env:ORACLE_IC_PATH) { $env:ORACLE_IC_PATH } else { "2118000" }
$workspace = if ($env:GITHUB_WORKSPACE) { $env:GITHUB_WORKSPACE } else { (Get-Location).Path }
$destRoot = if ($env:ORACLE_DIR) { $env:ORACLE_DIR } else { Join-Path $workspace ".oracle" }
$extractDir = Join-Path $destRoot "instantclient_21_18"
$baseUrl = "https://download.oracle.com/otn_software/nt/instantclient/$icPath"
$basicZip = "instantclient-basic-windows.x64-$icVersion.zip"
$sdkZip = "instantclient-sdk-windows.x64-$icVersion.zip"

New-Item -ItemType Directory -Force -Path $destRoot | Out-Null

$ociH = Join-Path $extractDir "sdk\include\oci.h"
$ociDll = Join-Path $extractDir "oci.dll"
$ociLib = Join-Path $extractDir "sdk\lib\msvc\oci.lib"

if ((Test-Path $ociH) -and (Test-Path $ociDll) -and (Test-Path $ociLib)) {
    Write-Host "[INFO] Oracle Instant Client already present: $extractDir"
} else {
    Write-Host "[INFO] Downloading Oracle Instant Client $icVersion into $destRoot"
    $tmp = Join-Path $destRoot "tmp"
    New-Item -ItemType Directory -Force -Path $tmp | Out-Null
    $basicPath = Join-Path $tmp $basicZip
    $sdkPath = Join-Path $tmp $sdkZip
    Invoke-WebRequest -Uri "$baseUrl/$basicZip" -OutFile $basicPath
    Invoke-WebRequest -Uri "$baseUrl/$sdkZip" -OutFile $sdkPath
    if (Test-Path $extractDir) {
        Remove-Item -Recurse -Force $extractDir
    }
    Expand-Archive -Path $basicPath -DestinationPath $destRoot -Force
    Expand-Archive -Path $sdkPath -DestinationPath $destRoot -Force
    Remove-Item -Recurse -Force $tmp
}

if (-not (Test-Path $ociH)) {
    throw "oci.h not found under $extractDir\sdk\include"
}
if (-not (Test-Path $ociLib)) {
    throw "oci.lib not found under $extractDir\sdk\lib\msvc"
}

$oracleHome = (Resolve-Path $extractDir).Path
$cgoCflags = "-I$oracleHome\sdk\include"
$cgoLdflags = "-L$oracleHome\sdk\lib\msvc -loci"

Write-Host "[INFO] ORACLE_HOME=$oracleHome"
Write-Host "[INFO] CGO_CFLAGS=$cgoCflags"
Write-Host "[INFO] CGO_LDFLAGS=$cgoLdflags"

if ($env:GITHUB_ENV) {
    Add-Content -Path $env:GITHUB_ENV -Value "ORACLE_HOME=$oracleHome"
    Add-Content -Path $env:GITHUB_ENV -Value "CGO_CFLAGS=$cgoCflags"
    Add-Content -Path $env:GITHUB_ENV -Value "CGO_LDFLAGS=$cgoLdflags"
}

$env:ORACLE_HOME = $oracleHome
$env:CGO_CFLAGS = $cgoCflags
$env:CGO_LDFLAGS = $cgoLdflags
$env:Path = "$oracleHome;$env:Path"
