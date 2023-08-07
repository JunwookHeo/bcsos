import time
import csv
import json
import requests

def get_blocks_from_height():
    num_block = 5
    outfile = f'blocks_2023_{num_block}.json'
    height = 797000
    
    with open(outfile, "w") as fo:
        for i, h in enumerate (range(height, height + num_block)):
            start = time.time()
            api_url = f'https://api.blockchair.com/bitcoin/raw/block/{h}'
            response = requests.get(api_url)
            res = response.json()
            json.dump(res, fo)
            fo.writelines('\n')
            done = time.time() - start
            print("Downloading {} : {}".format(i, done))

    with open(outfile, 'r') as openfile:
        # Reading from json file
        lines = openfile.readlines()
        print("==================================")
        for rec in lines:
            json_object = json.loads(rec)
            print(json_object)

get_blocks_from_height()