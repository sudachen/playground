---
layout: post
title: How fast is Sputnik VM
---

In december 2017, ETCDEV Team has added Sputnik VM to go-etherm client for ETC blockchain. 
I tried to compare performance of SputnikVM (Rust) against original Classic VM (Go). 

I benchmarked both of them over the subset of state tests from ethereumproject source tree and published following 
[Benchmark Visualizaion](https://github.com/sudachen/playground/blob/master/benchmarks/vm/README.md)

![Sputnk VM vs Classic VM](https://raw.githubusercontent.com/sudachen/playground/master/benchmarks/vm/_img/output_0_1.png)

What does it mean?

<!--more-->

Is optimized Rust code not so fast and runs with speed equal to Go?! Maybe Go part takes too much time and eleminate all Rust benifits?

Let's look at the Sputnik VM top calls gathered through Go pprof. 

![Sputnik VM Top Calls](https://raw.githubusercontent.com/sudachen/playground/master/benchmarks/vm/_img/output_0_11.png)

Looks like almost all the time is spent in the Rust code and in the CGO wrappers (Go foreign function interface). 
How much time exactly tooks CGO wrappers is not so important because for go-ethereum client both Rust and CGO is the same.

By the way, where Sputnik VM is faster and dramtically slower then the Go VM implementation?!

![Sputnik VM vs Classi VM by tests]({{site.baseurl}}/assets/2018-02-18-sputnik_fast_slow.png)

I theory, if run Sputnik VM on the big regular contracts from the real Blockchain it will take some benifits. 
I will test it latter with VMs from other clients to find how useful can be Sputnik VM. 
However now the benchmark shows that Sputnik VM does not have significant benifits against classic GO implementation of VM.

