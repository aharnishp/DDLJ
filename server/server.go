package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// Define the address to which you want to send the video file
	serverAddr := ""

	// Open the video file for reading
	videoFile, err := os.Open("../static/vi.mp4")
	if err != nil {
		fmt.Println("Error opening video file:", err)
		return
	}
	defer videoFile.Close()

	// Connect to the server
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		return
	}
	defer conn.Close()

	// Read and send the video file data
	buffer := make([]byte, 1024)
	for {
		n, err := videoFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading video file:", err)
			return
		}

		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Println("Error sending video data:", err)
			return
		}
	}

	fmt.Println("Video file sent successfully.")
}
