
@echo off
cd %~dp0

call :BENCHMARK classic 17
call :BENCHMARK sputnik 20

goto :EOF

:BENCHMARK
cd %1
echo benchamrking in %1
go run benchmark.go --pprof --cpuprof=benchmark.pprof --callgraph=%2 --result=benchmark.js
cd ..
