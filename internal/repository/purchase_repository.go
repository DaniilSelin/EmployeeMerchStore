package repository

import (
    "context"
    "fmt"
    "time"
    "github.com/jackc/pgx/v4/pgxpool"
)

type PurchasesRepository struct {
    db *pgxpool.Pool
}

func NewPurchasesRepository(db *pgxpool.Pool) *PurchasesRepository {
    return &PurchasesRepository{db: db}
}

func (pr *PurchasesRepository) BuyMerch(ctx context.Context, userID string, merchID int, quantity int) error {
    query := `
        INSERT INTO "MerchStore".purchases (user_id, merch_id, quantity, purchased_at)
        VALUES ($1, $2, $3, now())
        ON CONFLICT (user_id, merch_id)
        DO UPDATE SET 
            quantity = "MerchStore".purchases.quantity + EXCLUDED.quantity,
            purchased_at = now();
        `

    _, err := pr.db.Exec(ctx, query, userID, merchID, quantity)
    if err != nil {
        return fmt.Errorf("BuyMerch: %w", err)
    }
    return nil
}

func (pr *PurchasesRepository) GetMerchId(ctx context.Context, name string) (int, error) {
    query := `SELECT id FROM "MerchStore".merch WHERE name = $1 LIMIT 1`

    var merchID int
    
    err := pr.db.QueryRow(ctx, query, name).Scan(&merchID)
    if err != nil {
        if err.Error() == "no rows in result set" {
            return 0, fmt.Errorf("Merch with name %s not found", name)
        }
        return 0, fmt.Errorf("GetMerchId: %w", err)
    }
    return merchID, nil
}