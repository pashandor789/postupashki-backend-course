package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var crlf = []byte("\r\n")

func main() {
	data, _ := io.ReadAll(os.Stdin)

	end := bytes.Index(data, crlf)
	if end < 0 {
		fail("bad_start_line")
	}
	parseStartLine(string(data[:end]))
	data = data[end+2:]

	chunked := false
	contentLength := -1
	for {
		end := bytes.Index(data, crlf)
		if end < 0 {
			fail("bad_header")
		}
		line := string(data[:end])
		data = data[end+2:]
		if line == "" {
			break
		}

		name, value, found := strings.Cut(line, ":")
		if !found || name == "" || strings.ContainsAny(name, " \t") {
			fail("bad_header")
		}
		name = strings.ToLower(name)
		value = strings.Trim(value, " \t")
		fmt.Println("header", name, value)

		if name == "transfer-encoding" && strings.ToLower(value) == "chunked" {
			chunked = true
		}
		if name == "content-length" {
			n, err := strconv.Atoi(value)
			if err != nil || n < 0 {
				fail("bad_header")
			}
			contentLength = n
		}
	}

	var body []byte
	if chunked {
		body = parseChunked(data)
	} else if contentLength >= 0 {
		if len(data) < contentLength {
			fail("incomplete_body")
		}
		body = data[:contentLength]
	}

	fmt.Println("body.length", len(body))
	if len(body) > 0 && isPrintable(body) {
		fmt.Println("body.text", string(body))
	}
}

func parseStartLine(line string) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		fail("bad_start_line")
	}

	if strings.HasPrefix(parts[0], "HTTP/") {
		if parts[0] != "HTTP/1.1" || len(parts[1]) != 3 {
			fail("bad_start_line")
		}
		if _, err := strconv.Atoi(parts[1]); err != nil {
			fail("bad_start_line")
		}
		fmt.Println("type response")
		fmt.Println("version", parts[0])
		fmt.Println("status", parts[1])
		fmt.Println("reason", parts[2])
		return
	}

	if parts[2] != "HTTP/1.1" {
		fail("bad_start_line")
	}
	fmt.Println("type request")
	fmt.Println("method", parts[0])
	fmt.Println("target", parts[1])
	fmt.Println("version", parts[2])
}

func parseChunked(data []byte) []byte {
	var body []byte
	for {
		end := bytes.Index(data, crlf)
		if end < 0 {
			fail("bad_chunk")
		}
		sizeStr, _, _ := strings.Cut(string(data[:end]), ";")
		size, err := strconv.ParseInt(strings.TrimSpace(sizeStr), 16, 64)
		if err != nil || size < 0 {
			fail("bad_chunk")
		}
		data = data[end+2:]
		if size == 0 {
			return body
		}

		n := int(size)
		if len(data) < n+2 || !bytes.Equal(data[n:n+2], crlf) {
			fail("bad_chunk")
		}
		body = append(body, data[:n]...)
		data = data[n+2:]
	}
}

func isPrintable(b []byte) bool {
	for _, c := range b {
		if c < 0x20 || c > 0x7E {
			return false
		}
	}
	return true
}

func fail(reason string) {
	fmt.Println("error", reason)
	os.Exit(1)
}
