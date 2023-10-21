package main

import (
	"fmt"
	"net"
	"os"
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

	// Write data back to the client
	data := []byte("HTTP/1.1 200 OK\r\n\r\n")
	_, err = connection.Write(data)

	if err != nil {
		fmt.Println("Error writing back to connection")
		os.Exit(1)
	}
}
