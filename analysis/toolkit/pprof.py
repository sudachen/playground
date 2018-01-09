
import os.path
import subprocess
from tempfile import NamedTemporaryFile
from . import util
from . import branch
from . import bench
from .util import Return, Fail

BENCHMARK_PPROF = "benchmark.pprof"

def run(workdir,*args):
    try:
        with NamedTemporaryFile(mode="w+b") as o:
            util.verbose("in dir='{}'".format(workdir))
            util.verbose("\texecuting: {}"," ".join(args))
            with subprocess.Popen(args,stdout=o,cwd=workdir) as p:
                result = p.wait()
            if result == 0:
                o.seek(0)
                return Return(o.read())
            else:
                return Fail("with exit code {}".format(result))
    except subprocess.SubprocessError as e:
        return Fail("with SubprocessError({})".format(e))
    except OSError as e:
        return Fail("with OSError({})".format(e))

def pprof_to(branch_label,bench_label,fmt="png",count=20):

    if fmt not in ("png","pdf","svg"):
        raise ValueError("invalied format "+fmt)

    branch_dir = branch.look_for(branch_label).path_with(util.root_dir())
    bench_dir = os.path.join(branch_dir,os.path.dirname(bench.look_for(bench_label).file))
    return run(bench_dir,"go","tool","pprof","--"+fmt,"--nodecount={}".format(count),BENCHMARK_PPROF)
