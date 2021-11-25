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

# ns = [1, 2, 4, 8, 16]
# for i in ns:
#     print(i, pa(0.5, i)*100)

ns = [8, 4, 2, 1]
luni = [60.623815,  65.658131,  85.029631,  100.0]
luni = [1. - n / 100. for n in luni]

for n, l in zip(ns, luni):
    print(n, l, pa(l, n) * 100)

lexpo = [81.26491,  89.440182,  96.037027,  100.0]
lexpo = [1. - n / 100. for n in lexpo]

for n, l in zip(ns, lexpo):
    print(n, l, pa(l, n) * 100)
    