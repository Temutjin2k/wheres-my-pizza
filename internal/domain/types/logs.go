package types

const (
	// Info level actions
	ActionServiceStarted    = "service_started"
	ActionDBConnected       = "db_connected"
	ActionRabbitMQConnected = "rabbitmq_connected"
	ActionWorkerRegistered  = "worker_registered"
	ActionGracefulShutdown  = "graceful_shutdown"

	// Debug level actions
	ActionOrderReceived          = "order_received"
	ActionRequestReceived        = "request_received"
	ActionOrderPublished         = "order_published"
	ActionOrderProcessingStarted = "order_processing_started"
	ActionOrderCompleted         = "order_completed"
	ActionHeartbeatSent          = "heartbeat_sent"
	ActionNotificationReceived   = "notification_received"

	// Error level actions
	ActionValidationFailed         = "validation_failed"
	ActionDBTransactionFailed      = "db_transaction_failed"
	ActionDBQueryFailed            = "db_query_failed"
	ActionRabbitMQPublishFailed    = "rabbitmq_publish_failed"
	ActionWorkerRegistrationFailed = "worker_registration_failed"
	ActionMessageProcessingFailed  = "message_processing_failed"
	ActionDBConnectionFailed       = "db_connection_failed"
	ActionRabbitConnectionFailed   = "rabbitmq_connection_failed"
)
