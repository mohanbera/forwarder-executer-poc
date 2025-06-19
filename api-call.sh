#!/bin/bash

# Configuration
FILES_DIR="/home/administrator/Desktop/text-files/"  # directory containing files
API_URL="http://132.186.123.91:3003/hello"  # API endpoint URL

# Ensure the directory exists
if [ ! -d "$FILES_DIR" ]; then
  echo "Directory $FILES_DIR does not exist."
  exit 1
fi

# Get all files in the directory into an array
FILES=("$FILES_DIR"/*)

# Check if there are any files
if [ ${#FILES[@]} -eq 0 ]; then
  echo "No files found in $FILES_DIR."
  exit 1
fi

# Loop 100 times
for i in {1..100}; do
  # Pick a random file
  RANDOM_FILE="${FILES[RANDOM % ${#FILES[@]}]}"
  
  echo "[$i/100] Uploading: $RANDOM_FILE"

  # Send with curl
  curl -s -o /dev/null -w "%{http_code}\n" -F "upload=@${RANDOM_FILE}" "$API_URL"

  # Optional: add delay to prevent overloading the server
  # sleep 0.5
done

echo "All uploads completed."
