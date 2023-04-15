import time
import json

# pip install websocket_client
from websocket import create_connection


def connect():
    ws = create_connection("wss://wsaws.okx.com:8443/ws/v5/public")
    ws.send(
        json.dumps(
            {"op": "subscribe", "args": [{"channel": "bbo-tbt", "instId": "BTC-USDT"}]}
        )
    )
    data = ws.recv()
    data = json.loads(data)
    if data["event"] == "subscribe":
        print("Subscribed")
    elif data["event"] == "error":
        print("Error:", data["msg"])
    return ws


def measure_period(ws, updates=100):
    start_time = time.perf_counter()
    for _ in range(updates):
        ws.recv()
    print((time.perf_counter() - start_time) / updates * 1000, "ms")


def get_bid_ask(ws):
    while True:
        data = ws.recv()
        data = json.loads(data)["data"][0]
        bid = data["asks"][0][0]
        ask = data["bids"][0][0]
        ts = data["ts"]
        print(f"Bid: {bid}\tAsk: {ask}\tTimestamp: {ts}")


def main():
    ws = connect()
    try:
        # measure_period(ws)
        get_bid_ask(ws)
    finally:
        ws.close()


if __name__ == "__main__":
    main()
