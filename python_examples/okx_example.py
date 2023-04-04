import requests

host = "https://www.okx.com"
prefix = "/api/v5"
headers = {'Accept': 'application/json', 'Content-Type': 'application/json'}

url = '/market/tickers'
query_param = 'instType=SPOT'
r = requests.request('GET', host + prefix + url + '?' + query_param, headers=headers)

for pair in r.json()['data']:
    if pair['instId'] == 'BTC-USDT':
        print(pair)
        break

