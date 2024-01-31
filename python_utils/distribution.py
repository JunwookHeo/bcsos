import numpy as np
import matplotlib.pyplot as plt
import pandas as pd
import random

def random_select():
    df = pd.read_csv('rs.csv')
    df = df.astype({"Hash": str, "Time": int})
    
    df.hist(column='Time')
    #plt.hist(df['Time'], bins=10)
    plt.show()
    print(df.columns)
    

def exponential_select():
    x = np.random.exponential(8, 1000)
    plt.hist(x)

    plt.show()
    #plt.savefig("exponential4.png") 

def draw_threshold_time(ax):
    T0 = 6.93
    T1 = 13.86
    T2 = 27.73

    trans = ax.get_xaxis_transform()
    ax.axvline(T0, color="black", linestyle=":")
    plt.text(T0 - 1, -0.04, 'T0', verticalalignment='bottom', transform=trans)
    ax.axvline(T1, color="black", linestyle=":")
    plt.text(T1 - 1, -0.04, 'T1', verticalalignment='bottom', transform=trans)
    ax.axvline(T2, color="black", linestyle=":")
    plt.text(T2 - 1, -0.04, 'T2', verticalalignment='bottom', transform=trans)

def exponential_distribution():
    fig = plt.figure(figsize=(10, 5))
    ax = plt.axes()

    x = np.linspace(0, 100, 1000)
    y1 = np.full(1000, 1)
    y2 = np.exp(-0.1*x) * 10
    y3 = (1 - np.exp(-0.1*x)) * 100  

    ax.plot(x, y1, linestyle='-', label="Uniform access pattern")
    ax.plot(x, y2, linestyle='--', label="Exponentially decaying access pattern")
    # ax.plot(x, y3, linestyle=':', label="Exponentially decaying access pattern")
    draw_threshold_time(ax)
    
    ax.set_xticklabels([])
    ax.tick_params(right= False,top= False,left= True, bottom= False)
    
    plt.xlabel("Time", fontsize=14)
    plt.ylabel("Probability[%]", fontsize=14)
    plt.legend()
    plt.xlim(0, 100)
    plt.ylim(-0.2, 10)
    plt.tick_params(labelsize=14)    
    plt.savefig("AccessPatterns.png") 
    plt.show()

def probability_detecting():
    p_t = (31./1023.)
    p_e = 0.0842
    p1 = 1. - p_e
    p2 = 1. - p_t

    c1, c2 = [], []
    t1, t2 = 1, 1

    for i in range(100):
        t1, t2 = t1*p1, t2*p2   
        c1.append(1-t1)
        c2.append(1-t2)
        print("%d : %0.2f, %0.2f"%(i+1, (1-t1)*100, (1-t2)*100))

    print("P", (p1, p2))
    plt.figure(figsize=(10, 5))

    plt.plot(c1)
    plt.plot(c2)
    plt.legend(['Pd1 : {:.4f}'.format(p_e), 'Pd2 : {:.4f}'.format(p_t)], fontsize=16)
    plt.xlabel("Iteration", fontsize=16)
    plt.ylabel("Error Detection Probability", fontsize=16)
        
    plt.savefig('ErrDetectingProbability.png')
    plt.show()

def probability_detecting_sim():
    b_size = ""
    with open('./block_size.txt', 'r') as f:
        b_size = f.read()

    block_size = [eval(s) for s in b_size.split(' ')]
    # print(block_size)
    total = 0
    for s in block_size:
        total += s
    print("Avg size : ", total/len(block_size))

    random.seed()
    K = 31*1024
    d = 0
    p = 0.
    for i, s in enumerate(block_size):
        t = random.randint(0, s)
        if s <= K:
            k = 0
            p += 1.
        else:
            k = random.randint(0, s-K)
            p += float(K/s)
        if t >=k and t < k+K :
            d += 1
            print("Detected %d : %0.4f"%(d, d/(i+1)*100.))
    print('Expected Probability : %0.4f'%(p/len(block_size)*100.))
    plt.boxplot(block_size)
    plt.show()
    
    
# exponential_select()
# exponential_distribution()
probability_detecting()
# probability_detecting_sim()