#!/bin/bash

tmux has-session -t MLDC
if [ $? != 0 ]
then

    if compgen -G "./db_nodes/*.*" > /dev/null; then
        echo "pattern exists!"
        rm ./db_nodes/*.*
    fi

    tmux new-session -s MLDC -n "SC0" -d
    
    tmux split-window -v -t MLDC:0.0
    tmux split-window -v -t MLDC:0.0
    tmux split-window -v -t MLDC:0.2
    tmux split-window -v -t MLDC:0.0
    tmux split-window -v -t MLDC:0.2
    tmux split-window -v -t MLDC:0.4
    tmux split-window -v -t MLDC:0.6
    tmux select-layout even-vertical

    tmux split-window -h -t MLDC:0.0
    tmux split-window -h -t MLDC:0.2
    tmux split-window -h -t MLDC:0.4
    tmux split-window -h -t MLDC:0.6
    tmux split-window -h -t MLDC:0.8
    tmux split-window -h -t MLDC:0.10
    tmux split-window -h -t MLDC:0.12
    tmux split-window -h -t MLDC:0.14
    tmux send-keys -t MLDC:0.0 'go run blockchainnode.go -mode=ST -sc=0 -port=7001' C-m
    tmux send-keys -t MLDC:0.1 'go run blockchainnode.go -mode=ST -sc=0 -port=7002' C-m
    tmux send-keys -t MLDC:0.2 'go run blockchainnode.go -mode=ST -sc=0 -port=7003' C-m
    tmux send-keys -t MLDC:0.3 'go run blockchainnode.go -mode=ST -sc=0 -port=7004' C-m
    tmux send-keys -t MLDC:0.4 'go run blockchainnode.go -mode=ST -sc=0 -port=7005' C-m
    tmux send-keys -t MLDC:0.5 'go run blockchainnode.go -mode=ST -sc=0 -port=7006' C-m
    tmux send-keys -t MLDC:0.6 'go run blockchainnode.go -mode=ST -sc=0 -port=7007' C-m 
    tmux send-keys -t MLDC:0.7 'go run blockchainnode.go -mode=ST -sc=0 -port=7008' C-m
    tmux send-keys -t MLDC:0.8 'go run blockchainnode.go -mode=ST -sc=1 -port=7011' C-m
    tmux send-keys -t MLDC:0.9 'go run blockchainnode.go -mode=ST -sc=1 -port=7012' C-m
    tmux send-keys -t MLDC:0.10 'go run blockchainnode.go -mode=ST -sc=1 -port=7013' C-m
    tmux send-keys -t MLDC:0.11 'go run blockchainnode.go -mode=ST -sc=1 -port=7014' C-m
    tmux send-keys -t MLDC:0.12 'go run blockchainnode.go -mode=ST -sc=2 -port=7021' C-m
    tmux send-keys -t MLDC:0.13 'go run blockchainnode.go -mode=ST -sc=2 -port=7022' C-m
    tmux send-keys -t MLDC:0.14 'go run blockchainnode.go -mode=ST -sc=3 -port=7031' C-m 
    tmux send-keys -t MLDC:0.15 'cd ../blockchainsim' C-m
    tmux send-keys -t MLDC:0.15 'rm bc_sim.db' C-m
    tmux send-keys -t MLDC:0.15 'rm bc_sim.wallet' C-m
    tmux send-keys -t MLDC:0.15 'go run blockchainsim.go' C-m

    # tmux select-layout even-horizontal

    # tmux new-window -n "SC3" -t MLDC
    # tmux split-window -v -t MLDC:1
    # tmux split-window -v -t MLDC:1.1
    # tmux split-window -v -t MLDC:1.2
    # tmux split-window -h -t MLDC:1.0
    # tmux split-window -h -t MLDC:1.2
    # tmux split-window -h -t MLDC:1.4
    # tmux split-window -h -t MLDC:1.6
    # tmux send-keys -t MLDC:1.0 'go run blockchainnode.go -mode=ST -sc=1 -port=7011' C-m
    # tmux send-keys -t MLDC:1.1 'go run blockchainnode.go -mode=ST -sc=1 -port=7012' C-m
    # tmux send-keys -t MLDC:1.2 'go run blockchainnode.go -mode=ST -sc=1 -port=7013' C-m
    # tmux send-keys -t MLDC:1.3 'go run blockchainnode.go -mode=ST -sc=1 -port=7014' C-m
    # tmux send-keys -t MLDC:1.4 'go run blockchainnode.go -mode=ST -sc=2 -port=7021' C-m
    # tmux send-keys -t MLDC:1.5 'go run blockchainnode.go -mode=ST -sc=2 -port=7022' C-m
    # tmux send-keys -t MLDC:1.6 'go run blockchainnode.go -mode=ST -sc=3 -port=7031' C-m 
    # tmux send-keys -t MLDC:1.7 'cd ../blockchainsim' C-m    
    # tmux send-keys -t MLDC:1.7 'rm bc_sim.db' C-m
    # tmux send-keys -t MLDC:1.7 'rm bc_sim.wallet' C-m
    # tmux send-keys -t MLDC:1.7 'go run blockchainsim.go' C-m

fi
tmux attach -t MLDC
