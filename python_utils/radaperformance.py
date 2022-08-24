import matplotlib
import matplotlib.pyplot as plt

# matplotlib.use( 'WebAgg' )

# Libraries
import pandas as pd
from math import pi

def getReplications():
    group = ['IPFS-based','CUB','Jidar','MLDC']
    colors = ['g', 'b', 'y', 'r']
    # Set data
    df = pd.DataFrame({
    'group': group,
    'Security': [1.2, 0.3, 0.4, 1.9],
    'Decentralisation': [0.7, 0.3, 1.5, 1.9],
    'NE': [0.3, 1.4, 0.7, 1.7],
    'CE' : [1.8, 1.7, 1.1, 1.8],
    'SE': [1.8, 1.4, 1.6, 1.5],    
    })

    print(df)
    return 'Replication', group, colors, df

def getRedactions():
    group = ['Chameleon hash','MOF-BC','muchain','LiTiChain']
    colors = ['g', 'b', 'y', 'r']
    # Set data
    df = pd.DataFrame({
    'group': group,
    'Security': [1.1, 1.3, 1.2, 0.4],
    'Decentralisation': [0.3, 0.5, 1.8, 1.7],
    'NE': [1.1, 1.8, 1.7, 1.7],
    'CE': [0.2, 1.2, 0.5, 1.1],
    'SE': [1.0, 1.1, 0.2, 1.6],
    })

    print(df)
    return 'Redaction', group, colors, df

def getContents():
    group = ['MOF-BC','CoinPrune','Mimblewimble','Mina']
    colors = ['g', 'b', 'y', 'r']
    # Set data
    df = pd.DataFrame({
    'group': group,
    'Security': [1.3, 1.2, 1.9, 2.0],
    'Decentralisation': [0.5, 0.3, 1.8, 1.8],
    'NE': [1.8, 1.8, 0.4, 0.2],
    'CE': [1.2, 0.2, 0.1, 0.1],
    'SE': [1.1, 0.3, 1.2, 2.0],
    })

    print(df)
    return 'Content', group, colors, df

# ------- PART 1: Create background
 
# title, group, colors, df = getReplications()
# title, group, colors, df = getRedactions()
# title, group, colors, df = getContents()

foptimisations = [getReplications, getRedactions, getContents]

plt.subplots(1, 3, figsize=(15, 5.5))
plt.subplots_adjust(wspace=0.5, hspace=1)

for i, foptimise in enumerate(foptimisations):
    title, group, colors, df = foptimise()

    # number of variable
    categories=list(df)[1:]
    N = len(categories)
    
    # What will be the angle of each axis in the plot? (we divide the plot / number of variable)
    angles = [n / float(N) * 2 * pi for n in range(N)]
    angles += angles[:1]
    
    # Initialise the spider plot
    ax = plt.subplot(1, 3, i+1, polar=True)
    
    # If you want the first axis to be on top:
    ax.set_theta_offset(pi / 2)
    ax.set_theta_direction(-1)
    
    # Draw one axe per variable + add labels
    plt.xticks(angles[:-1], categories)
    
    # Draw ylabels
    ax.set_rlabel_position(0)
    plt.yticks([0,1,2], ["Low","Medium","High"], color="grey", size=7)
    plt.ylim(0,2)
    

    # ------- PART 2: Add plots
    
    # Plot each individual = each line of the data
    # I don't make a loop, because plotting more than 3 groups makes the chart unreadable
    alpha = 0.0

    for i, g in enumerate (group):
        values=df.loc[i].drop('group').values.flatten().tolist()
        values += values[:1]
        ax.plot(angles, values, linewidth=1, linestyle='solid', label=group[i])
        ax.fill(angles, values, colors[i], alpha=alpha)
        
    # Add legend
    plt.legend(loc='right', bbox_to_anchor=(0.9, -0.22), fontsize=13)
    plt.title(title+"-based", fontsize=16, pad=20)
    plt.tick_params(labelsize=14) 


# Show the graph
plt.savefig("Performance.png")
plt.show()