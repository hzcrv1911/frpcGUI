@echo off
echo 正在配置frpcgui项目所需的环境变量...

REM 设置Go环境变量
set GO_HOME=D:\Apps\Go
set PATH=%GO_HOME%\bin;%PATH%

REM 设置MinGW环境变量
set MINGW_HOME=D:\Apps\llvm-mingw-20251104-msvcrt-x86_64
set PATH=%MINGW_HOME%\bin;%PATH%

REM 设置Windows SDK环境变量（请根据实际安装路径调整版本号）
set WindowsSdkVerBinPath=C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\

REM 添加WiX Toolset到PATH（如果通过dotnet tool安装，通常不需要手动设置）
REM set WIX_HOME=C:\Program Files (x86)\WiX Toolset v6.0.2\bin
REM set PATH=%WIX_HOME%;%PATH%

echo 环境变量已设置完成！
echo.
echo 当前环境变量：
echo GO_HOME=%GO_HOME%
echo MINGW_HOME=%MINGW_HOME%
echo WindowsSdkVerBinPath=%WindowsSdkVerBinPath%
echo.
echo PATH已更新，包含以下路径：
echo %GO_HOME%\bin
echo %MINGW_HOME%\bin
echo.
echo 验证安装：
echo.
echo Go版本：
go version
echo.
echo GCC版本：
gcc --version
echo.
echo 请确保已安装WiX Toolset v6.0.2
echo.
echo 注意：这些环境变量仅在当前命令提示符会话中有效。
echo 如需永久设置，请通过系统属性中的环境变量设置。
pause