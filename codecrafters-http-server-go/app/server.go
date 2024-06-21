package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221:", err)
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Server is now listening on port 4221...")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	req := make([]byte, 1024)
	n, err := conn.Read(req)
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	request := string(req[:n])

	switch {
	case isGetRootRequest(request):
		writeResponse(conn, "HTTP/1.1 200 OK\r\n\r\n")
	case isGetEchoRequest(request):
		path := extractEchoPath(request)
		writeResponse(conn, fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(path), path))
	case isGetUserAgentRequest(request):
		userAgent := extractUserAgent(request)
		writeResponse(conn, fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent))
	case isGetFileRequest(request):
		dir := os.Args[2]
		path := extractFilePath(request)
		data, err := os.ReadFile(dir + path)
		if err != nil {
			writeResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
		} else {
			writeResponse(conn, fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data))
		}
	case isPostFileRequest(request):
		dir := os.Args[2]
		path := extractFilePath(request)
		body := extractPostBody(request)
		err := os.WriteFile(dir+path, []byte(body), 0666)
		if err != nil {
			writeResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
		} else {
			writeResponse(conn, "HTTP/1.1 201 Created\r\n\r\n")
		}
	default:
		writeResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}
}

func isGetRootRequest(request string) bool {
	match, _ := regexp.MatchString("GET / HTTP/1.1", request)
	return match
}

func isGetEchoRequest(request string) bool {
	match, _ := regexp.MatchString("^GET /echo/[A-Za-z0-9\\-._~%]+ HTTP/1\\.1", request)
	return match
}

func isGetUserAgentRequest(request string) bool {
	match, _ := regexp.MatchString("^GET /user-agent HTTP/1\\.1", request)
	return match
}

func isGetFileRequest(request string) bool {
	match, _ := regexp.MatchString("^GET /files/[A-Za-z0-9\\-._~%]+ HTTP/1\\.1", request)
	return match
}

func isPostFileRequest(request string) bool {
	match, _ := regexp.MatchString("^POST /files/[A-Za-z0-9\\-._~%]+ HTTP/1\\.1", request)
	return match
}

func extractEchoPath(request string) string {
	path := regexp.MustCompile("^GET /echo/([A-Za-z0-9\\-._~%]+) HTTP/1\\.1").FindStringSubmatch(request)[1]
	return path
}

func extractUserAgent(request string) string {
	userAgent := regexp.MustCompile("User-Agent: (.*)").FindStringSubmatch(request)[1]
	userAgent = strings.Trim(userAgent, "\r\n")
	return userAgent
}

func extractFilePath(request string) string {
	path := regexp.MustCompile("^(GET|POST) /files/([A-Za-z0-9\\-._~%]+) HTTP/1\\.1").FindStringSubmatch(request)[2]
	return path
}

func extractPostBody(request string) string {
	body := regexp.MustCompile("\r\n\r\n(.*)").FindStringSubmatch(request)[1]
	body = strings.Trim(body, "\x00")
	return body
}

func writeResponse(conn net.Conn, response string) {
	conn.Write([]byte(response))
}
