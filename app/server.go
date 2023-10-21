package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path"
	"strings"
)

type Request struct {
	Method         string
	Path           string
	Version        string
	Host           string
	UserAgent      string
	ContentLength  string
	AcceptEncoding string
	Body           string
}

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

	// trims the null bytes (\x00) from the beginning and end of the byte slice
	buffer = bytes.Trim(buffer, "\x00")

	// Read from buffer and set response correctly
	request := string(buffer)
	requestLines := strings.Split(request, "\r\n")

	//headerMethod, headerRequestPath := extractHttpMethodAndRequestPath(requestLines)

	responseData := formatRequestBufferToRequestStruct(buffer)

	bodyItem := extractUrlFromHttpHeaderPath(responseData.Path)
	// fmt.Printf("RESPONSE ITEM DATA : %v with type \n", responseData)

	switch responseData.Method {
	case "POST":
		if strings.HasPrefix(responseData.Path, "/files") {
			var response string
			fileName := strings.TrimPrefix(responseData.Path, "/files/")
			bodyRequest := []byte(requestLines[6])

			err := os.WriteFile(path.Join(*directory, fileName), bodyRequest, fs.ModeTemporary)
			if err != nil {
				fmt.Println("Writing to file failed", err.Error())
				os.Exit(1)
			}

			response = "HTTP/1.1 201 Created\r\n\r\n"
			writeHttpResponse(connection, response)
		}
	case "GET":
		if responseData.Path == "/" {
			writeHttpResponse(connection, "HTTP/1.1 200 OK\r\n\r\n")
		} else if strings.HasPrefix(responseData.Path, "/echo") {
			bodyResponse := string(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(bodyItem), bodyItem))

			writeHttpResponse(connection, bodyResponse)
		} else if strings.HasPrefix(responseData.Path, "/user-agent") {
			value := extractUserAgent(requestLines)
			bodyResponse := string(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(value), value))

			writeHttpResponse(connection, bodyResponse)
		} else if strings.HasPrefix(responseData.Path, "/files") {
			var response string
			fileName := strings.TrimPrefix(responseData.Path, "/files/")
			contents, err := os.ReadFile(path.Join(*directory, fileName))
			if err != nil {
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			} else {
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(contents), string(contents))
			}

			writeHttpResponse(connection, response)
		} else {
			writeHttpResponse(connection, "HTTP/1.1 404 Not Found\r\n\r\n")
		}
	default:
		fmt.Println("Not implemented method")
		os.Exit(1)
	}
}

func formatRequestBufferToRequestStruct(buffer []byte) Request {
	requestLines := strings.Split(string(buffer), "\r\n")
	headerLineItem := strings.Split(requestLines[0], " ")

	requestResponse := Request{}
	requestResponse.Method = headerLineItem[0]
	requestResponse.Path = headerLineItem[1]
	requestResponse.Version = headerLineItem[2]

	fmt.Printf("RESPONSE ITEM DATA ONE : %v \n", requestLines)

	for _, line := range requestLines[1:] {
		if line == "" {
			break
		}

		headerParts := strings.Split(line, ": ")

		switch headerParts[0] {
		case "Host":
			requestResponse.Host = headerParts[1]
		case "User-Agent":
			requestResponse.UserAgent = headerParts[1]
		case "Accept-Encoding":
			requestResponse.AcceptEncoding = headerParts[1]
		case "Content-Length":
			requestResponse.ContentLength = headerParts[1]
		default:
			fmt.Printf("Unhandled case  key: %s value: %s\n", headerParts[0], headerParts[1])
		}
	}

	fmt.Printf("Type: %T, Value: %+v\n", requestResponse, requestResponse)

	if len(requestLines) > len(requestLines[1:])+1 {
		requestResponse.Body = requestLines[len(requestLines[1:])+1]
	}

	return requestResponse
}

func writeHttpResponse(conn net.Conn, response string) {
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}

func extractHttpMethodAndRequestPath(headers []string) (
	httpMethod string,
	requestPath string,
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
