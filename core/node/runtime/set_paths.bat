@echo off
REM ReserveChain Runtime Path Setup
REM Scans the local runtime/ folder for bundled runtimes and writes runtime\paths.cfg.
REM Exports GO_EXE, PHP_EXE, SQLITE_EXE for the current command session.

setlocal ENABLEDELAYEDEXPANSION

REM Determine project root from this script's location (runtime\set_paths.bat -> root)
set "RUNTIME_DIR=%~dp0"
cd /d "%RUNTIME_DIR%.."
set "ROOT=%cd%"
set "RUNTIME_DIR=%ROOT%\runtime"

REM Initialise variables
set "GO_EXE="
set "PHP_EXE="
set "SQLITE_EXE="

REM Detect Go runtime
if exist "%RUNTIME_DIR%\go\bin\go.exe" (
    set "GO_EXE=%RUNTIME_DIR%\go\bin\go.exe"
)

REM Detect PHP runtime
if exist "%RUNTIME_DIR%\php\php.exe" (
    set "PHP_EXE=%RUNTIME_DIR%\php\php.exe"
)

REM Detect SQLite runtime
if exist "%RUNTIME_DIR%\sqlite\sqlite3.exe" (
    set "SQLITE_EXE=%RUNTIME_DIR%\sqlite\sqlite3.exe"
)

REM Write simple config file
set "CFG_FILE=%RUNTIME_DIR%\paths.cfg"
> "%CFG_FILE%" echo ROOT=%ROOT%
>> "%CFG_FILE%" echo GO_EXE=%GO_EXE%
>> "%CFG_FILE%" echo PHP_EXE=%PHP_EXE%
>> "%CFG_FILE%" echo SQLITE_EXE=%SQLITE_EXE%

echo =========================================
echo   ReserveChain Runtime Path Setup
echo =========================================
echo Root:   %ROOT%
echo Runtime:%RUNTIME_DIR%
echo.
echo GO_EXE     = %GO_EXE%
echo PHP_EXE    = %PHP_EXE%
echo SQLITE_EXE = %SQLITE_EXE%
echo.
echo Config written to: %CFG_FILE%
echo.

REM Export variables to caller
endlocal & (
  set "ROOT=%ROOT%"
  if defined GO_EXE set "GO_EXE=%GO_EXE%"
  if defined PHP_EXE set "PHP_EXE=%PHP_EXE%"
  if defined SQLITE_EXE set "SQLITE_EXE=%SQLITE_EXE%"
)
