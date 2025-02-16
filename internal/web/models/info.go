package models

type InfoResponse struct {
	Coins       int32        `json:"coins"`
	Inventory   []*Item      `json:"inventory"`
	CoinHistory *CoinHistory `json:"coinHistory"`
}
