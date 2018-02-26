
@echo off
cd %~dp0

if .%1. == .clean. goto :CLEAN

call :BENCHMARK classic 17
call :BENCHMARK sputnik 20

goto :EOF

:BENCHMARK
cd %1
echo benchamrking in %1
go run benchmark.go --pprof --mprof --cpuprof=benchmark.pprof --memprof=benchmark.mprof --result=benchmark.js
cd ..
exit /B

:CLEAN
for /D %%i in (classic, sputnik) do (
	call :RM %%i\benchmark.js
	call :RM %%i\benchmark.pprof
	call :RM %%i\benchmark.exe
)
goto :EOF

:RM
if exist %1 del /Q %1
exit /B
