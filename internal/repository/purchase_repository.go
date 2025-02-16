package repository

import (
    "context"
    "fmt"

    "EmployeeMerchStore/internal/models"
    "github.com/jackc/pgx/v4/pgxpool"
)

type PurchasesRepository struct {
    db *pgxpool.Pool
}

func NewPurchasesRepository(db *pgxpool.Pool) *PurchasesRepository {
    return &PurchasesRepository{db: db}
}

func (pr *PurchasesRepository) BuyMerch(ctx context.Context, userID string, merchID int, quantity, price int) error {
    tx, err := pr.db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to start transaction: %w", err)
    }
    defer func() {
        if err != nil {
            tx.Rollback(ctx)
        }
    }()

    // Добавляем покупку
    purchaseQuery := `
        INSERT INTO "MerchStore".purchases (user_id, merch_id, quantity, purchased_at)
        VALUES ($1, $2, $3, now())
        ON CONFLICT (user_id, merch_id)
        DO UPDATE SET 
            quantity = "MerchStore".purchases.quantity + EXCLUDED.quantity,
            purchased_at = now();
    `
    _, err = tx.Exec(ctx, purchaseQuery, userID, merchID, quantity)
    if err != nil {
        return fmt.Errorf("BuyMerch: failed to insert/update purchase: %w", err)
    }

    // Обновляем баланс пользователя
    updateBalanceQuery := `
        UPDATE "MerchStore".users
        SET balance = balance - $1
        WHERE id = $2
    `
    totalCost := price * quantity
    _, err = tx.Exec(ctx, updateBalanceQuery, totalCost, userID)
    if err != nil {
        return fmt.Errorf("BuyMerch: failed to update user balance: %w", err)
    }

    // Записываем в ledger
    ledgerQuery := `
        INSERT INTO "MerchStore".ledger (user_id, movement_type, amount, reference_id, created_at)
        VALUES ($1, 'purchase', $2, $3, now());
    `
    _, err = tx.Exec(ctx, ledgerQuery, userID, price, merchID)
    if err != nil {
        return fmt.Errorf("BuyMerch: failed to insert into ledger: %w", err)
    }

    // Фиксируем транзакцию
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

func (pr *PurchasesRepository) GetMerchId(ctx context.Context, name string) (int, int, error) {
    query := `SELECT id, price FROM "MerchStore".merch WHERE name = $1 LIMIT 1`

    var merchID int
    var price int
    
    err := pr.db.QueryRow(ctx, query, name).Scan(&merchID, &price)
    if err != nil {
        if err.Error() == "no rows in result set" {
            return 0, 0, fmt.Errorf("Merch with name %s not found", name)
        }
        return 0, 0, fmt.Errorf("GetMerchId: %w", err)
    }
    return merchID, price, nil
}

func (pr *PurchasesRepository) GetUserMerch(ctx context.Context, userID string) ([]*models.UserMerch, error) {
    query := `
        SELECT p.merch_id, m.name, m.price, p.quantity, p.purchased_at
        FROM "MerchStore".purchases p
        JOIN "MerchStore".merch m ON p.merch_id = m.id
        WHERE p.user_id = $1
        ORDER BY p.purchased_at DESC;
    `

    rows, err := pr.db.Query(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("GetUserMerch: %w", err)
    }
    defer rows.Close()

    var merchList []*models.UserMerch
    for rows.Next() {
        var um models.UserMerch
        err := rows.Scan(&um.MerchID, &um.Name, &um.Price, &um.Quantity, &um.PurchasedAt)
        if err != nil {
            return nil, fmt.Errorf("GetUserMerch scan: %w", err)
        }
        merchList = append(merchList, &um)
    }

    if rows.Err() != nil {
        return nil, fmt.Errorf("GetUserMerch rows error: %w", rows.Err())
    }

    return merchList, nil
}
