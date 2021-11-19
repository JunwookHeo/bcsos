import numpy as np
import matplotlib.pyplot as plt
import pandas as pd

def random_select():
    df = pd.read_csv('rs.csv')
    df = df.astype({"Hash": str, "Time": int})
    
    df.hist(column='Time');
    #plt.hist(df['Time'], bins=10)
    plt.show()
    print(df.columns)
    

def exponential_select():
    x = np.random.exponential(8, 1000)
    plt.hist(x)

    plt.show()
    #plt.savefig("exponential4.png") 
    
exponential_select()