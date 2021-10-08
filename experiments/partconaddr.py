import numpy as np

MAXNODES = 16

def genNodes():
    nodes = np.random.choice(range(MAXNODES), 8, replace=False)
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
    
def contentsAddressing(nodes, r):
    contnodes = []
    for c in range(MAXNODES):
        dist = nodes^c
        distidx = np.array(dist).argsort()
        orderedNodes = nodes[distidx[::]]
        contnodes.append(orderedNodes)
    return contnodes
            
def printStatus(nodes, contnodes):
    for n in range(MAXNODES):
        if n in nodes:
            print(n, end='\t')

        else:
            print('_', end='\t')
    print()
    print()

    for i, cns in enumerate( contnodes):
        for n in range(MAXNODES):
            if n in cns[:4]:
                print(i, end='\t')
            else:
                print('_', end='\t')
        print()


nodes = genNodes()
nodes.sort()
contnodes = contentsAddressing(nodes, 2)
printStatus(nodes, contnodes)


