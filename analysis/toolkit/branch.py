
import os.path
from .util import ALL


class Branch(object):
    __slots__ = ["label","dir","pfx"]

    def __init__(self,label,dir,pfx):
        self.label = label
        self.dir = dir
        self.pfx = pfx

    def path_with(self,root):
        return os.path.join(root,"branch",self.dir)


branches = [
    Branch(label="classic",dir="classic",pfx="classic_"),
    Branch(label="sputnik",dir="sputnik",pfx="sputnik_"),
]


class UnknownBranchError(Exception):
    def __init__(self,branch_label):
        Exception.__init__(self,"unknown branch {}".format(branch_label))


def list_of(labels):
    if labels is ALL:
        return (b for b in branches)
    else:
        return (look_for(l) for l in labels)


def look_for(label):
    for b in branches:
        if b.label == label:
            return b
    raise UnknownBranchError(label)

