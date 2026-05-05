@echo off
setlocal

set "ROOT=%~dp0.."
set "SRC=%ROOT%\src"
set "PY_SRC=%ROOT%\src1\main.py"
set "PY_EXE=%ROOT%\.venv\Scripts\python.exe"
set "EXE=%ROOT%\bin\solver.exe"

if exist "%EXE%" (
    "%EXE%" %*
    exit /b %errorlevel%
)

where go >nul 2>nul
if not errorlevel 1 (
    pushd "%SRC%"
    go run . %*
    set "CODE=%errorlevel%"
    popd
    exit /b %CODE%
)

if exist "%PY_EXE%" (
    "%PY_EXE%" "%PY_SRC%" %*
    exit /b %errorlevel%
)

if exist "%PY_SRC%" (
    py -3 "%PY_SRC%" %*
    exit /b %errorlevel%
)

echo Jalankan build terlebih dahulu, atau install Go.
exit /b 1
