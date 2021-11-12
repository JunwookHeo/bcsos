#!/bin/bash

if compgen -G "./db_nodes/*.db" > /dev/null; then
    echo "pattern exists!"
    rm ./db_nodes/*.db
fi
# Storage Class == 0
screen -dmS n7001 go run storagesrv.go -mode=pan -type=0 -port=7001
screen -dmS n7002 go run storagesrv.go -mode=pan -type=0 -port=7002
screen -dmS n7003 go run storagesrv.go -mode=pan -type=0 -port=7003
screen -dmS n7004 go run storagesrv.go -mode=pan -type=0 -port=7004
screen -dmS n7005 go run storagesrv.go -mode=pan -type=0 -port=7005
screen -dmS n7006 go run storagesrv.go -mode=pan -type=0 -port=7006
screen -dmS n7007 go run storagesrv.go -mode=pan -type=0 -port=7007
screen -dmS n7008 go run storagesrv.go -mode=pan -type=0 -port=7008

# Storage Class == 1
screen -dmS n7011 go run storagesrv.go -mode=pan -type=1 -port=7011
screen -dmS n7012 go run storagesrv.go -mode=pan -type=1 -port=7012
screen -dmS n7013 go run storagesrv.go -mode=pan -type=1 -port=7013
screen -dmS n7014 go run storagesrv.go -mode=pan -type=1 -port=7014

# Storage Class == 2
screen -dmS n7021 go run storagesrv.go -mode=pan -type=2 -port=7021
screen -dmS n7022 go run storagesrv.go -mode=pan -type=2 -port=7022

# Storage Class == 3
screen -dmS n7033 go run storagesrv.go -mode=pan -type=3 -port=7033

screen -ls
