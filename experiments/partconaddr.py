import numpy as np

MAXNODES = 16

def genSoredNodes(num):
    nodes = np.random.choice(range(MAXNODES), num, replace=False)
    nodes = np.array([0, 2, 5, 6, 8, 9, 11, 13, 15])
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

def _getOrderedNodesFromContent(nodes, cid):
    dist = nodes^cid
    distidx = np.array(dist).argsort()
    orderedNodes = nodes[distidx[::]]

    return orderedNodes

def contentsAddressingbyGroup(nodes, rf):
    contnodes = []
    
    subgroups = _getSubLists(nodes, rf)
        
    for c in range(MAXNODES):
        tmpnodes = []
        for g in subgroups:
            orderedNodes = _getOrderedNodesFromContent(g, c)            
            tmpnodes.append(orderedNodes[0:1])

        contnodes.append(tmpnodes)
        
    return contnodes

def contensAddressing(nodes, r):
    contnodes = []
            
    for c in range(MAXNODES):
        orderedNodes = _getOrderedNodesFromContent(nodes, c)            
        contnodes.append(orderedNodes[0:r])
        
    return contnodes

def printStatus(nodes, contnodes, r):
    for n in range(MAXNODES):
        if n in nodes:
            print(n, end='\t')

        else:
            print('_', end='\t')
    print()
    print()

    for i, cns in enumerate( contnodes):
        for n in range(MAXNODES):
            if n in cns[:r]:
                print(i, end='\t')
            else:
                print('_', end='\t')
        print()


nodes = genSoredNodes(8)

contnodes = contensAddressing(nodes, 5)
printStatus(nodes, contnodes, 5)
print()

contnodes = contentsAddressingbyGroup(nodes, 2)
printStatus(nodes, contnodes, 4)


