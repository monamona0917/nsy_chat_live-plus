$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent $PSScriptRoot
$FrontendRoot = Join-Path $RepoRoot "replive-web"
$OutDir = Join-Path $RepoRoot "dist"
$BackendOutFile = Join-Path $OutDir "replive.exe"
$FrontendOutFile = Join-Path $OutDir "replive_web.exe"
$ReadmeFile = Join-Path $OutDir "使用说明.txt"

Set-Location $FrontendRoot
if (-not (Test-Path "node_modules")) {
  npm install --package-lock=false
}
npm run compile

Set-Location $RepoRoot
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o $BackendOutFile .
go build -o $FrontendOutFile ./replive-web

@"
使用方法：

1. 双击 replive.exe
   这是后端程序，负责登录、同步数据和提供本地 API。

2. 双击 replive_web.exe
   这是前端程序，会自动打开浏览器访问聊天页面。

普通用户不需要安装 Node/npm/Bun，也不需要运行命令。
如果浏览器没有自动打开，请手动访问：
http://127.0.0.1:5173/

默认前端会连接后端：
http://127.0.0.1:8888/
"@ | Set-Content -Encoding UTF8 $ReadmeFile

Write-Host "Windows backend built: $BackendOutFile"
Write-Host "Windows frontend built: $FrontendOutFile"
Write-Host "README written: $ReadmeFile"
