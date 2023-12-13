import cv2

skip_ratio = 1

VIDEO_PATH = "/home/aharnish/Documents/cc/DDLJ/client/media/part0.mp4"

# Open video file
video = cv2.VideoCapture(VIDEO_PATH)
imW = video.get(cv2.CAP_PROP_FRAME_WIDTH)
imH = video.get(cv2.CAP_PROP_FRAME_HEIGHT)

frame_count = 0

while(video.isOpened()):

    frame_count += 1

    # Acquire frame and resize to expected shape [1xHxWx3]
    ret, frame = video.read()
    print("frame got", frame_count)
    if not ret:
      print('Reached the end of the video!')
      break
    if(frame_count % skip_ratio != 0):
        continue