#!/bin/bash

ldflags="-s -w"

echo "building the rook-operator"
GOOS=linux go build -ldflags "${ldflags}" github.com/rook/rook-operator

#echo "building the container"
#docker build -t rook/operator .
