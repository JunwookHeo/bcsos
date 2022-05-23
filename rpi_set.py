## https://stackoverflow.com/questions/35821184/implement-an-interactive-shell-over-ssh-in-python-using-paramiko

import paramiko
import time

import subprocess
print(subprocess.check_output("ls -la", shell=True).decode('utf-8'))
exit

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect("node1.local", 22, "mldc", "mldc", timeout=5)

channel = ssh.get_transport().open_session()
channel.get_pty()
channel.invoke_shell()

cmds = ["ls -la", "cd client", "./blockchainnode"]

for cmd in cmds:
    channel.send(cmd + '\n')
    while True:
        if channel.recv_ready():
            output = channel.recv(1024)
            print(output.decode('utf-8'), end='')
        else:
            time.sleep(0.5)
            if not(channel.recv_ready()):
                break

ssh.close()

