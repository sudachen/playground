@echo off
cd %~dp0

for /D %%i in (classic, sputnik) do (
	call :RM %%i\benchmark.js
	call :RM %%i\benchmark.pprof
	call :RM %%i\benchmark.exe
)

goto :EOF

:RM
if exist %1 del /Q %1
exit /B




