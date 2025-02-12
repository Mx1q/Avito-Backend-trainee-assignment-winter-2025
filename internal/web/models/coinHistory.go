package models

type CoinHistory struct {
	Received []*CoinReceivedTransfer `json:"received,omitempty"`
	Sent     []*CoinSentTransfer     `json:"sent,omitempty"`
}

type CoinReceivedTransfer struct {
	FromUser string `json:"fromUser,omitempty"`
	Amount   int32  `json:"amount,omitempty"`
}

type CoinSentTransfer struct {
	ToUser string `json:"toUser,omitempty"`
	Amount int32  `json:"amount,omitempty"`
}
