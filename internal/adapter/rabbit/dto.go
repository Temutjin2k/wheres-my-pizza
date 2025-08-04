package rabbit

import "github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"

// Order represents the structure of an order to be published
type Order struct {
	OrderNumber     string         `json:"order_number"`
	CustomerName    string         `json:"customer_name"`
	OrderType       string         `json:"order_type"`
	TableNumber     *int           `json:"table_number,omitempty"` // pointer to allow null in JSON
	DeliveryAddress *string        `json:"delivery_address,omitempty"`
	Items           []OrderItem    `json:"items"`
	TotalAmount     float64        `json:"total_amount"`
	Priority        int            `json:"priority"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// OrderItem represents an item in the order
type OrderItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func FromInternalToPublishOrder(m *models.CreateOrder) *Order {
	if m == nil {
		return nil
	}

	// Convert order items
	publishItems := make([]OrderItem, 0, len(m.Items))
	for _, item := range m.Items {
		publishItems = append(publishItems, OrderItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	return &Order{
		OrderNumber:     m.Number,
		CustomerName:    m.CustomerName,
		OrderType:       m.Type,
		TableNumber:     m.TableNumber,
		DeliveryAddress: m.DeliveryAddress,
		Items:           publishItems,
		TotalAmount:     m.TotalAmount,
		Priority:        m.Priority,
		Metadata:        nil,
	}
}
