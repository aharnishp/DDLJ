package main

import (
    "fmt"
    "net"
    "os"
    "io"
)

func main() {
    // Connect to the server
    serverAddr := "192.168.70.59:8080" // Change to the server's address
    conn, err := net.Dial("tcp", serverAddr)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        return
    }
    defer conn.Close()

    // Create a file to store the received video
    file, err := os.Create("media/received_video.mp4")
    if err != nil {
        fmt.Println("Error creating the file:", err)
        return
    }
    defer file.Close()

    // Create a buffer to read and write data
    buffer := make([]byte, 1024)

    for {
        // Read a chunk of data from the server
        bytesRead, err := conn.Read(buffer)
        if err == io.EOF {
            break
        } else if err != nil {
            fmt.Println("Error reading data from server:", err)
            return
        }

        // Write the data chunk to the file
        _, err = file.Write(buffer[:bytesRead])
        if err != nil {
            fmt.Println("Error writing data to file:", err)
            return
        }
    }

    fmt.Println("Video file received and saved to 'media/received_video.mp4'")
}
