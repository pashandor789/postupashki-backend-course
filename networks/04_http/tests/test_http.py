import unittest

from harness import expect, fields, run

TIMEOUT = 15
CRLF = "\r\n"


def message(start, header_lines=(), body=""):
    """Собирает сообщение: строка начала, поля заголовка, пустая строка, тело."""
    return start + CRLF + "".join(line + CRLF for line in header_lines) + CRLF + body


def parse(msg):
    code, out, err = run(stdin=msg, timeout=TIMEOUT)
    if code != 0:
        raise AssertionError(
            "корректное сообщение отвергнуто: код возврата %d\nвывод: %s\nstderr: %s"
            % (code, out, err))
    return out


def values(stdout, key):
    """Значения всех строк с указанным ключом в порядке вывода."""
    result = []
    for line in stdout.splitlines():
        line = line.strip()
        if line == key:
            result.append("")
        elif line.startswith(key + " "):
            result.append(line[len(key) + 1:].strip())
    return result


def header_pairs(stdout):
    """Поля заголовка как список пар «имя, значение» в порядке вывода."""
    pairs = []
    for item in values(stdout, "header"):
        parts = item.split(None, 1)
        pairs.append((parts[0], parts[1].strip() if len(parts) == 2 else ""))
    return pairs


def expect_error(case, msg, reason):
    code, out, err = run(stdin=msg, timeout=TIMEOUT)
    if code == 0:
        case.fail("ожидался ненулевой код возврата и строка %r, получен код 0\nвывод: %s"
                  % ("error " + reason, out))
    expect(case, fields(out), {"error": reason})


GET = message("GET /index.html HTTP/1.1", ["Host: example.org", "Accept: text/html"])
OK = message("HTTP/1.1 200 OK", ["Server: demo", "Content-Length: 13"], "Hello, world!")


class RequestLine(unittest.TestCase):
    def test_method_target_and_version(self):
        expect(self, fields(parse(GET)), {
            "type": "request",
            "method": "GET",
            "target": "/index.html",
            "version": "HTTP/1.1",
        })

    def test_target_keeps_query_string(self):
        msg = message("GET /search?q=hello%20world&page=2 HTTP/1.1", ["Host: example.org"])
        expect(self, fields(parse(msg)), {"target": "/search?q=hello%20world&page=2"})

    def test_method_is_printed_as_written(self):
        msg = message("DELETE /items/7 HTTP/1.1", ["Host: example.org"])
        expect(self, fields(parse(msg)), {"type": "request", "method": "DELETE"})

    def test_request_has_no_status_line_keys(self):
        got = fields(parse(GET))
        self.assertNotIn("status", got, "для запроса поле status выводить не следует")
        self.assertNotIn("reason", got, "для запроса поле reason выводить не следует")


class StatusLine(unittest.TestCase):
    def test_status_and_reason(self):
        expect(self, fields(parse(OK)), {
            "type": "response",
            "version": "HTTP/1.1",
            "status": "200",
            "reason": "OK",
        })

    def test_reason_may_contain_spaces(self):
        msg = message("HTTP/1.1 404 Not Found", ["Content-Length: 0"])
        expect(self, fields(parse(msg)), {"status": "404", "reason": "Not Found"})

    def test_reason_may_be_empty(self):
        msg = message("HTTP/1.1 204 ", ["Server: demo"])
        got = values(parse(msg), "reason")
        self.assertEqual([""], got,
                         "при пустом поясняющем тексте ожидалась одна строка reason "
                         "с пустым значением, получено %r" % (got,))

    def test_response_has_no_request_line_keys(self):
        got = fields(parse(OK))
        self.assertNotIn("method", got, "для ответа поле method выводить не следует")
        self.assertNotIn("target", got, "для ответа поле target выводить не следует")


class Headers(unittest.TestCase):
    def test_names_are_lowercased(self):
        msg = message("GET / HTTP/1.1", ["Host: example.org", "X-Request-Id: 42"])
        got = header_pairs(parse(msg))
        self.assertEqual([("host", "example.org"), ("x-request-id", "42")], got,
                         "ожидались имена полей в нижнем регистре, получено %r" % (got,))

    def test_value_is_trimmed(self):
        msg = message("GET / HTTP/1.1", ["Host:    example.org   ", "Accept:\ttext/html"])
        got = dict(header_pairs(parse(msg)))
        expect(self, got, {"host": "example.org", "accept": "text/html"})

    def test_value_may_contain_colon(self):
        msg = message("GET / HTTP/1.1", ["Host: example.org:8080"])
        got = dict(header_pairs(parse(msg)))
        expect(self, got, {"host": "example.org:8080"})

    def test_repeated_names_are_printed_in_order(self):
        msg = message("HTTP/1.1 200 OK",
                      ["Set-Cookie: a=1", "Content-Length: 0", "Set-Cookie: b=2"])
        got = header_pairs(parse(msg))
        self.assertEqual([("set-cookie", "a=1"), ("content-length", "0"),
                          ("set-cookie", "b=2")], got,
                         "ожидались все вхождения полей в исходном порядке, получено %r"
                         % (got,))

    def test_message_without_headers(self):
        msg = message("GET / HTTP/1.1")
        out = parse(msg)
        expect(self, fields(out), {"type": "request", "target": "/", "body.length": "0"})
        self.assertEqual([], header_pairs(out),
                         "в сообщении нет полей заголовка, строки header выводить не следует")


