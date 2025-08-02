package models

type Order struct {
	Field string
}

type CreateOrderRequest struct {
	CustomerName string
	OrderType    string
	Items        []OrderItem
}

type OrderItem struct {
	Name     string
	Quantity int
	Price    float64
}

type OrderCreatedInfo struct {
	OrderNumber string
	Status      string
	TotalAmount float64
}
