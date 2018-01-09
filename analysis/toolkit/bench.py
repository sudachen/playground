
import os
import os.path
import json
import copy
from . import branch
from . import exec
from . import util


class Benchmark(object):
    __slots__ = ["label","file","pfx","env"]

    def __init__(self, label, file, pfx, env=exec.Env()):
        self.label = label
        self.file = file
        self.pfx = pfx
        self.env = env

    def path_with(self,root):
        return os.path.join(root,os.path.dirname(self.file))


class GoBenchmark(Benchmark):

    def __init__(self, label, name, pfx, env=exec.Env()):
        file = os.path.join("benchmarks",name,"benchmark.go")
        Benchmark.__init__(self,label,file,pfx,env)


benchmarks = [
    GoBenchmark(label="VM",name="benchvm",pfx="vm_")
]


class BenchmarkError(Exception):
    def __init__(self,text):
        super(Exception, self).__init__(self, text)


class UnknownBenchmarkError(BenchmarkError):
    def __init__(self,benchmark_label):
        super(BenchmarkError, self).__init__(self, "unknown benchmark {}".format(benchmark_label))


class ExecutionBenchmarkError(BenchmarkError):
    def __init__(self,benchmark_label,reason):
        super(BenchmarkError, self).__init__(self, "benchmark {} failed: {}".format(benchmark_label,reason))


class MsgKind(object):
    def __str__(self):
        if self is MsgError:
            return "Error"
        if self is MsgInfo:
            return "Info"
        if self is MsgDebug:
            return "Debug"
        if self is MsgOpt:
            return "Opt"
        raise ValueError()


MsgError = MsgKind()
MsgDebug = MsgKind()
MsgInfo = MsgKind()
MsgOpt = MsgKind()


class Message(object):
    __slots__ = ["kind", "text"]

    def __init__(self,kind,text):
        self.kind = kind
        self.text = text

    def __repr__(self):
        return 'Message(kind="{}", text="{}")'.format(
            self.kind,
            self.text
        )


class Task(object):
    __slots__ = ["label", "total", "active", "count", "error", "children", "messages"]

    def __init__(self, label, total, active, count, error, children, messages):
        self.label = label
        self.total = total
        self.active = active
        self.count = count
        self.error = error
        self.children = children
        self.messages = messages

    def __repr__(self):
        return 'Task(label="{}", total={}, active={}, count={}, error={}, children={}, messages={})'.format(
            self.label,
            self.total,
            self.active,
            self.count,
            repr(self.error),
            self.children,
            self.messages
        )


class PprofRow(object):
    __slots__ = ('flat','flatP','sumP','cum','cumP','function')
    columns = ("flat","flat%","sum%","cum","cum%","function")

    def __init__(self, flat, flatP, sumP, cum, cumP, function):
        self.flat = flat
        self.flatP = flatP
        self.sumP = sumP
        self.cum = cum
        self.cumP = cumP
        self.function = function

    def __repr__(self):
        return "PprofRow(flat={}, flatp={}, sumP={}, cum={}, cumP={} function='{}')".format(
            self.flat,self.flatP,self.sumP,self.cum,self.cumP,self.function)

    def __getitem__(self, item):
        if item == "flat":
            return self.flat
        if item == "flat%":
            return self.flatP
        if item == "sum%":
            return self.sumP
        if item == "cum":
            return self.cum
        if item == "cum%":
            return self.cumP
        if function == "function":
            return self.cumP
        raise KeyError(item)

    def __iter__(self):
        yield self.flat
        yield self.flatP
        yield self.sumP
        yield self.cum
        yield self.cumP
        yield self.function

    def __len__(self):
        return len(self.columns)

class PprofUnit(object):
    __slots__ = ['label']

    def __init__(self, label):
        self.label = label

    def __repr__(self):
        return self.label


Msec = PprofUnit("ms")
Usec = PprofUnit("us")
Sec = PprofUnit("s")


class Pprof(object):
    __slots__ = ['label', 'options', 'unit', 'rows', 'errors']

    def __init__(self, label, options, unit, rows, errors):
        self.label = label
        self.options = options
        if unit == 'Msec':
            self.unit = Msec
        elif unit == 'Usec':
            self.unit = Usec
        elif unit == 'Sec':
            self.unit = Sec
        self.rows = rows
        self.errors = errors

    def __repr__(self):
        return "Pprof(label='{}', options='{}', unit='{}', rows={}, errors={})".format(
            self.label, self.options, self.unit, self.rows, self.errors)


def load_results(f):

    def decode_object(m):
        if "kind" in m:
            kind = m["kind"]
            if kind == "MsgError":
                kind = MsgError
            elif kind == "MsgInfo":
                kind = MsgInfo
            elif kind == "MsgDebug":
                kind = MsgDebug
            elif kind == "MsgOpt":
                kind = MsgOpt
            else:
                raise ValueError()
            return Message(kind,m["text"])
        elif "flat%" in m:
            return PprofRow(
                float(m["flat"]),
                float(m["flat%"]),
                float(m["sum%"]),
                float(m["cum"]),
                float(m["cum%"]),
                m["function"]
            )
        elif "rows" in m:
            return Pprof(
                m["label"],
                m["options"],
                m["unit"],
                m.get("rows",None),
                m.get("errors",None)
            )
        elif "label" in m:
            t = Task(
                m["label"],
                int(m["total"]),
                int(m["active"]),
                int(m["count"]),
                m.get("error",None),
                m.get("children",None),
                m.get("messages",None),
            )
            if m["label"] == '.':
                return (t,m.get("pprof",None))
            return t
        return m

    return json.load(f,object_hook=decode_object)


def look_for(label):
    for b in benchmarks:
        if b.label == label:
            return b
    raise UnknownBenchmarkError(label)


class Result(object):
    __slots__ = ["branch", "results", "pprof"]

    def __init__(self, branch, results, pprof):
        self.branch = branch
        self.results = results
        self.pprof = pprof

    def __repr__(self):
        return "Result(branch='{}', results={}, pprof={})".format(
            self.branch,
            self.results,
            self.pprof
        )


def load(label,file):
    r, ppf = load_results(file)
    return Result(label, r, {i.label:i for i in ppf} )

def execute(root, the_branch, the_bench, temp=None, pprof=None):
    branch_dir = the_branch.path_with(root)
    file = os.path.join(branch_dir, the_bench.file)
    ext = os.path.splitext(file)[1][1:]
    e = exec.lookup_for(ext)
    status, stdout, stderr = e.execute_bench(file, the_bench.env, temp, pprof)
    if status is not exec.Success:
        raise ExecutionBenchmarkError(the_bench.label, status.reason)
    return load(the_branch.label,stdout)

def run_benchmark(bench_label, branches, root=util.root_dir(), temp=None, pprof=None):
    "it does benchmark for specified branches or for ALL branches if specified ALL"
    the_bench = look_for(bench_label)
    results = (execute(root, b, the_bench, temp, pprof) for b in branch.list_of(branches))
    return results

