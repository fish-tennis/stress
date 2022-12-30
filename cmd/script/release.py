import os
import shutil
import argparse
import sys
import threading

from fileutil import copytree, tar_file, Sftp, SSH

hosts = [
    {
        "host": '175.97.174.5',
        'user': 'root',
        'psw': 'YcNbashanghaUsh2t[23Us'
    },
    {
        "host": '175.99.7.99',
        'user': 'root',
        'psw': 'YcNbashanghaUsh2t[23Us'
    },
    {
        "host": '175.97.189.129',
        'user': 'root',
        'psw': 'YcNbashanghaUsh2t[23Us'
    },
    {
        "host": '175.97.139.67',
        'user': 'root',
        'psw': 'YcNbashanghaUsh2t[23Us'
    },
]


class Release(object):

    def __init__(self) -> None:
        self.root = os.path.realpath("../")
        self.packdir = 'tmp/release'
        self.exe = 'webapp'
        self.exclude = ['req_short_cut.json', 'server.json', 'state.json']
        self.conf = self.packdir

    def build(self):
        os.chdir(self.root)
        os.environ.setdefault('CGO_ENABLED', '0')
        os.environ.setdefault('GOOS', 'linux')
        return os.system(f'go build -o {self.exe}')

    def _copy(self):
        pass

    def copy_file(self):
        dst = self.packdir
        shutil.rmtree(dst)
        os.makedirs(dst, exist_ok=True)

        shutil.copyfile(self.exe, os.path.join(dst, self.exe))

        copytree('conf', self.conf, self.exclude)
        self._copy()

        tar_file(dst)

    def upload(self):
        l = []
        for info in hosts:
            args = (info['host'], info['user'], info['psw'])
            t = threading.Thread(target=upload_once, args=args)
            l.append(t)
            t.start()
        for t in l:
            t.join()


class ReleaseWin(Release):

    def __init__(self) -> None:
        super().__init__()
        self.packdir = 'tmp/winapp'
        self.exclude = []
        self.exe = 'webapp.exe'
        self.conf = os.path.join(self.packdir, 'conf')
        self.copystatic = True

    def build(self):
        os.chdir(self.root)
        return os.system(f"go build -o {self.exe}")

    def _copy(self):
        copytree('static', os.path.join(self.packdir, 'static'), self.exclude)

    def upload(self):
        return


def get_ins(platfrom='windows'):
    print('release ', platfrom)
    if platfrom == 'linux':
        return Release()
    return ReleaseWin()


def upload_once(host, user, psw):
    print(f'upload to {host}')
    while True:
        try:
            sftp = Sftp(host, user, psw)
            remotedir = '/root/tmp'
            ssh = SSH(host, user, psw)
            ssh.exec('mkdir -p ' + remotedir)
            localfile = os.path.join(os.path.realpath('../'), 'release.tar.gz')
            sftp.upload(remotedir, localfile)

            cmds = [
                f'cd {remotedir}', 'tar xf release.tar.gz', 'mv webapp ../data',
                'chmod +x ../data/webapp'
            ]
            cmd = ' && '.join(cmds)
            ssh.exec(cmd)
            return
        except Exception as e:
            print(f'{host} bad network, retry', e)




def chmod():
    remotedir = '/root/tmp'
    cmds = [
        f'cd {remotedir}', 'tar xf release.tar.gz', 'mv webapp ../data',
        'chmod +x ../data/webapp'
    ]
    cmd = ' && '.join(cmds)

    for info in hosts:
        ssh = SSH(info['host'], info['user'], info['psw'])
        ssh.exec(cmd)

def clear():
    remotedir = '/root/data'
    cmds = [
        f'cd {remotedir}', 
        'sh clear.sh',
        'rm nohup.out',
    ]
    cmd = ' && '.join(cmds)

    for info in hosts:
        ssh = SSH(info['host'], info['user'], info['psw'])
        ssh.exec(cmd)

def copy():
    remotedir = '/root/data'
    cmds = [
        f'cp /root/tmp/robot.b3 {remotedir}',
    ]
    cmd = ' && '.join(cmds)

    for info in hosts:
        # ssh = SSH(info['host'], info['user'], info['psw'])
        # ssh.exec(cmd)
        root = os.path.realpath('../')
        conf = os.path.join(root, 'conf')
        file = os.path.join(conf, 'robot.b3')
        print(file)
        sftp = Sftp(info['host'], info['user'], info['psw'])
        sftp.upload(remotedir, file)


def release(env:str):
    ins = get_ins(env)
    if ins.build() != 0:
        sys.exit(1)
    ins.copy_file()
    ins.upload()

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('-env', '--env', default='windows')
    parser.add_argument('-cmd', '--cmd', default='release')
    args = parser.parse_args()
    cmds = {
        'chmod':chmod,
        'clear':clear,
        'copy':copy,
    }
    
    fn = cmds.get(args.cmd)
    if not fn:
        release(args.env)
    else:
        fn()
    return



if __name__ == '__main__':
    main()

    # build()
    # pack_file()
    # tar_file()