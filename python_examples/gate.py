import time
import json

# pip install websocket_client
from websocket import create_connection


def connect():
    ws = create_connection("wss://api.gateio.ws/ws/v4/")
    ws.send(
        json.dumps(
            {
                "time": int(time.time()),
                "channel": "spot.book_ticker",
                "event": "subscribe",  # "unsubscribe" for unsubscription
                "payload": ["BTC_USDT"],
            }
        )
    )
    ws.recv()
    return ws


def measure_period(ws, updates=100):
    start_time = time.perf_counter()
    for _ in range(updates):
        ws.recv()
    print((time.perf_counter() - start_time) / updates * 1000, "ms")


def get_bid_ask(ws):
    while True:
        data = ws.recv()
        data = json.loads(data)
        bid = data["result"]["b"]
        ask = data["result"]["a"]
        ts = data["result"]["t"]
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
