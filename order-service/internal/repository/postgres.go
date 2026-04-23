package repository

import (
	"ap2/order-service/internal/domain"
	"database/sql"
	"fmt"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Save(o *domain.Order) error {
	query := `INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, o.CreatedAt)
	if err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) FindByID(id string) (*domain.Order, error) {
	query := `SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var o domain.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find order: %w", err)
	}
	return &o, nil
}

func (r *PostgresOrderRepository) Update(o *domain.Order) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, o.Status, o.ID)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) FindByStatus(status string) ([]*domain.Order, error) {
	query := `SELECT id, customer_id, item_name, amount, status, created_at 
	          FROM orders WHERE ($1 = '' OR status = $1)`
	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0) // ← всегда пустой массив, не nil
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, &o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return orders, nil
}
