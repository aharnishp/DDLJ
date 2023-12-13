package main

import (
	"fmt"
	"os"
	"os/exec"
)

var (
	serverAddress = "localhost:8080"
	csvFilePath   = "./files/data.csv" // Path to your CSV file
)

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

func main() {
	runAnalysisService()

}
