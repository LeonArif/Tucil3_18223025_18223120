@echo off
setlocal

set "ROOT=%~dp0.."
set "SRC=%ROOT%\src"
set "OUT=%ROOT%\bin\solver.exe"

where go >nul 2>nul
if errorlevel 1 (
    echo Go tidak ditemukan di PATH.
    echo Install Go agar binary solver.exe bisa dibangun.
    exit /b 1
)

pushd "%SRC%"
go build -o "%OUT%" .
set "CODE=%errorlevel%"
popd
if %CODE%==0 (
    echo Build sukses: %OUT%
)
exit /b %CODE%
