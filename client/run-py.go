package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

var (
	serverAddress = "localhost:8080"
	csvFilePath   = "./files/data.csv" // Path to your CSV file
)

func runAnalysisService() {
	// Set the environment variables
	os.Setenv("VIRTUAL_ENV", "/home/aharnish/Documents/cc/DDLJ/client/py-tensorflow/tflite1-env")
	// os.Setenv("VIRTUAL_ENV", "/home/aharnish/Documents/cc/DDLJ/client/py-tensorflow/tflite1-env")
	// os.Setenv("VIRTUAL_ENV", "/home/aharnish/Documents/cc/TensorFlow-Lite-Object-Detection-on-Android-and-Raspberry-Pi/tflite1-env")
	os.Setenv("PATH", os.Getenv("PATH")+":"+os.Getenv("VIRTUAL_ENV")+"/bin")

	fmt.Println(os.Getenv("PATH"))
	fmt.Println(os.Getenv("VIRTUAL_ENV"))

	// Launch the Python script
	cmd := exec.Command("python", "py-tensorflow/cc-video-in.py")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("Update: Service Started")
	// time.Sleep(3 * time.Second)
	// fmt.Println("Update: Service done")
}

func main() {
	runAnalysisService()

}
