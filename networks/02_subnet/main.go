package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	if os.Args[1] == "subnet" {
		subnet(os.Args[2])
	} else {
		route(os.Args[2], os.Args[3])
	}
}

func subnet(spec string) {
	addr, prefix := parseCIDR(spec)
	mask := maskOf(prefix)
	network := addr & mask
	broadcast := network | ^mask

	first, last := network+1, broadcast-1
	hosts := uint64(1)<<(32-prefix) - 2
	broadcastStr := formatIP(broadcast)
	if prefix == 31 {
		first, last, hosts, broadcastStr = network, broadcast, 2, "none"
	}
	if prefix == 32 {
		first, last, hosts, broadcastStr = network, network, 1, "none"
	}

	fmt.Println("network", formatIP(network))
	fmt.Println("broadcast", broadcastStr)
	fmt.Println("netmask", formatIP(mask))
	fmt.Println("prefix", prefix)
	fmt.Println("first", formatIP(first))
	fmt.Println("last", formatIP(last))
	fmt.Println("hosts", hosts)
}

func route(path, destination string) {
	dst := parseIP(destination)
	file, err := os.Open(path)
	if err != nil {
		fail(err.Error())
	}

	bestPrefix, bestVia := -1, ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 || strings.HasPrefix(parts[0], "#") {
			continue
		}
		network, prefix := parseCIDR(parts[0])
		if dst&maskOf(prefix) == network&maskOf(prefix) && prefix > bestPrefix {
			bestPrefix, bestVia = prefix, parts[1]
		}
	}

	if bestPrefix < 0 {
		fmt.Println("unreachable true")
		os.Exit(1)
	}
	fmt.Println("via", bestVia)
	fmt.Println("prefix", bestPrefix)
}

func parseCIDR(spec string) (uint32, int) {
	addr, prefixStr, _ := strings.Cut(spec, "/")
	prefix, err := strconv.Atoi(prefixStr)
	if err != nil || prefix < 0 || prefix > 32 {
		fail("bad prefix: " + spec)
	}
	return parseIP(addr), prefix
}

func parseIP(s string) uint32 {
	ip := net.ParseIP(s).To4()
	if ip == nil {
		fail("bad address: " + s)
	}
	return binary.BigEndian.Uint32(ip)
}

func maskOf(prefix int) uint32 {
	return ^uint32(0) << (32 - prefix)
}

func formatIP(v uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", v>>24, v>>16&0xFF, v>>8&0xFF, v&0xFF)
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
