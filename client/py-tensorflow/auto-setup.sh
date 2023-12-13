# git clone https://github.com/EdjeElectronics/TensorFlow-Lite-Object-Detection-on-Android-and-Raspberry-Pi.git

# mv TensorFlow-Lite-Object-Detection-on-Android-and-Raspberry-Pi tflite1
# cd tflite1

sudo pip3 install virtualenv
python3 -m venv tflite1-env
source tflite1-env/bin/activate

pip3 install tensorflow opencv-python

echo "The following process may end up in error on new ubuntu."

bash get_pi_requirements.sh

