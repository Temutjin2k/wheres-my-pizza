package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type orderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *orderRepository {
	return &orderRepository{
		pool: pool,
	}
}

func (r *orderRepository) Create(ctx context.Context, req *models.CreateOrder) (*models.Order, error) {
	var order models.Order

	// Start a transaction
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert the order
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (
			number, 
			customer_name, 
			type, 
			table_number, 
			delivery_address, 
			total_amount, 
			priority, 
			status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING 
			id, created_at, updated_at, number, customer_name, 
			type, table_number, delivery_address, total_amount, 
			priority, status, processed_by, completed_at`,
		req.Number,
		req.CustomerName,
		req.Type,
		req.TableNumber,
		req.DeliveryAddress,
		req.TotalAmount,
		req.Priority,
		req.Status,
	).Scan(
		&order.ID,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.Number,
		&order.CustomerName,
		&order.Type,
		&order.TableNumber,
		&order.DeliveryAddress,
		&order.TotalAmount,
		&order.Priority,
		&order.Status,
		&order.ProcessedBy,
		&order.CompletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Insert order items
	for _, item := range req.Items {
		_, err := tx.Exec(ctx,
			`INSERT INTO order_items (
				order_id, 
				name, 
				quantity, 
				price
			) VALUES ($1, $2, $3, $4)`,
			order.ID,
			item.Name,
			item.Quantity,
			item.Price,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &order, nil
}

func (r *orderRepository) GetAndIncrementSequence(ctx context.Context, date string) (int, error) {
	var seq int

	// Configure transaction with serializable isolation level
	txOptions := pgx.TxOptions{
		IsoLevel:       pgx.Serializable,  // Highest isolation level
		AccessMode:     pgx.ReadWrite,     // Default, but explicit is better
		DeferrableMode: pgx.NotDeferrable, // Not deferrable for this case
	}

	// Start a transaction
	tx, err := r.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Use the new pgx syntax for querying
	err = tx.QueryRow(ctx,
		`INSERT INTO order_sequences (date, last_value) 
		VALUES ($1, 1)
		ON CONFLICT (date) DO UPDATE 
		SET last_value = order_sequences.last_value + 1,
		    updated_at = $2
		RETURNING last_value`,
		date, time.Now().UTC(),
	).Scan(&seq)

	if err != nil {
		return 0, fmt.Errorf("failed to get/increment sequence: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return seq, nil
}
