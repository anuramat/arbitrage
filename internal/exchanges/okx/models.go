package okx

// type loginRequest struct {
// 	Op   string `json:"op"`
// 	Args []struct {
// 		APIKey     string `json:"apiKey"`
// 		Passphrase string `json:"passphrase"`
// 		Timestamp  string `json:"timestamp"`
// 		Sign       string `json:"sign"`
// 	} `json:"args"`
// }

// type loginResponse struct {
// 	Event string `json:"event"`
// 	Code  string `json:"code"`
// 	Msg   string `json:"msg"`
// 	Data  []struct {
// 		APIKey string `json:"apiKey"`
// 	} `json:"data"`
// }

type subscribeRequest struct {
	Op   string             `json:"op"`
	Args []subscriptionArgs `json:"args"`
}

type subscriptionArgs struct {
	Channel string `json:"channel"`
	InstID  string `json:"instId"`
}

type bookSnapshotUpdate struct {
	Arg    subscriptionArgs   `json:"arg"`
	Action string             `json:"action"`
	Data   []bookSnapshotData `json:"data"`
}

type bookSnapshotData struct {
	Asks     [][]string `json:"asks"`
	Bids     [][]string `json:"bids"`
	Ts       string     `json:"ts"`
	Checksum int        `json:"checksum"`
}
