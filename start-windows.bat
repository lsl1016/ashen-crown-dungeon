@echo off
cd /d %~dp0
echo Starting Ashen Crown V0.6 server...
start "Ashen Crown Server" /min "dist\ashen-crown-windows-amd64.exe" -web "web" -data "data"
timeout /t 2 /nobreak >nul
start "" http://localhost:8080
echo Game opened at http://localhost:8080
echo If the browser reports that it cannot connect, wait a moment and refresh once.
