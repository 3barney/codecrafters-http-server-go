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

const (
	HostHeader           = "Host"
	UserAgentHeader      = "User-Agent"
	AcceptEncodingHeader = "Accept-Encoding"
	ContentLengthHeader  = "Content-Length"
)

const (
	GetMethod  = "GET"
	PostMethod = "POST"
)

const (
	RootPath      = "/"
	EchoPath      = "/echo"
	UserAgentPath = "/User-Agent"
	FilesPath     = "/files"
)

const FileNamePath = "/files/"

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

	// trims the null bytes (\x00) at end since we fill it to 1024
	buffer = bytes.Trim(buffer, "\x00")

	httpRequestStruct := requestBufferToRequestStruct(buffer)

	switch httpRequestStruct.Method {
	case PostMethod:
		if strings.HasPrefix(httpRequestStruct.Path, FilesPath) {
			fileName := strings.TrimPrefix(httpRequestStruct.Path, FileNamePath)
			handlePostFilesPath(connection, httpRequestStruct, directory, fileName)
		}
	case GetMethod:
		switch {
		case httpRequestStruct.Path == RootPath:
			handleRootPath(connection)
		case strings.HasPrefix(httpRequestStruct.Path, EchoPath):
			handleEchoPath(connection, httpRequestStruct.Path)
		case strings.HasPrefix(httpRequestStruct.Path, UserAgentPath):
			handleUserAgentPath(connection, httpRequestStruct.UserAgent)
		case strings.HasPrefix(httpRequestStruct.Path, FilesPath):
			fileName := strings.TrimPrefix(httpRequestStruct.Path, FileNamePath)
			handleGetFilesPath(connection, directory, fileName)
		default:
			handleUnknownPath(connection)
		}
	default:
		fmt.Println("Not implemented method")
		os.Exit(1)
	}
}

func requestBufferToRequestStruct(buffer []byte) Request {
	requestLines := strings.Split(string(buffer), "\r\n")
	headerLineItem := strings.Split(requestLines[0], " ")

	requestResponse := Request{
		Method:  headerLineItem[0],
		Path:    headerLineItem[1],
		Version: headerLineItem[2],
	}

	for _, line := range requestLines[1:] {
		if line == "" {
			break
		}

		name, value := parseHeader(line)

		headerMap := map[string]*string{
			HostHeader:           &requestResponse.Host,
			UserAgentHeader:      &requestResponse.UserAgent,
			AcceptEncodingHeader: &requestResponse.AcceptEncoding,
			ContentLengthHeader:  &requestResponse.ContentLength,
		}

		if destination, exists := headerMap[name]; exists {
			*destination = value
		} else {
			fmt.Printf("Unhandled case  key: %s value: %s\n", name, value)
		}
	}

	fmt.Printf("Type: %T, Value: %+v\n", requestResponse, requestResponse)

	if requestResponse.Method == "POST" && len(requestLines) >= 6 {
		requestResponse.Body = requestLines[6]
	}

	return requestResponse
}

func parseHeader(header string) (string, string) {
	parts := strings.SplitN(header, ": ", 2)
	if len(parts) < 2 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func writeHttpResponse(conn net.Conn, response string) {
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}

func handleRootPath(connection net.Conn) {
	writeHttpResponse(connection, "HTTP/1.1 200 OK\r\n\r\n")
}

func handleEchoPath(connection net.Conn, path string) {
	bodyItem := strings.TrimPrefix(path, "/echo/")
	bodyResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(bodyItem), bodyItem)
	writeHttpResponse(connection, bodyResponse)
}

func handleUserAgentPath(connection net.Conn, userAgent string) {
	bodyResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
	writeHttpResponse(connection, bodyResponse)
}

func handleGetFilesPath(connection net.Conn, directory *string, fileName string) {
	contents, err := os.ReadFile(path.Join(*directory, fileName))
	if err != nil {
		writeHttpResponse(connection, "HTTP/1.1 404 Not Found\r\n\r\n")
		return
	}

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(contents), string(contents))
	writeHttpResponse(connection, response)
}

func handlePostFilesPath(connection net.Conn, request Request, directory *string, fileName string) {

	err := os.WriteFile(path.Join(*directory, fileName), []byte(request.Body), fs.ModeTemporary)
	if err != nil {
		fmt.Println("Writing to file failed", err.Error())
		os.Exit(1)
	}

	response := "HTTP/1.1 201 Created\r\n\r\n"
	writeHttpResponse(connection, response)
}

func handleUnknownPath(connection net.Conn) {
	writeHttpResponse(connection, "HTTP/1.1 404 Not Found\r\n\r\n")
}
