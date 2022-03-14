import requests
api_url = "https://api.blockchair.com/bitcoin/blocks?a=year,avg(size)"
response = requests.get(api_url)
res = response.json()
for d in res['data']:
    print(d['year'], ':', d['avg(size)'])
    

api_url = "https://api.blockchair.com/ethereum/blocks?a=year,avg(size)"
response = requests.get(api_url)
res = response.json()
for d in res['data']:
    print(d['year'], ':', d['avg(size)'])