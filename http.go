 package main
  
import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    []byte
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	log.Println("Listening on :8080 ...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		// عالج كل اتصال في goroutine
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// نستخدم bufio.Reader عشان نقرأ سطر سطر وبمرونة
	reader := bufio.NewReader(conn)

	// 1) اقرأ سطر الطلب: GET /path HTTP/1.1
	startLine, err := readLine(reader)
	if err != nil {
		if err != io.EOF {
			log.Println("read start line error:", err)
		}
		return
	}
	if startLine == "" {
		// اتصال فاضي
		return
	}
	method, path, version, err := parseStartLine(startLine)
	if err != nil {
		writeSimpleResponse(conn, 400, "Bad Request", "Invalid start line")
		return
	}

	// 2) اقرأ الهيدرز إلى أن نوصل لسطر فاضي
	headers, err := readHeaders(reader)
	if err != nil {
		writeSimpleResponse(conn, 400, "Bad Request", "Invalid headers")
		return
	}

	// 3) لو فيه Content-Length نقرأ body بهالطول
	var body []byte
	if clStr, ok := headers["content-length"]; ok {
		cl, err := strconv.Atoi(strings.TrimSpace(clStr))
		if err != nil || cl < 0 {
			writeSimpleResponse(conn, 400, "Bad Request", "Invalid Content-Length")
			return
		}
		body = make([]byte, cl)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			writeSimpleResponse(conn, 400, "Bad Request", "Could not read body")
			return
		}
	}

	req := Request{
		Method:  method,
		Path:    path,
		Version: version,
		Headers: headers,
		Body:    body,
	}

	// 4) راوتر بسيط حسب الـ path
	switch {
	case req.Method == "GET" && req.Path == "/":
		home(conn, &req)
	case req.Method == "GET" && req.Path == "/hello":
		hello(conn, &req)
	case req.Method == "POST" && req.Path == "/echo":
		echo(conn, &req)
	default:
		writeSimpleResponse(conn, 404, "Not Found", "Route not found")
	}
}

func home(conn net.Conn, req *Request) {
	html := `<!doctype html>
<html>
<head><meta charset="utf-8"><title>Home</title></head>
<body>
<h1>It works 🎉</h1>
<p>You built HTTP over TCP in Go.</p>
<p>Try <a href="/hello">/hello</a> or POST to <code>/echo</code>.</p>
</body></html>`
	writeResponse(conn, 200, "OK", "text/html; charset=utf-8", []byte(html))
}

func hello(conn net.Conn, req *Request) {
	msg := "Hello from raw TCP HTTP!\n"
	writeResponse(conn, 200, "OK", "text/plain; charset=utf-8", []byte(msg))
}

func echo(conn net.Conn, req *Request) {
	// يرجّع نفس البودي اللي استقبلناه
	ct := req.Headers["content-type"]
	if ct == "" {
		ct = "application/octet-stream"
	}
	writeResponse(conn, 200, "OK", ct, req.Body)
}

// ===== Helpers =====

func readLine(r *bufio.Reader) (string, error) {
	// يقرأ حتى \n (ويسيب \r لو موجود)
	line, err := r.ReadString('\n')
	if err != nil {
		// ممكن err == io.EOF إذا مافي سطر كامل
		return "", err
	}
	// نشيل \r\n أو \n
	return strings.TrimRight(line, "\r\n"), nil
}

func parseStartLine(line string) (method, path, version string, err error) {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid start line")
	}
	return parts[0], parts[1], parts[2], nil
}

func readHeaders(r *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)
	for {
		hline, err := readLine(r)
		if err != nil {
			return nil, err
		}
		if hline == "" {
			// سطر فاضي = نهاية الهيدرز
			break
		}
		// صيغة: Key: Value
		colon := strings.Index(hline, ":")
		if colon <= 0 {
			// هيدر غير صحيح
			continue
		}
		key := strings.ToLower(strings.TrimSpace(hline[:colon]))
		val := strings.TrimSpace(hline[colon+1:])
		headers[key] = val
	}
	return headers, nil
}

func writeSimpleResponse(conn net.Conn, status int, statusText, body string) {
	writeResponse(conn, status, statusText, "text/plain; charset=utf-8", []byte(body+"\n"))
}

func writeResponse(conn net.Conn, status int, statusText, contentType string, body []byte) {
	// نبني الاستجابة يدويًا
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "HTTP/1.1 %d %s\r\n", status, statusText)
	fmt.Fprintf(sb, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(sb, "Content-Length: %d\r\n", len(body))
	// لأبسطية: نخلي Connection: close
	fmt.Fprintf(sb, "Connection: close\r\n")
	fmt.Fprintf(sb, "\r\n") // نهاية الهيدرز
	headerBytes := []byte(sb.String())

	// نكتب الهيدر ثم البودي
	_, _ = conn.Write(headerBytes)
	_, _ = conn.Write(body)
}
