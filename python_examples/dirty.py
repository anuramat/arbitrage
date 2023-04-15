import requests
from decimal import Decimal


def gate_bid_ask():
    host = "https://api.gateio.ws"
    prefix = "/api/v4"
    headers = {"Accept": "application/json", "Content-Type": "application/json"}

    url = "/spot/tickers"
    query_param = "currency_pair=BTC_USDT"
    data = requests.request(
        "GET", host + prefix + url + "?" + query_param, headers=headers
    ).json()[0]

    bid = data["highest_bid"]
    ask = data["lowest_ask"]
    bid, ask = Decimal(bid), Decimal(ask)

    return bid, ask


def okex_bid_ask():
    host = "https://aws.okx.com"
    prefix = "/api/v5"
    headers = {"Accept": "application/json", "Content-Type": "application/json"}

    url = "/market/ticker"
    query_param = "instId=BTC-USDT"
    data = requests.request(
        "GET", host + prefix + url + "?" + query_param, headers=headers
    ).json()["data"][0]

    bid = data["bidPx"]
    ask = data["askPx"]
    bid, ask = Decimal(bid), Decimal(ask)

    return bid, ask


def main():
    while True:
        okex_fee = Decimal("0.001")
        gate_fee = Decimal("0.001")
        gate_bid, gate_ask = gate_bid_ask()
        okex_bid, okex_ask = okex_bid_ask()
        print(
            gate_bid * (1 - gate_fee) - okex_ask * (1 + okex_fee),
            okex_bid * (1 - okex_fee) - gate_ask * (1 + gate_fee),
        )


if __name__ == "__main__":
    main()
