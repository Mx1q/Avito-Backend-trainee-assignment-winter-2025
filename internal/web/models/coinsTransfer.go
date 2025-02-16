package models

type CoinsTransfer struct {
	ToUser string `json:"toUser,omitempty"`
	Amount int32  `json:"amount,omitempty"`
}

//func ToCoinsTransferEntity(transfer *CoinsTransfer) *entity.Auth {
//	return &entity.Auth{
//		Username: auth.Username,
//		Password: auth.Password,
//	}
//}
