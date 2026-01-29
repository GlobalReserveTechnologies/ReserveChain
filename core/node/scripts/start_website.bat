@echo off
setlocal ENABLEDELAYEDEXPANSION
title ReserveChain DevNet Website

REM ---------------------------------------------------------------------------
REM  ReserveChain DevNet - Website / Workstation Launcher (Windows)
REM  This script starts the PHP built-in server to serve the marketing site
REM  and workstation SPA from the /public directory.
REM  Assumptions:
REM    - You run it from the project root via: scripts\start_website.bat
REM    - PHP is located at: runtime\php\php.exe
REM ---------------------------------------------------------------------------

REM Resolve project root
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%.."
set "ROOT=%CD%"

set "HOST=127.0.0.1"
set "PORT=8090"

echo ===============================================
echo   ReserveChain DevNet - Website Starter
echo ===============================================
echo   Project root : %ROOT%
echo   PHP runtime  : %ROOT%\runtime\php\php.exe
echo   Doc root     : %ROOT%\public
echo   URL          : http://%HOST%:%PORT%/
echo -----------------------------------------------
echo   Press CTRL+C to stop the server.
echo ===============================================
echo.

REM Check PHP runtime
if not exist "%ROOT%\runtime\php\php.exe" (
  echo [ERROR] PHP runtime not found at:
  echo         %ROOT%\runtime\php\php.exe
  echo.
  echo Please place php.exe under runtime\php\ and try again.
  echo.
  pause
  goto :EOF
)

REM Check public directory
if not exist "%ROOT%\public" (
  echo [ERROR] Public web root not found:
  echo         %ROOT%\public
  echo.
  echo Please ensure the public/ directory exists.
  echo.
  pause
  goto :EOF
)

"%ROOT%\runtime\php\php.exe" -S %HOST%:%PORT% -t public

echo.
echo -----------------------------------------------
echo  PHP server has exited.
echo -----------------------------------------------
echo.
pause
endlocal
