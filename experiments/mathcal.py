import math

def pa(p, n):
    def sump(p, n):
        s = 0
        for i in range(n):
            s += pow(p, i)/math.factorial(i)
        #print('sump : ', s)
        return s
    P_A = math.exp(-p*n) * sump(p*n, n)  
    return P_A

ns = [1, 2, 3, 10, 16]
for i in ns:
    print(i, pa(0.1, i)*100)
