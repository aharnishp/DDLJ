package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	serverAddress = "localhost:8080"
	csvFilePath   = "./files/data.csv" // Path to your CSV file
)

func introduce() {

	art := []string{
		"DDDD   DDDD  L   JJJJJJJJ ",
		"D   D  D   D L        J   ",
		"D   D  D   D L        J   ",
		"D   D  D   D L    J   J   ",
		"DDDD   DDDD  LLLL JJJJJ   ",
	}

	for _, line := range art {

		fmt.Println(line)
	}

	fmt.Print("Distributed Deep Learning Jobs\n")
	fmt.Print("A project developed at Ahmedabad University by Aharnish, Jevin, Mohnish, and Yansi\n")
}

func main() {

	introduce()
	if len(os.Args) > 1 {
		serverAddress = string(os.Args[0])
		csvFilePath = string(os.Args[1])
	} else {
		fmt.Println("Go Client worker")
		fmt.Println("Args: ServerAddress CSVFilePath")

		fmt.Println("Using Default Server Address and file path")

	}
	// Connect to the server
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		return
	}
	defer conn.Close()

	// Receive video segments from the server
	// for {
	fileName, fileSize, err := receiveFileFromServer(conn)
	if err != nil {
		fmt.Println("Error receiving video segment:", err)

	}

	fmt.Printf("Received %s (%d bytes)\n", fileName, fileSize)
	// }

	// run the python service
	runAnalysisService()

	// Send the CSV file to the server
	sendFileToServer(conn, csvFilePath)

	// Wait for acknowledgment from the server
	acknowledgment, err := receiveAcknowledgment(conn)
	if err != nil {
		fmt.Println("Error receiving acknowledgment:", err)
	} else {
		fmt.Println(acknowledgment)
	}
}

func runAnalysisService() {
	c := exec.Command("python ./client/service/main.py")

	if err := c.Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}
func receiveFileFromServer(conn net.Conn) (string, int64, error) {
	reader := bufio.NewReader(conn)

	// Read the file name and size from the server
	fileInfo, err := reader.ReadString('\n')
	if err != nil {
		return "", 0, err
	}

	// Split the fileInfo string into name and size
	infoParts := strings.Fields(fileInfo)
	if len(infoParts) != 2 {
		return "", 0, fmt.Errorf("invalid file info received")
	}

	fileName := infoParts[0]
	fileSize, err := strconv.ParseInt(infoParts[1], 10, 64)
	if err != nil {
		return "", 0, err
	}

	// Create a file to write the segment data
	filePath := filepath.Join("media/", fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	// Read and write the file data
	_, err = io.CopyN(file, reader, fileSize)
	if err != nil {
		return "", 0, err
	}

	return filePath, fileSize, nil
}

func sendFileToServer(conn net.Conn, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	fileName := filepath.Base(filePath)

	conn.Write([]byte(fmt.Sprintf("%s %d\n", fileName, fileSize)))

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			break
		}
		conn.Write(buffer[:n])
	}

	conn.Write([]byte("\nEOF\n"))
}

func receiveAcknowledgment(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	acknowledgment, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return acknowledgment, nil
}
