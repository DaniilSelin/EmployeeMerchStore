package repository

import (
	"context"
	"fmt"

    "EmployeeMerchStore/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type MerchRepository struct {
	db *pgxpool.Pool
}

func NewMerchRepository(db *pgxpool.Pool) *MerchRepository {
	return &MerchRepository{db: db}
}

func (mr *MerchRepository) GetMerch(ctx context.Context, id int) (models.Merch, error) {
	query := `SELECT id, name, price, description, created_at FROM "MerchStore".merch WHERE id = $1`
	var merch models.Merch
	if err := mr.db.QueryRow(ctx, query, id).Scan(&merch.ID, &merch.Name, &merch.Price, &merch.Description, &merch.CreatedAt); err != nil {
		return models.Merch{}, fmt.Errorf("GetMerch: %w", err)
	}
	return merch, nil
}

func (mr *MerchRepository) CreateMerch(ctx context.Context, name string, price int, description string) (int, error) {
	query := `INSERT INTO "MerchStore".merch (name, price, description) VALUES ($1, $2, $3) RETURNING id`

    var merchID int 
    err := mr.db.QueryRow(ctx, query, name, price, description).Scan(&merchID)
    if err != nil {
        return 0, fmt.Errorf("CreateMerch: %w", err)
    }

    return merchID, nil
}

func (mr *MerchRepository) UpdateMerch(ctx context.Context, id, name string, price int, description string) error {
	query := `UPDATE "MerchStore".merch SET name = $1, price = $2, description = $3 WHERE id = $4`
	ct, err := mr.db.Exec(ctx, query, name, price, description, id)
	if err != nil {
		return fmt.Errorf("UpdateMerch: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("UpdateMerch: no merch found with id %s", id)
	}
	return nil
}

func (mr *MerchRepository) DeleteMerch(ctx context.Context, id string) error {
	query := `DELETE FROM "MerchStore".merch WHERE id = $1`
	ct, err := mr.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DeleteMerch: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("DeleteMerch: no merch found with id %s", id)
	}
	return nil
}
