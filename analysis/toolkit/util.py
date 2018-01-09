
import os.path

ALL = type("",(),{})


def make_temp_dir_with(file=None):
    if file is None:
        dirname = os.path.dirname(os.path.dirname(__file__))
    else:
        dirname = analysis_dir()
    dirname = os.path.join(dirname,".temp")
    if not os.path.exists(dirname):
        os.mkdir(dirname)
    return dirname


def analysis_dir():
    dir = os.path.dirname(__file__)
    dir = os.path.dirname(dir)
    return dir


def root_dir():
    dir =  os.path.dirname(analysis_dir())
    return dir


def verbose(fmt,*args):
    print(fmt.format(*args))

class SuccessObject(object):
    pass

Success = SuccessObject()

class Return(object):
    __slots__ = ["value"]

    def __init__(self,value):
        self.value = value

class Fail(object):
    __slots__ = ["reason"]

    def __init__(self,reason):
        self.reason = reason

    def __repr__(self):
        return "Fail(reason='{}')".format(self.reason)
