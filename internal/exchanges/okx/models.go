package okx

type LoginRequest struct {
	Op   string `json:"op"`
	Args []struct {
		APIKey     string `json:"apiKey"`
		Passphrase string `json:"passphrase"`
		Timestamp  string `json:"timestamp"`
		Sign       string `json:"sign"`
	} `json:"args"`
}

type LoginResponse struct {
	Event string `json:"event"`
	Code  string `json:"code"`
	Msg   string `json:"msg"`
	Data  []struct {
		APIKey string `json:"apiKey"`
	} `json:"data"`
}

type SubscribeRequest struct {
	Op   string             `json:"op"`
	Args []SubscriptionArgs `json:"args"`
}

type SubscriptionArgs struct {
	Channel string `json:"channel"`
	InstID  string `json:"instId"`
}

type BookSnapshotResponse struct {
	Arg    SubscriptionArgs   `json:"arg"`
	Action string             `json:"action"`
	Data   []BookSnapshotData `json:"data"`
}

type BookSnapshotData struct {
	Asks     [][]string `json:"asks"`
	Bids     [][]string `json:"bids"`
	Ts       string     `json:"ts"`
	Checksum int        `json:"checksum"`
}
