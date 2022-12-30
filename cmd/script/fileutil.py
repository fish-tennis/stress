

import shutil
import os
import tarfile

import pysftp
import paramiko

def ensure_dir(dname):
    if os.path.isfile(dname):
        dname = os.path.dirname(dname)

    if os.path.exists(dname):
        return
    os.makedirs(dname)


def copytree(src, dst, exclude=[]):
    for root, _, files in os.walk(src):
        for file in files:
            if file in exclude:
                continue
            dname = root.split(src)[1][1:]
            dstd = os.path.join(dst, dname)
            ensure_dir(dstd)
            shutil.copyfile(os.path.join(root, file), os.path.join(dstd, file))
            print(f'copy {file} from {root} to {dstd}')

def tar_file(dname='', suffix='') -> str:
    dname = os.path.realpath(dname)
    os.chdir(dname)
    tarname = dname + '.tar.gz'
    if suffix:
        tarname = '{}_{}.tar.gz'.format(dname, suffix)
    tar = tarfile.open(tarname, 'w:gz')
    print(tarname)
    for root, dir, files in os.walk(dname):
        for file in files:
            dn = root.split(dname)[1][1:]
            fullpath = os.path.join(dn, file)
            print('tar', file)
            tar.add(fullpath)
    tar.close()
    return tarname

class Sftp(object):

    def __init__(self, host: str, user: str, psw: str) -> None:
        cnopts = pysftp.CnOpts()
        cnopts.hostkeys = None
        self.sftp = pysftp.Connection(
            host, username=user, password=psw, cnopts=cnopts)

    def upload(self, remoteDir, localfile):
        with self.sftp.cd(remoteDir):
            print(f'开始传包 {localfile} to {remoteDir}')
            self.sftp.put(localfile)
            print('传包完成')


class SSH(object):

    def __init__(self, host: str, user: str, psw: str) -> None:
        self.client = paramiko.SSHClient()
        self.client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        self.client.connect(hostname=host,
                            port=22,
                            username=user,
                            password=psw)

    def exec(self, cmd, show=False):
        print('remote cmd: ', cmd)
        _, stdout, stderr = self.client.exec_command(cmd, timeout=30.0)
        res, err = stdout.read(), stderr.read()
        msg = res if res else err
        msg = msg.decode('utf-8')
        if show:
            print(msg)
        return msg