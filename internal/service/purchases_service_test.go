package service

import (
    "context"
    "errors"
    "testing"

    "EmployeeMerchStore/internal/models"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockPurchasesRepo struct {
	mock.Mock
}

func (m *MockPurchasesRepo) GetUserMerch(ctx context.Context, id string) ([]*models.UserMerch, error) {
    args := m.Called(ctx, id)
    
    // когда есть покупки
    if args.Get(0) != nil {
        return args.Get(0).([]*models.UserMerch), args.Error(1)
    }
    
    // когда покупок нет
    return []*models.UserMerch{}, args.Error(1)
}

func (m *MockPurchasesRepo) GetMerchId(ctx context.Context, name string) (int, int, error) {
    args := m.Called(ctx, name)
    return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockPurchasesRepo) BuyMerch(ctx context.Context, userId string, merchId int, quantity, price int) error {
    args := m.Called(ctx, userId, merchId, quantity, price)
    return args.Error(0)
}

func TestGetUserMerch_Success(t *testing.T) {
	mockRepo := new(MockPurchasesRepo)
	mockUserRepo := new(MockUserRepo)
	purchasesService := NewPurchasesService(mockRepo, mockUserRepo)

	expectedMerch := []*models.UserMerch{
		{MerchID: 1, Name: "T-Shirt", Quantity: 1},
		{MerchID: 2, Name: "Mug", Quantity: 2},
	}

	mockRepo.On("GetUserMerch", mock.Anything, "user-id").Return(expectedMerch, nil)

	merch, err := purchasesService.GetUserMerch(context.Background(), "user-id")
	assert.NoError(t, err)
	assert.Equal(t, expectedMerch, merch)

	mockRepo.AssertExpectations(t)
}

func TestGetUserMerch_Error(t *testing.T) {
	mockRepo := new(MockPurchasesRepo)
	mockUserRepo := new(MockUserRepo)
	purchasesService := NewPurchasesService(mockRepo, mockUserRepo)

	mockRepo.On("GetUserMerch", mock.Anything, "user-id").Return(nil, errors.New("DB error"))

	merch, err := purchasesService.GetUserMerch(context.Background(), "user-id")
	assert.Error(t, err)
	assert.Nil(t, merch)

	mockRepo.AssertExpectations(t)
}

func TestBuyMerch_Success(t *testing.T) {
    mockRepo := new(MockPurchasesRepo)
	mockUserRepo := new(MockUserRepo)
    purchasesService := NewPurchasesService(mockRepo, mockUserRepo)

    mockRepo.On("GetMerchId", mock.Anything, "T-Shirt").Return(1, 1, nil).Once()
    mockUserRepo.On("GetBalance", mock.Anything, "user-id").Return(100, nil).Once()
    mockRepo.On("BuyMerch", mock.Anything, "user-id", 1, 1, 1).Return(nil).Once()

    err := purchasesService.BuyMerch(context.Background(), "user-id", "T-Shirt")
    assert.NoError(t, err)

    mockRepo.AssertExpectations(t)
}

func TestBuyMerch_GetMerchId_Error(t *testing.T) {
	mockRepo := new(MockPurchasesRepo)
	mockUserRepo := new(MockUserRepo)
	purchasesService := NewPurchasesService(mockRepo, mockUserRepo)

	mockRepo.On("GetMerchId", mock.Anything, "T-Shirt").Return(0, 0, errors.New("not found"))

	err := purchasesService.BuyMerch(context.Background(), "user-id", "T-Shirt")
	assert.Error(t, err)

	mockRepo.AssertExpectations(t)
}

func TestBuyMerch_BuyMerch_Error(t *testing.T) {
	mockRepo := new(MockPurchasesRepo)
	mockUserRepo := new(MockUserRepo)
	purchasesService := NewPurchasesService(mockRepo, mockUserRepo)

	mockRepo.On("GetMerchId", mock.Anything, "T-Shirt").Return(0, 0, errors.New("not found")).Once()
    mockUserRepo.AssertNotCalled(t, "GetBalance", mock.Anything, "user-id")

	err := purchasesService.BuyMerch(context.Background(), "user-id", "T-Shirt")
	assert.Error(t, err)

	mockRepo.AssertExpectations(t)
}
