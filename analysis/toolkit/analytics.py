
class Group(object):
    __slots__ = ["label"]


class Test(object):
    __slots__ = ["label","branches"]

    def __init__(self,label):
        self.label = label
        self.branches = []

    def __repr__(self):
        return 'Test(label="{}", branches={})'.format(self.label,self.branches)


class Branch(object):
    __slots__ = ["label","total","active","count","error"]

    def __init__(self,label,total,active,count,error):
        self.label = label
        self.total = total
        self.active = active
        self.count = count
        self.error = error

    def __repr__(self):
        return 'Branch(label="{}", total={}, active={}, error={})'.format(self.label,self.total,self.active,repr(self.error))


def extract_tree(branch,r,R):
    if not r.children:
        t = R.get(r.label,None)
        if t is None:
            t = Test(r.label)
            R[r.label] = t
        t.branches.append(Branch(branch,r.total,r.active,r.count,None))
    else:
        for c in r.children:
            extract_tree(branch,c,R)


def extract(results, R=None):
    if R is None:
        R = {}
    for r in results:
        extract_tree(r.branch,r.results,R)
    return R


class Dif(object):
    __slots__ = ["label","total","active","norm_active"]

    def __init__(self,label,b0,b1):
        self.label = label
        self.total = b1.total - b0.total
        self.active = b1.active - b0.active
        self.norm_active = float(b1.active)/b1.count - float(b0.active)/b0.count

    def __repr__(self):
        return "Dif(label='{}', total={}, active={}, norm_active={})".format(self.label,self.total,self.active,self.norm_active)


def percent_of(base,value):
    base = base
    value = value
    return int(value/base*100)


class Percent(object):
    __slots__ = ["label","passive","active"]

    def __init__(self,label,base,value):
        self.label = label
        self.passive = percent_of(base.total-base.active,value.total-value.active)
        self.active = percent_of(base.active,value.active)

    def __repr__(self):
        return "Percent(label='{}', active={}%, passive={}%)".format(self.label,self.active,self.passive)


def collect_branch(t,label):
    for i in t.branches:
        if i.label == label:
            return i


def collect_task(t,master,target):
    b0 = collect_branch(t,master)
    if b0 == None: return None
    b1 = collect_branch(t,target)
    if b1 == None: return None
    return Percent(t.label,b0,b1)


def collect(R,master,target):
    return (i for i in (collect_task(v,master,target) for v in R.values()) if i != None)
