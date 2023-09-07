# Ethereum JSON-RPC rrom https://ethereum.github.io/execution-apis/api-documentation/
# Ethereum node list from https://ethereumnodes.com/

import requests
import json

START_BLOCK_HASH = "0xd66bb57fdf5582f6163c60517f55ca8dde2445c57de8e21450d75aaae0221c97" 
NUM_BLOCKS = 720
NAME_BFILE = f"blocks_eth_{NUM_BLOCKS}.json"

def dn_ethereum_blocks():
    url = "https://eth.llamarpc.com"
    # url = "https://cloudflare-eth.com/"
    headers = {'content-type': 'application/json'}
 
    # Example echo method
    payload_grb = {
        "jsonrpc": "2.0",
        "method": "debug_getRawBlock",
        "params": [""],
        "id": 0
    }
    payload_gbh = {
    "jsonrpc": "2.0",
    "method": "eth_getBlockByHash",
    "params": ["", True],
    "id": 0
    }
    payload_gbn = {
        "jsonrpc": "2.0",
        "method": "eth_getBlockByNumber",
        "params": ["0x11240e9",True],
        "id": 0
    }
    payload_bn = {
        "jsonrpc": "2.0",
        "method": "eth_blockNumber",
        "params": ["0x045afad9411245341822400d8f5accba192ff062dfe6ea1d9ccc0b666da0b053"],
        "id": 0
    }

    payload_gbh['params'][0] = START_BLOCK_HASH

    block_numbers = []
    block_hashes = []
    for i in range(NUM_BLOCKS):
        # retreive blocks with hash
        response = requests.post(url, data=json.dumps(payload_gbh), headers=headers).json()
        block = response['result']
        print(block['hash'], block['parentHash'], block['number'])
        payload_gbh['params'][0] = block['parentHash']

        # get raw blocks with their block number
        block_numbers.insert(0, block['number'])
        block_hashes.insert(0, block['hash'])
        

    with open(NAME_BFILE, "w") as outfile:        
        for i, (num, hash) in enumerate(zip(block_numbers, block_hashes)):
            # get raw blocks with their block number
            payload_grb['params'][0] = num
            raw_block = requests.post(url, data=json.dumps(payload_grb), headers=headers).json()
            raw_block['id'] = hash
            
            print("Downloading", i, num)
            # json_object = json.dumps(raw_block, indent=4)
            # outfile.write(json_object)      
            json.dump(raw_block, outfile)
            outfile.writelines('\n')

    with open(NAME_BFILE, 'r') as openfile:
        # Reading from json file
        lines = openfile.readlines()
        print("==================================")
        for rec in lines:
            json_object = json.loads(rec)
            print(json_object['id'], len(json_object['result'])/2)
 
def anylysis_btc_eth_blocks():
    import matplotlib.pyplot as plt
    import pandas as pd

    NUM_BLOCK = 720
    BTC_BLOCKS = f'ppos_btc_{NUM_BLOCK}.csv'
    ETH_BLOCKS = f'ppos_eth_{NUM_BLOCK}.csv'

    fig = plt.figure(figsize=(10, 5))

    dfb = pd.read_csv(BTC_BLOCKS, header=None)
    dfb.columns = ['Encoding', 'Decoding', 'Size']

    plt.scatter(dfb['Size']/1000, dfb['Encoding'], c='#1f77b4', marker='*')
    plt.scatter(dfb['Size']/1000, dfb['Decoding'], c='#1f77b4', marker='D')

    dfe = pd.read_csv(ETH_BLOCKS, header=None)
    dfe.columns = ['Encoding', 'Decoding', 'Size']

    plt.scatter(dfe['Size']/1000, dfe['Encoding'], c='#ff7f0e', marker='*')
    plt.scatter(dfe['Size']/1000, dfe['Decoding'], c='#ff7f0e', marker='D')

    plt.yscale("log")

    print('Encording', dfb['Encoding'].mean(), dfe['Encoding'].mean())
    print('Decoding', dfb['Decoding'].mean(), dfe['Decoding'].mean())
    print('Size', dfb['Size'].mean(), dfe['Size'].mean())
    
    plt.savefig('CompareBctEth.png')
    # plt.show()

def anylysis_btc_eth_blocks_2():
    import matplotlib.pyplot as plt
    import pandas as pd

    NUM_BLOCK = 720
    BTC_BLOCKS = f'ppos_btc_{NUM_BLOCK}.csv'
    ETH_BLOCKS = f'ppos_eth_{NUM_BLOCK}.csv'

    fig = plt.figure(figsize=(10, 5))
    
    dfb = pd.read_csv(BTC_BLOCKS, header=None)
    dfb.columns = ['Encoding', 'Decoding', 'Size']

    plt.scatter(dfb['Encoding'], dfb['Decoding'], s=dfb['Size']/10000, alpha=0.5)

    dfe = pd.read_csv(ETH_BLOCKS, header=None)
    dfe.columns = ['Encoding', 'Decoding', 'Size']

    plt.scatter(dfe['Encoding'], dfe['Decoding'], s=dfe['Size']/10000, alpha=0.5)

    plt.yscale("log")
    plt.xscale("log")

    plt.savefig('CompareBctEth_2.png')
    # plt.show()

if __name__ == "__main__":
    # dn_ethereum_blocks()
    anylysis_btc_eth_blocks()
    anylysis_btc_eth_blocks_2()

