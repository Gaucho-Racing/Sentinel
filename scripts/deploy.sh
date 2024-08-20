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

echo "Building container for Sentinel v$1"
# Build the docker container
docker build -t gauchoracing/sentinel:"$1" -t gauchoracing/sentinel:latest --platform linux/amd64,linux/arm64 --push --progress=plain .

echo "Container deployed successfully"

# Check if GitHub CLI is installed
if ! command -v gh &> /dev/null
then
    echo "GitHub CLI (gh) is not installed. Please install it to proceed."
    exit 1
fi

# Create a release tag
git tag -s v$1 -m "Release version $1"
git push origin v$1

# Create a release
gh release create v$1 --generate-notes

echo "Package released successfully for version $1"