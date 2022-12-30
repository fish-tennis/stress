import subprocess
import os

def run():
    os.chdir("../")
    # if not os.path.exists('./tmp/log'):
    #     os.makedirs('./tmp/log')

    log = open("stdout.log", "a")
    p = subprocess.Popen(
        ["./app.exe"],
        stderr=log,
        stdout=log,
        close_fds=True,
    )
    print(p.pid, "started")


if __name__ == '__main__':
    run()