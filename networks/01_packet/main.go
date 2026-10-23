package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	input, _ := io.ReadAll(os.Stdin)
	hexStr := strings.Map(func(r rune) rune {
		if '0' <= r && r <= '9' || 'a' <= r && r <= 'f' {
			return r
		}
		return -1
	}, strings.ToLower(string(input)))
	data, err := hex.DecodeString(hexStr)
	if err != nil || len(data) < 14 {
		fmt.Fprintln(os.Stderr, "bad input")
		os.Exit(1)
	}

	etherType := binary.BigEndian.Uint16(data[12:14])
	fmt.Println("eth.dst", formatMAC(data[0:6]))
	fmt.Println("eth.src", formatMAC(data[6:12]))
	fmt.Printf("eth.ethertype 0x%04x\n", etherType)
	if etherType != 0x0800 {
		return
	}

	ip := data[14:]
	ihl := int(ip[0]&0x0F) * 4
	totalLength := int(binary.BigEndian.Uint16(ip[2:4]))
	fragField := binary.BigEndian.Uint16(ip[6:8])
	protocol := ip[9]

	var ipFlags []string
	if fragField&0x4000 != 0 {
		ipFlags = append(ipFlags, "DF")
	}
	if fragField&0x2000 != 0 {
		ipFlags = append(ipFlags, "MF")
	}

	fmt.Println("ip.version", ip[0]>>4)
	fmt.Println("ip.ihl_bytes", ihl)
	fmt.Println("ip.total_length", totalLength)
	fmt.Printf("ip.id 0x%04x\n", binary.BigEndian.Uint16(ip[4:6]))
	fmt.Println("ip.flags", joinOrNone(ipFlags))
	fmt.Println("ip.frag_offset", int(fragField&0x1FFF)*8)
	fmt.Println("ip.ttl", ip[8])
	fmt.Println("ip.protocol", protocol)
	fmt.Println("ip.src", formatIP(ip[12:16]))
	fmt.Println("ip.dst", formatIP(ip[16:20]))
	fmt.Println("ip.checksum_valid", checksum(ip[:ihl]) == 0xFFFF)

	t := ip[ihl:]
	headerLen := 0
	switch protocol {
	case 6:
		headerLen = int(t[12]>>4) * 4
		names := []string{"FIN", "SYN", "RST", "PSH", "ACK", "URG"}
		var tcpFlags []string
		for i, name := range names {
			if t[13]&(1<<i) != 0 {
				tcpFlags = append(tcpFlags, name)
			}
		}
		fmt.Println("tcp.src_port", binary.BigEndian.Uint16(t[0:2]))
		fmt.Println("tcp.dst_port", binary.BigEndian.Uint16(t[2:4]))
		fmt.Println("tcp.seq", binary.BigEndian.Uint32(t[4:8]))
		fmt.Println("tcp.ack", binary.BigEndian.Uint32(t[8:12]))
		fmt.Println("tcp.data_offset_bytes", headerLen)
		fmt.Println("tcp.flags", joinOrNone(tcpFlags))
		fmt.Println("tcp.window", binary.BigEndian.Uint16(t[14:16]))
	case 17:
		headerLen = 8
		fmt.Println("udp.src_port", binary.BigEndian.Uint16(t[0:2]))
		fmt.Println("udp.dst_port", binary.BigEndian.Uint16(t[2:4]))
		fmt.Println("udp.length", binary.BigEndian.Uint16(t[4:6]))
	}

	fmt.Println("payload.length", totalLength-ihl-headerLen)
}

func checksum(header []byte) uint32 {
	var sum uint32
	for i := 0; i+1 < len(header); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(header[i : i+2]))
	}
	for sum>>16 != 0 {
		sum = sum&0xFFFF + sum>>16
	}
	return sum
}

func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, ",")
}

func formatMAC(b []byte) string {
	var parts []string
	for _, v := range b {
		parts = append(parts, fmt.Sprintf("%02x", v))
	}
	return strings.Join(parts, ":")
}

func formatIP(b []byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}
