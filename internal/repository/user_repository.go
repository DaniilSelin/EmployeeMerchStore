package repository

import ()

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetPswdHash(ctx context, id string) (string, error) {
	query := `SELECT password FROM "MerchStore".users WHERE id = $1`
	
	var hashPswd string
	
	if err := ur.db.QueryRow(ctx, query, id).Scan(&hashPswd); err != nil {
		return "", fmt.Errorf("Auth: %w", err)
	}
	
	return hashPswd, nil
}

func (ur *UserRepository) GetUserId(ctx context, id string) (string, error) {
	return
}


func (ur *UserRepository) GetBalance(ctx context.Context, id string) (float64, error) {
	query := `SELECT balance FROM "MerchStore".users WHERE id = $1`
	
	var balance float64
	
	if err := ur.db.QueryRow(ctx, query, id).Scan(&balance); err != nil {
		return 0, fmt.Errorf("GetBalance: %w", err)
	}
	
	return balance, nil
}

func (ur *UserRepository) CreateUser(ctx context.Context, id, username, hashPswd string, balance float64) (error) {
	query := `INSERT INTO "MerchStore".users (id, username, password, balance) VALUES ($1, $2, $3, $4)`
	if _, err := ur.db.Exec(ctx, query, id, username, hashPswd, balance); err != nil {
		return "", fmt.Errorf("CreateUser: %w", err)
	}
	return nil
}

func (ur *UserRepository) UpdatePswdHash(ctx context.Context, id, newHashPswd string) error {
	query := `UPDATE "MerchStore".users SET password = $1 WHERE id = $2`
	ct, err := ur.db.Exec(ctx, query, newHashPswd, id)
	if err != nil {
		return fmt.Errorf("UpdatePswdHash: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("UpdatePswdHash: no user found with id %s", id)
	}
	return nil
}

func (ur *UserRepository) DeleteUser(ctx context.Context, id string) (error) {
	query := `DELETE FROM "MerchStore".users WHERE id = $1`
	ct, err := ur.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", id, err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no user found with id %s", id)
	}
	return nil
}