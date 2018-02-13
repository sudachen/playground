
import sys
import os.path

sys.path.append(os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))),'benchmark','py'))

from benchmark import *

set_root_dir(os.path.dirname(__file__))

class Branch(object):

    __slots__ = ('label')

    def __init__(self, label):
        self.label = label

    def dirname(self, bench_label):
        return os.path.join(os.path.dirname(__file__), bench_label, self.label)

    def temp(self, bench_label):
        return os.path.join(os.path.dirname(__file__), bench_label, self.label)

    def execute(self, bench, *a, **k):
        return bench.execute(self, *a, **k)


class Test(object):
    __slots__ = ('label', 'vars')

    def __init__(self, label, L):
        self.label = label
        self.vars = [None]*L

    def __repr__(self):
        return 'Test(label="{}", vars={})'.format(self.label, self.vars)


class Var(object):
    __slots__ = ('label', 'total', 'active', 'count', 'error')

    def __init__(self, label, total, active, count, error):
        self.label = label
        self.total = total
        self.active = active
        self.count = count
        self.error = error

    def __repr__(self):
        return 'Var(label="{}", total={}, active={}, error={})'.\
            format(self.label, self.total, self.active, repr(self.error))


def extract_tree(label, r, m, L, x):
    if not r.children:
        t = m.get(r.label, None)
        if t is None:
            t = Test(r.label,L)
            m[r.label] = t
        t.vars[x] = Var(label, r.total, r.active, r.count, r.error)
    else:
        for c in r.children:
            extract_tree(label, c, m, L, x)


def extract(*results):
    m = {}
    L = len(results)
    x = 0
    for r in results:
        extract_tree(r.branch.label, r.results, m, L, x)
        x += 1
    return m


def collect_task(t, transform, *idxs):
    b = []
    for i in idxs:
        x = t.vars[i]
        if x is None:
            return None
        b.append(x)
    return transform(*b)


def collect(r, transform, *idxs):
    if len(idxs) == 0:
        idxs = [0,1]
    return (i for i in (collect_task(v, transform, *idxs) for v in r.values()) if i is not None)


def percent_of(base, value):
    if base == 0 or value == 0:
        raise ValueError
    return int(value/base*100)


def active_percent_of(base, value):
    try:
        return percent_of(base.active, value.active)
    except ValueError:
        return None
