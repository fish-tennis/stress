import os
from fileutil import copytree


pbdir = '../../network/pb'
scdir = '../../../../../../union/src/proto/pb'
# copy pb
def clear_pb():
    for f in os.listdir(pbdir):
        os.remove(os.path.join(pbdir, f))

    


if __name__ == '__main__':
    clear_pb()
    copytree(scdir, pbdir)