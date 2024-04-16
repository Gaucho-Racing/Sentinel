#!/bin/bash

# Check if version is provided
if [ -z "$1" ]
  then
    echo "No version number provided"
    exit 1
fi

# Check if docker is installed
if ! [ -x "$(command -v docker)" ]; then
  echo 'Error: docker is not installed.' >&2
  exit 1
fi

echo "Building container for GR Sentinel v$1"
# Build the docker container
docker build -t gauchoracing/sentinel:"$1" -t gauchoracing/sentinel:latest --platform linux/amd64,linux/arm64 --push --progress=plain .