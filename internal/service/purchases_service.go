package service

import (
	"context"
	"fmt"

	"EmployeeMerchStore/internal/repository"
	"EmployeeMerchStore/internal/models"
)

type PurchasesService struct {
	PurchasesRepo repository.PurchasesRepositoryInterface
}

func NewPurchasesService(PurchasesRepo repository.PurchasesRepositoryInterface) *PurchasesService {
	return &PurchasesService{
		PurchasesRepo: PurchasesRepo,
	}
}

func (ps *PurchasesService) GetUserMerch(ctx context.Context, id string) ([]*models.UserMerch, error) {
	merchList, err := ps.PurchasesRepo.GetUserMerch(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get balabnce: %w", err)
	}

	return merchList, nil
}

func (ps *PurchasesService) BuyMerch(ctx context.Context, userId, nameMerch string) error {    merchID, err := ps.PurchasesRepo.GetMerchId(ctx, nameMerch)
    if err != nil {
        return fmt.Errorf("failed to get merch id for '%s': %w", nameMerch, err)
    }

    if err := ps.PurchasesRepo.BuyMerch(ctx, userId, merchID, 1); err != nil {
        return fmt.Errorf("failed to buy merch: %w", err)
    }

    return nil
}
