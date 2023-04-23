package gate

// response error codes
const (
	invalidFormat = 1 + iota
	invalidArgs
	serverSideError
)

type genericResponse struct {
	Time    int64  `json:"time"`
	TimeMs  int64  `json:"time_ms"`
	Channel string `json:"channel"`
	Event   string `json:"event"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type subscriptionResponse struct {
	genericResponse
	Error *responseError `json:"error"`
}

type tickerUpdate struct {
	genericResponse
	Result struct {
		TimeMs       int64  `json:"t"`
		UpdateID     int64  `json:"u"`
		CurrencyPair string `json:"s"`
		BidPrice     string `json:"b"`
		AskPrice     string `json:"a"`
		BidAmount    string `json:"B"`
		AskAmount    string `json:"A"`
	} `json:"result"`
}
