# arbitrage

specifying currency pairs for exchanges:

```toml
[gate]
currencyPairs = ["BTC_USDT"]
[okx]
currencyPairs = ["BTC_USDT"]
[whitebit]
currencyPairs = ["BTC_USDT"]
```

specifying currency pairs for all exchagnes:

```toml
[all]
currencyPairs = ["BTC_USDT"]
```

exchange specific currency pairs will be ignored, if section "all" is defined
