package main

import (
	// "bufio"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	listenAddress   = ":8080"
	videoPath       = "./media/vi.mp4" // Change this to your video file path
	segmentSize     = 1024 * 1024      // 1 MB segments
	outputDirectory = "./media/"
	duration        = 5
	filePath        = "./files/file0.csv"
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

func setup() {
	// Define command-line flags and set default values
	flag.StringVar(&listenAddress, "listen", "8080", "Listen address for the server")
	flag.StringVar(&videoPath, "video", "./media/vi.mp4", "Path to the video file")
	flag.IntVar(&segmentSize, "segmentSize", 1024*1024, "Segment size in bytes")
	flag.StringVar(&outputDirectory, "outputDir", "./media/", "Output directory")
	flag.IntVar(&duration, "duration", 20, "Segment duration in seconds")

	// Parse command-line arguments
	flag.Parse()
	listenAddress = ":" + listenAddress

}

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d%02d/%02d", year, month, day)
}

func runWebService(serverPort string, startTrigger chan<- eventTrigger) {
	router := gin.Default()
	router.Delims("{[{", "}]}")
	router.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "DDLJ",
		})
	})

	router.GET("/dummyService", func(c *gin.Context) {

		startTrigger <- eventTrigger{Type: "executeService", Payload: true}
		c.JSON(http.StatusOK, gin.H{
			"message": "Your Service has started",
		})
	})

	router.GET("/downloadFile", func(c *gin.Context) {

		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {

			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "file0.csv"))
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Expires", "0")
		c.Header("Cache-Control", "must-revalidate")
		c.Header("Pragma", "public")

		c.File(filePath)
	})

	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("video")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set the file path for storing
		filePath := "../media/" + file.Filename

		// Save the file to the specified directory
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
		})
	})

	router.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "./templates/upload/upload.html", gin.H{"message": "OK"})
	})

	router.Run(":" + serverPort)
}

type eventTrigger struct {
	Type    string
	Payload bool
}

func main() {

	introduce()
	setup()
	// Start the server
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Slave Connection Server is listening on port", listenAddress)

	connectionManager := &ConnectionManager{}

	var executeService = make(chan eventTrigger)

	go runWebService("8000", executeService)

	for {
		go connectionManager.AcceptConnections(listener)

		fmt.Println("Number of active connections: " + fmt.Sprint(len(connectionManager.connections)))
		select {
		case msg, ok := <-executeService:
			if !ok {
				fmt.Println("Error: Execute service issue")
				break
			}

			switch msg.Payload {
			case true:
				splitVideo(videoPath, outputDirectory, len(connectionManager.connections))
				for i := 0; i < len(connectionManager.connections); i++ {
					handleClient(connectionManager.connections[i], fmt.Sprintf("./media/part%d.mp4", i), fmt.Sprintf("./files/file%d.csv", i))
				}
				// executeService <- eventTrigger{Type: "executeService", Payload: false}

			case false:
				time.Sleep(1 * time.Second)
				continue
			}

		default:
			time.Sleep(1 * time.Second)
			continue
		}

	}

}

// ConnectionManager represents a structure to manage connections
type ConnectionManager struct {
	connections []net.Conn
	mutex       sync.Mutex
}

// AcceptConnections continuously accepts connections and updates the connection array
func (cm *ConnectionManager) AcceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Add the connection to the array
		cm.addConnection(conn)

		// Handle the connection concurrently
		fmt.Println("Accepted connection from", conn.RemoteAddr())
		go cm.handleConnection(conn)
	}
}

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

		// If the error is not a timeout, the connection is likely closed
		return false
	}

	return true
}

// handleConnection handles the tasks associated with a connection
func (cm *ConnectionManager) handleConnection(conn net.Conn) {
	defer conn.Close()
	for {

		if !isConnectionActive(conn) {
			fmt.Println("Connection broken with client")
			cm.removeConnection(conn)
			break
		}

		time.Sleep(1 * time.Second)
	}
}

// addConnection adds a connection to the connection array
func (cm *ConnectionManager) addConnection(conn net.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.connections = append(cm.connections, conn)
	fmt.Println("Number of connections:", len(cm.connections))
}

func (cm *ConnectionManager) removeConnection(conn net.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	var index = 0
	for i, c := range cm.connections {
		if conn == c {
			index = i
			break
		}
	}
	cm.connections = append(cm.connections[:index], cm.connections[index+1:]...)
	fmt.Println("Number of connections:", len(cm.connections))
}

