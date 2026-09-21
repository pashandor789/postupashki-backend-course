import socket
import struct
import subprocess
import threading
import time
import unittest

from harness import expect, run

HOST = "127.0.0.1"
CLASS_IN = 1
RESPONSE_FLAGS = 0x8180
QUESTION_POINTER = b"\xc0\x0c"

TYPE_CODES = {"A": 1, "NS": 2, "CNAME": 5, "MX": 15, "TXT": 16, "AAAA": 28}


def encode_name(name):
    data = b""
    for label in name.rstrip(".").split("."):
        data += bytes([len(label)]) + label.encode("ascii")
    return data + b"\x00"


def read_name(message, offset):
    labels = []
    end = None
    for _ in range(128):
        length = message[offset]
        if length & 0xC0 == 0xC0:
            if end is None:
                end = offset + 2
            offset = struct.unpack("!H", message[offset:offset + 2])[0] & 0x3FFF
            continue
        offset += 1
        if length == 0:
            break
        labels.append(message[offset:offset + length].decode("ascii"))
        offset += length
    if end is None:
        end = offset
    return (".".join(labels) + "." if labels else "."), end


def record(name, rtype, ttl, rdata):
    return name + struct.pack("!HHIH", rtype, CLASS_IN, ttl, len(rdata)) + rdata


def text_strings(*parts):
    data = b""
    for part in parts:
        raw = part.encode("utf-8")
        data += bytes([len(raw)]) + raw
    return data


class Query:
    """Разобранный запрос, полученный тестовым сервером."""

    def __init__(self, raw):
        self.raw = raw
        head = struct.unpack("!HHHHHH", raw[:12])
        self.id, self.flags = head[0], head[1]
        self.qdcount, self.ancount, self.nscount, self.arcount = head[2:]
        self.name, end = read_name(raw, 12)
        self.qtype, self.qclass = struct.unpack("!HH", raw[end:end + 4])
        self.question = raw[12:end + 4]


def respond(query, answers=(), rcode=0, authority=(), additional=()):
    """Собирает байты ответа; имя вопроса лежит по смещению 12, как в запросе."""
    header = struct.pack("!HHHHHH", query.id, RESPONSE_FLAGS | rcode,
                         1, len(answers), len(authority), len(additional))
    return (header + query.question + b"".join(answers)
            + b"".join(authority) + b"".join(additional))


class DnsServer:
    """UDP-сервер на локальном адресе: считает запросы и отдаёт заготовленные байты."""

    def __init__(self, handler):
        self.handler = handler
        self.requests = []
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.socket.bind((HOST, 0))
        self.socket.settimeout(0.05)
        self.port = self.socket.getsockname()[1]
        self._stop = threading.Event()
        self._thread = threading.Thread(target=self._serve, daemon=True)
        self._thread.start()

    def _serve(self):
        while not self._stop.is_set():
            try:
                raw, peer = self.socket.recvfrom(4096)
            except socket.timeout:
                continue
            except OSError:
                return
            query = Query(raw)
            self.requests.append(query)
            reply = self.handler(query)
            if reply is None:
                continue
            if isinstance(reply, (bytes, bytearray)):
                reply = [reply]
            for packet in reply:
                self.socket.sendto(packet, peer)

    def close(self):
        self._stop.set()
        self._thread.join(2)
        self.socket.close()


def blocks(stdout):
    """Разбирает вывод на блоки query/status/answer/end."""
    result = []
    current = None
    for line in stdout.splitlines():
        line = line.strip()
        if not line:
            continue
        parts = line.split(None, 1)
        key = parts[0]
        rest = parts[1].strip() if len(parts) > 1 else ""
        if key == "query":
            if current is not None:
                result.append(current)
            current = {"query": rest, "status": None, "answer": [], "ended": False}
        elif current is None:
            continue
        elif key == "status":
            current["status"] = rest
        elif key == "answer":
            current["answer"].append(rest)
        elif key == "end":
            current["ended"] = True
            result.append(current)
            current = None
    if current is not None:
        result.append(current)
    return result


def printed(block):
    """Приводит блок вывода к виду «ключ значение» для сравнения."""
    got = {"query": block["query"], "записей в ответе": str(len(block["answer"]))}
    if block["status"] is not None:
        got["status"] = block["status"]
    for index, value in enumerate(block["answer"], 1):
        got["answer %d" % index] = value
    return got


