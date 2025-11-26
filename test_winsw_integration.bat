@echo off
echo Testing WinSW Integration for frpcgui
echo.

REM Check if required files exist
echo Checking required files...
if not exist "frpcgui.exe" (
    echo ERROR: frpcgui.exe not found
    goto :error
)
if not exist "frpc.exe" (
    echo ERROR: frpc.exe not found
    goto :error
)
if not exist "winsw.exe" (
    if not exist "winsw-x64.exe" (
        echo ERROR: winsw.exe or winsw-x64.exe not found
        goto :error
    )
    set WINSW_EXE=winsw-x64.exe
) else (
    set WINSW_EXE=winsw.exe
)

echo All required files found.
echo.

REM Create a test configuration
echo Creating test configuration...
echo [common] > test_config.ini
echo server_addr = 127.0.0.1 >> test_config.ini
echo server_port = 7000 >> test_config.ini
echo token = test_token >> test_config.ini
echo. >> test_config.ini
echo [test_proxy] >> test_config.ini
echo type = tcp >> test_config.ini
echo local_ip = 127.0.0.1 >> test_config.ini
echo local_port = 8080 >> test_config.ini
echo remote_port = 8080 >> test_config.ini

echo Test configuration created.
echo.

REM Test WinSW configuration generation
echo Testing WinSW configuration generation...
powershell -Command "& {Add-Type -AssemblyName System.Xml; [xml]$config = Get-Content 'test_config.ini.winsw.xml' -ErrorAction SilentlyContinue; if ($config) { Write-Host 'WinSW configuration found:'; $config.OuterXml } else { Write-Host 'No WinSW configuration found' }}"

echo.
echo Testing completed.
echo.
echo To test the full integration:
echo 1. Run frpcgui.exe
echo 2. Import the test_config.ini file
echo 3. Try to start the service
echo 4. Check if the service is installed and running
echo.
goto :end

:error
echo.
echo Please ensure all required files are in the same directory.
echo Required files:
echo - frpcgui.exe
echo - frpc.exe
echo - winsw.exe or winsw-x64.exe
echo.
exit /b 1

:end
echo Test script completed.