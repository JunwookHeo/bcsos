# Ethereum JSON-RPC rrom https://ethereum.github.io/execution-apis/api-documentation/
# Ethereum node list from https://ethereumnodes.com/

import requests
import json

START_BLOCK_HASH = "0xd66bb57fdf5582f6163c60517f55ca8dde2445c57de8e21450d75aaae0221c97" 
NUM_BLOCKS = 10
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
            print(json_object['id'], len(json_object['result']))
 
if __name__ == "__main__":
    dn_ethereum_blocks()