def asked(query):
    """Приводит полученный сервером запрос к виду «ключ значение»."""
    return {
        "вопросов в запросе": str(query.qdcount),
        "записей в разделе ответа запроса": str(query.ancount),
        "признак ответа": str((query.flags >> 15) & 1),
        "код операции": str((query.flags >> 11) & 0x0F),
        "бит рекурсии": str((query.flags >> 8) & 1),
        "имя вопроса": query.name,
        "тип вопроса": str(query.qtype),
        "класс вопроса": str(query.qclass),
    }


class DnsTestCase(unittest.TestCase):
    def serve(self, handler):
        server = DnsServer(handler)
        self.addCleanup(server.close)
        return server

    def ask(self, server, lines, timeout=30):
        stdin = "".join(line + "\n" for line in lines)
        code, out, err = run(args=[HOST, server.port], stdin=stdin, timeout=timeout)
        if code != 0:
            self.fail("программа завершилась с кодом %d, ожидался 0\n"
                      "вывод:\n%s\nstderr:\n%s" % (code, out, err))
        result = blocks(out)
        if len(result) != len(lines):
            self.fail("ожидалось блоков «query ... end»: %d, получено: %d\n"
                      "вывод:\n%s" % (len(lines), len(result), out))
        for block in result:
            if not block["ended"]:
                self.fail("блок «query %s» не завершён строкой end\n"
                          "вывод:\n%s" % (block["query"], out))
        return result

    def one(self, server, line, timeout=30):
        return self.ask(server, [line], timeout=timeout)[0]


class QueryFormat(DnsTestCase):
    def handler(self, query):
        return respond(query, [record(QUESTION_POINTER, 1, 60,
                                      socket.inet_aton("10.0.0.1"))])

    def test_query_contains_exactly_one_question(self):
        server = self.serve(self.handler)
        self.one(server, "example.com A")
        expect(self, asked(server.requests[0]), {
            "вопросов в запросе": "1",
            "записей в разделе ответа запроса": "0",
        })

    def test_query_asks_for_recursion(self):
        server = self.serve(self.handler)
        self.one(server, "example.com A")
        expect(self, asked(server.requests[0]), {
            "признак ответа": "0",
            "код операции": "0",
            "бит рекурсии": "1",
        })

    def test_query_carries_name_and_internet_class(self):
        server = self.serve(self.handler)
        self.one(server, "www.example.com A")
        expect(self, asked(server.requests[0]), {
            "имя вопроса": "www.example.com.",
            "тип вопроса": "1",
            "класс вопроса": "1",
        })

    def test_type_names_are_translated_to_codes(self):
        server = self.serve(lambda query: respond(query))
        names = ["A", "AAAA", "CNAME", "MX", "TXT", "NS"]
        self.ask(server, ["example.com " + name for name in names])
        got = {name: str(query.qtype)
               for name, query in zip(names, server.requests)}
        expect(self, got, {name: str(TYPE_CODES[name]) for name in names})

    def test_response_with_foreign_identifier_is_ignored(self):
        def handler(query):
            wrong = bytearray(respond(query, [record(QUESTION_POINTER, 1, 60,
                                                     socket.inet_aton("10.9.9.9"))]))
            wrong[0:2] = struct.pack("!H", query.id ^ 0x5555)
            right = respond(query, [record(QUESTION_POINTER, 1, 60,
                                           socket.inet_aton("10.1.1.1"))])
            return [bytes(wrong), right]

        server = self.serve(handler)
        block = self.one(server, "example.com A")
        expect(self, printed(block), {
            "status": "NOERROR",
            "записей в ответе": "1",
            "answer 1": "A 10.1.1.1 60",
        })


