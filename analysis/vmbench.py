
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.ticker as mtick
import toolkit as tk


def plot_bench_hist(Q, base, target):
    Max = None
    while True:
        L = [[x.active for x in tk.collect(tk.extract(q), target, tk.time_percent_of, base)
              if Max is None or x.active < Max] for q in Q]
        P = pd.Series([v for q in L for v in q])
        if P.std() < 33:
            break
        Max = P.max()

    P = pd.Series([v for q in L for v in q])
    total, mean, median, std = P.size, P.mean(), P.median(), P.std()
    S = [pd.Series([x for x in q if x >= mean-std*3 and x <= mean+std*3]) for q in L]
    drop = total-sum(q.size for q in S)

    n_bins=int(std*1.5)
    alpha=0.5
    fig,ax = plt.subplots(figsize=(16,6))
    for s in S: s.hist(bins=n_bins, alpha=alpha)
    ax.xaxis.set_major_formatter(mtick.PercentFormatter())
    ax.set_ylabel('Tests count', fontsize=18)
    ax.set_xlabel(
        'How many time takes {} VM against {} VM.  (median {:.0f}, μ {:.0f}, σ {:.0f})'.
            format(target, base, median, mean, std),
        fontsize=18)
    ax.set_title(
        'Comparision of {}/{} VMs benchmarks. ({} repeats, {} result total, {} dropped)'.
            format(target, base, len(L), total, drop),
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


def run_benchmark(*branches, temp=None):
    for r in tk.run_benchmarks("VM", ["classic", "sputnik"], temp=temp):
        print("PASSED")


if __name__ == "__main__":
    run_benchmark(temp=tk.make_temp_dir_with(__file__))
