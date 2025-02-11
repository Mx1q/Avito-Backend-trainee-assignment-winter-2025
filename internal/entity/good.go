package entity

type Good struct {
	Name     string
	Price    int32
	Quantity int32
}

type IGoodRepository interface {
	GetInventory(username string) ([]Good, error)
}

type IGoodService interface {
	GetInventory(username string) ([]Good, error)
}
