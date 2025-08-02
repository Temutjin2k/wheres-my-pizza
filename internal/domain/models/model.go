package models

type Order struct {
	Field string
}

type CreateOrderRequest struct {
	CustomerName    string
	OrderType       string
	Items           []OrderItem
	TableNumber     *int    // Only for dine_in
	DeliveryAddress *string // Only for delivery
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
