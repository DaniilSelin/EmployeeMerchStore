package repository

import (
	"context"
	"fmt"

    "EmployeeMerchStore/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type LedgerRepository struct {
	db *pgxpool.Pool
}

func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (lr *LedgerRepository) SendMoney(ctx context.Context, fromUser, toUser string, amount int) error {
	tx, err := lr.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction start failed: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Обновляем балансы
	updateQuery := `
        UPDATE "MerchStore".users 
        SET balance = balance + 
            CASE 
                WHEN id = $1 THEN -$3::integer
                WHEN id = $2 THEN $3::integer
            END 
        WHERE id IN ($1, $2)`

	_, err = tx.Exec(ctx, updateQuery, fromUser, toUser, amount)
	if err != nil {
		return fmt.Errorf("balance update failed: %w", err)
	}

	// Логируем списание у отправителя
	_, err = tx.Exec(ctx, `
		INSERT INTO "MerchStore".ledger (user_id, reference_id_usr, movement_type, amount) 
		VALUES ($1, $2, 'transfer_out', $3)`, fromUser, toUser, amount)
	if err != nil {
		return fmt.Errorf("failed to log sender transaction: %w", err)
	}

	// Логируем зачисление у получателя
	_, err = tx.Exec(ctx, `
		INSERT INTO "MerchStore".ledger (user_id, reference_id_usr, movement_type, amount) 
		VALUES ($1, $2, 'transfer_in', $3)`, toUser, fromUser, amount)
	if err != nil {
		return fmt.Errorf("failed to log recipient transaction: %w", err)
	}

	// Фиксируем транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}


func (lr *LedgerRepository) GetUserTransactions(ctx context.Context, userID string, limit, offset int) (*[]models.Ledger, error) {
	query := `
		SELECT id, user_id, movement_type, amount, reference_id, reference_id_usr, created_at
		FROM "MerchStore".ledger
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := lr.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.Ledger
	for rows.Next() {
		var entry models.Ledger
		err := rows.Scan(&entry.ID, &entry.UserID, &entry.MovementType, &entry.Amount, &entry.ReferenceID, &entry.Reference_id_usr, &entry.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &transactions, nil
}

/*
С курсором, так и не потестил

func (lr *LedgerRepository) GetUserTransactions(ctx context.Context, userID string, cursor *time.Time, limit int) ([]LedgerEntry, error) {
	query := `
		SELECT id, user_id, movement_type, amount, reference_id, created_at
		FROM "MerchStore".ledger
		WHERE user_id = $1 AND ($2::TIMESTAMP IS NULL OR created_at < $2)
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := lr.db.Query(ctx, query, userID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer rows.Close()

	var transactions []LedgerEntry
	for rows.Next() {
		var entry LedgerEntry
		err := rows.Scan(&entry.ID, &entry.UserID, &entry.MovementType, &entry.Amount, &entry.ReferenceID, &entry.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return transactions, nil
}

*/