package types

const (
	// Info level actions
	ActionServiceStarted    = "service_started"
	ActionServiceStop       = "service_stop"
	ActionDBConnected       = "db_connected"
	ActionRabbitMQConnected = "rabbitmq_connected"

	// Debug level actions
	ActionOrderReceived        = "order_received"
	ActionRequestReceived      = "request_received"
	ActionOrderPublished       = "order_published"
	ActionNotificationReceived = "notification_received"

	// Error level actions
	ActionValidationFailed      = "validation_failed"
	ActionDBTransactionFailed   = "db_transaction_failed"
	ActionDBQueryFailed         = "db_query_failed"
	ActionRabbitMQPublishFailed = "rabbitmq_publish_failed"
)
