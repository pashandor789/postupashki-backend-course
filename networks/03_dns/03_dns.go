package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var typeCodes = map[string]uint16{
	"A":     1,
	"NS":    2,
	"CNAME": 5,
	"MX":    15,
	"TXT":   16,
	"AAAA":  28,
}

type Record struct {
	Type  string
	Value string
	TTL   uint32
}

type cacheEntry struct {
	record  []Record
	expires time.Time
}

func buildQuery(name string, qtype uint16) ([]byte, uint16, error) {
	if len(name) == 0 {
		return nil, 0, fmt.Errorf("Пусто")
	}
	result := []byte{}
	header := []byte{
		0x00, 0x00,
		0x01, 0x00,
		0x00, 0x01,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
	}
	id := uint16(rand.Intn(65536))
	binary.BigEndian.PutUint16(header[0:2], id)

	en := encodeName(name)
	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, qtype)

	result = append(result, header...)
	result = append(result, en...)
	result = append(result, typeBytes...)
	result = append(result, 0x00, 0x01)

	return result, id, nil
}

func encodeName(name string) []byte {
	name = strings.TrimSuffix(name, ".")
	parts := strings.Split(name, ".")
	result := []byte{}
	for _, part := range parts {
		result = append(result, byte(len(part)))
		result = append(result, []byte(part)...)
	}
	result = append(result, 0x00)
	return result

}

func readResponse(conn net.Conn, id uint16) ([]byte, error) {
	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		if n < 12 {
			continue
		}
		if binary.BigEndian.Uint16(buf[0:2]) != id {
			continue
		}
		return buf[:n], nil
	}
}

func rcodeName(rcode uint16) string {
	switch rcode {
	case 0:
		return "NOERROR"
	case 1:
		return "FORMERR"
	case 2:
		return "SERVFAIL"
	case 3:
		return "NXDOMAIN"
	case 5:
		return "REFUSED"
	default:
		return fmt.Sprintf("RCODE%d", rcode)
	}
}

func formatIPv6(data []byte) string {
	if len(data) != 16 {
		return "invalid"
	}

	groups := make([]uint16, 8)
	for i := 0; i < 8; i++ {
		groups[i] = binary.BigEndian.Uint16(data[i*2 : i*2+2])
	}

	maxStart, maxLen := -1, 0
	curStart, curLen := -1, 0
	for i := 0; i < 8; i++ {
		if groups[i] == 0 {
			if curStart == -1 {
				curStart, curLen = i, 1
			} else {
				curLen++
			}
			if curLen > maxLen {
				maxStart, maxLen = curStart, curLen
			}
		} else {
			curStart, curLen = -1, 0
		}
	}

	parts := []string{}
	i := 0
	for i < 8 {
		if i == maxStart {
			if i == 0 {
				parts = append(parts, "")
			}
			parts = append(parts, "")
			i += maxLen
			if i == 8 {
				parts = append(parts, "")
			}
		} else {
			parts = append(parts, fmt.Sprintf("%x", groups[i]))
			i++
		}
	}

	result := strings.Join(parts, ":")

	for strings.Contains(result, ":::") {
		result = strings.ReplaceAll(result, ":::", "::")
	}
	return result
}

func readName(msg []byte, offset int) (string, int) {
	labels := []string{}
	end := -1
	jumps := 0

	for {
		b := msg[offset]

		if b&0xC0 == 0xC0 {
			jumps++
			if jumps > 128 {
				break
			}
			if end == -1 {
				end = offset + 2
			}
			offset = int(binary.BigEndian.Uint16(msg[offset:offset+2]) & 0x3FFF)
		} else if b == 0 {
			if end == -1 {
				end = offset + 1
			}
			break
		} else {
			l := int(b)
			labels = append(labels, string(msg[offset+1:offset+1+l]))
			offset += 1 + l
		}

	}
	return strings.Join(labels, ".") + ".", end
}

func parseAnswers(msg []byte) []Record {
	qdcount := int(binary.BigEndian.Uint16(msg[4:6]))
	ancount := int(binary.BigEndian.Uint16(msg[6:8]))
	offset := 12
	for i := 0; i < qdcount; i++ {
		_, offset = readName(msg, offset)
		offset += 4
	}
	records := []Record{}
	for i := 0; i < ancount; i++ {
		_, offset = readName(msg, offset)
		rtype := binary.BigEndian.Uint16(msg[offset : offset+2])
		ttl := binary.BigEndian.Uint32(msg[offset+4 : offset+8])
		rdlen := int(binary.BigEndian.Uint16(msg[offset+8 : offset+10]))
		offset += 10
		rdataOffset := offset

		if rtype == 1 {
			if rdlen == 4 {
				value := net.IP(msg[rdataOffset : rdataOffset+4]).String()
				records = append(records, Record{"A", value, ttl})
			}
		} else if rtype == 28 {
			if rdlen == 16 {
				value := formatIPv6(msg[rdataOffset : rdataOffset+16])
				records = append(records, Record{"AAAA", value, ttl})
			}
		} else if rtype == 5 || rtype == 2 {
			name, _ := readName(msg, rdataOffset)
			typeName := "CNAME"
			if rtype == 2 {
				typeName = "NS"
			}
			records = append(records, Record{typeName, name, ttl})
		} else if rtype == 15 {
			if rdlen >= 2 {
				priority := binary.BigEndian.Uint16(msg[rdataOffset : rdataOffset+2])
				name, _ := readName(msg, rdataOffset+2)
				value := fmt.Sprintf("%d %s", priority, name)
				records = append(records, Record{"MX", value, ttl})
			}
		} else if rtype == 16 {
			result := ""
			pos := rdataOffset
			for pos < rdataOffset+rdlen {
				l := int(msg[pos])
				pos++
				result += string(msg[pos : pos+l])
				pos += l
			}
			records = append(records, Record{"TXT", result, ttl})
		}

		offset += rdlen
	}
	return records
}

func main() {
	addr := os.Args[1]
	port := os.Args[2]

	scanner := bufio.NewScanner(os.Stdin)
	conn, err := net.Dial("udp", net.JoinHostPort(addr, port))
	if err != nil {
		log.Fatal("ошибка в dial")
	}
	defer conn.Close()

	cache := map[string]cacheEntry{}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		typeName := parts[1]
		key := strings.ToLower(name) + " " + typeName
		code, ok := typeCodes[typeName]
		if !ok {
			continue
		}
		query, id, err := buildQuery(name, code)
		if err != nil {
			continue
		}
		fmt.Println("query", name, typeName)

		entry, found := cache[key]

		if found && time.Now().Before(entry.expires) {
			fmt.Println("status NOERROR")
			for _, r := range entry.record {
				fmt.Println("answer", r.Type, r.Value, r.TTL)
			}
			fmt.Println("end")
			continue
		}

		conn.Write(query)

		response, err := readResponse(conn, id)
		if err != nil {
			fmt.Println("status TIMEOUT")
			fmt.Println("end")
			os.Exit(1)
		}

		flags := binary.BigEndian.Uint16(response[2:4])
		rcode := flags & 0x0F
		fmt.Println("status", rcodeName(rcode))
		records := parseAnswers(response)
		for _, r := range records {
			fmt.Println("answer", r.Type, r.Value, r.TTL)
		}
		if rcode == 0 && len(records) > 0 {
			minTTL := records[0].TTL
			for _, r := range records {
				if r.TTL < minTTL {
					minTTL = r.TTL
				}
			}
			if minTTL > 0 {
				cache[key] = cacheEntry{records, time.Now().Add(time.Duration(minTTL) * time.Second)}
			}
		}
		fmt.Println("end")

	}

}
