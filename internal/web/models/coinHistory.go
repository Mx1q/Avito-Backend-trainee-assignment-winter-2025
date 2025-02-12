package models

import "Avito-Backend-trainee-assignment-winter-2025/internal/entity"

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

func ToCoinSentTransferTransport(sentTransfer *entity.User) *CoinSentTransfer {
	return &CoinSentTransfer{
		ToUser: sentTransfer.Username,
		Amount: sentTransfer.Coins,
	}
}

func ToCoinReceivedTransferTransport(sentTransfer *entity.User) *CoinReceivedTransfer {
	return &CoinReceivedTransfer{
		FromUser: sentTransfer.Username,
		Amount:   sentTransfer.Coins,
	}
}

func ToCoinsHistoryTransport(history *entity.CoinsHistory) *CoinHistory {
	coinHistory := new(CoinHistory)

	coinHistory.Received = make([]*CoinReceivedTransfer, len(history.Received))
	for i := 0; i < len(history.Received); i++ {
		coinHistory.Received[i] = ToCoinReceivedTransferTransport(history.Received[i])
	}

	coinHistory.Sent = make([]*CoinSentTransfer, len(history.Sent))
	for i := 0; i < len(history.Sent); i++ {
		coinHistory.Sent[i] = ToCoinSentTransferTransport(history.Sent[i])
	}

	return coinHistory
}
