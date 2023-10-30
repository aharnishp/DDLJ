#include <stdio.h>
#include <tensorflow/c/c_api.h>
#include <opencv2/opencv.hpp>
#include <fstream>




const char* input_tensor_name = "image_tensor:0";
const int input_width = 1024;
const int input_height = 1024;
const int input_channels = 3;

const char* tags[] = { "serve" }; // Use the default tag "serve"


// void save_detection_results_image(const cv::Mat& original_image, float* boxes, float* scores, int* classes, int num_detections, const char* output_filename) {
//     cv::Mat image = original_image.clone();
//     for (int i = 0; i < num_detections; i++) {
//         // Retrieve box coordinates
//         float ymin = boxes[i * 4] * image.rows;
//         float xmin = boxes[i * 4 + 1] * image.cols;
//         float ymax = boxes[i * 4 + 2] * image.rows;
//         float xmax = boxes[i * 4 + 3] * image.cols;

//         // Draw bounding box and label
//         cv::rectangle(image, cv::Point(xmin, ymin), cv::Point(xmax, ymax), cv::Scalar(0, 255, 0), 2);
//         cv::putText(image, std::to_string(classes[i]), cv::Point(xmin, ymin - 10), cv::FONT_HERSHEY_SIMPLEX, 0.5, cv::Scalar(0, 255, 0), 2);
//     }

//     // Save the image with bounding boxes to a file
//     cv::imwrite(output_filename, image);
// }

void save_detection_results(const char* filename, float* boxes, float* scores, int* classes, int num_detections) {
    std::ofstream output_file(filename);
    if (!output_file.is_open()) {
        fprintf(stderr, "Error opening output file for writing.\n");
        return;
    }

    for (int i = 0; i < num_detections; i++) {
        output_file << "Detected object " << i << ": Class " << classes[i]
                    << ", Score " << scores[i] << ", Box (" << boxes[i * 4]
                    << ", " << boxes[i * 4 + 1] << ", " << boxes[i * 4 + 2]
                    << ", " << boxes[i * 4 + 3] << ")\n";
    }

    output_file.close();
}

void print_detection_results(float* boxes, float* scores, int* classes, int num_detections) {
    for (int i = 0; i < num_detections; i++) {
        printf("Detected object %d: Class %d, Score %f, Box (%f, %f, %f, %f)\n",
            i, classes[i], scores[i], boxes[i * 4], boxes[i * 4 + 1], boxes[i * 4 + 2], boxes[i * 4 + 3]);
    }
}

int main(int argc, char** argv) {
    // Initialize TensorFlow library
    TF_Status* status = TF_NewStatus();
    TF_Graph* graph = TF_NewGraph();
    TF_SessionOptions* session_opts = TF_NewSessionOptions();

    TF_Session* session = TF_LoadSessionFromSavedModel(
        session_opts, NULL, "models/ssd_resnet152_v1_fpn_1024x1024_coco17_tpu-8/saved_model", tags, 1, graph, NULL, status);

    if (TF_GetCode(status) != TF_OK) {
        fprintf(stderr, "Error loading the model: %s\n", TF_Message(status));
        return 1;
    }

    // Load and process the input JPEG image using OpenCV
    // cv::Mat image = cv::imread("input_image.jpg");

    // Read the image data from stdin (pipe)
    std::vector<uint8_t> image_data;
    char buffer[1024];
    size_t bytesRead;
    while ((bytesRead = fread(buffer, 1, sizeof(buffer), stdin)) > 0) {
        image_data.insert(image_data.end(), buffer, buffer + bytesRead);
    }

    // Create an OpenCV Mat from the image data
    cv::Mat image = cv::imdecode(image_data, cv::IMREAD_COLOR);

    if (image.empty()) {
        fprintf(stderr, "Error loading the input image.\n");
        return 1;
    }


    // Resize the image to the model's input size
    // cv::resize(image, image, cv::Size(input_width, input_height));

    // Preprocess the image as needed for your specific model
    // Convert to float32 and normalize pixel values (0-255 to 0-1)
    image.convertTo(image, CV_32F, 1.0 / 255.0);


    // Prepare the input tensor
    TF_Tensor* input_tensor = TF_NewTensor(
        TF_FLOAT, NULL, 0, image.data, input_width * input_height * input_channels * sizeof(float), NULL, NULL);

    // Set up the input map
    TF_Output inputs[1];
    inputs[0].oper = TF_GraphOperationByName(graph, input_tensor_name);
    inputs[0].index = 0;

    TF_Tensor* input_tensors[1] = { input_tensor };

    // Set up the output map
    TF_Output output;
    output.oper = TF_GraphOperationByName(graph, "detection_boxes:0");
    output.index = 0;

    
    // Run inference
    TF_Tensor* output_tensors[4];
    TF_SessionRun(session, NULL, inputs, input_tensors, 1, &output, output_tensors, 4, NULL, 0, NULL, status);

    if (TF_GetCode(status) != TF_OK) {
        fprintf(stderr, "Error running inference: %s\n", TF_Message(status));
        return 1;
    }

    // Extract detection results
    float* boxes = (float*)TF_TensorData(output_tensors[0]);
    float* scores = (float*)TF_TensorData(output_tensors[1]);
    int* classes = (int*)TF_TensorData(output_tensors[3]);
    int num_detections = TF_Dim(output_tensors[0], 1);

    // Print the detected objects
    print_detection_results(boxes, scores, classes, num_detections);

    save_detection_results("out.txt", boxes, scores, classes, num_detections);
    // save_detection_results_image(image, boxes, scores, classes, num_detections, "out.jpg");

    // Clean up TensorFlow resources
    TF_DeleteTensor(input_tensor);
    for (int i = 0; i < 4; i++) {
        TF_DeleteTensor(output_tensors[i]);
    }
    TF_CloseSession(session, status);
    TF_DeleteSession(session, status);
    TF_DeleteSessionOptions(session_opts);
    TF_DeleteGraph(graph);
    TF_DeleteStatus(status);

    return 0;
}
