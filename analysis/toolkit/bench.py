
import os
import os.path
import json
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
        if self is MsgPprof:
            return "Pprof"
        raise ValueError()


MsgError = MsgKind()
MsgDebug = MsgKind()
MsgPprof = MsgKind()
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
        elif "label" in m:
            return Task(
                m["label"],
                int(m["total"]),
                int(m["active"]),
                int(m["count"]),
                m.get("error",None),
                m.get("children",None),
                m.get("messages",None),
            )
        return m

    return json.load(f,object_hook=decode_object)


def look_for(label):
    for b in benchmarks:
        if b.label == label:
            return b
    raise UnknownBenchmarkError(label)


class Result(object):
    __slots__ = ["branch", "results"]

    def __init__(self, branch, results):
        self.branch = branch
        self.results = results


def execute(root, the_branch, the_bench, temp=None):
    branch_dir = the_branch.path_with(root)
    file = os.path.join(branch_dir, the_bench.file)
    ext = os.path.splitext(file)[1][1:]
    e = exec.lookup_for(ext)
    status, stdout, stderr = e.execute_bench(file, the_bench.env, temp)
    if status is not exec.Success:
        raise ExecutionBenchmarkError(the_bench.label, status.reason)
    return Result(the_branch.label, load_results(stdout))


def run_benchmarks(bench_label, branches, root=util.root_dir(), temp=None):
    "it does benchmark for specified branches or for ALL branches if specified ALL"
    the_bench = look_for(bench_label)
    results = (execute(root, b, the_bench, temp) for b in branch.list_of(branches))
    return results

