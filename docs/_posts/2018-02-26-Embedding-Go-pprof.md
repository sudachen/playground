---
layout: post
title: Embedding Go Pprof
---

As you know Golang has time-based profiling subsytem and visualization tool called pprof. 
The pprof is not friendly for use in a manner different than command line tool.
However there is a way to embed it in an application to generate some reports in tests/benchmarks cases.
I use it in my benchmark module to generate benchmarking report contaning top calls and callgrapth of the significant functions.

<!--more-->

[pprof/pprof.go:26](https://github.com/google/pprof/blob/db0be723d40dfbb90c702d493a71d398173358e7/pprof.go#L26)
```golang
func main() {
    if err := driver.PProf(&driver.Options{}); err != nil {
    	fmt.Fprintf(os.Stderr, "pprof: %v\n", err)
    	os.Exit(2)
    }
}
```

It's the main module of pprof tool. The package pprof/driver is a frontend to the pprof internals which
adopts the оptions and calls to the pprof/internal/driver package.

[pprof/driver/driver.go:31](https://github.com/google/pprof/blob/db0be723d40dfbb90c702d493a71d398173358e7/driver/driver.go#L31)
```golang
func PProf(o *Options) error {
    return internaldriver.PProf(o.internalOptions())
}
```

So, there are two resonable questions: what exactly is the pprof/driver.Option and how can I play with it? 
Since all functionality is in the internal package, the only way to use this funtionality,
without modification the original code, to use the both pprof/driver.Option and pprof/driver.Pprof.

The pprof/driver.Options is a set of objects which can control the pprof actions

[pprof/driver/driver.go:55](https://github.com/google/pprof/blob/db0be723d40dfbb90c702d493a71d398173358e7/driver/driver.go#L55)
```golang
type Options struct {
    Writer  Writer      
    Flagset FlagSet
    Fetch   Fetcher
    Sym     Symbolizer
    Obj     ObjTool
    UI      UI
}
```

The fields **Options.Sym** and **Options.Obj** are used to customize symbolyzation. It's not nesessary to use these fields in most cases.

The **Options.UI** presents an object manages user interactions

[pprof/driver/driver.go:189](https://github.com/google/pprof/blob/db0be723d40dfbb90c702d493a71d398173358e7/driver/driver.go#L189)
```golang
type UI interface {
    ReadLine(prompt string) (string, error)
    Print(...interface{})
    PrintErr(...interface{})
    IsTerminal() bool
    SetAutoComplete(complete func(string) string)
}
```

The **Options.Flagset** presents an object with logic similar to the standard flag.FlagSet.
Pprof can access to commandline arguments and options via this object.

[pprof/driver/driver.go:72](https://github.com/google/pprof/blob/db0be723d40dfbb90c702d493a71d398173358e7/driver/driver.go#L72)
```golang
type FlagSet interface {
    Bool(name string, def bool, usage string) *bool
    Int(name string, def int, usage string) *int
    Float64(name string, def float64, usage string) *float64
    String(name string, def string, usage string) *string
    BoolVar(pointer *bool, name string, def bool, usage string)
    IntVar(pointer *int, name string, def int, usage string)
    Float64Var(pointer *float64, name string, def float64, usage string)
    StringVar(pointer *string, name string, def string, usage string)
    StringList(name string, def string, usage string) *[]*string
    ExtraUsage() string
    Parse(usage func()) []string
}
```

The **Options.Fetch** presents object wich fetches pprof/profile.Profile. Only this object can really know from where is the perf-profile.

[pprof/driver/driver.go:108](https://github.com/google/pprof/blob/db0be723d40dfbb90c702d493a71d398173358e7/driver/driver.go#L108)
```golang
type Fetcher interface {
    Fetch(src string, duration, timeout time.Duration)(*profile.Profile, string, error)
}
```

The **Options.Writer** - as I see the idea of this filed is setup an object handling file operations directd by --output flag. (spoiler: it does not work yet)

[pprof/driver/driver.go:66](https://github.com/google/pprof/blob/db0be723d40dfbb90c702d493a71d398173358e7/driver/driver.go#L66)
```golang
type Writer interface {
    Open(name string) (io.WriteCloser, error)
}
```

Now, let be a little magic... but before I still need a specific objects.
One which implemets pprof/driver.Fetcher and other - pprof/driver.FlagSet

Local profile fetcher 

[benchmark/ppftool/fetcher.go:11](https://github.com/sudachen/benchmark/blob/master/ppftool/fetcher.go#L11)
```golang
type fetcher struct {
    b []byte
}

func (f *fetcher) Fetch(src string, duration, timeout time.Duration)(*profile.Profile, string, error) {
    	p, err := profile.ParseData(f.b)
    	return p, "", err
    }
    return nil, "", fmt.Errorf("unknown source %s", src)
}

func Fetcher(b []byte) driver.Fetcher {
    return &fetcher{b}
}
```

and custom flagset

[benchmark/ppftool/flagset.go:8](https://github.com/sudachen/benchmark/blob/master/ppftool/flagset.go#L8)
```golang
type FlagSet struct {
    *flag.FlagSet
    args []string
}

func (f *FlagSet) StringList(name string, def string, usage string)*[]*string {
    return &[]*string{f.FlagSet.String(name, def, usage)}
}

func (f *FlagSet) ExtraUsage() string {
    return ""
}

func (f *FlagSet) Parse(usage func()) []string {
    f.FlagSet.Usage = func() {}
    f.FlagSet.Parse(f.Args)
    return f.FlagSet.Args()
}

func Flagset(a ...string) driver.FlagSet {
    return &FlagSet{
                flag.NewFlagSet("ppf", flag.ContinueOnError),
                append(a,"")}
}
```

Now ... The MAGIC!

[playground/docs/samples/pprof/magic1.go:12](https://github.com/sudachen/playground/blob/master/docs/samples/pprof/magic1.go#L12)
```golang
func main() {
	var bf bytes.Buffer

	pprof.StartCPUProfile(&bf)
	for s:= ""; len(s) < 100000;  {
		s = s + fmt.Sprintf("%d", len(s))
	}
	pprof.StopCPUProfile()

	driver.PProf(&driver.Options{
		Fetch:   ppftool.Fetcher(bf.Bytes()),
		Flagset: ppftool.Flagset("-top", "-nodecount=5"),
	})
}
```

```text
>go run magic1.go

Main binary filename not available. 
Type: cpu 
Time: Feb 22, 2018 at 5:08pm (-03) 
Duration: 501.76ms, Total samples = 530ms (105.63%) 
Showing nodes accounting for 380ms, 71.70% of 530ms total 
Showing top 5 nodes out of 61 
      flat  flat%   sum%        cum   cum% 
     190ms 35.85% 35.85%      190ms 35.85%  runtime.memmove 
     100ms 18.87% 54.72%      190ms 35.85%  runtime.scanobject 
      40ms  7.55% 62.26%       40ms  7.55%  runtime.heapBits.bits (inline)
      30ms  5.66% 67.92%       30ms  5.66%  runtime.memclrNoHeapPointers
      20ms  3.77% 71.70%       20ms  3.77%  runtime.gcmarknewobject 
```

It works! ... but is not useful.

Since **Options.Writer** is not really used by the pprof/internal/driver I can't redirect output gracefully.

[pprof/internal/driver/driver.go:139](https://github.com/google/pprof/blob/db0be723d40dfbb90c702d493a71d398173358e7/internal/driver/driver.go#L139)
```golang
func generateReport(p *profile.Profile, cmd []string, vars variables, o *plugin.Options) error {
    ...

    // Output to specified file.
    o.UI.PrintErr("Generating report in ", output)
    out, err := os.Create(output)
    if err != nil {
    	return err
    }
    if _, err := src.WriteTo(out); err != nil {
    	out.Close()
    	return err
    }
    return out.Close()
}
```

So I will use external file, util it's not fixed. It's dirty but it works.

[playground/docs/samples/pprof/magic2.go:12](https://github.com/sudachen/playground/blob/master/docs/samples/pprof/magic2.go#L12)
```golang
func main() {
    ... the same code here

    driver.PProf(&driver.Options{
    	Fetch:   ppftool.Fetcher(bf.Bytes()),
        Flagset: ppftool.Flagset("-top", "-nodecount=5", "-unit=s",
                                 "-output=pprof.output.txt"),
    	UI:      ppftool.FakeUi(), // supress errors and unwanted messages
    })
}

```

Now ``go run magic2.go`` prints nothing, but creates file pprof.output.txt containing profiling output. 
To be useful this output need to be parsed and represented in some regular structure.

I can represent profiling output as a sort of structured report. I think, it's quite useful.

[benchmark/ppftool/report.go:9](https://github.com/sudachen/benchmark/blob/master/ppftool/report.go#L9)
```golang
type Report struct {
    Unit
    Rows
    ...
}

type Rows []*Row

type Row struct {
    Function string

    Flat, FlatPercent, SumPercent, Cum, CumPercent float64
}

type Unit byte
const (
    Second 	 Unit = 0
    Millisecond Unit = 1
    Microsecond Unit = 2
    ...
)

```

And, of couse, I need to convert text output of the pprof into structured records.

[benchmark/ppftool/report.go:87](https://github.com/sudachen/benchmark/blob/master/ppftool/report.go#L63)
```golang
func (r *Report) Write(b []byte) {
    ...

    tf := func(s string) (x []string) {
        ... // extracting fields into array x
    }

    skip := true
    for _, l := range strings.Split(string(b), "\n") {
    	a := tf(l)
    	if skip && "flat flat% sum% cum cum%" == strings.Join(a, " ") {
    	    skip = false
    	}
    	if !skip && len(a) > 5 {
    	    i := &Row{}
    	    fmt.Sscanf(a[0], "%f", &i.Flat)
    	    fmt.Sscanf(a[1], "%f", &i.FlatPercent)
    	    fmt.Sscanf(a[2], "%f", &i.SumPercent)
    	    fmt.Sscanf(a[3], "%f", &i.Cum)
    	    fmt.Sscanf(a[4], "%f", &i.CumPercent)
    	    i.Function = a[5]
    	    r.Rows = append(r.Rows, i)
    	}
    }
    return
}
```

Ok, compile all together now

[playground/docs/samples/pprof/magic3.go:14](https://github.com/sudachen/playground/blob/master/docs/samples/pprof/magic3.go#L14)
```golang
func main() {
    var bf bytes.Buffer

    pprof.StartCPUProfile(&bf)
    for s:= ""; len(s) < 100000;  {
    	s = s + fmt.Sprintf("%d", len(s))
    }
    pprof.StopCPUProfile()

    tempfile := "pprof.output.txt"
    unit := util.DefaultUnit
    rpt := &util.Report{Unit: unit}

    driver.PProf(&driver.Options{
    	Fetch:   ppftool.Fetcher(bf.Bytes()),
        Flagset: ppftool.Flagset("-top", "-nodecount=5", "-unit="+unit.String(),
                                 "-output="+tempfile),
    	UI:      ppftool.FakeUi(),
    })

    if b, err := ioutil.ReadFile(tempfile); err == nil {
    	rpt.WriteTop(b)
    	os.Remove(tempfile)
    }

    fmt.Printf("%10s %11s %s\n", "flat", "%flat", "function")
    for _, row := range rpt.Rows {
        fmt.Printf("%10.3f %10.3f%% %s\n",
                   row.Flat, row.FlatPercent, row.Function)
    }
}
``` 

Bingo! But still is not friendly. The good idea is to make a function implementing top command with constraints.

Pprof constraints ... yes, pprof can limit samples by filters like -show, -hide, -focus. So I need a struct represented filtering options.

[benchmark/ppftool/options.go:10](https://github.com/sudachen/benchmark/blob/master/ppftool/options.go#L10)
```golang
type Options struct {
    Unit // Second, Millisecond, Microsecond
    ...
    Count     int      // -nodecount=
    CumSort   bool     // -cum=
    TagShow   []string // -tagshow=
    TagHide   []string // -taghide=
    TagIgnore []string // -tagignore=
    TagFocus  []string // -tagfocus=
    Ignore    []string // -ignore=
    Focus     []string // -focus=
    Show      []string // -show=
    Hide      []string // -hide=
    ...
}

func (o *Options) flagset(c ...string) driver.FlagSet {
    if o.Count > 0 {
    	c = append(c, fmt.Sprintf("-nodecount=%d", o.Count))
    }

    // ... processing filters here
	
    return Flagset(c...)
}
```

Now the top command can be implemented easily.
All I need is call pprof with set of arguments generated from Options and then convert temporal
file to structured Report.

[benchmark/ppftool/top.go:11](https://github.com/sudachen/benchmark/blob/master/ppftool/top.go#L11)
```golang
func Top(b []byte, o *Options) (*Report, error) {
    tempfile := TempFileName()
    rpt := &Report{Unit: o.Unit}

    err := driver.PProf(&driver.Options{
    	Fetch:   &fetcher{b},
    	Flagset: o.flagset("-top", "-output="+tempfile),
    	UI:      &ui{report: rpt},
    })

    if err != nil {
    	return nil, err
    }

    if b, err := ioutil.ReadFile(tempfile); err != nil {
    	return nil, err
    } else {
    	rpt.WriteTop(b)
    }

    os.Remove(tempfile)

    return rpt, nil
}
```

And now, finally, frendrly magic appears

[playground/docs/samples/pprof/magic4.go:12](https://github.com/sudachen/playground/blob/master/docs/samples/pprof/magic4.go#L12)
```golang
func main() {
    var bf bytes.Buffer

    pprof.StartCPUProfile(&bf)
    for s:= ""; len(s) < 100000;  {
    	s = s + fmt.Sprintf("%d",len(s))
    }
    pprof.StopCPUProfile()

    rpt, err := ppftool.Top(
                    bf.Bytes(),
                    &ppftool.Options{Count: 5, Hide: []string{"runtime\\."}})

    if err != nil {
    	fmt.Println(err)
    	os.Exit(2)
    }

    fmt.Printf("%10s %11s %s\n","flat","%flat","function")
    for _, row := range rpt.Rows {
        fmt.Printf("%10.3f %10.3f%% %s\n",
                   row.Flat,row.FlatPercent,row.Function)
    }
}
```

```text
>go run magic4.go           
      flat       %flat function
     0.180     28.570% main.main
     0.010      1.590% sync.(*Pool).pinSlow
     0.010      1.590% sync.poolCleanup
     0.000      0.000% fmt.Sprintf
     0.000      0.000% fmt.newPrinter
```
