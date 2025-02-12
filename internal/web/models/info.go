package models

type Info struct {
	Coins       int32       `json:"coins,omitempty"`
	Inventory   []*Item     `json:"inventory,omitempty"`
	CoinHistory CoinHistory `json:"coinHistory,omitempty"`
}
