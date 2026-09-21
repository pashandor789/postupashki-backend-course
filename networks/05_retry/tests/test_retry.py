"""Тесты задания 05_retry: повторная отправка запроса."""
import http.server
import socket
import socketserver
import threading
import time
import unittest

from harness import expect, run

BASE_MS = 200
FACTOR = 2
CAP_MS = 2000
SLACK_MS = 300
LEEWAY_MS = 60
STARTUP_S = 5.0
JITTER_RUNS = 5
GROWTH_RUNS = 5


def ceiling_ms(attempt):
    """Расчётный интервал перед попыткой с указанным номером."""
    return min(CAP_MS, BASE_MS * FACTOR ** (attempt - 1))


def budget_s(max_attempts):
    """Предел времени на один запуск: наибольшая возможная сумма пауз плюс запас."""
    worst = sum(ceiling_ms(n) for n in range(2, max_attempts + 1))
    return worst / 1000.0 + STARTUP_S


def closed_port_url():
    """Адрес порта, на котором заведомо никто не слушает."""
    probe = socket.socket()
    probe.bind(("127.0.0.1", 0))
    port = probe.getsockname()[1]
    probe.close()
    return "http://127.0.0.1:%d/resource" % port


def _handler_class(owner):
    class Handler(http.server.BaseHTTPRequestHandler):
        protocol_version = "HTTP/1.1"

        def respond(self):
            length = int(self.headers.get("Content-Length") or 0)
            if length:
                self.rfile.read(length)
            status, retry_after = owner.record(self.command, self.headers)
            body = b"scripted response"
            self.send_response(status)
            if retry_after is not None:
                self.send_header("Retry-After", str(retry_after))
            self.send_header("Content-Type", "text/plain")
            self.send_header("Content-Length", str(len(body)))
            self.send_header("Connection", "close")
            self.end_headers()
            self.wfile.write(body)

        do_GET = respond
        do_POST = respond
        do_PUT = respond
        do_DELETE = respond

        def log_message(self, fmt, *args):
            pass

    return Handler


class _Server(http.server.ThreadingHTTPServer):
    daemon_threads = True

    def server_bind(self):
        """Привязка без обратного разрешения имени: getfqdn на localhost занимает секунды."""
        socketserver.TCPServer.server_bind(self)
        self.server_name, self.server_port = self.server_address[:2]


class ScriptedServer:
    """Сервер HTTP на localhost, отвечающий по заданному сценарию.

    Сценарий — список ответов: код или пара «код, значение Retry-After в секундах».
    Последний ответ сценария повторяется для всех дальнейших запросов.
    """

    def __init__(self, script):
        self.script = [s if isinstance(s, tuple) else (s, None) for s in script]
        self.requests = []
        self._lock = threading.Lock()
        self._httpd = _Server(("127.0.0.1", 0), _handler_class(self))
        self._thread = threading.Thread(
            target=self._httpd.serve_forever, args=(0.05,), daemon=True)
        self._thread.start()

    @property
    def url(self):
        host, port = self._httpd.server_address[:2]
        return "http://%s:%d/resource" % (host, port)

    def record(self, method, headers):
        with self._lock:
            index = len(self.requests)
            self.requests.append({
                "time": time.monotonic(),
                "method": method,
                "headers": {k.lower(): v for k, v in headers.items()},
            })
        return self.script[min(index, len(self.script) - 1)]

    def gaps_ms(self):
        """Промежутки между моментами получения соседних запросов."""
        moments = [r["time"] for r in self.requests]
        return [int(round((b - a) * 1000)) for a, b in zip(moments, moments[1:])]

    def methods(self):
        return [r["method"] for r in self.requests]

    def stop(self):
        self._httpd.shutdown()
        self._httpd.server_close()


