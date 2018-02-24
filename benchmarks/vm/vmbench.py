
import base64
import tempfile
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.ticker as mtick
import graphviz as gv
from IPython.display import Markdown, display, Image
from toolkit import *


def plot_bench_hist(base, target, transform):
    L = list(collect(extract(target, base),transform))
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
        'How fast is {} VM against {} VM.  (median {:.0f}, μ {:.0f}, σ {:.0f})'.
            format(target.branch.label, base.branch.label, median, mean, std),
        fontsize=18)
    ax.set_title('{} results, {} usable, {} dropped'.\
                 format(len(L), total, drop),
                 fontsize=18)

    title = 'Comparision of {}/{} VMs benchmarks over {} test'.\
            format(target.branch.label.title(), base.branch.label.title(), len(L))

    display(Markdown("# "+title))
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
    S = base64.b64decode(target.pprof[what].image).decode()
    tfn = tempfile.mktemp(suffix='.png', prefix='pgraphviz-')
    display(Image(gv.Source(S, format='png', engine='dot').render(tfn)))

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

def bench_and_report(base_name, target_name):
    bench  = Benchmark("vm")
    base   = Branch(base_name).load_or_execute(bench, pprof=BENCHMARK_PPROF, mprof=BENCHMARK_MPROF, callgraph=True)
    target = Branch(target_name).load_or_execute(bench, pprof=BENCHMARK_PPROF, mprof=BENCHMARK_MPROF, callgraph=True)

    def transform(b,t):
        if b.active > 0 and t.active > 0:
            r = (t.active/b.active)
            return int(r*100-100) if r >= 1 else int(-1/r*100+100)

    plot_bench_hist(base, target, transform)
    plot_fast_slow_tests(base, target, transform)
    plot_pprof('TOP calls', 'top', base, target)
    plot_pprof_image('Top calls on {} VM'.format(base_name.title()), 'top', base)
    plot_pprof_image('Top calls on {} VM'.format(target_name.title()), 'top', target)
    plot_pprof('TOP allocs', 'alloc', base, target)
    plot_pprof_image('Top allocs on {} VM'.format(base_name.title()), 'alloc', base)
    plot_pprof_image('Top allocs on {} VM'.format(target_name.title()), 'alloc', target)


TEXT = r'''{
 "cells": [
  {
   "cell_type": "code",
   "execution_count": null,
   "metadata": {
    "collapsed": true
   },
   "outputs": [],
   "source": [
    "import vmbench\n",
    "vmbench.bench_and_report('classic','sputnik')"
   ]
  }
 ],
 "metadata": {
  "kernelspec": {
   "display_name": "Python 3",
   "language": "python",
   "name": "python3"
  },
  "language_info": {
   "codemirror_mode": {
    "name": "ipython",
    "version": 3
   },
   "file_extension": ".py",
   "mimetype": "text/x-python",
   "name": "python",
   "nbconvert_exporter": "python",
   "pygments_lexer": "ipython3",
   "version": "3.6.3"
  }
 },
 "nbformat": 4,
 "nbformat_minor": 2
}
'''


def _remove_sytel_scoped(body):
    while True:
        i = body.find("<style scoped>")
        if i >= 0:
            j = body.find("</style>", i)
            body = body[:i] + body[j+8:]
        else:
            return body


if __name__ == '__main__':
    import os, io, nbformat
    from nbconvert.preprocessors import ExecutePreprocessor
    from nbconvert import MarkdownExporter
    from traitlets.config import Config
    nb = nbformat.reads(TEXT,nbformat.NO_CONVERT)
    ep = ExecutePreprocessor(timeout=600, kernel_name='python3')
    ep.preprocess(nb, {'metadata': {'path': '.'}})
    c = Config()
    c.ExtractOutputPreprocessor.output_filename_template = '_img/{unique_key}_{cell_index}_{index}{extension}'
    me = MarkdownExporter(config=c)
    (body, r) = me.from_notebook_node(nb)
    README = open("README.md","w")
    README.write(_remove_sytel_scoped(body))
    if not os.path.isdir('_img'):
        os.mkdir('_img')
    for n, d in r.get('outputs',{}).items():
        with open(n,"wb") as img:
            img.write(d)
