package whitebit

import "encoding/json"

type subscriptionRequest struct {
	ID     int64  `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type subscriptionResponse struct {
	ID     int `json:"id"`
	Result struct {
		Status string `json:"status"`
	} `json:"result"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

type depthUpdate struct {
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

type depthUpdateData struct {
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
}