class Output:
    """Разбор вывода программы по строкам событий."""

    def __init__(self, code, stdout, stderr):
        self.code = code
        self.stdout = stdout
        self.stderr = stderr
        self.attempts = []
        self.sleeps = []
        self.result = None
        self.reported = None
        for line in stdout.splitlines():
            parts = line.split()
            if len(parts) >= 4 and parts[0] == "attempt" and parts[2] in ("status", "error"):
                self.attempts.append({
                    "number": parts[1],
                    "kind": parts[2],
                    "value": " ".join(parts[3:]),
                })
            elif len(parts) == 2 and parts[0] == "sleep_ms":
                try:
                    self.sleeps.append(int(parts[1]))
                except ValueError:
                    pass
            elif len(parts) >= 4 and parts[0] == "result" and parts[2] == "attempts":
                self.result = parts[1]
                self.reported = parts[3]

    def summary(self):
        statuses = [a["value"] if a["kind"] == "status" else "error" for a in self.attempts]
        return {
            "result": self.result or "(строка result не напечатана)",
            "attempts": self.reported or "(строка result не напечатана)",
            "attempt_lines": str(len(self.attempts)),
            "sleep_lines": str(len(self.sleeps)),
            "statuses": ",".join(statuses) if statuses else "(нет строк attempt)",
            "numbers": ",".join(a["number"] for a in self.attempts) or "(нет строк attempt)",
            "exit_code": str(self.code),
        }

    def context(self):
        lines = [l.strip() for l in self.stdout.splitlines() if l.strip()]
        text = "(вывод программы) " + " | ".join(lines) if lines \
            else "(программа ничего не напечатала)"
        if self.stderr.strip():
            text += " (stderr) " + self.stderr.strip().splitlines()[0]
        return text


def launch(url, options=(), timeout=30):
    code, stdout, stderr = run([url, *options], timeout=timeout)
    return Output(code, stdout, stderr)


class ClientCase(unittest.TestCase):
    def server(self, script):
        srv = ScriptedServer(script)
        self.addCleanup(srv.stop)
        return srv


class Outcome(ClientCase):
    def test_success_on_first_attempt(self):
        srv = self.server([200])
        out = launch(srv.url, timeout=budget_s(1))
        expect(self, out.summary(), {
            "result": "success",
            "attempts": "1",
            "attempt_lines": "1",
            "sleep_lines": "0",
            "statuses": "200",
        }, out.context())

    def test_server_error_is_retried_until_success(self):
        srv = self.server([503, 503, 200])
        out = launch(srv.url, timeout=budget_s(3))
        expect(self, out.summary(), {
            "result": "success",
            "attempts": "3",
            "attempt_lines": "3",
            "sleep_lines": "2",
            "statuses": "503,503,200",
            "numbers": "1,2,3",
        }, out.context())

    def test_bad_request_is_not_retried(self):
        srv = self.server([400])
        out = launch(srv.url, timeout=budget_s(1))
        expect(self, out.summary(), {
            "result": "failure",
            "attempts": "1",
            "attempt_lines": "1",
            "sleep_lines": "0",
            "statuses": "400",
        }, out.context())

    def test_not_found_is_not_retried(self):
        srv = self.server([404])
        out = launch(srv.url, timeout=budget_s(1))
        expect(self, out.summary(), {
            "result": "failure",
            "attempts": "1",
            "statuses": "404",
        }, out.context())

    def test_too_many_requests_is_retried(self):
        srv = self.server([429, 200])
        out = launch(srv.url, timeout=budget_s(2))
        expect(self, out.summary(), {
            "result": "success",
            "attempts": "2",
            "statuses": "429,200",
        }, out.context())

    def test_attempt_limit_is_respected(self):
        srv = self.server([503])
        out = launch(srv.url, ["--max-attempts", "3"], timeout=budget_s(3))
        expect(self, out.summary(), {
            "result": "failure",
            "attempts": "3",
            "attempt_lines": "3",
            "sleep_lines": "2",
            "statuses": "503,503,503",
        }, out.context())

    def test_default_attempt_limit_is_five(self):
        srv = self.server([503])
        out = launch(srv.url, timeout=budget_s(5))
        expect(self, out.summary(), {
            "result": "failure",
            "attempts": "5",
            "attempt_lines": "5",
            "sleep_lines": "4",
        }, out.context())

    def test_single_attempt_makes_no_pause(self):
        srv = self.server([503])
        out = launch(srv.url, ["--max-attempts", "1"], timeout=budget_s(1))
        expect(self, out.summary(), {
            "result": "failure",
            "attempts": "1",
            "attempt_lines": "1",
            "sleep_lines": "0",
        }, out.context())

    def test_exit_code_is_zero_on_success(self):
        srv = self.server([200])
        out = launch(srv.url, timeout=budget_s(1))
        expect(self, out.summary(), {"result": "success", "exit_code": "0"}, out.context())

    def test_exit_code_is_not_zero_on_failure(self):
        srv = self.server([400])
        out = launch(srv.url, timeout=budget_s(1))
        expect(self, out.summary(), {"result": "failure"}, out.context())
        self.assertNotEqual(
            0, out.code,
            "при итоге failure ожидался ненулевой код возврата, получен 0")