func connectAvailableClients(listener net.Listener, stopAccepting chan struct{}, dataChannel chan []net.Conn) {
	var connections []net.Conn
	clientNumber := 0
	ln := 0
	fmt.Println("loop number" + fmt.Sprint(ln))
	for {
		ln += 1
		fmt.Println("loop number" + fmt.Sprint(ln))
		select {

		case <-stopAccepting:
			dataChannel <- connections
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}
			fmt.Println("Accepted Client : " + fmt.Sprint(clientNumber))
			connections = append(connections, conn)
			dataChannel <- connections

			clientNumber += 1

		}

	}
}

func handleClient(conn net.Conn, videoSegmentPath string, filePath string) {

	fmt.Printf("Client %s connected.\n", conn.RemoteAddr())

	sendFileToClient(conn, videoSegmentPath)

	// Receive the CSV file from the client
	receivedCSVPath := filePath
	receiveFileFromClient(conn, receivedCSVPath)
	fmt.Printf("Received CSV file from client %s.\n", conn.RemoteAddr())

	// // Acknowledge the client
	// conn.Write([]byte("CSV file received. Acknowledged.\n"))
}

func splitVideo(inputVideo string, outputDirectory string, numParts int) {

	// Ensure the output directory exists.
	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		os.Mkdir(outputDirectory, os.ModeDir)
	}

	// Calculate the duration of the input video.
	ffprobeCmd := exec.Command("ffprobe", "-i", inputVideo, "-show_entries", "format=duration", "-v", "quiet", "-of", "csv=p=0")
	durationBytes, err := ffprobeCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error getting video duration: %v\n", err)
		return
	}

	duration, err := strconv.ParseFloat(string(durationBytes[:4]), 64)

	fmt.Println("Video Duration: " + fmt.Sprint(duration))

	if err != nil {
		fmt.Printf("Error parsing duration of video")
	}
	fmt.Printf("Input video duration: %f seconds\n", duration)

	// Calculate the duration of each part.

	partDuration := duration / float64(numParts)

	fmt.Println("Part Duration: " + fmt.Sprint(partDuration))
	// Split the video into equal parts.
	for i := 0; i < numParts; i++ {
		outputFile := filepath.Join(outputDirectory, fmt.Sprintf("part%d.mp4", i))
		trimStart := fmt.Sprintf("%.2f", float64(i)*float64(partDuration))
		trimEnd := fmt.Sprintf("%.2f", float64(i+1)*float64(partDuration))

		fmt.Println("Trim Start" + fmt.Sprint(trimStart) + ":::: trim end:  " + fmt.Sprintf(trimEnd))

		splitCmd := exec.Command("ffmpeg", "-i", inputVideo, "-ss", trimStart, "-to", trimEnd, "-c", "copy", outputFile)
		err := splitCmd.Run()
		if err != nil {
			fmt.Printf("Error splitting video: %v\n", err)
			return
		}
		fmt.Printf("Part %d: %s to %s\n", i+1, trimStart, trimEnd)
	}
	return

}

func sendFileToClient(conn net.Conn, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	fileName := filepath.Base(filePath)

	// Send the file size and name to the client
	conn.Write([]byte(fmt.Sprintf("%s %d\n", fileName, fileSize)))

	buffer := make([]byte, segmentSize)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			break
		}
		conn.Write(buffer[:n])
	}

	return
}

// func receiveFileFromClient(conn net.Conn, filePath string) {
// 	file, err := os.Create(filePath)
// 	if err != nil {
// 		fmt.Println("Error creating file:", err)
// 		return
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(conn)

// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if line == "EOF" {
// 			file.WriteString(line + "\n")
// 			break
// 		}
// 		fmt.Println(line)
// 		file.WriteString(line + "\n")
// 	}

// 	return

// }

func receiveFileFromClient(conn net.Conn, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Set a deadline for the entire file transfer process
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // Adjust the timeout duration as needed
	// conn.SetReadDeadline(time.Unix(0, 0)) // time.Now().Add(10 * time.Second))
	fmt.Println("Before io.Copy")

	// conn.SetReadDeadline(time.Unix(0, 0)) // time.Now().Add(10 * time.Second))
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, err = io.Copy(file, conn)
	if err != nil {
		fmt.Println("Error receiving file content:", err)
		return
	}

	fmt.Println("After io.Copy")

	if err != nil {
		fmt.Println("Error receiving file content:", err)
		return
	}

	fmt.Println("File received successfully.")
}

func flushStorage() {

}
