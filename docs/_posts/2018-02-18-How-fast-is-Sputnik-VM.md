---
layout: post
title: How fast is Sputnik VM
---

In december 2017, ETCDEV Team has added Sputnik VM to go-etherm client for ETC blockchain. 
I tried to compare performance of SputnikVM (Rust) against original Classic VM (Go). 

I benchmarked both of them over subset of state tests from ethereumproject source tree and published following 
[Benchmark Visualizaion in iPython Notebook](https://github.com/sudachen/playground/blob/master/benchmarks/vm/vmbench.ipynb)

![Sputnk VM vs Classic VM]({{site.baseurl}}/assets/posts/2018-02-18/sputnik_vs_classic.png)

What does it mean?

<!--more-->

Optimized Rust code is not so fast and runs with tspeed equal to GO?! Hm, maybe Go part takes so match time and eleminate all Rust benifits?

Let's look at the Sputnk VM top calls gathered through Go pprof. 

![Sputnik VM Top Calls]({{site.baseurl}}/assets/posts/2018-02-18/sputnik_top_calls.png)

Hm, looks like all time spended in the Rust code and in the CGO (Go foreign function interface). 
How many times exactly tooks CGO is not so important because for go-ethereum client both Rust and CGO is the same.

By the way, where Sputnik VM is faster and dramtically slower then the Go VM implementation?!

![Sputnik VM vs Classi VM by tests]({{site.baseurl}}/assets/posts/2018-02-18/sputnik_fast_slow.png)