class AnswerSection(DnsTestCase):
    def test_address_record_value_and_ttl(self):
        server = self.serve(lambda query: respond(query, [
            record(encode_name("www.example.com"), 1, 300,
                   socket.inet_aton("93.184.216.34")),
        ]))
        block = self.one(server, "www.example.com A")
        expect(self, printed(block), {
            "query": "www.example.com A",
            "status": "NOERROR",
            "записей в ответе": "1",
            "answer 1": "A 93.184.216.34 300",
        })

    def test_compressed_name_in_answer_section(self):
        server = self.serve(lambda query: respond(query, [
            record(QUESTION_POINTER, 1, 120, socket.inet_aton("198.51.100.7")),
        ]))
        block = self.one(server, "www.example.com A")
        expect(self, printed(block), {
            "записей в ответе": "1",
            "answer 1": "A 198.51.100.7 120",
        })

    def test_compressed_name_inside_record_data(self):
        server = self.serve(lambda query: respond(query, [
            record(QUESTION_POINTER, 5, 3600,
                   b"\x03cdn" + QUESTION_POINTER),
        ]))
        block = self.one(server, "example.com CNAME")
        expect(self, printed(block), {
            "записей в ответе": "1",
            "answer 1": "CNAME cdn.example.com. 3600",
        })

    def test_cname_chain_precedes_address_record(self):
        target = encode_name("cdn.example.net")
        server = self.serve(lambda query: respond(query, [
            record(QUESTION_POINTER, 5, 3600, target),
            record(target, 1, 60, socket.inet_aton("203.0.113.8")),
        ]))
        block = self.one(server, "www.example.com A")
        expect(self, printed(block), {
            "status": "NOERROR",
            "записей в ответе": "2",
            "answer 1": "CNAME cdn.example.net. 3600",
            "answer 2": "A 203.0.113.8 60",
        })

    def test_ipv6_address_is_printed_in_short_form(self):
        server = self.serve(lambda query: respond(query, [
            record(QUESTION_POINTER, 28, 900,
                   socket.inet_pton(socket.AF_INET6, "2001:db8::1")),
        ]))
        block = self.one(server, "example.com AAAA")
        expect(self, printed(block), {
            "записей в ответе": "1",
            "answer 1": "AAAA 2001:db8::1 900",
        })

    def test_mail_exchange_record_prints_priority(self):
        server = self.serve(lambda query: respond(query, [
            record(QUESTION_POINTER, 15, 1800,
                   struct.pack("!H", 10) + b"\x04mail" + QUESTION_POINTER),
        ]))
        block = self.one(server, "example.com MX")
        expect(self, printed(block), {
            "записей в ответе": "1",
            "answer 1": "MX 10 mail.example.com. 1800",
        })

    def test_text_record_is_printed_without_quotes(self):
        server = self.serve(lambda query: respond(query, [
            record(QUESTION_POINTER, 16, 240,
                   text_strings("v=spf1 ", "-all")),
        ]))
        block = self.one(server, "example.com TXT")
        expect(self, printed(block), {
            "записей в ответе": "1",
            "answer 1": "TXT v=spf1 -all 240",
        })

    def test_name_server_records_keep_order(self):
        server = self.serve(lambda query: respond(query, [
            record(QUESTION_POINTER, 2, 172800, b"\x03ns2" + QUESTION_POINTER),
            record(QUESTION_POINTER, 2, 172800, b"\x03ns1" + QUESTION_POINTER),
        ]))
        block = self.one(server, "example.com NS")
        expect(self, printed(block), {
            "записей в ответе": "2",
            "answer 1": "NS ns2.example.com. 172800",
            "answer 2": "NS ns1.example.com. 172800",
        })

    def test_empty_answer_section_prints_no_records(self):
        server = self.serve(lambda query: respond(query))
        block = self.one(server, "example.com MX")
        expect(self, printed(block), {
            "query": "example.com MX",
            "status": "NOERROR",
            "записей в ответе": "0",
        })

    def test_other_sections_are_not_printed(self):
        authority = [record(QUESTION_POINTER, 2, 172800,
                            b"\x03ns1" + QUESTION_POINTER)]
        additional = [record(b"\x03ns1" + QUESTION_POINTER, 1, 172800,
                             socket.inet_aton("192.0.2.53"))]
        server = self.serve(lambda query: respond(
            query,
            [record(QUESTION_POINTER, 1, 60, socket.inet_aton("192.0.2.1"))],
            authority=authority, additional=additional))
        block = self.one(server, "example.com A")
        expect(self, printed(block), {
            "записей в ответе": "1",
            "answer 1": "A 192.0.2.1 60",
        })


class ResponseCodes(DnsTestCase):
    def check(self, rcode, status):
        server = self.serve(lambda query: respond(query, rcode=rcode))
        block = self.one(server, "nowhere.example A")
        expect(self, printed(block), {
            "query": "nowhere.example A",
            "status": status,
            "записей в ответе": "0",
        })

    def test_name_does_not_exist(self):
        self.check(3, "NXDOMAIN")

    def test_server_failure(self):
        self.check(2, "SERVFAIL")

    def test_request_refused(self):
        self.check(5, "REFUSED")

    def test_format_error(self):
        self.check(1, "FORMERR")

    def test_error_status_does_not_stop_the_program(self):
        def handler(query):
            if query.name.startswith("missing."):
                return respond(query, rcode=3)
            return respond(query, [record(QUESTION_POINTER, 1, 60,
                                          socket.inet_aton("192.0.2.10"))])

        server = self.serve(handler)
        first, second = self.ask(server, ["missing.example A", "present.example A"])
        expect(self, printed(first), {"status": "NXDOMAIN", "записей в ответе": "0"})
        expect(self, printed(second), {
            "status": "NOERROR",
            "answer 1": "A 192.0.2.10 60",
        })