class Requests(ClientCase):
    def test_default_method_is_get(self):
        srv = self.server([200])
        out = launch(srv.url, timeout=budget_s(1))
        expect(self, out.summary(), {"attempts": "1"}, out.context())
        self.assertEqual(
            ["GET"], srv.methods(),
            "без ключа --method ожидался метод GET, сервер получил %r" % (srv.methods(),))

    def test_requested_method_reaches_server(self):
        srv = self.server([200])
        out = launch(srv.url, ["--method", "POST"], timeout=budget_s(1))
        expect(self, out.summary(), {"attempts": "1"}, out.context())
        self.assertEqual(
            ["POST"], srv.methods(),
            "с ключом --method POST ожидался метод POST, сервер получил %r" % (srv.methods(),))

    def test_every_printed_attempt_reaches_server(self):
        srv = self.server([503, 503, 200])
        out = launch(srv.url, timeout=budget_s(3))
        expect(self, out.summary(), {"attempt_lines": "3"}, out.context())
        self.assertEqual(
            3, len(srv.requests),
            "напечатано 3 попытки, а сервер получил запросов: %d" % len(srv.requests))


class Idempotency(ClientCase):
    def test_post_without_key_is_not_retried(self):
        srv = self.server([503])
        out = launch(srv.url, ["--method", "POST"], timeout=budget_s(1))
        expect(self, out.summary(), {
            "result": "failure",
            "attempts": "1",
            "attempt_lines": "1",
            "sleep_lines": "0",
            "statuses": "503",
        }, out.context())
        self.assertEqual(
            1, len(srv.requests),
            "POST без ключа идемпотентности не повторяется, "
            "ожидался один запрос, сервер получил %d" % len(srv.requests))

    def test_post_with_key_is_retried(self):
        srv = self.server([503, 200])
        out = launch(srv.url, ["--method", "POST", "--idempotency-key", "k-42"],
                     timeout=budget_s(2))
        expect(self, out.summary(), {
            "result": "success",
            "attempts": "2",
            "statuses": "503,200",
        }, out.context())

    def test_idempotency_key_is_sent_in_header(self):
        srv = self.server([503, 200])
        out = launch(srv.url, ["--method", "POST", "--idempotency-key", "k-42"],
                     timeout=budget_s(2))
        expect(self, out.summary(), {"attempts": "2"}, out.context())
        got = [r["headers"].get("idempotency-key") for r in srv.requests]
        self.assertEqual(
            ["k-42", "k-42"], got,
            "ожидался заголовок Idempotency-Key со значением k-42 в каждом запросе, "
            "получено %r" % (got,))


