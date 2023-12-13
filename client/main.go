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
	"time"
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

// func isConnectionActive(conn net.Conn) bool {
// 	// Set a deadline for the read or write operation
// 	conn.SetReadDeadline(time.Now().Add(time.Second))

// 	// Attempt to read a small amount of data
// 	buffer := make([]byte, 1)
// 	_, err := conn.Read(buffer)

// 	// Reset the deadline
// 	conn.SetReadDeadline(time.Time{})

// 	// Check for errors
// 	if err != nil {
// 		// Check if the error indicates a timeout, which means the connection is still active
// 		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
// 			return true
// 		}

// 		// If the error is not a timeout, the connection is likely closed
// 		return false
// 	}

// 	return true
// }

func isConnectionActive(conn net.Conn) bool {
	// Set a deadline for the read or write operation
	conn.SetReadDeadline(time.Now().Add(time.Second))

	// Attempt to read a small amount of data
	buffer := make([]byte, 1)
	_, err := conn.Read(buffer)

	// Reset the deadline
	conn.SetReadDeadline(time.Time{})

	// Check for errors
	if err != nil {
		// Check if the error indicates a timeout, which means the connection is still active
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return true
		}

		// If the error is not a timeout, print the error
		fmt.Println("Error checking connection:", err)

		return false
	}

	return true
}

func main() {

	introduce()
	if len(os.Args) > 1 {
		serverAddress = string(os.Args[1])
		csvFilePath = string(os.Args[2])
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

	for {

		// if !isConnectionActive((conn)) {
		// 	fmt.Println("Error: Connection Lost")
		// 	break
		// }

		fileName, fileSize, err := receiveFileFromServer(conn)
		if err != nil {
			fmt.Println("Error receiving video segment:", err)

		}

		fmt.Printf("Received %s (%d bytes)\n", fileName, fileSize)
		// }

		// run the python service
		fmt.Println("Video is now being processed.")
		runAnalysisService() // dummy analysis service
		fmt.Println("Video was processed.")
		// Send the CSV file to the server
		sendFileToServer(conn, csvFilePath)

		// // Wait for acknowledgment from the server
		// acknowledgment, err := receiveAcknowledgment(conn)
		// if err != nil {
		// 	fmt.Println("Error receiving acknowledgment:", err)
		// } else {
		// 	fmt.Println(acknowledgment)
		// }
	}
}

func runAnalysisService() {

	// Get the absolute path of the current directory
	currentDir, err1 := os.Getwd()
	if err1 != nil {
		fmt.Println("Error getting current directory:", err1)
		return
	}
	// fmt.Println(currentDir)

	// Set the virtual environment path relative to the current directory
	virtualEnvPath1 := fmt.Sprintf("%s/py-tensorflow/tflite1-env", currentDir)
	// virtualEnvPath := fmt.Sprintf("%s/py-tensorflow/tflite1-env", currentDir)

	// fmt.Println(virtualEnvPath1)

	// Set the environment variables
	os.Setenv("VIRTUAL_ENV", virtualEnvPath1)
	os.Setenv("PATH", os.Getenv("PATH")+":"+os.Getenv("VIRTUAL_ENV")+"/bin")

	// // Set the environment variables
	// os.Setenv("VIRTUAL_ENV", "/home/aharnish/Documents/cc/TensorFlow-Lite-Object-Detection-on-Android-and-Raspberry-Pi/tflite1-env")
	// os.Setenv("PATH", os.Getenv("PATH")+":"+os.Getenv("VIRTUAL_ENV")+"/bin")

	fmt.Println(os.Getenv("PATH"))
	fmt.Println(os.Getenv("VIRTUAL_ENV"))

	// Launch the Python script
	cmd := exec.Command("python", "py-tensorflow/cc-video-in.py")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	// Get the exit code of the Python script
	exitCode := 0
	if err != nil {
		// Error occurred, extract exit code if possible
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			fmt.Println("Error running Python script:", err)
		}
	}

	// Print the exit code
	fmt.Printf("Python script exited with code %d\n", exitCode)

	fmt.Println("GO PROCEEDED FURTHER.")
}

// fmt.Println("Update: Service Started")
// time.Sleep(3 * time.Second)
// fmt.Println("Update: Service done")}

func receiveFileFromServer(conn net.Conn) (string, int64, error) {
	reader := bufio.NewReader(conn)

	// Read the file name and size from the server
	fileInfo, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", 0, err
	}

	// Split the fileInfo string into name and size
	infoParts := strings.Fields(fileInfo)
	if len(infoParts) != 2 {
		return "", 0, fmt.Errorf("invalid file info received")
	}

	// fileName := infoParts[0]
	fileSize, err := strconv.ParseInt(infoParts[1], 10, 64)
	if err != nil {
		return "", 0, err
	}

	// Create a file to write the segment data
	filePath := filepath.Join("media/", "part0.mp4") // fileName)
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

// func sendFileToServer(conn net.Conn, filePath string) {
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		fmt.Println("Error opening file:", err)
// 		return
// 	}
// 	defer file.Close()

// 	fileInfo, _ := file.Stat()
// 	fileSize := fileInfo.Size()
// 	fileName := filepath.Base(filePath)

// 	conn.Write([]byte(fmt.Sprintf("%s %d\n", fileName, fileSize)))

// 	buffer := make([]byte, 1024)
// 	for {
// 		n, err := file.Read(buffer)
// 		if err != nil {
// 			break
// 		}
// 		fmt.Println(buffer)
// 		conn.Write(buffer[:n])
// 	}

// 	conn.Write([]byte("\nEOF\n"))
// 	return

// }

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

	// Send the file name and size to the server
	conn.Write([]byte(fmt.Sprintf("%s %d\n", fileName, fileSize)))

	// Use io.Copy to copy the file content to the connection
	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error sending file content:", err)
		return
	}

	// Signal the end of file to the server
	conn.Write([]byte("\nEOF\n"))

	fmt.Println("File sent successfully.")
}

func receiveAcknowledgment(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	acknowledgment, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return acknowledgment, nil
}
