package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var typeCodes = map[string]uint16{"A": 1, "NS": 2, "CNAME": 5, "MX": 15, "TXT": 16, "AAAA": 28}

var rcodeNames = map[int]string{0: "NOERROR", 1: "FORMERR", 2: "SERVFAIL", 3: "NXDOMAIN", 5: "REFUSED"}

type cacheEntry struct {
	answers []string
	expires time.Time
}

func main() {
	server := net.JoinHostPort(os.Args[1], os.Args[2])
	cache := map[string]cacheEntry{}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		name, qtype := parts[0], parts[1]
		fmt.Println("query", name, qtype)

		key := strings.ToLower(name) + " " + qtype
		if entry, ok := cache[key]; ok && time.Now().Before(entry.expires) {
			fmt.Println("status NOERROR")
			printAnswers(entry.answers)
			continue
		}

		resp, err := exchange(server, name, typeCodes[qtype])
		if err != nil {
			fmt.Println("status TIMEOUT")
			fmt.Println("end")
			os.Exit(1)
		}

		rcode, answers, minTTL := parseResponse(resp)
		status, ok := rcodeNames[rcode]
		if !ok {
			status = fmt.Sprintf("RCODE%d", rcode)
		}
		fmt.Println("status", status)
		printAnswers(answers)

		if rcode == 0 && len(answers) > 0 && minTTL > 0 {
			cache[key] = cacheEntry{answers, time.Now().Add(time.Duration(minTTL) * time.Second)}
		}
	}
}

func printAnswers(answers []string) {
	for _, a := range answers {
		fmt.Println("answer", a)
	}
	fmt.Println("end")
}

func exchange(server, name string, qtype uint16) ([]byte, error) {
	conn, err := net.Dial("udp", server)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	id := uint16(rand.Intn(65536))
	query := binary.BigEndian.AppendUint16(nil, id)
	query = append(query, 0x01, 0x00, 0, 1, 0, 0, 0, 0, 0, 0)
	for _, label := range strings.Split(strings.TrimSuffix(name, "."), ".") {
		query = append(query, byte(len(label)))
		query = append(query, label...)
	}
	query = append(query, 0)
	query = binary.BigEndian.AppendUint16(query, qtype)
	query = binary.BigEndian.AppendUint16(query, 1)

	if _, err := conn.Write(query); err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		if n >= 12 && binary.BigEndian.Uint16(buf[0:2]) == id {
			return buf[:n], nil
		}
	}
}

func parseResponse(msg []byte) (int, []string, uint32) {
	rcode := int(msg[3] & 0x0F)
	qdcount := int(binary.BigEndian.Uint16(msg[4:6]))
	ancount := int(binary.BigEndian.Uint16(msg[6:8]))

	offset := 12
	for i := 0; i < qdcount; i++ {
		_, offset = readName(msg, offset)
		offset += 4
	}

	var answers []string
	var minTTL uint32
	for i := 0; i < ancount; i++ {
		_, offset = readName(msg, offset)
		rtype := binary.BigEndian.Uint16(msg[offset : offset+2])
		ttl := binary.BigEndian.Uint32(msg[offset+4 : offset+8])
		rdlen := int(binary.BigEndian.Uint16(msg[offset+8 : offset+10]))
		start := offset + 10
		offset = start + rdlen

		value, typeName := formatRecord(msg, rtype, start, rdlen)
		if typeName == "" {
			continue
		}
		answers = append(answers, fmt.Sprintf("%s %s %d", typeName, value, ttl))
		if len(answers) == 1 || ttl < minTTL {
			minTTL = ttl
		}
	}
	return rcode, answers, minTTL
}

func formatRecord(msg []byte, rtype uint16, start, length int) (string, string) {
	data := msg[start : start+length]
	switch rtype {
	case 1:
		return net.IP(data).String(), "A"
	case 28:
		return net.IP(data).String(), "AAAA"
	case 2:
		name, _ := readName(msg, start)
		return name, "NS"
	case 5:
		name, _ := readName(msg, start)
		return name, "CNAME"
	case 15:
		name, _ := readName(msg, start+2)
		return fmt.Sprintf("%d %s", binary.BigEndian.Uint16(data[0:2]), name), "MX"
	case 16:
		var text string
		for i := 0; i < len(data); i += int(data[i]) + 1 {
			text += string(data[i+1 : i+1+int(data[i])])
		}
		return text, "TXT"
	}
	return "", ""
}

func readName(msg []byte, offset int) (string, int) {
	var labels []string
	for {
		length := int(msg[offset])
		if length&0xC0 == 0xC0 {
			pointer := int(binary.BigEndian.Uint16(msg[offset:offset+2]) & 0x3FFF)
			rest, _ := readName(msg, pointer)
			labels = append(labels, strings.TrimSuffix(rest, "."))
			return strings.Join(labels, ".") + ".", offset + 2
		}
		offset++
		if length == 0 {
			return strings.Join(labels, ".") + ".", offset
		}
		labels = append(labels, string(msg[offset:offset+length]))
		offset += length
	}
}
