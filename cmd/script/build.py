import os


def build():
    os.chdir('../')
    os.system("go build -o app.exe")


if __name__ == '__main__':
    build()