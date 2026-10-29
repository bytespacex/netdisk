@echo off
REM Build netdisk for Windows (amd64) -> dist\netdisk-windows_amd64.exe
REM Version is stamped from the date (YY.MM.DD); set VERSION to override.
setlocal
cd /d "%~dp0"

if "%VERSION%"=="" (
  for /f %%i in ('powershell -NoProfile -Command "Get-Date -Format yy.MM.dd"') do set "VERSION=%%i"
)

if not exist dist mkdir dist

set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

echo netdisk %VERSION%
go build -trimpath -ldflags="-s -w -X main.Version=%VERSION%" -o dist\netdisk-windows_amd64.exe .
echo done -^> dist\
endlocal
