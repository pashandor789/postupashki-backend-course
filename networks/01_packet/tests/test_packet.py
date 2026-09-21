import unittest

from harness import expect, fields, run

SYN = (
    "001b213c4d5e0025645d1e220800450000281c46400040062805c0a800025db8"
    "d822a86a005012345678000000005002faf000000000"
)
SYN_PADDED = SYN + "00" * 18

ACK_DATA = (
    "aabbccddeeff1122334455660800450000340a0b0000390663ac0a0000050a00"
    "0009c73801bb000003e8000007d0501801f60000000068656c6c6f20776f726c"
    "6421"
)

UDP_DNS = (
    "00005e00530100005e00530208004500003d1c46000078116a40ac10000a0808"
    "0808cfda003500290000abcd0100000100000000000003777777076578616d70"
    "6c6503636f6d0000010001"
)

FRAG_OPTS = (
    "001b213c4d5e0025645d1e220800460000401c4620b940062561c0a80101c0a8"
    "0102940400000000000000000000000000000000000000000000000000000000"
    "0000000000000000000000000000"
)

BAD_CKS = (
    "001b213c4d5e0025645d1e220800450000281c460000400689eac0a800020101"
    "010104d2005000000001000000015010040000000000"
)

ARP = (
    "ffffffffffff0025645d1e22080600010800060400010025645d1e22c0a80002"
    "000000000000c0a80001"
)


def parse(dump):
    code, out, err = run(stdin=dump + "\n")
    if code != 0:
        raise AssertionError("программа завершилась с кодом %d\nstderr: %s" % (code, err))
    return fields(out)


class EthernetLayer(unittest.TestCase):
    def test_addresses_and_type(self):
        expect(self, parse(SYN), {
            "eth.dst": "00:1b:21:3c:4d:5e",
            "eth.src": "00:25:64:5d:1e:22",
            "eth.ethertype": "0x0800",
        }, SYN)

    def test_non_ip_frame_stops_after_ethernet(self):
        got = parse(ARP)
        expect(self, got, {"eth.ethertype": "0x0806"}, ARP)
        self.assertNotIn("ip.src", got,
                         "для кадра с ethertype 0x0806 поля уровня IP выводить не следует")


class IPv4Layer(unittest.TestCase):
    def test_basic_fields(self):
        expect(self, parse(SYN), {
            "ip.version": "4",
            "ip.ihl_bytes": "20",
            "ip.total_length": "40",
            "ip.id": "0x1c46",
            "ip.ttl": "64",
            "ip.protocol": "6",
            "ip.src": "192.168.0.2",
            "ip.dst": "93.184.216.34",
        }, SYN)

    def test_flags_do_not_fragment(self):
        expect(self, parse(SYN), {"ip.flags": "DF", "ip.frag_offset": "0"}, SYN)

    def test_flags_more_fragments_and_offset(self):
        expect(self, parse(FRAG_OPTS), {
            "ip.flags": "MF",
            "ip.frag_offset": "1480",
            "ip.ihl_bytes": "24",
        }, FRAG_OPTS)

    def test_checksum_valid(self):
        expect(self, parse(SYN), {"ip.checksum_valid": "true"}, SYN)

    def test_checksum_invalid(self):
        expect(self, parse(BAD_CKS), {"ip.checksum_valid": "false"}, BAD_CKS)


class TcpLayer(unittest.TestCase):
    def test_ports_and_flags_syn(self):
        expect(self, parse(SYN), {
            "tcp.src_port": "43114",
            "tcp.dst_port": "80",
            "tcp.seq": "305419896",
            "tcp.ack": "0",
            "tcp.data_offset_bytes": "20",
            "tcp.flags": "SYN",
            "tcp.window": "64240",
        }, SYN)

    def test_flags_push_ack(self):
        expect(self, parse(ACK_DATA), {"tcp.flags": "PSH,ACK"}, ACK_DATA)

    def test_payload_length(self):
        expect(self, parse(ACK_DATA), {
            "ip.total_length": "52",
            "payload.length": "12",
        }, ACK_DATA)

    def test_empty_payload(self):
        expect(self, parse(SYN), {"payload.length": "0"}, SYN)

    def test_frame_padding_is_not_payload(self):
        expect(self, parse(SYN_PADDED), {"payload.length": "0"}, SYN_PADDED)


class UdpLayer(unittest.TestCase):
    def test_ports_and_length(self):
        expect(self, parse(UDP_DNS), {
            "ip.protocol": "17",
            "udp.src_port": "53210",
            "udp.dst_port": "53",
            "udp.length": "41",
            "payload.length": "33",
        }, UDP_DNS)

    def test_no_tcp_fields(self):
        got = parse(UDP_DNS)
        self.assertNotIn("tcp.src_port", got,
                         "для дейтаграммы UDP поля уровня TCP выводить не следует")


class InputFormat(unittest.TestCase):
    def test_spaces_and_newlines_are_ignored(self):
        spaced = " ".join(SYN[i:i + 4] for i in range(0, len(SYN), 4))
        expect(self, parse(spaced), {"ip.src": "192.168.0.2"}, spaced)

    def test_uppercase_hex(self):
        expect(self, parse(SYN.upper()), {"ip.dst": "93.184.216.34"}, SYN.upper())


if __name__ == "__main__":
    unittest.main(verbosity=2)
