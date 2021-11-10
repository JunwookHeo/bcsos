import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns


x = np.random.exponential(8, 1000)
#plt.hist(x)

sns.histplot(x, stat='probability')
plt.show()
#plt.savefig("exponential4.png") 