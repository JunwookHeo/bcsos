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

def get_bitcoin_transactions(name):
    # df = pd.DataFrame()

    preid = 0
    limit = 1
    offset = 11
    for i in range(offset):
        api_url = f"https://api.blockchair.com/{name}/transactions?q=time(2022-01)&s=id(asc)&limit={limit}&offset={i}"
        response = requests.get(api_url)
        res = response.json()
        print('============Bitcoin===========')
        # print('transaction hash : {}'.format(res['data'][0]['id']))
        print('transaction hash : {}'.format(res['data'][-1]['id']))
        preid = int(res['data'][-1]['id'])

        # print(res)
        for d in res['data']:
            api_url = f"https://api.blockchair.com/{name}/raw/transaction/{d['hash']}"
            response = requests.get(api_url)
            transaction = response.json()
            print(transaction['data'][d['hash']]['decoded_raw_transaction'])

def get_bitcoin_blocks(name):
    # df = pd.DataFrame()

    preid = 0
    limit = 1
    offset = 2
    for i in range(offset):
        api_url = f"https://api.blockchair.com/{name}/blocks?q=time(2022-01)&s=id(asc)&limit={limit}&offset={i}"
        response = requests.get(api_url)
        res = response.json()
        print('============Bitcoin===========')
        # print('transaction hash : {}'.format(res['data'][0]['id']))
        print('transaction hash : {}'.format(res['data'][-1]['id']))
        preid = int(res['data'][-1]['id'])

        # print(res)
        for d in res['data']:
            api_url = f"https://api.blockchair.com/{name}/raw/block/{d['hash']}"
            response = requests.get(api_url)
            block = response.json()
            # print(block)

def get_transaction_from_url():
    import wget
    url = "https://gz.blockchair.com/bitcoin/transactions/"
    file = "blockchair_bitcoin_transactions_20221009.tsv.gz"
    response = wget.download(f"{url}{file}", file)

def get_bitcoin_transactions2():
    import csv
    import json
    path = "./blockchainsim/iotdata/blockchair_bitcoin_transactions_20221009.tsv"
    outfile = 'transactions.json'
    with open(outfile, "w") as fo:
        with open(path, newline='') as fi:
            header = next(fi)
            print(header)
            lines = csv.reader(fi, delimiter='\t')
            for i, rec in enumerate(lines):            
                api_url = f'https://blockchain.info/rawtx/{rec[1]}'
                response = requests.get(api_url)
                res = response.json()
                json.dump(res, fo)
                fo.writelines('\n')
                print(res)
                if i > 0:
                    break
    with open(outfile, 'r') as openfile:
        # Reading from json file
        lines = openfile.readlines()
        print("==================================")
        for rec in lines:
            json_object = json.loads(rec)
            print(json_object)

def get_blocks_from_height():
    import csv
    import json
    outfile = 'blocks.json'
    height = 756450
    with open(outfile, "w") as fo:
        for h in range(height, height + 2):
            api_url = f'https://api.blockchair.com/bitcoin/raw/block/{h}'
            response = requests.get(api_url)
            res = response.json()
            json.dump(res, fo)
            fo.writelines('\n')
            print(res)

    with open(outfile, 'r') as openfile:
        # Reading from json file
        lines = openfile.readlines()
        print("==================================")
        for rec in lines:
            json_object = json.loads(rec)
            print(json_object)

# df = blockchain_stats('bitcoin')
# show_plot(df, 'BitcoinState')

# df = blockchain_stats('ethereum')
# show_plot(df, 'EthereumState')

# df = blockchain_stats('litecoin')
# show_plot(df, 'LitecoinState')

# df = blockchain_stats('zcash')
# show_plot(df, 'ZcashState')


# get_bitcoin_transactions('bitcoin')
# get_bitcoin_blocks('bitcoin')
# get_transaction_from_url()


# api_url = f"https://api.blockchair.com/bitcoin/raw/transaction/00d42e0fa72b7a3742e27dcc961c9ff265d297d0a8ca6aff62896733d14f6672"
# response = requests.get(api_url)
# res = response.json()
# print(res)

get_blocks_from_height()