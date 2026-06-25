param(
  [string[]]$Targets = @(
    "windows/amd64",
    "linux/amd64",
    "linux/arm64",
    "darwin/amd64",
    "darwin/arm64"
  ),
  [switch]$SkipWeb,
  [switch]$RunTests,
  [string]$GoBin = "",
  [string]$OutDir = ""
)

$ErrorActionPreference = "Stop"

function Find-CommandPath {
  param([string[]]$Names)
  foreach ($Name in $Names) {
    $Command = Get-Command $Name -ErrorAction SilentlyContinue
    if ($Command) {
      return $Command.Source
    }
  }
  return ""
}

function Invoke-Native {
  param(
    [string]$FilePath,
    [string[]]$Arguments = @()
  )
  & $FilePath @Arguments
  if ($LASTEXITCODE -ne 0) {
    throw "Command failed with exit code ${LASTEXITCODE}: $FilePath $($Arguments -join ' ')"
  }
}

function Split-Target {
  param([string]$Target)
  $Parts = $Target.Split("/")
  if ($Parts.Count -ne 2 -or [string]::IsNullOrWhiteSpace($Parts[0]) -or [string]::IsNullOrWhiteSpace($Parts[1])) {
    throw "Invalid target '$Target'. Expected format: goos/goarch"
  }
  return @{
    GOOS = $Parts[0]
    GOARCH = $Parts[1]
  }
}

function Get-BinaryName {
  param(
    [string]$BaseName,
    [string]$GOOS,
    [string]$GOARCH
  )
  $Name = "${BaseName}_${GOOS}_${GOARCH}"
  if ($GOOS -eq "windows") {
    $Name = "$Name.exe"
  }
  return $Name
}

$RepoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $RepoRoot

if ([string]::IsNullOrWhiteSpace($OutDir)) {
  $OutDir = Join-Path $RepoRoot "dist\release"
}
$CacheDir = Join-Path $RepoRoot "dist\.gocache"
$TmpDir = Join-Path $RepoRoot "dist\.gotmp"
New-Item -ItemType Directory -Force -Path $OutDir, $CacheDir, $TmpDir | Out-Null

if ([string]::IsNullOrWhiteSpace($GoBin)) {
  $GoBin = Find-CommandPath @("go.exe", "go")
}
if ([string]::IsNullOrWhiteSpace($GoBin)) {
  throw "go compiler not found. Please install Go or pass -GoBin C:\path\to\go.exe"
}

$env:CGO_ENABLED = "0"
$env:GOCACHE = $CacheDir
$env:GOTMPDIR = $TmpDir

Write-Host "using go: $GoBin"
Invoke-Native $GoBin @("version")
Write-Host "output dir: $OutDir"
Write-Host "targets: $($Targets -join ', ')"
Write-Host "CGO_ENABLED=$env:CGO_ENABLED"

if ($RunTests) {
  Write-Host "running backend tests..."
  Invoke-Native $GoBin @("test", "./...")
} else {
  Write-Host "backend tests: skipped (pass -RunTests to enable)"
}

if (-not $SkipWeb) {
  $FrontendRoot = Join-Path $RepoRoot "replive-web"
  Set-Location $FrontendRoot
  $PackageManager = Find-CommandPath @("bun.exe", "bun", "npm.cmd", "npm")
  if ([string]::IsNullOrWhiteSpace($PackageManager)) {
    throw "bun or npm not found. Install one of them, or run build_all.bat -SkipWeb"
  }

  if (-not (Test-Path "node_modules")) {
    Write-Host "installing frontend dependencies..."
    if ((Split-Path -Leaf $PackageManager) -like "bun*") {
      Invoke-Native $PackageManager @("install")
    } else {
      Invoke-Native $PackageManager @("install", "--package-lock=false")
    }
  }

  Write-Host "building frontend assets..."
  if ((Split-Path -Leaf $PackageManager) -like "bun*") {
    Invoke-Native $PackageManager @("run", "build")
  } else {
    Invoke-Native $PackageManager @("run", "build")
  }
  Set-Location $RepoRoot
}

$BuiltFiles = New-Object System.Collections.Generic.List[string]
foreach ($Target in $Targets) {
  $Parsed = Split-Target $Target
  $GOOS = $Parsed.GOOS
  $GOARCH = $Parsed.GOARCH

  $env:GOOS = $GOOS
  $env:GOARCH = $GOARCH
  $BackendOut = Join-Path $OutDir (Get-BinaryName "replive" $GOOS $GOARCH)
  Write-Host "building backend $Target -> $BackendOut"
  Invoke-Native $GoBin @("build", "-o", $BackendOut, ".")
  $BuiltFiles.Add($BackendOut)

  if (-not $SkipWeb) {
    $WebOut = Join-Path $OutDir (Get-BinaryName "replive_web" $GOOS $GOARCH)
    Write-Host "building web launcher $Target -> $WebOut"
    Invoke-Native $GoBin @("build", "-o", $WebOut, "./replive-web")
    $BuiltFiles.Add($WebOut)
  }
}

Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "build finished:"
foreach ($File in $BuiltFiles) {
  Write-Host "  $File"
}
