package service

import (
	"context"
	"fmt"

	"EmployeeMerchStore/internal/repository"
	"EmployeeMerchStore/internal/models"
)

type LedgerService struct {
    LedgerRepo repository.LedgerRepositoryInterface
    UserRepo   repository.UserRepositoryInterface
}

func NewLedgerService(ledgerRepo repository.LedgerRepositoryInterface, userRepo repository.UserRepositoryInterface) *LedgerService {
    return &LedgerService{
        LedgerRepo: ledgerRepo,
        UserRepo:   userRepo,
    }
}

func (ls *LedgerService) SendMoney(ctx context.Context, fromUser, toUser string, amount int) error {
    if amount <= 0 {
        return fmt.Errorf("amount must be positive")
    }

    toUserID, _, err := ls.UserRepo.GetUserCredentials(ctx, toUser)
    if err != nil {
        return fmt.Errorf("failed to get recipient id for username '%s': %w", toUser, err)
    }

    senderBalance, err := ls.UserRepo.GetBalance(ctx, fromUser)
    if err != nil {
        return fmt.Errorf("failed to get sender balance: %w", err)
    }
    if senderBalance < amount {
        return fmt.Errorf("insufficient balance: available %d, required %d", senderBalance, amount)
    }

    if err := ls.LedgerRepo.SendMoney(ctx, fromUser, toUserID, amount); err != nil {
        return fmt.Errorf("failed to send money: %w", err)
    }

    return nil
}

func (ls *LedgerService) GetUserTransactions(ctx context.Context, id string) ([]*models.Ledger, []*models.Ledger, error) {
    transactionsAll, err := ls.LedgerRepo.GetUserTransactions(ctx, id, 100, 0) // пример: limit 100, offset 0
    if err != nil {
        return nil, nil, fmt.Errorf("failed to get user transactions: %w", err)
    }

    var transactionsIn []*models.Ledger
    var transactionsOut []*models.Ledger

    for _, entry := range *transactionsAll {
        switch entry.MovementType {
        case "transfer_in":
            transactionsIn = append(transactionsIn, &entry) // Append the pointer
        case "transfer_out":
            transactionsOut = append(transactionsOut, &entry) // Append the pointer
        }
    }


    return transactionsIn, transactionsOut, nil
}
