package repository

import (
	"context"
	"EmployeeMerchStore/internal/models"
)

type LedgerRepositoryInterface interface {
	SendMoney(ctx context.Context, fromUser, toUser string, amount int) error
	GetUserTransactions(ctx context.Context, userID string, limit, offset int) (*[]models.Ledger, error)
}

type UserRepositoryInterface interface {
	GetUserCredentials(ctx context.Context, username string) (string, string, error)
	GetBalance(ctx context.Context, id string) (int, error)
	CreateUser(ctx context.Context, id, username, hashPswd string, balance int) error
}

type PurchasesRepositoryInterface interface {
	BuyMerch(ctx context.Context, userID string, merchID int, quantity, price int) error
	GetMerchId(ctx context.Context, name string) (int, int, error)
	GetUserMerch(ctx context.Context, userID string) ([]*models.UserMerch, error)
}

type MerchRepositoryInterface interface {
	GetMerch(ctx context.Context, id int) (models.Merch, error)
	CreateMerch(ctx context.Context, name string, price int, description string) (int, error)
	UpdateMerch(ctx context.Context, id int, name string, price int, description string) error
	DeleteMerch(ctx context.Context, id int) error
}