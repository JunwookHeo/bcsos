## https://stackoverflow.com/questions/35821184/implement-an-interactive-shell-over-ssh-in-python-using-paramiko

import paramiko
from scp import SCPClient
import time, os
from os import system
import argparse

USER = 'mldc'
PASSWD = 'mldc'
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

def putBinary(nodes):
    for node in nodes:
        print("Put out/client %s"%(node.toString()))
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        ssh.connect(node.url, 22, USER, PASSWD, timeout=5)

        with SCPClient(ssh.get_transport()) as scp:
            scp.put('out/client', recursive=True)

        ssh.close()

def getResult(path, nodes):
    for node in nodes:
        print("Get result from %s"%(node.toString()))
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        try:
            ssh.connect(node.url, 22, USER, PASSWD, timeout=5)
        except Exception as e:
            print("%s connection error : "%(node.url), e)
            ssh.close()
            continue

        with SCPClient(ssh.get_transport(), sanitize=lambda x: x) as scp:
            nodepath= "%s/%s"%(path, node.url)    
            os.mkdir(nodepath)
            scp.get(remote_path='~/client/db_nodes/*.*', local_path=nodepath, recursive=True)

        ssh.close()

def tmux(command):
    system('tmux %s' % command)

def tmux_shell(pane, command):
    tmux('send-keys -t %d "%s" "C-m"' % (pane, command))

def runSim(pane, node):
    tmux_shell(pane, 'sshpass -p %s ssh -o StrictHostKeyChecking=no mldc@%s' %(PASSWD, node.url))
    tmux_shell(pane, 'cd client')
    tmux_shell(pane, 'rm -rf db_nodes')
    tmux_shell(pane, './blockchainnode %s %s'%(node.sc, node.port))

### Configuration of GUI with tmux
### Split window and connect to each RPI node with ssh
### Then run client
def connectNodes(nodes):    
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

parser = argparse.ArgumentParser()
parser.add_argument('--get_result', type=str, default="yes", help='Get result data')
args = parser.parse_args()
print(args.get_result)

if args.get_result.lower() == 'yes':
    ### Get nodes from rpi.nodes
    nodes = getNodes()

    outputname = time.strftime("MLDC_%Y%m%d_%H%M%S")
    os.mkdir(outputname)
    getResult(outputname, nodes)

    print(outputname)
elif args.get_result.lower() == 'no':
    ### Get nodes from rpi.nodes
    nodes = getNodes()
    ### push output file to nodes
    putBinary(nodes)

    ### Run client
    connectNodes(nodes)

