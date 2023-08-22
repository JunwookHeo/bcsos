import matplotlib.pyplot as plt

file_5b = './redact/chain_result_5b.txt'
# file_10b = './redact/chain_result_10b.txt'

with open(file_5b, 'r') as f:
    data_5b = [float(line.split(' ')[-1])*1000 for line in f.read().splitlines()]
    print(data_5b)

# with open(file_10b, 'r') as f:
#     data_10b = [float(line.split(' ')[-1])*1000 for line in f.read().splitlines()]
#     print(data_10b)


plt.figure(figsize=(10, 5))
plt.plot(data_5b, label='5 Blocks')
# plt.plot(data_10b, label='10 Blocks')
plt.xticks(fontsize=16)
plt.yticks(fontsize=16)
# plt.legend()
plt.ylabel('Verification time[msec]' ,fontsize=16)
plt.xlabel('Number of modifications', fontsize=16)
plt.savefig('VerifyTimeChain.png')
