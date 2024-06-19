package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

const CRLF = "\r\n"

func main() {
	// Print logs for debugging
	fmt.Println("Logs from your program will appear here!")

	// Use flag to parse the directory argument
	dir := flag.String("directory", "", "enter a directory")
	flag.Parse()

	// Bind to port 4221
	ln, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("Server is listening on port 4221")

	// Accept connections in a loop
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle each connection in a separate goroutine
		go func() {
			defer conn.Close()

			// Read the request
			buf := make([]byte, 1024)
			_, err = conn.Read(buf)
			if err != nil {
				fmt.Println("Error reading request:", err)
				return
			}

			req := string(buf)
			lines := strings.Split(req, CRLF)
			if len(lines) < 1 {
				fmt.Println("Invalid request")
				return
			}

			// Parse the request line
			requestLine := strings.Split(lines[0], " ")
			if len(requestLine) < 2 {
				fmt.Println("Invalid request line")
				return
			}

			method := requestLine[0]
			path := requestLine[1]

			fmt.Println("Request path:", path)

			var res string

			switch {
			case path == "/":
				res = "HTTP/1.1 200 OK\r\n\r\n/"
			case strings.HasPrefix(path, "/echo/"):
				msg := strings.TrimPrefix(path, "/echo/")
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(msg), msg)
			case path == "/user-agent":
				if len(lines) > 2 {
					msg := strings.Split(lines[2], " ")[1]
					res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(msg), msg)
				} else {
					res = "HTTP/1.1 404 Not Found\r\n\r\n"
				}
			case strings.HasPrefix(path, "/files/") && *dir != "":
				filename := strings.TrimPrefix(path, "/files/")
				filepath := *dir + filename
				fmt.Println("Filepath:", filepath)

				if method == "GET" {
					if file, err := os.ReadFile(filepath); err == nil {
						content := string(file)
						res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
					} else {
						res = "HTTP/1.1 404 Not Found\r\n\r\n"
					}
				} else if method == "POST" {
					body := strings.Split(req, CRLF+CRLF)[1]
					if err := os.WriteFile(filepath, []byte(body), 0644); err == nil {
						fmt.Println("Wrote file")
						res = "HTTP/1.1 201 Created\r\n\r\n"
					} else {
						res = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
					}
				}
			default:
				res = "HTTP/1.1 404 Not Found\r\n\r\n"
			}

			// Send the response
			_, err = conn.Write([]byte(res))
			if err != nil {
				fmt.Println("Error writing response:", err)
				return
			}
		}()
	}
}
