package okx

import "errors"

var ErrOrderbookDesync = errors.New("orderbook desync detected")

type subscribeRequest struct {
	Op   string            `json:"op"`
	Args []subscriptionArg `json:"args"`
}

type subscriptionArg struct {
	Channel string `json:"channel"`
	InstID  string `json:"instId"`
}

type subscriptionResponse struct {
	Event   string          `json:"event"`
	Channel string          `json:"channel"`
	Arg     subscriptionArg `json:"arg"`
	Code    int             `json:"code,string"`
	Msg     string          `json:"msg"`
}

type bookSnapshotUpdate struct {
	Arg    subscriptionArg    `json:"arg"`
	Action string             `json:"action"`
	Data   []bookSnapshotData `json:"data"`
}

type bookSnapshotData struct {
	Asks     [][4]string `json:"asks"`
	Bids     [][4]string `json:"bids"`
	Ts       int64       `json:"ts,string"`
	Checksum int32       `json:"checksum"`
}
