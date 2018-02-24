import sys

sys.path = ["C:/Projects/_GoPkg_/src/github.com/sudachen/benchmark/py"] + sys.path

import vmbench as tk

vmbench = tk.Benchmark("vm")
classic = tk.Branch("classic").load_or_execute(vmbench, pprof=tk.BENCHMARK_PPROF, mprof=tk.BENCHMARK_MPROF, callgraph=True)
sputnik = tk.Branch("sputnik").load_or_execute(vmbench, pprof=tk.BENCHMARK_PPROF, mprof=tk.BENCHMARK_MPROF, callgraph=True)


