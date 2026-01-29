@echo off
setlocal ENABLEDELAYEDEXPANSION
title ReserveChain DevNet Node

REM ---------------------------------------------------------------------------
REM  ReserveChain DevNet - Node Launcher (Windows)
REM  This script starts a single DevNet node using the bundled Go runtime.
REM  Assumptions:
REM    - You run it from the project root via: scripts\start_node.bat
REM    - Go is located at: runtime\go\bin\go.exe
REM    - Config is: config\devnet.yaml
REM ---------------------------------------------------------------------------

REM Resolve project root (one level up from scripts\)
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%.."
set "ROOT=%CD%"

echo ===============================================
echo   ReserveChain DevNet - Node Starter (Windows)
echo ===============================================
echo   Project root : %ROOT%
echo   Go runtime   : %ROOT%\runtime\go\bin\go.exe
echo   Config file  : %ROOT%\config\devnet.yaml
echo -----------------------------------------------
echo   Press CTRL+C to stop the node cleanly.
echo ===============================================
echo.

REM Check Go runtime
if not exist "%ROOT%\runtime\go\bin\go.exe" (
  echo [ERROR] Go runtime not found at:
  echo         %ROOT%\runtime\go\bin\go.exe
  echo.
  echo Please place go.exe under runtime\go\bin\ and try again.
  echo.
  pause
  goto :EOF
)

REM Check config file
if not exist "%ROOT%\config\devnet.yaml" (
  echo [ERROR] Config file not found:
  echo         %ROOT%\config\devnet.yaml
  echo.
  echo Please ensure devnet.yaml exists and is valid YAML.
  echo.
  pause
  goto :EOF
)

REM Start node (foreground). CTRL+C will be delivered to the Go process
REM so it can perform a clean shutdown and flush state.
"%ROOT%\runtime\go\bin\go.exe" run ./cmd/node

echo.
echo -----------------------------------------------
echo  Node process has exited.
echo -----------------------------------------------
echo.
pause
endlocal