class Backoff(ClientCase):
    def test_pause_does_not_exceed_computed_interval(self):
        srv = self.server([503])
        out = launch(srv.url, ["--max-attempts", "5"], timeout=budget_s(5))
        expect(self, out.summary(), {"attempt_lines": "5"}, out.context())
        for index, gap in enumerate(srv.gaps_ms()):
            number = index + 2
            limit = ceiling_ms(number)
            self.assertLessEqual(
                gap, limit + SLACK_MS,
                "перед попыткой %d расчётный интервал равен %d мс, "
                "допускается не более %d мс с запасом на накладные расходы, "
                "сервер получил запрос через %d мс" % (number, limit, limit + SLACK_MS, gap))

    def test_pause_matches_reported_value(self):
        srv = self.server([503, 503, 200])
        out = launch(srv.url, timeout=budget_s(3))
        expect(self, out.summary(), {"sleep_lines": "2", "attempt_lines": "3"}, out.context())
        for index, (reported, gap) in enumerate(zip(out.sleeps, srv.gaps_ms())):
            self.assertGreaterEqual(
                gap, reported - LEEWAY_MS,
                "напечатано sleep_ms %d, а следующий запрос пришёл через %d мс: "
                "фактическая пауза короче напечатанной" % (reported, gap))
            self.assertLessEqual(
                gap, reported + SLACK_MS,
                "напечатано sleep_ms %d, а следующий запрос пришёл через %d мс: "
                "фактическая пауза длиннее напечатанной" % (reported, gap))

    def test_pause_grows_with_attempt_number(self):
        first, fourth = [], []
        for _ in range(GROWTH_RUNS):
            srv = self.server([503])
            out = launch(srv.url, ["--max-attempts", "4"], timeout=budget_s(4))
            expect(self, out.summary(), {"sleep_lines": "3"}, out.context())
            first.append(out.sleeps[0])
            fourth.append(out.sleeps[2])
        self.assertGreater(
            max(fourth), max(first),
            "расчётный интервал перед попыткой 4 равен %d мс, перед попыткой 2 — %d мс; "
            "за %d запусков наибольшая пауза перед попыткой 2 составила %d мс, "
            "перед попыткой 4 — %d мс, роста интервала не видно"
            % (ceiling_ms(4), ceiling_ms(2), GROWTH_RUNS, max(first), max(fourth)))

    def test_pause_never_exceeds_upper_limit(self):
        srv = self.server([503])
        out = launch(srv.url, ["--max-attempts", "7"], timeout=budget_s(7))
        expect(self, out.summary(), {"attempt_lines": "7", "sleep_lines": "6"}, out.context())
        for index, reported in enumerate(out.sleeps):
            self.assertLessEqual(
                reported, CAP_MS,
                "верхний предел интервала — %d мс, перед попыткой %d напечатано "
                "sleep_ms %d" % (CAP_MS, index + 2, reported))
        for index, gap in enumerate(srv.gaps_ms()):
            self.assertLessEqual(
                gap, CAP_MS + SLACK_MS,
                "верхний предел интервала — %d мс, а запрос попытки %d пришёл "
                "через %d мс" % (CAP_MS, index + 2, gap))

    def test_full_jitter_varies_pauses(self):
        seen = []
        for _ in range(JITTER_RUNS):
            srv = self.server([503, 200])
            out = launch(srv.url, timeout=budget_s(2))
            expect(self, out.summary(), {"sleep_lines": "1"}, out.context())
            seen.append(out.sleeps[0])
        self.assertGreater(
            len(set(seen)), 1,
            "при полном случайном разбросе пауза выбирается из отрезка от 0 до %d мс "
            "и в разных запусках различается; за %d запусков получено одно и то же "
            "значение %d мс" % (ceiling_ms(2), JITTER_RUNS, seen[0]))

    def test_retry_after_sets_minimum_pause(self):
        srv = self.server([(503, 1), 200])
        out = launch(srv.url, timeout=budget_s(2) + 1)
        expect(self, out.summary(), {
            "result": "success",
            "attempts": "2",
            "sleep_lines": "1",
        }, out.context())
        self.assertGreaterEqual(
            out.sleeps[0], 1000,
            "ответ содержал Retry-After: 1, ожидалась пауза не меньше 1000 мс, "
            "напечатано sleep_ms %d" % out.sleeps[0])
        gap = srv.gaps_ms()[0]
        self.assertGreaterEqual(
            gap, 1000 - LEEWAY_MS,
            "ответ содержал Retry-After: 1, а следующий запрос пришёл через %d мс" % gap)

    def test_retry_after_replaces_random_spread(self):
        srv = self.server([(503, 1), 200])
        out = launch(srv.url, timeout=budget_s(2) + 1)
        expect(self, out.summary(), {"attempts": "2", "sleep_lines": "1"}, out.context())
        self.assertLessEqual(
            out.sleeps[0], 1000 + SLACK_MS,
            "при заданном Retry-After разброс не применяется: ожидалась пауза "
            "около 1000 мс, напечатано sleep_ms %d" % out.sleeps[0])


class ConnectionFailure(ClientCase):
    def test_connection_error_is_retried(self):
        out = launch(closed_port_url(), ["--max-attempts", "2"], timeout=budget_s(2))
        expect(self, out.summary(), {
            "result": "failure",
            "attempts": "2",
            "attempt_lines": "2",
            "sleep_lines": "1",
            "statuses": "error,error",
        }, out.context())

    def test_connection_error_is_reported_with_text(self):
        out = launch(closed_port_url(), ["--max-attempts", "1"], timeout=budget_s(1))
        expect(self, out.summary(), {
            "result": "failure",
            "attempts": "1",
            "statuses": "error",
        }, out.context())
        self.assertTrue(
            out.attempts[0]["value"].strip(),
            "после слова error ожидается текст с причиной, получена строка без текста")


if __name__ == "__main__":
    unittest.main(verbosity=2)
