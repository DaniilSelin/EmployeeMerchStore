package repository

import (
	"fmt"
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetUserCredentials(ctx context.Context, username string) (string, string, error) {
    var id, hash string
    query := `SELECT id, password FROM "MerchStore".users WHERE username = $1`
    err := ur.db.QueryRow(ctx, query, username).Scan(&id, &hash)
    if err != nil {
        return "", "", fmt.Errorf("GetUserCredentials: %w", err)
    }
    return id, hash, nil
}

func (ur *UserRepository) GetBalance(ctx context.Context, id string) (int, error) {
	query := `SELECT balance FROM "MerchStore".users WHERE id = $1`
	
	var balance int
	
	if err := ur.db.QueryRow(ctx, query, id).Scan(&balance); err != nil {
		return 0, fmt.Errorf("GetBalance: %w", err)
	}
	
	return balance, nil
}

func (ur *UserRepository) CreateUser(ctx context.Context, id, username, hashPswd string, balance float64) (error) {
	query := `INSERT INTO "MerchStore".users (id, username, password, balance) VALUES ($1, $2, $3, $4)`
	if _, err := ur.db.Exec(ctx, query, id, username, hashPswd, balance); err != nil {
		return fmt.Errorf("CreateUser: %w", err)
	}
	return nil
}