import matplotlib
import matplotlib.pyplot as plt

# matplotlib.use( 'WebAgg' )

# Libraries
import pandas as pd
from math import pi

def getReplications():
    group = ['IPFS-based','CUB','Jidar','MLDC']
    # colors = ['r', 'g', 'b', 'y']
    # Set data
    df = pd.DataFrame({
    'group': group,
    'C': [2, 1, 1, 2],
    'I': [1, 1, 1, 1],
    'A': [0.3, 0.3, 0.3, 0.3],
    'DM': [1, 0.3, 1, 1],
    'DD': [0.3, 0.3, 0.3, 0.3],
    'NE': [0.3, 0.3, 0.3, 0.3],
    'CE': [1, 1, 0.3, 1],
    'SE': [2, 2, 2, 2],
    })

    print(df)
    return 'Replication', group, df

def getRedactions():
    group = ['Chameleon hash-based','Polynomial-based', 'RSA-based', 'MOF-BC',r'$\mu$chain','LiTiChain']
    # colors = ['r', 'g', 'b', 'y']
    # Set data
    df = pd.DataFrame({
    'group': group,
    'C': [0.3, 0.3, 0.3, 0.3, 1, 1],
    'I': [0.3, 0.3, 0.3, 0.3, 0.3, 0.3],
    'A': [1, 1, 1, 1, 1, 1],
    'DM': [0.3, 0.3, 0.3, 0.3, 1, 1],
    'DD': [1, 1, 1, 1, 1, 1],
    'NE': [0.3, 1, 1, 1, 1, 1],
    'CE': [0.3, 0.3, 1, 1, 0.3, 1],
    'SE': [2, 2, 2, 2, 1, 2],
    })

    print(df)
    return 'Redaction', group, df

def getContents():
    group = ['MOF-BC','CoinPrune','Mimblewimble','Mina']
    # colors = ['r', 'g', 'b', 'y']
    # Set data
    df = pd.DataFrame({
    'group': group,
    'C': [0.3, 1, 2, 2],
    'I': [0.3, 1, 2, 2],
    'A': [1, 0.3, 1, 1],
    'DM': [0.3, 1, 1, 1],
    'DD': [1, 0.3, 1, 1],
    'NE': [2, 2, 0.3, 0.3],
    'CE': [0.3, 0.3, 0.3, 0.3],
    'SE': [2, 2, 2, 2],
    })

    print(df)
    return 'Content', group, df

# ------- PART 1: Create background
 
# title, group, colors, df = getReplications()
# title, group, colors, df = getRedactions()
# title, group, colors, df = getContents()

foptimisations = [getReplications, getRedactions, getContents]

plt.subplots(1, 3, figsize=(15, 8))
# plt.subplots_adjust(wspace=0.5, hspace=1)

for i, foptimise in enumerate(foptimisations):
    colors = ['m', 'b', 'g', 'r', 'y', 'c']

    title, group, df = foptimise()

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
    ax.set_rlabel_position(150)
    plt.yticks([0,1,2], ["Negative", "Neutral", "Positive"], color="grey", size=7)
    plt.ylim(0,2)
    

    # ------- PART 2: Add plots
    
    # Plot each individual = each line of the data
    # I don't make a loop, because plotting more than 3 groups makes the chart unreadable
    alpha = 0.0
    linestyles = ["solid", "dashed", "dashdot", "dotted", "dashed", "dashdot"]
    markers=['+','v','*','o', '<', '>']

    for i, g in enumerate (group):
        values=df.loc[i].drop('group').values.flatten().tolist()
        values += values[:1]
        values = [x-i*0.015 for x in values]
        ax.plot(angles, values, linewidth=2, linestyle=linestyles[i%len(group)], label=group[i], color=colors[i%len(group)], marker=markers[i%len(group)])
        # ax.fill(angles, values, colors[i], alpha=alpha)
        
    # Add legend
    plt.legend(loc='right', bbox_to_anchor=(0.9, -0.3), fontsize=13)
    plt.title(title+"-based", fontsize=16, pad=10)
    plt.tick_params(labelsize=14) 


# Show the graph
plt.tight_layout()
plt.savefig("Performance.png")
plt.show()