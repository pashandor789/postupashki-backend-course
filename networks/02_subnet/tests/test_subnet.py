import os
import tempfile
import unittest

from harness import expect, fields, run

TIMEOUT = 15

TABLE = """\
# таблица маршрутов
0.0.0.0/0 eth0

10.0.0.0/8 eth1
10.1.0.0/16 eth2
10.1.2.0/24 eth3
192.168.0.0/16 eth4
203.0.113.7/32 eth5
"""

TABLE_WITHOUT_DEFAULT = """\
10.0.0.0/8 eth1
192.168.0.0/16 eth4
"""


def subnet(case, spec):
    code, out, err = run(["subnet", spec], timeout=TIMEOUT)
    if code != 0:
        case.fail("subnet %s: программа завершилась с кодом %d\nstderr: %s"
                  % (spec, code, err))
    return fields(out)


def write_table(case, text):
    handle, path = tempfile.mkstemp(prefix="routes-", suffix=".txt")
    with os.fdopen(handle, "w", encoding="utf-8") as stream:
        stream.write(text)
    case.addCleanup(os.unlink, path)
    return path


def route(case, table_text, destination):
    path = write_table(case, table_text)
    code, out, err = run(["route", path, destination], timeout=TIMEOUT)
    return code, fields(out), err


def route_ok(case, table_text, destination):
    code, got, err = route(case, table_text, destination)
    if code != 0:
        case.fail("route %s: программа завершилась с кодом %d, ожидался 0\nstderr: %s"
                  % (destination, code, err))
    return got


class SubnetProperties(unittest.TestCase):
    def test_network_and_broadcast(self):
        expect(self, subnet(self, "192.168.10.0/24"), {
            "network": "192.168.10.0",
            "broadcast": "192.168.10.255",
            "prefix": "24",
        })

    def test_netmask_and_host_range(self):
        expect(self, subnet(self, "192.168.10.0/24"), {
            "netmask": "255.255.255.0",
            "first": "192.168.10.1",
            "last": "192.168.10.254",
            "hosts": "254",
        })

    def test_address_inside_block_is_reduced_to_network(self):
        expect(self, subnet(self, "192.168.10.37/24"), {
            "network": "192.168.10.0",
            "broadcast": "192.168.10.255",
            "first": "192.168.10.1",
        })

    def test_prefix_not_aligned_to_octet(self):
        expect(self, subnet(self, "10.1.2.200/22"), {
            "network": "10.1.0.0",
            "netmask": "255.255.252.0",
            "broadcast": "10.1.3.255",
            "first": "10.1.0.1",
            "last": "10.1.3.254",
            "hosts": "1022",
        })

    def test_prefix_30_point_to_point(self):
        expect(self, subnet(self, "172.16.5.5/30"), {
            "network": "172.16.5.4",
            "netmask": "255.255.255.252",
            "broadcast": "172.16.5.7",
            "first": "172.16.5.5",
            "last": "172.16.5.6",
            "hosts": "2",
        })

    def test_prefix_31_has_no_broadcast(self):
        expect(self, subnet(self, "10.0.0.4/31"), {
            "network": "10.0.0.4",
            "netmask": "255.255.255.254",
            "broadcast": "none",
            "first": "10.0.0.4",
            "last": "10.0.0.5",
            "hosts": "2",
        })

    def test_prefix_31_accepts_second_address_of_pair(self):
        expect(self, subnet(self, "10.0.0.5/31"), {
            "network": "10.0.0.4",
            "first": "10.0.0.4",
            "last": "10.0.0.5",
            "hosts": "2",
        })

    def test_prefix_32_single_address(self):
        expect(self, subnet(self, "203.0.113.9/32"), {
            "network": "203.0.113.9",
            "netmask": "255.255.255.255",
            "broadcast": "none",
            "first": "203.0.113.9",
            "last": "203.0.113.9",
            "hosts": "1",
        })

    def test_prefix_0_covers_whole_address_space(self):
        expect(self, subnet(self, "0.0.0.0/0"), {
            "network": "0.0.0.0",
            "netmask": "0.0.0.0",
            "broadcast": "255.255.255.255",
            "first": "0.0.0.1",
            "last": "255.255.255.254",
            "hosts": "4294967294",
        })

    def test_short_prefix_8(self):
        expect(self, subnet(self, "10.20.30.40/8"), {
            "network": "10.0.0.0",
            "netmask": "255.0.0.0",
            "broadcast": "10.255.255.255",
            "hosts": "16777214",
        })

    def test_short_prefix_12(self):
        expect(self, subnet(self, "172.20.30.40/12"), {
            "network": "172.16.0.0",
            "netmask": "255.240.0.0",
            "broadcast": "172.31.255.255",
            "last": "172.31.255.254",
            "hosts": "1048574",
        })

    def test_invalid_prefix_length_is_rejected(self):
        code, out, err = run(["subnet", "10.0.0.0/33"], timeout=TIMEOUT)
        self.assertNotEqual(
            code, 0,
            "для длины префикса 33 ожидался ненулевой код возврата, получен 0\n"
            "stdout: %s" % out)


class RouteSelection(unittest.TestCase):
    def test_longest_prefix_wins(self):
        expect(self, route_ok(self, TABLE, "10.1.2.5"),
               {"via": "eth3", "prefix": "24"})

    def test_shorter_prefix_used_when_longer_does_not_match(self):
        expect(self, route_ok(self, TABLE, "10.1.9.9"),
               {"via": "eth2", "prefix": "16"})

    def test_aggregate_route_used_when_subnets_do_not_match(self):
        expect(self, route_ok(self, TABLE, "10.9.9.9"),
               {"via": "eth1", "prefix": "8"})

    def test_default_route_used_when_nothing_else_matches(self):
        expect(self, route_ok(self, TABLE, "8.8.8.8"),
               {"via": "eth0", "prefix": "0"})

    def test_host_route_wins_over_network_route(self):
        expect(self, route_ok(self, TABLE, "203.0.113.7"),
               {"via": "eth5", "prefix": "32"})

    def test_line_order_does_not_change_result(self):
        reordered = "\n".join(reversed(TABLE.splitlines())) + "\n"
        expect(self, route_ok(self, reordered, "10.1.2.5"),
               {"via": "eth3", "prefix": "24"})

    def test_comments_and_blank_lines_are_ignored(self):
        noisy = ("\n# первая запись\n\n10.1.2.0/24 eth3\n"
                 "\n#10.1.2.0/24 eth9\n\n0.0.0.0/0 eth0\n\n")
        expect(self, route_ok(self, noisy, "10.1.2.5"),
               {"via": "eth3", "prefix": "24"})

    def test_missing_route_reports_unreachable(self):
        code, got, err = route(self, TABLE_WITHOUT_DEFAULT, "8.8.8.8")
        expect(self, got, {"unreachable": "true"})
        self.assertNotEqual(
            code, 0,
            "при отсутствии подходящей записи ожидался ненулевой код возврата, получен 0\n"
            "stderr: %s" % err)

    def test_match_present_gives_zero_exit_code(self):
        path = write_table(self, TABLE_WITHOUT_DEFAULT)
        code, out, err = run(["route", path, "192.168.1.1"], timeout=TIMEOUT)
        self.assertEqual(
            code, 0,
            "при найденном маршруте ожидался код возврата 0, получен %d\nstderr: %s"
            % (code, err))
        expect(self, fields(out), {"via": "eth4", "prefix": "16"})


if __name__ == "__main__":
    unittest.main(verbosity=2)
