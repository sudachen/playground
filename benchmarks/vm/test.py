import vmbench as tk

vmbench = tk.Benchmark("vm")
classic = tk.Branch("classic").execute(vmbench, pprof=tk.BENCHMARK_PPROF)
#sputnik = tk.Branch("sputnik").execute(vmbench, pprof=tk.BENCHMARK_PPROF)


