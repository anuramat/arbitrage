package whitebit

import "encoding/json"

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
	// depth update fields
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
	// ping response fields
	ID     int64     `json:"id"`
	Result string    `json:"result"`
	Error  *struct{} `json:"error"`
}

type depthUpdateData struct {
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
}
