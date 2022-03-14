# MLDC
This is a simulator for Multi-Level Distributed Caching on the Blockchain for Storage Optimisation.

# Requirements
* go version 1.16 or above
* Linux 64bit/Raspberry 64bit/Windows 10 or above 64bit
* tmux 3.0a

# Run
* run blockchainnode/runsim2.sh
* connect http://localhost:8080/ and click "Start Test" button

## Runing individual nodes
* cd blockchainnode
* go run blockchainnode.go -mode=pan -type=0 -port=7001'
  * type is the storage class, 3 is the highest node
  * mode : dev or pan

## Runing Simulator server
* cd blockchainsim
* go run blockchainsim.go




