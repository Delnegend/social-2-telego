#!/usr/sh

docker run --rm -v "$(pwd)":/app -w /app golang:alpine go build -o app