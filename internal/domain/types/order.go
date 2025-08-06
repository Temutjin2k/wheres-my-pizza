package types

import "time"

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

	CookingTimeDineIn   time.Duration = time.Second * 8
	CookingTimeTakeOut  time.Duration = time.Second * 10
	CookingTimeDelivery time.Duration = time.Second * 12
)
