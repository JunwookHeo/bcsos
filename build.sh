#!/bin/bash

DIR=./out
if [ -d "$DIR" ] 
then
    rm -rf "$DIR"
    echo "rmdir"
fi

mkdir "$DIR"
echo "create"

cd blockchainnode
/bin/bash buildnode.sh

cd ../blockchainsim
/bin/bash buildsim.sh
