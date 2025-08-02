package handler

type CreateOrderRequest struct {
	CustomerName    string      `json:"customer_name"`
	OrderType       string      `json:"order_type"`
	Items           []OrderItem `json:"items"`
	TableNumber     *int        `json:"table_number,omitempty"`     // Only for dine_in
	DeliveryAddress *string     `json:"delivery_address,omitempty"` // Only for delivery
}

type OrderItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type CreateOrderResponse struct {
	OrderNumber string  `json:"order_number"`
	Status      string  `json:"status"`
	TotalAmount float64 `json:"total_amount"`
}
