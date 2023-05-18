package whitebit

import (
	"encoding/json"
	"errors"
)

var ErrOrderbookDesync = errors.New("orderbook desync detected")

type request struct {
	ID     int64  `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type subscriptionResponse struct {
	ID     int `json:"id"`
	Result struct {
		Status string `json:"status"`
	} `json:"result"`
	Error *responseError `json:"error"`
}

type responseError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type depthUpdate struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct{}       `json:"error"`
}

type depthUpdateData struct {
	Asks [][2]string `json:"asks"`
	Bids [][2]string `json:"bids"`
}
