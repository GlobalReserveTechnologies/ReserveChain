@echo off
setlocal ENABLEDELAYEDEXPANSION
title ReserveChain DevNet - Start All

REM ---------------------------------------------------------------------------
REM  ReserveChain DevNet - Start All (Windows)
REM  This script opens two windows:
REM    - Node (Go)
REM    - Website (PHP)
REM ---------------------------------------------------------------------------

set "SCRIPT_DIR=%~dp0"

echo ===============================================
echo   ReserveChain DevNet - Start All
echo ===============================================
echo   This will launch:
echo     - Node (Go)
echo     - Website (PHP)
echo -----------------------------------------------
echo   You can close this window after launch.
echo ===============================================
echo.

REM Start node
start "ReserveChain DevNet Node" cmd /c "%SCRIPT_DIR%start_node.bat"

REM Small delay to let node boot
timeout /t 2 /nobreak >nul

REM Start website
start "ReserveChain DevNet Website" cmd /c "%SCRIPT_DIR%start_website.bat"

echo All components launched.
echo You can visit: http://127.0.0.1:8090/workstation
echo.
pause
endlocal
