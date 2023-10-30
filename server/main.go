package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

var (
	listenAddress   = ":8080"
	videoPath       = "./media/vi.mp4" // Change this to your video file path
	segmentSize     = 1024 * 1024      // 1 MB segments
	outputDirectory = "./media/"
	duration        = 5
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
	flag.StringVar(&listenAddress, "listen", ":8080", "Listen address for the server")
	flag.StringVar(&videoPath, "video", "./media/vi.mp4", "Path to the video file")
	flag.IntVar(&segmentSize, "segmentSize", 1024*1024, "Segment size in bytes")
	flag.StringVar(&outputDirectory, "outputDir", "./media/", "Output directory")
	flag.IntVar(&duration, "duration", 5, "Segment duration in seconds")

	// Parse command-line arguments
	flag.Parse()
	listenAddress = ":" + listenAddress

}

func main() {

	introduce()
	setup()
	var connections []net.Conn
	// Start the server
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	for {

		fmt.Println("Server is listening on port", listenAddress)

		dataChannel := make(chan []net.Conn)
		stopAccepting := make(chan struct{})

		go connectAvailableClients(listener, stopAccepting, dataChannel)

		waitFor(duration)
		close(stopAccepting)

		connections = <-dataChannel

		if len(connections) == 0 {
			fmt.Println("No Client Available: Waiting...")
		} else {
			break
		}

	}
	splitVideo(videoPath, outputDirectory, len(connections))

	for i := 0; i < len(connections); i++ {
		handleClient(connections[i], fmt.Sprintf("./media/part%d.mp4", i), fmt.Sprintf("./files/file%d.csv", i))
	}
}

func waitFor(duration int) {
	for i := 0; i < duration; i++ {
		fmt.Println("Waiting for client: " + fmt.Sprint(duration-i) + " seconds left")
		time.Sleep(1 * time.Second)
	}
}

func connectAvailableClients(listener net.Listener, stopAccepting chan struct{}, dataChannel chan []net.Conn) {
	var connections []net.Conn

	for {
		clientNumber := 0
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
	defer conn.Close()

	fmt.Printf("Client %s connected.\n", conn.RemoteAddr())

	sendFileToClient(conn, videoSegmentPath)

	// Receive the CSV file from the client
	receivedCSVPath := filePath
	receiveFileFromClient(conn, receivedCSVPath)
	fmt.Printf("Received CSV file from client %s.\n", conn.RemoteAddr())

	// Acknowledge the client
	conn.Write([]byte("CSV file received. Acknowledged.\n"))
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
	if err != nil {
		fmt.Printf("Error parsing duration of video")
	}
	fmt.Printf("Input video duration: %f seconds\n", duration)

	// Calculate the duration of each part.

	partDuration := duration / float64(numParts)

	// Split the video into equal parts.
	for i := 0; i < numParts; i++ {
		outputFile := filepath.Join(outputDirectory, fmt.Sprintf("part%d.mp4", i))
		trimStart := fmt.Sprintf("%.2f", float64(i)*float64(partDuration))
		trimEnd := fmt.Sprintf("%.2f", float64(i+1)*float64(partDuration))

		splitCmd := exec.Command("ffmpeg", "-i", inputVideo, "-ss", trimStart, "-to", trimEnd, "-c", "copy", outputFile)
		err := splitCmd.Run()
		if err != nil {
			fmt.Printf("Error splitting video: %v\n", err)
			return
		}
		fmt.Printf("Part %d: %s to %s\n", i+1, trimStart, trimEnd)
	}

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
}

func receiveFileFromClient(conn net.Conn, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "EOF" {
			conn.Write([]byte("\n"))
			break
		}
		file.WriteString(line + "\n")
	}
}
