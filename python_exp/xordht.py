import numpy as np

def genNodes():
    nodes = np.random.choice(range(16), 8, replace=False)
    return nodes

def calNeighbours(nodes):
    neighbours = []
    for i, n in enumerate(nodes):
        distance = np.array([])
        for t in nodes:
            distance = np.append(distance, n^t)
            #print(f'{n:04b} xor {t:04b} : {n ^ t}')
        #print(f'distance {n} : {distance}')
        didx = np.array(distance).argsort()
        #print(f'sorted index {n} : {didx}')
        orderedNodes = nodes[didx[::]]
        #print(f'ordered list : {orderedNodes}')
        neighbours.append(orderedNodes)
    return neighbours

def printNeighbours(neighbours):
    print('=======Neighbour list========')
    for n in neighbours:
        print(n)

def checkClosestNodes(nodes, cid, r):
    print(f'CID = {cid}, nodes = {nodes}')
    distance = nodes^cid
    didx = np.array(distance).argsort()
    orderedNodes = nodes[didx[::]]
    print(f'distance : {distance} - ordered nodes : {orderedNodes}')

def checkContentsAddressing(neighbours, cid, r):
    print(f'CID = {cid}, Replication Factor = {r}')
    for n in neighbours:
        distance = n^cid
        didx = np.array(distance).argsort()
        orderedNodes = n[didx[::]]
        print(f'distrance : {distance} - Nodes : {orderedNodes}')
    

nodes = genNodes()
#nodes = np.array([0, 2, 5, 6, 8, 9, 13, 15])
nodes.sort()
print(nodes)
neighbours = calNeighbours(nodes)
printNeighbours(neighbours)
checkClosestNodes(nodes, np.random.randint(16), 3)


