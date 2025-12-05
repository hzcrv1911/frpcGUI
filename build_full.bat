@echo off
setlocal enabledelayedexpansion
set GOARCH_x64=amd64
set GOARCH_x86=386
set BUILDDIR=%~dp0
cd /d %BUILDDIR% || exit /b 1

REM 默认构建所有架构
set FRPCGUI_TARGET=%~1
if "%FRPCGUI_TARGET%" == "" set FRPCGUI_TARGET=x64 x86

:packages
	echo [+] Downloading packages
	go mod tidy || goto :error

:resources
	echo [+] Generating resources
	for /f %%a in ('go generate') do set %%a
	if not defined VERSION exit /b 1

:build
	echo [+] Building program
	set MOD=github.com/hzcrv1911/frpcgui
	set GO111MODULE=on
	set CGO_ENABLED=0
	for %%a in (%FRPCGUI_TARGET%) do (
		if defined GOARCH_%%a (
			set GOARCH=!GOARCH_%%a!
		) else (
			set GOARCH=%%a
		)
		go build -trimpath -ldflags="-H windowsgui -s -w -X %MOD%/pkg/version.BuildDate=%BUILD_DATE%" -o bin\%%a\frpcgui.exe .\cmd\frpcgui || goto :error
		REM 清理并复制 assets 目录到当前架构目录
		if exist bin\%%a\assets rmdir /S /Q bin\%%a\assets
		if exist assets (
			robocopy assets bin\%%a\assets /E /NFL /NDL /NJH /NJS /nc /ns /np
			if !errorlevel! gtr 7 goto :error
		)
	)

:create_archives
	echo [+] Creating portable archives
	for %%a in (%FRPCGUI_TARGET%) do (
		tar -ac -C bin\%%a -f bin\frpcgui-%VERSION%-%%a.zip frpcgui.exe assets
	)

:success
	echo [+] Success
	echo [+] Portable executables created in bin\ directory
	echo [+] Archives created: frpcgui-%VERSION%-*.zip
	exit /b 0

:error
	echo [-] Failed with error %errorlevel%.
	exit /b %errorlevel%
