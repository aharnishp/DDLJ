package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// Define the port the server will listen on
	port := ":8080"

	// Listen for incoming connections
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port", port)

	for {
		// Accept a connection from a client
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		defer conn.Close()

		// Handle the connection
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Println("Client connected:", conn.RemoteAddr())

	// Open the video file
	videoFile, err := os.Open("../static/vi.mp4") // Change "video.mp4" to the actual path of your video file
	if err != nil {
		fmt.Println("Error opening video file:", err)
		return
	}
	defer videoFile.Close()

	// Create a buffer to read and send the video data
	buffer := make([]byte, 1024)

	for {
		// Read a chunk of data from the video file
		bytesRead, err := videoFile.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error reading video file:", err)
			return
		}

		// Send the data chunk to the client
		_, err = conn.Write(buffer[:bytesRead])
		if err != nil {
			fmt.Println("Error sending data to client:", err)
			return
		}
	}

	fmt.Println("Video file sent to client:", conn.RemoteAddr())
}
