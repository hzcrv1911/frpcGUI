@echo off
echo =================================================================
echo WARNING: This script will DELETE ALL GIT HISTORY.
echo It will keep only the current files as a single "Initial commit".
echo This operation is destructive and cannot be undone.
echo =================================================================
echo.

echo.
echo 0. Detecting current branch...
for /f "tokens=*" %%i in ('git branch --show-current') do set BRANCH_NAME=%%i
if "%BRANCH_NAME%"=="" (
    echo Error: Could not detect current branch name.
    goto :error
)
echo Detected current branch: %BRANCH_NAME%

set /p "confirm=Are you sure you want to reset history for branch '%BRANCH_NAME%'? Type 'yes' to confirm: "
if /i not "%confirm%"=="yes" (
    echo Operation cancelled.
    pause
    exit /b
)

echo.
echo 1. Creating new orphan branch 'latest_branch'...
call git checkout --orphan latest_branch
if %errorlevel% neq 0 goto :error

echo.
echo 2. Adding all files...
call git add -A
if %errorlevel% neq 0 goto :error

echo.
echo 3. Committing changes...
call git commit -am "Initial commit"
if %errorlevel% neq 0 goto :error

echo.
echo 4. Deleting old '%BRANCH_NAME%' branch...
call git branch -D %BRANCH_NAME%
if %errorlevel% neq 0 goto :error

echo.
echo 5. Renaming current branch to '%BRANCH_NAME%'...
call git branch -m %BRANCH_NAME%
if %errorlevel% neq 0 goto :error

echo.
echo 6. Force pushing to remote 'origin'...
call git push -f origin %BRANCH_NAME%
if %errorlevel% neq 0 goto :error

echo.
echo 7. Cleaning up local git garbage (reducing .git folder size)...
call git reflog expire --expire=now --all
call git gc --prune=now --aggressive

echo.
echo ==========================================
echo Success! Git history has been reset.
echo ==========================================
pause
exit /b

:error
echo.
echo !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
echo An error occurred. Please check the output above.
echo !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
pause
exit /b 1
