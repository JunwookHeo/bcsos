import requests
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker


def blockchain_stats(name):
    df = pd.DataFrame()

    api_url = f"https://api.blockchair.com/{name}/blocks?a=year,sum(size),avg(size),avg(transaction_count)"
    response = requests.get(api_url)
    res = response.json()
    print('============Bitcoin===========')
    year = []
    total_size = []
    block_size = []
    num_transaction = []
    sum = 0
    for d in res['data']:
        print(d['year'], ':', d['sum(size)'], d['avg(size)'], d['avg(transaction_count)'])
        year.append(d['year'])
        sum += d['sum(size)']
        total_size.append(sum)
        block_size.append(d['avg(size)'])
        num_transaction.append(d['avg(transaction_count)'])

    df['Year'] = year
    df['Total_Block_Size'] = total_size
    df['Avg_Block_Size'] = block_size
    df['Avg_Num_Transaction'] = num_transaction

    # api_url = "https://api.blockchair.com/bitcoin/transactions?a=year,avg(size)"
    # response = requests.get(api_url)
    # res = response.json()
    # print('============Bitcoin===========')
    # transaction_size = []
    # for d in res['data']:
    #     print(d['year'], ':', d['avg(size)'])
    #     transaction_size.append(d['avg(size)'])
        
    # df['Avg_Transaction_Size'] = transaction_size

    # api_url = "https://api.blockchair.com/bitcoin/stats"
    # response = requests.get(api_url)
    # res = response.json()
    # print('============Bitcoin State===========')
    # print('Num Blocks/Day : ', res['data']['blocks_24h'])
    # print('Num Transactions/Day : ', res['data']['transactions_24h'])

    print(df)
    return df

def show_plot(df, title):
    fig = plt.figure(figsize=(12, 4))
    
    ax1 = plt.subplot(1,3,1)
    plt.plot(df.Year, df.Total_Block_Size, "g.-", label="Total Block Size")
    plt.xticks(rotation=90)
    plt.xlabel("Year")
    plt.ylabel("Total Block Size[GByte]")
    ax1.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: format(int(x/1000000000), ',')))

    ax2 = plt.subplot(1,3,2)
    plt.plot(df.Year, df.Avg_Block_Size, "b.-", label="Avg Block Size") 
    plt.xticks(rotation=90)
    plt.xlabel("Year")
    plt.ylabel("Avg Block Size[KByte]")
    ax2.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: format(int(x/1000), ',')))

    ax3 = plt.subplot(1,3,3)
    plt.plot(df.Year, df.Avg_Num_Transaction, "r.-", label="Avg Number of Transactions")
    plt.xticks(rotation=90)
    plt.xlabel("Year")
    plt.ylabel("Avg # of Transactions per Block")
    ax3.get_yaxis().set_major_formatter(ticker.FuncFormatter(lambda x, p: format(int(x), ',')))

    plt.tight_layout()
    plt.savefig(title) 
    # plt.show()

# df = blockchain_stats('bitcoin')
# show_plot(df, 'BitcoinState')

# df = blockchain_stats('ethereum')
# show_plot(df, 'EthereumState')

# df = blockchain_stats('litecoin')
# show_plot(df, 'LitecoinState')

df = blockchain_stats('zcash')
show_plot(df, 'ZcashState')
