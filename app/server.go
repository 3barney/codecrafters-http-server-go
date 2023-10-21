package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Bind to a port and start listening for connections
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	// defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	// Handle connection
	handleClientConnection(conn)
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
	fmt.Printf(" This is the request for http endpoint \n", request)

	requestLines := strings.Split(request, "\r\nn")

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
			fmt.Printf("Header request path %s \n", headerRequestPath)
			fmt.Printf("Body Response %s \n", bodyResponse)

			_, err = connection.Write([]byte(bodyResponse))
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

func readHttpHeadersFromRequestLine(headers []string) (httpHeaderMethod string, httpRequestPath string) {
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
