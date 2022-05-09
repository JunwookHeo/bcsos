#!/bin/bash

tmux has-session -t MLDC
if [ $? != 0 ]
then

    if compgen -G "./db_nodes/*.*" > /dev/null; then
        echo "pattern exists!"
        rm ./db_nodes/*.*
    fi

    tmux new-session -s MLDC -n "SCs" -d
    tmux split-window -v -t MLDC:0
    tmux split-window -v -t MLDC:0.1
    tmux split-window -v -t MLDC:0.2
    tmux select-layout even-vertical

    tmux send-keys -t MLDC:0.0 'go run blockchainnode.go -mode=ST -sc=0 -port=7001' C-m
    tmux send-keys -t MLDC:0.1 'go run blockchainnode.go -mode=ST -sc=1 -port=7011' C-m
    tmux send-keys -t MLDC:0.2 'go run blockchainnode.go -mode=ST -sc=2 -port=7021' C-m
    tmux send-keys -t MLDC:0.3 'go run blockchainnode.go -mode=ST -sc=3 -port=7031' C-m

    tmux new-window -n "SIM" -t MLDC
    tmux send-keys -t MLDC:1 'cd ../blockchainsim' C-m
    tmux send-keys -t MLDC:1 'rm bc_sim.db' C-m
    tmux send-keys -t MLDC:1 'rm bc_sim.wallet' C-m
    tmux send-keys -t MLDC:1 'go run blockchainsim.go' C-m

fi
tmux attach -t MLDC
