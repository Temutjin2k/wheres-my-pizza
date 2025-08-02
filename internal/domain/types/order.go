package types

const (
	StatusOrderReceived  = "received"
	StatusOrderCooking   = "cooking"
	StatusOrderReady     = "ready"
	StatusOrderCompleted = "completed"
	StatusOrderCancelled = "cancelled"
)

// Must be one of: `'dine_in'`, `'takeout'`, or `'delivery'`.

const (
	OrderTypeDineIn   = "dine_in"
	OrderTypeTakeOut  = "takeout"
	OrderTypeDelivery = "delivery"
)
