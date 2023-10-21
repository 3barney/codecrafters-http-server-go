package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

var directory *string

func main() {
	directory = flag.String("directory", "/", "File store location")
	flag.Parse()

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Bind to a port and start listening for connections
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	fmt.Println("Directory: ", *directory)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		// Handle connections
		go handleClientConnection(conn)
	}
}

func handleClientConnection(connection net.Conn) {
	defer connection.Close()

	// Read and process data from the client
	buffer := make([]byte, 1024)

	_, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading connection contents")
		os.Exit(1)
	}

	// Read from buffer and set response correctly
	request := string(buffer)
	requestLines := strings.Split(request, "\r\n")

	headerMethod, headerRequestPath := readHttpHeadersFromRequestLine(requestLines)
	bodyItem := extractUrlFromHttpHeaderPath(headerRequestPath)

	switch headerMethod {
	case "GET":
		if headerRequestPath == "/" {
			_, err = connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				os.Exit(1)
			}
		} else if strings.HasPrefix(headerRequestPath, "/echo") {
			bodyResponse := string(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(bodyItem), bodyItem))

			_, err = connection.Write([]byte(bodyResponse))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				os.Exit(1)
			}
		} else if strings.HasPrefix(headerRequestPath, "/user-agent") {
			value := extractUserAgent(requestLines)
			bodyResponse := string(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(value), value))

			_, err = connection.Write([]byte(bodyResponse))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				os.Exit(1)
			}
		} else if strings.HasPrefix(headerRequestPath, "/files") {
			var response string
			fileName := strings.TrimPrefix(headerRequestPath, "/files/")
			contents, err := os.ReadFile(path.Join(*directory, fileName))
			if err != nil {
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			} else {
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(contents), string(contents))
			}

			_, err = connection.Write([]byte(response))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				os.Exit(1)
			}
		} else {
			_, err = connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				os.Exit(1)
			}
		}
	default:
		fmt.Println("Not implemented method")
		os.Exit(1)
	}
}

/*
*
filename := subpaths[2]
+ 				fileContent, err := os.ReadFile(path.Join(*dir, filename))
+ 				if err != nil {
+ 					response = "HTTP/1.1 404 Not Found\r\n\r\n"
+ 				} else {
+ 					response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", len(fileContent))
+ 					response += string(fileContent) + "\r\n"
1
+ 				}
*
*/
func readHttpHeadersFromRequestLine(headers []string) (
	httpHeaderMethod string,
	httpRequestPath string,
) {
	headerItems := strings.Split(headers[0], " ")
	method := headerItems[0]
	path := headerItems[1]

	return method, path
}

func extractUrlFromHttpHeaderPath(headerPath string) string {
	pathItems := strings.TrimPrefix(headerPath, "/echo/")

	fmt.Println(pathItems)
	return pathItems
}

func extractUserAgent(requestLines []string) string {
	if len(requestLines) < 2 {
		fmt.Println("Missing user agent property")
		os.Exit(1)
	}

	return strings.TrimPrefix(requestLines[2], "User-Agent: ")
}
