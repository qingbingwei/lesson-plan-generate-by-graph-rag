@echo off
setlocal

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0stop-win.ps1" %*
exit /b %errorlevel%
