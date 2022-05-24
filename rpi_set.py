## https://stackoverflow.com/questions/35821184/implement-an-interactive-shell-over-ssh-in-python-using-paramiko

import paramiko
from scp import SCPClient
import time
from os import system

class Node:
    def __init__(self, url, sc, port):
        self.url = url
        self.sc = sc
        self.port = port
    def toString(self):
        return 'url={} sc={} port={}'.format(self.url, self.sc, self.port)

def getNodes():
    nodes = []
    with open('rpi.nodes') as fp:
        for line in fp.readlines():
            try:
                url, sc, port = line.split()
                if not url.startswith('#'):
                    nodes.append(Node(url, sc, port))
            except ValueError as e:
                print(e)
    
    return nodes

def copyBinary(nodes):
    for node in nodes:
        print(node.toString())
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        ssh.connect(node.url, 22, "mldc", "mldc", timeout=5)

        with SCPClient(ssh.get_transport()) as scp:
            scp.put('out/client', recursive=True)

        ssh.close()

def runSim(pane, node):
    tmux_shell(pane, 'sshpass -p mldc ssh mldc@%s' %(node.url))
    tmux_shell(pane, 'cd client')
    tmux_shell(pane, 'rm -rf db_nodes')
    tmux_shell(pane, './blockchainnode %s %s'%(node.sc, node.port))
    

### Get nodes from rpi.nodes
nodes = getNodes()
### Copy output file to nodes
# copyBinary(nodes)

def tmux(command):
    system('tmux %s' % command)

def tmux_shell(pane, command):
    tmux('send-keys -t %d "%s" "C-m"' % (pane, command))

tmux('kill-server')
tmux('new-session -s MLDC -n "RPI" -d')

num_vert = min(7, len(nodes)-1)
for i in range(num_vert):
    tmux('split-window -v')
    tmux('select-layout even-vertical')

num_hori = min(8, len(nodes) - 8)
for i in range(num_hori):     
    tmux('split-window -h -t %d'%(i*2))

for i, node in enumerate(nodes):
    runSim(i, node)    

tmux('attach -t MLDC')

