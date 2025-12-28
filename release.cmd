@echo off
setlocal enabledelayedexpansion

echo Config233-Go Auto Release Script
echo ================================

REM Check if version is provided
if "%1"=="" (
    echo Usage: %0 ^<version^>
    echo Example: %0 v1.0.0
    exit /b 1
)

set VERSION=%1

echo Releasing version %VERSION%
echo.

REM Check git status
echo Checking git status...
git status --porcelain
if errorlevel 1 (
    echo Error: Git command failed
    exit /b 1
)
if not errorlevel 0 (
    echo Error: Working directory is not clean. Please commit or stash changes.
    exit /b 1
)

REM Run tests
echo Running tests...
go test ./...
if errorlevel 1 (
    echo Error: Tests failed
    exit /b 1
)

REM Build
echo Building...
go build .
if errorlevel 1 (
    echo Error: Build failed
    exit /b 1
)

REM Create git tag
echo Creating git tag %VERSION%...
git tag -a %VERSION% -m "Release %VERSION%"
if errorlevel 1 (
    echo Error: Failed to create git tag
    exit /b 1
)

REM Push tag
echo Pushing tag to remote...
git push origin %VERSION%
if errorlevel 1 (
    echo Error: Failed to push tag
    exit /b 1
)

REM Push main branch
echo Pushing main branch...
git push origin main
if errorlevel 1 (
    echo Error: Failed to push main branch
    exit /b 1
)

echo.
echo Release %VERSION% completed successfully!
echo The module will be available at: https://pkg.go.dev/config233-go@%VERSION%
echo.