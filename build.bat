@echo off
chcp 65001 >nul
echo ========================================
echo FRPCGUI Portable Build Script
echo ========================================
echo.

REM Add Go to PATH if not already present
set GO_PATH=D:\Apps\Go\bin
if not exist "%GO_PATH%\go.exe" (
    echo [ERROR] Go not found at %GO_PATH%, please install Go 1.24.0 or higher
    echo Download URL: https://golang.org/dl/
    pause
    exit /b 1
)

REM Temporarily add Go to PATH for this session
set PATH=%GO_PATH%;%PATH%

REM Verify Go is now available
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Failed to add Go to PATH
    pause
    exit /b 1
)

echo [INFO] Using Go from: %GO_PATH%
go version
echo.

REM Check if MinGW is installed
where gcc >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARNING] MinGW GCC not found, some features may not work properly
    echo Recommended to install MinGW: https://github.com/mstorsjo/llvm-mingw/releases
    echo.
)

echo [INFO] Starting portable build...
echo.

setlocal enabledelayedexpansion
set GOARCH_x64=amd64
set GOARCH_x86=386
set BUILDDIR=%~dp0
cd /d %BUILDDIR% || exit /b 1

REM 默认构建所有架构
set FRPCGUI_TARGET=%~1
if "%FRPCGUI_TARGET%" == "" set FRPCGUI_TARGET=x64 x86

:packages
	echo [1/4] Downloading dependencies...
	go mod tidy || goto :error

:resources
	echo [2/4] Generating resource files...
	for /f %%a in ('go generate') do set %%a
	if not defined VERSION (
		echo [ERROR] Failed to get version information
		goto :error
	)
	echo [INFO] Version: %VERSION%

:build
	echo [3/4] Compiling program...
	set MOD=github.com/hzcrv1911/frpcgui
	set GO111MODULE=on
	set CGO_ENABLED=0
	for %%a in (%FRPCGUI_TARGET%) do (
		echo [Building] %%a architecture...
		if defined GOARCH_%%a (
			set GOARCH=!GOARCH_%%a!
		) else (
			set GOARCH=%%a
		)
		go build -trimpath -ldflags="-H windowsgui -s -w -X %MOD%/pkg/version.BuildDate=%BUILD_DATE%" -o bin\%%a\frpcgui.exe .\cmd\frpcgui || goto :error
		echo [Done] bin\%%a\frpcgui.exe
		REM 清理并复制 assets 目录到当前架构目录
		if exist bin\%%a\assets rmdir /S /Q bin\%%a\assets
		if exist assets (
			robocopy assets bin\%%a\assets /E /NFL /NDL /NJH /NJS /nc /ns /np
			if !errorlevel! gtr 7 goto :error
		)
	)

:create_archives
	echo [4/4] Creating archives...
	for %%a in (%FRPCGUI_TARGET%) do (
		echo [Packaging] frpcgui-%VERSION%-%%a.zip...
		tar -ac -C bin\%%a -f bin\frpcgui-%VERSION%-%%a.zip frpcgui.exe assets
	)

:success
	echo.
	echo ========================================
	echo Build successful!
	echo ========================================
	echo.
	echo Executable locations:
	for %%a in (%FRPCGUI_TARGET%) do (
		echo   - bin\%%a\frpcgui.exe
	)
	echo.
	echo Assets locations:
	if exist bin\assets echo   - bin\assets\winsw.exe
	if exist bin\assets echo   - bin\assets\frpc.exe
	for %%a in (%FRPCGUI_TARGET%) do (
		echo   - bin\frpcgui-%VERSION%-%%a.zip
	)
	echo.
	echo You can run frpcgui.exe directly or distribute the zip files to other users.
	echo.
	pause
	exit /b 0

:error
	echo.
	echo [ERROR] Build failed with error code: %errorlevel%
	echo.
	pause
	exit /b %errorlevel%