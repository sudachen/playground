
import sys
import os.path
import base64
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.ticker as mtick
from IPython.display import Markdown, display, Image

sys.path.append(os.path.dirname(os.path.dirname(__file__)))

from toolkit import *

def plot_bench_hist(base, target, transform):
    L = list(collect(extract(base, target),transform))
    P = pd.Series(L)

    mean, median, std = P.mean(), P.median(), P.std()
    S = pd.Series([x for x in L if x > mean-std*3 and x < mean+std*3])
    total = S.size
    drop = len(L)-total

    n_bins = int(S.std()*1.5)
    alpha = 0.5
    fig,ax = plt.subplots(figsize=(16,6))
    S.hist(bins=n_bins, alpha=alpha)
    ax.xaxis.set_major_formatter(mtick.PercentFormatter())
    ax.set_ylabel('Tests count', fontsize=18)
    ax.set_xlabel(
        'How many time takes {} VM against {} VM.  (median {:.0f}, μ {:.0f}, σ {:.0f})'.
            format(target.branch.label, base.branch.label, median, mean, std),
        fontsize=18)
    ax.set_title(
        'Comparision of {}/{} VMs benchmarks. ({} results, {} usable, {} dropped)'.
            format(target.branch.label, base.branch.label, len(L), total, drop),
        fontsize=18)

    plt.show()


def strip(f):
    rules = [
        ("github.com/sudachen/playground/branch/sputnik/vm/vendor/github.com/ethereumproject/",""),
        ("github.com/ethereumproject/",""),
        ("github.com/sudachen/","")
    ]
    for r in rules:
        if f.startswith(r[0]):
            f = r[1] + f[len(r[0]):]
    return f


def plot_pprof2(label1, p1, label2, p2):
    h = int(len(p1.rows)*4.5/10+0.5)
    fig, (ax1, ax2) = plt.subplots(nrows=2, ncols=4, figsize=(15,h), sharey='row')
    for label, p, ax in ((label1, p1, ax1),(label2, p2, ax2)):
        attr = ('cum','cum%','flat','flat%')
        for i in range(4):
            df = pd.DataFrame.from_records([(strip(x.function), x[attr[i]]) for x in p.rows],columns=(label, attr[i]))
            df.set_index([label], inplace=True)
            df.plot(kind='barh', ax=ax[i]).invert_yaxis()
    plt.tight_layout()
    plt.show()


def plot_pprof(title, what, base, target):
    display(Markdown("# "+title))
    plot_pprof2(base.branch.label, base.pprof[what], target.branch.label, target.pprof[what])


def plot_pprof_image(title, what, target):
    display(Markdown("# "+title))
    S = base64.b64decode(target.pprof[what].image)
    display(Image(S))


def plot_fast_slow_tests(base, target, transform):
    title = "How fast and slow is {} VM against {} VM by tests". \
        format(target.branch.label.title(), base.branch.label.title())
    good = lambda i: i.vars[0].active != 0  and i.vars[1].active != 0
    f1 = lambda b,t: (transform(b,t), b.label, t.label)
    Q = sorted([f1(i.vars[0], i.vars[1]) + (i.label, float(i.vars[0].active)/1000000000, float(i.vars[1].active)/1000000000) \
                for i in extract(base, target).values() if good(i) ] )
    f2 = lambda x: str(int(-x)) + '% faster' if x < 0 else str(int(x)) + '% slower'
    d = lambda x: (x[3],f2(x[0]),'{:.3f}s'.format(x[5]),'{:.3f}s'.format(x[4]))
    X = [d(x) for x in Q[:15]] + [('...','','','')] + [d(x) for x in Q[-15:]]
    pFld = '{} vs {}'.format(target.branch.label.title(),base.branch.label.title())
    fields = ['Name',pFld,target.branch.label.title(),base.branch.label.title()]
    F = pd.DataFrame(X,columns=fields).set_index(fields)
    display(Markdown("# "+title))
    display(F)
