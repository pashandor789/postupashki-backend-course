"""Запуск программы студента и разбор её вывода."""
import os
import pathlib
import subprocess

TASK_DIR = pathlib.Path(os.environ.get("HW_TASK_DIR")
                        or pathlib.Path(__file__).resolve().parent.parent)
BUILT = False


def _build():
    global BUILT
    if BUILT:
        return
    BUILT = True
    script = TASK_DIR / "build.sh"
    if script.exists():
        r = subprocess.run(["sh", str(script)], cwd=TASK_DIR,
                           capture_output=True, text=True, timeout=300)
        if r.returncode != 0:
            raise AssertionError("build.sh завершился с кодом %d\n%s%s"
                                 % (r.returncode, r.stdout, r.stderr))


def run(args=(), stdin="", timeout=30):
    """Запускает run.sh и возвращает (код возврата, stdout, stderr)."""
    script = TASK_DIR / "run.sh"
    if not script.exists():
        raise AssertionError(
            "в каталоге %s нет файла run.sh — создайте его по образцу run.sh.example"
            % TASK_DIR.name)
    _build()
    env = dict(os.environ, PYTHONUNBUFFERED="1")
    r = subprocess.run(["sh", str(script), *map(str, args)], cwd=TASK_DIR,
                       input=stdin, capture_output=True, text=True,
                       timeout=timeout, env=env)
    return r.returncode, r.stdout, r.stderr


def fields(stdout):
    """Разбирает вывод вида «ключ значение» в словарь."""
    out = {}
    for line in stdout.splitlines():
        line = line.strip()
        if not line:
            continue
        parts = line.split(None, 1)
        if len(parts) == 2:
            out[parts[0]] = parts[1].strip()
    return out


def expect(case, got, want, hexdump=""):
    """Сверяет подмножество ключей и печатает понятную разницу."""
    missing = [k for k in want if k not in got]
    wrong = {k: (want[k], got[k]) for k in want if k in got and got[k] != want[k]}
    if not missing and not wrong:
        return
    lines = []
    if missing:
        lines.append("нет строк с ключами: " + ", ".join(missing))
    for k, (w, g) in wrong.items():
        lines.append("%s: ожидалось %r, получено %r" % (k, w, g))
    if hexdump:
        lines.append("входной дамп: " + hexdump[:120] + ("..." if len(hexdump) > 120 else ""))
    case.fail("\n".join(lines))
