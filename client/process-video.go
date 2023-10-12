package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"gocv.io/x/gocv"
	"github.com/tensorflow/tensorflow/tensorflow/go"
)

func main() {
	// Load the pre-trained TensorFlow model for object detection.
	modelPath := "./models/ssd_resnet152_v1_fpn_1024x1024_coco17_tpu-8/saved_model/saved_model.pb"
	model, err := tensorflow.LoadSavedModel(modelPath, []string{"serve"}, nil)
	if err != nil {
		log.Fatalf("Failed to load the model: %v", err)
	}
	defer model.Session.Close()

	// Open the video file.
	videoPath := "./media/received_video.mp4"
	video, err := gocv.VideoCaptureFile(videoPath)
	if err != nil {
		log.Fatalf("Error opening video: %v", err)
		return
	}
	defer video.Close()

	window := gocv.NewWindow("Object Detection")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := video.Read(&img); !ok {
			fmt.Printf("Cannot read video feed\n")
			return
		}

		inputTensor, outputTensor, err := prepareInputOutput(model, img)
		if err != nil {
			log.Fatalf("Error preparing input/output: %v", err)
		}

		session := model.Session
		result, err := session.Run(
			map[tensorflow.Output]*tensorflow.Tensor{
				model.Graph.Operation(outputTensor).Output(0): outputTensor,
			},
			map[tensorflow.Output]*tensorflow.Tensor{
				model.Graph.Operation(inputTensor).Output(0): inputTensor,
			},
			nil,
		)
		if err != nil {
			log.Fatalf("Error running session: %v", err)
		}

		detectedObjects, err := postprocessResult(result)
		if err != nil {
			log.Fatalf("Error post-processing result: %v", err)
		}

		for _, obj := range detectedObjects {
			drawBoundingBox(&img, obj)
		}

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

func prepareInputOutput(model *tensorflow.SavedModel, img gocv.Mat) (*tensorflow.Tensor, *tensorflow.Tensor, error) {
	inputOp := "serving_default_input_tensor_name"
	outputOp := "serving_default_output_tensor_name"

	graph := model.Graph
	inputTensor := graph.Operation(inputOp).Output(0)
	outputTensor := graph.Operation(outputOp).Output(0)

	tensor, err := gocv.ImgToTensor(img)
	if err != nil {
		return nil, nil, err
	}

	return tensor, outputTensor, nil
}

func postprocessResult(result map[tensorflow.Output]*tensorflow.Tensor) ([]Object, error) {
	// Implement post-processing logic for your specific object detection model here.
	// Extract object detection results from the output tensor and convert them into
	// a list of objects with their coordinates and class labels.
}

func drawBoundingBox(img *gocv.Mat, obj Object) {
	gocv.Rectangle(img, image.Rect(obj.X, obj.Y, obj.X+obj.Width, obj.Y+obj.Height), color.RGBA{0, 255, 0, 0}, 2)
	label := fmt.Sprintf("%s: %.2f", obj.Class, obj.Confidence)
	gocv.PutText(img, label, image.Point{obj.X, obj.Y - 10}, gocv.FontHersheySimplex, 0.5, color.RGBA{0, 255, 0, 0}, 2)
}

type Object struct {
	Class      string
	Confidence float32
	X          int
	Y          int
	Width      int
	Height     int
}
