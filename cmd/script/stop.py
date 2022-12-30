import psutil


def stop():
    for pid in psutil.pids():
        try:
            p = psutil.Process(pid)
        except Exception:
            continue
        if p.name().find('app.exe') == -1:
            continue
        print(p.name(), p.memory_info().rss / 1024 / 1024)
        p.kill()


if __name__ == '__main__':
    stop()