package repository

import (
	"ap2/payment-service/internal/domain"
	"database/sql"
	"fmt"
)

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Save(p *domain.Payment) error {
	query := `INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status, p.CreatedAt)
	if err != nil {
		return fmt.Errorf("save payment: %w", err)
	}
	return nil
}

func (r *PostgresPaymentRepository) FindByOrderID(orderID string) (*domain.Payment, error) {
	query := `SELECT id, order_id, transaction_id, amount, status, created_at FROM payments WHERE order_id = $1`
	row := r.db.QueryRow(query, orderID)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find payment: %w", err)
	}
	return &p, nil
}
