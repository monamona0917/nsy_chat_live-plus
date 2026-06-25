@echo off
setlocal enabledelayedexpansion

cd /d "%~dp0"

if not exist dist mkdir dist

if "%GO_BIN%"=="" (
    where go.exe >nul 2>nul
    if not errorlevel 1 (
        set "GO_BIN=go.exe"
    ) else (
        echo go compiler not found
        exit /b 1
    )
)

for /f "delims=" %%i in ('"%GO_BIN%" env GOOS') do set "HOST_GOOS=%%i"
for /f "delims=" %%i in ('"%GO_BIN%" env GOARCH') do set "HOST_GOARCH=%%i"

if /i not "%HOST_GOOS%"=="windows" (
    echo build_win.bat must run on Windows, current host: %HOST_GOOS%/%HOST_GOARCH%
    exit /b 1
)

if "%CGO_ENABLED%"=="" set "CGO_ENABLED=0"
set "OUTPUT_PATH=dist\replive_%HOST_GOOS%_%HOST_GOARCH%.exe"

echo using go: %GO_BIN%
"%GO_BIN%" version
echo building native windows binary: %HOST_GOOS%/%HOST_GOARCH%
echo CGO_ENABLED=%CGO_ENABLED%

"%GO_BIN%" test ./... -run TestDoesNotExist
if errorlevel 1 exit /b 1

"%GO_BIN%" build -o "%OUTPUT_PATH%" .
if errorlevel 1 exit /b 1

copy /y "%OUTPUT_PATH%" "replive.exe" >nul

echo build finished:
echo   dist: %CD%\%OUTPUT_PATH%
echo   runtime: %CD%\replive.exe
