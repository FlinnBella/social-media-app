#!/bin/bash
set -e 

echo "Testing reels API!"

curl -X POST http://localhost:8080/api/v1/video/reels \
-H "Content-Type: multipart/form-data" \
-F "image=@tests/test_images/test_image_1.jpg" \
-F "image_name=test_image_1.jpg" \
-F "prompt=This is a test image"