class Body(unittest.TestCase):
    def test_body_from_content_length(self):
        expect(self, fields(parse(OK)), {"body.length": "13", "body.text": "Hello, world!"})

    def test_body_length_zero_without_content_length(self):
        msg = message("GET / HTTP/1.1", ["Host: example.org"])
        got = parse(msg)
        expect(self, fields(got), {"body.length": "0"})
        self.assertEqual([], values(got, "body.text"),
                         "при пустом теле строку body.text выводить не следует")

    def test_bytes_after_declared_length_are_ignored(self):
        msg = message("HTTP/1.1 200 OK", ["Content-Length: 5"], "helloXXXXX")
        expect(self, fields(parse(msg)), {"body.length": "5", "body.text": "hello"})

    def test_body_with_crlf_is_not_split_into_headers(self):
        body = "alpha" + CRLF + CRLF + "beta"
        msg = message("HTTP/1.1 200 OK", ["Content-Length: 13"], body)
        out = parse(msg)
        expect(self, fields(out), {"body.length": "13"})
        self.assertEqual([("content-length", "13")], header_pairs(out),
                         "последовательность CRLF внутри тела не начинает новые поля "
                         "заголовка, получено %r" % (header_pairs(out),))

    def test_non_printable_body_has_no_text(self):
        body = "ab\x00cd"
        msg = message("HTTP/1.1 200 OK", ["Content-Length: 5"], body)
        out = parse(msg)
        expect(self, fields(out), {"body.length": "5"})
        self.assertEqual([], values(out, "body.text"),
                         "тело содержит непечатный байт, строку body.text выводить "
                         "не следует, получено %r" % (values(out, "body.text"),))


class ChunkedBody(unittest.TestCase):
    def test_chunks_are_joined(self):
        body = "4" + CRLF + "Wiki" + CRLF + "5" + CRLF + "pedia" + CRLF + "0" + CRLF + CRLF
        msg = message("HTTP/1.1 200 OK", ["Transfer-Encoding: chunked"], body)
        expect(self, fields(parse(msg)), {"body.length": "9", "body.text": "Wikipedia"})

    def test_zero_chunk_ends_body(self):
        body = ("5" + CRLF + "hello" + CRLF + "0" + CRLF
                + "X-Checksum: 0a1b" + CRLF + CRLF)
        msg = message("HTTP/1.1 200 OK", ["Transfer-Encoding: chunked"], body)
        out = parse(msg)
        expect(self, fields(out), {"body.length": "5", "body.text": "hello"})
        self.assertNotIn("x-checksum", dict(header_pairs(out)),
                         "поля концевика после части нулевой длины выводить не следует")

    def test_chunk_size_is_hexadecimal(self):
        alphabet = "abcdefghijklmnopqrstuvwxyz"
        body = "1A" + CRLF + alphabet + CRLF + "0" + CRLF + CRLF
        msg = message("HTTP/1.1 200 OK", ["Transfer-Encoding: chunked"], body)
        expect(self, fields(parse(msg)), {"body.length": "26", "body.text": alphabet})

    def test_chunk_extension_is_ignored(self):
        body = "5;name=value" + CRLF + "hello" + CRLF + "0" + CRLF + CRLF
        msg = message("HTTP/1.1 200 OK", ["Transfer-Encoding: chunked"], body)
        expect(self, fields(parse(msg)), {"body.length": "5", "body.text": "hello"})


class BadMessage(unittest.TestCase):
    def test_bad_start_line(self):
        expect_error(self, message("GARBAGE", ["Host: example.org"]), "bad_start_line")

    def test_version_other_than_1_1_is_rejected(self):
        expect_error(self, message("GET / HTTP/1.0", ["Host: example.org"]),
                     "bad_start_line")

    def test_line_feed_without_carriage_return_is_rejected(self):
        msg = "GET / HTTP/1.1\nHost: example.org\n\n"
        expect_error(self, msg, "bad_start_line")

    def test_header_without_colon(self):
        expect_error(self, message("GET / HTTP/1.1", ["Host example.org"]), "bad_header")

    def test_space_before_colon_is_rejected(self):
        expect_error(self, message("GET / HTTP/1.1", ["Host : example.org"]), "bad_header")

    def test_incomplete_body(self):
        msg = message("HTTP/1.1 200 OK", ["Content-Length: 10"], "abc")
        expect_error(self, msg, "incomplete_body")

    def test_chunked_body_without_zero_chunk(self):
        body = "5" + CRLF + "hello" + CRLF
        msg = message("HTTP/1.1 200 OK", ["Transfer-Encoding: chunked"], body)
        expect_error(self, msg, "bad_chunk")

    def test_chunk_size_is_not_hexadecimal(self):
        body = "zz" + CRLF + "hello" + CRLF + "0" + CRLF + CRLF
        msg = message("HTTP/1.1 200 OK", ["Transfer-Encoding: chunked"], body)
        expect_error(self, msg, "bad_chunk")

    def test_chunk_shorter_than_declared(self):
        body = "9" + CRLF + "hello" + CRLF + "0" + CRLF + CRLF
        msg = message("HTTP/1.1 200 OK", ["Transfer-Encoding: chunked"], body)
        expect_error(self, msg, "bad_chunk")


if __name__ == "__main__":
    unittest.main(verbosity=2)