class Caching(DnsTestCase):
    def address_server(self, ttl, address="192.0.2.77"):
        return self.serve(lambda query: respond(query, [
            record(QUESTION_POINTER, query.qtype, ttl,
                   socket.inet_pton(socket.AF_INET6, address)
                   if query.qtype == 28 else socket.inet_aton(address)),
        ]))

    def check_requests(self, server, want):
        got = str(len(server.requests))
        if got != str(want):
            self.fail("сервер должен был получить запросов: %d, получено: %s\n"
                      "имена и типы запросов: %s"
                      % (want, got,
                         ", ".join("%s/%d" % (q.name, q.qtype) for q in server.requests)))

    def test_repeated_question_is_answered_from_cache(self):
        server = self.address_server(600)
        first, second = self.ask(server, ["example.com A", "example.com A"])
        self.check_requests(server, 1)
        expect(self, printed(second), {
            "query": "example.com A",
            "status": "NOERROR",
            "записей в ответе": "1",
            "answer 1": printed(first)["answer 1"],
        })

    def test_cached_answer_is_reused_for_ipv6_address(self):
        server = self.address_server(600, "2001:db8::2")
        self.ask(server, ["example.com AAAA", "example.com AAAA"])
        self.check_requests(server, 1)

    def test_types_do_not_share_a_cache_entry(self):
        def handler(query):
            if query.qtype == 28:
                rdata = socket.inet_pton(socket.AF_INET6, "2001:db8::3")
            else:
                rdata = socket.inet_aton("192.0.2.3")
            return respond(query, [record(QUESTION_POINTER, query.qtype, 600, rdata)])

        server = self.serve(handler)
        blocks_out = self.ask(server, ["example.com A", "example.com AAAA"])
        self.check_requests(server, 2)
        expect(self, printed(blocks_out[1]), {"answer 1": "AAAA 2001:db8::3 600"})

    def test_zero_lifetime_is_not_cached(self):
        server = self.address_server(0)
        first, _ = self.ask(server, ["example.com A", "example.com A"])
        expect(self, printed(first), {"answer 1": "A 192.0.2.77 0"})
        self.check_requests(server, 2)

    def test_missing_name_is_not_cached(self):
        server = self.serve(lambda query: respond(query, rcode=3))
        self.ask(server, ["nowhere.example A", "nowhere.example A"])
        self.check_requests(server, 2)

    def test_expired_entry_is_requested_again(self):
        delay = 2.0

        def handler(query):
            if query.name.startswith("slow."):
                time.sleep(delay)
                return respond(query, [record(QUESTION_POINTER, 1, 600,
                                              socket.inet_aton("192.0.2.9"))])
            return respond(query, [record(QUESTION_POINTER, 1, 1,
                                          socket.inet_aton("192.0.2.8"))])

        server = self.serve(handler)
        lines = ["short.example A", "slow.example A", "short.example A"]
        result = self.ask(server, lines, timeout=60)
        self.check_requests(server, 3)
        expect(self, printed(result[2]), {
            "query": "short.example A",
            "status": "NOERROR",
            "answer 1": "A 192.0.2.8 1",
        })


class SilentServer(DnsTestCase):
    def silent(self):
        return self.serve(lambda query: None)

    def attempt(self, server):
        started = time.monotonic()
        try:
            code, out, err = run(args=[HOST, server.port],
                                 stdin="example.com A\n", timeout=20)
        except subprocess.TimeoutExpired:
            self.fail("программа не завершилась за 20 секунд; предел ожидания ответа "
                      "не должен превышать пяти секунд")
        spent = time.monotonic() - started
        if spent > 10:
            self.fail("программа работала %.1f с; предел ожидания ответа не должен "
                      "превышать пяти секунд" % spent)
        return code, out, err

    def test_silent_server_gives_timeout_status(self):
        code, out, _ = self.attempt(self.silent())
        result = blocks(out)
        if len(result) != 1:
            self.fail("ожидался один блок «query ... end», получено: %d\n"
                      "вывод:\n%s" % (len(result), out))
        expect(self, printed(result[0]), {
            "query": "example.com A",
            "status": "TIMEOUT",
            "записей в ответе": "0",
        })

    def test_silent_server_gives_nonzero_exit_code(self):
        code, out, _ = self.attempt(self.silent())
        if code == 0:
            self.fail("при отсутствии ответа ожидался ненулевой код возврата, "
                      "получено 0\nвывод:\n%s" % out)


if __name__ == "__main__":
    unittest.main(verbosity=2)
