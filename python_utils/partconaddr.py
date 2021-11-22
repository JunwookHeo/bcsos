import numpy as np

MAXNODES = 16

def genSoredNodes(num):
    nodes = np.random.choice(range(MAXNODES), num, replace=False)
    nodes = np.array([0, 1, 2, 3, 6, 7, 12, 15])
    #nodes = np.array([2, 4, 6, 7, 8, 11, 15])
    nodes.sort()
    return nodes

def _getSubLists(nodes, rf):
    subgroups = []
    
    numgroups = 0x1<<rf
    maxsubnodes = MAXNODES>>rf
    mask = (MAXNODES-1)^(maxsubnodes-1)
    
    for g in range(numgroups):
        subnodes = [n for n in nodes if(n&mask == g*maxsubnodes)]
        subnodes = np.array(subnodes)
        subgroups.append(subnodes)
    return subgroups

def _getOrderedNodesFromContentbyScalar(nodes, cid):
    if len(nodes) == 0: return []
    
    print(nodes)
    dist = cid - nodes
    print(dist)
    dist = [d if(d > 0) else -d + 0.5 for d in dist]
    print(dist)
    distidx = np.array(dist).argsort()
    orderedNodes = nodes[distidx[::]]

    return orderedNodes

def _getOrderedNodesFromContentbyXOR(nodes, cid):
    if len(nodes) == 0: return []
    
    dist = nodes^cid
    distidx = np.array(dist).argsort()
    orderedNodes = nodes[distidx[::]]

    return orderedNodes

def contentsAddressingbyGroup(nodes, rf, xor=True):
    contnodes = []
    
    subgroups = _getSubLists(nodes, rf)
        
    for c in range(MAXNODES):
        tmpnodes = []
        for g in subgroups:
            if(xor):
                orderedNodes = _getOrderedNodesFromContentbyXOR(g, c)
            else:
                orderedNodes = _getOrderedNodesFromContentbyScalar(g, c)
            tmpnodes.append(orderedNodes[0:1])

        contnodes.append(tmpnodes)
        
    return contnodes

def contensAddressing(nodes, r, xor=True):
    contnodes = []
            
    for c in range(MAXNODES):
        if(xor):
            orderedNodes = _getOrderedNodesFromContentbyXOR(nodes, c)
        else:
            orderedNodes = _getOrderedNodesFromContentbyScalar(nodes, c)
        contnodes.append(orderedNodes[0:r])
        
    return contnodes

def _getMaxDigitofNID(maxnode):
    cnt = 0
    while(maxnode > 1):
        maxnode >>= 1
        cnt += 1
    
    return cnt

def _printEmptyNode():
    numdigit = _getMaxDigitofNID(MAXNODES)
    for _ in range(numdigit):
        print('_', end='')
    print(' ', end='')

def _printSeperator():
    numdigit = _getMaxDigitofNID(MAXNODES)
    for _ in range(MAXNODES):
        for _ in range(numdigit):
            print('-', end='')
        print('-', end='')
    print()

def printStatus(nodes, contnodes, r):        
    for n in range(MAXNODES):
        if n in nodes:
            print(f'{n:04b}', end=' ')

        else:
            #print('_', end=' ')
            _printEmptyNode()
    print()
    _printSeperator()

    for i, cns in enumerate( contnodes):
        for n in range(MAXNODES):
            if n in cns[:r]:
                print(f'{i:04b}', end=' ')
            else:
                #print('_', end=' ')
                _printEmptyNode()
        print()


nodes = genSoredNodes(8)

contnodes = contensAddressing(nodes, 4)
printStatus(nodes, contnodes, 4)
print()

# contnodes = contentsAddressingbyGroup(nodes, 2)
# printStatus(nodes, contnodes, 4)

