#!/bin/bash

echo "Build MLDC Simulator Node"

CGO_ENABLED=1

DIR=./out
if [ -d "$DIR" ] 
then
    rm -rf "$DIR"
    echo "rmdir"
fi

while read -p "Select Target (1. ARM64 Linux, 2.AMD64 Windows): " opt1
do
    if [ $opt1 -eq 1 ]
    then
        GOARCH=arm64
        GOOS=linux
        if [[ $OSTYPE == 'darwin'* ]]; then
            CC=aarch64-unknown-linux-gnu-gcc
        elif [ $OSTYPE == 'linux'* ]]; then
            CC=aarch64-linux-gnu-gcc
        fi
                
        OUT_FILE=blockchainnode
        break
    elif [ $opt1 -eq 2 ]    
    then
        GOARCH=amd64
        GOOS=windows
        CC=x86_64-w64-mingw32-gcc
        OUT_FILE=blockchainnode.exe
        break
    else
        echo "Choose the right one"
    fi
done

env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=$CGO_ENABLED CC=$CC go build -o $DIR/$OUT_FILE

if [ -d "$DIR" ] 
then
    mv "$DIR" ../out/client
    echo "mv dir"
fi