import glob
import os


def clear():
    dname = '../*.log'
    for f in glob.glob(dname):
        os.remove(f)
    # os.remove('../app.log')
    # os.remove('../stdout.log')


if __name__ == '__main__':
    clear()
