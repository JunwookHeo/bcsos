#!/bin/bash

if compgen -G "./db_nodes/*.db" > /dev/null; then
    echo "pattern exists!"
    rm ./db_nodes/*.db
fi

screen -dmS n7001 go run storagesrv.go -mode=pan -type=0 -port=7001

screen -dmS n7002 go run storagesrv.go -mode=pan -type=1 -port=7002

screen -dmS n7003 go run storagesrv.go -mode=pan -type=2 -port=7003

screen -ls
