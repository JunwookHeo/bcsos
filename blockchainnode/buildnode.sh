#!/bin/bash

echo "Build MLDC Simulator Node"

CGO_ENABLED=1

while read -p "Select Target (1. ARM64 Linux, 2. AMD64 Linux, 3.AMD64 Windows): " opt1
do
    if [ $opt1 -eq 1 ]
    then
        GOARCH=arm64
        GOOS=linux
        CC=aarch64-linux-gnu-gcc
        break
    elif [ $opt1 -eq 2 ]
    then
        GOARCH=amd64
        GOOS=linux
        CC=""
        break
    elif [ $opt1 -eq 3 ]
    then
        GOARCH=amd64
        GOOS=windows
        CC=x86_64-w64-mingw32-gcc
        break
    else
        echo "Choose the right one"
    fi
done

env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=$CGO_ENABLED CC=$CC go build
