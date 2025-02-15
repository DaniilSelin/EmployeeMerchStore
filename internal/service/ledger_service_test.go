package service

import (
    "context"
    "errors"
    "testing"

    "EmployeeMerchStore/internal/models"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockLedgerRepo struct {
	mock.Mock
}

func (m *MockLedgerRepo) SendMoney(ctx context.Context, fromUser, toUser string, amount int) error {
	args := m.Called(ctx, fromUser, toUser, amount)
	return args.Error(0)
}

func (m *MockLedgerRepo) GetUserTransactions(ctx context.Context, id string, limit, offset int) (*[]models.Ledger, error) {
	args := m.Called(ctx, id, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Ledger), args.Error(1)
}

func TestSendMoney_Success(t *testing.T) {
	mockLedgerRepo := new(MockLedgerRepo)
	mockUserRepo := new(MockUserRepo)
	ledgerService := NewLedgerService(mockLedgerRepo, mockUserRepo)

	mockUserRepo.On("GetUserCredentials", mock.Anything, "recipient").Return("user-id-2", "some-pass", nil)
	mockUserRepo.On("GetBalance", mock.Anything, "sender").Return(100, nil)
	mockLedgerRepo.On("SendMoney", mock.Anything, "sender", "user-id-2", 50).Return(nil)

	err := ledgerService.SendMoney(context.Background(), "sender", "recipient", 50)
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}

func TestSendMoney_InsufficientFunds(t *testing.T) {
	mockLedgerRepo := new(MockLedgerRepo)
	mockUserRepo := new(MockUserRepo)
	ledgerService := NewLedgerService(mockLedgerRepo, mockUserRepo)

	mockUserRepo.On("GetBalance", mock.Anything, "sender").Return(1000, nil).Once()   // Получаем баланс отправителя (1000)
	mockUserRepo.On("GetBalance", mock.Anything, "recipient").Return(0, nil).Once()  // Получаем баланс получателя (0)
	mockUserRepo.On("GetUserCredentials", mock.Anything, "sender").Return("senderID", "", nil).Once()  // Получаем данные отправителя
	mockUserRepo.On("GetUserCredentials", mock.Anything, "recipient").Return("recipientID", "", nil).Once()  // Получаем данные получателя

	mockLedgerRepo.On("SendMoney", mock.Anything, "sender", "recipient", 1050).Return(nil).Once()

	err := ledgerService.SendMoney(context.Background(), "sender", "recipient", 1050)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")

	mockUserRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}




func TestSendMoney_InvalidAmount(t *testing.T) {
	mockLedgerRepo := new(MockLedgerRepo)
	mockUserRepo := new(MockUserRepo)
	ledgerService := NewLedgerService(mockLedgerRepo, mockUserRepo)

	err := ledgerService.SendMoney(context.Background(), "sender", "recipient", -10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be positive")
}

func TestSendMoney_RecipientNotFound(t *testing.T) {
	mockLedgerRepo := new(MockLedgerRepo)
	mockUserRepo := new(MockUserRepo)
	ledgerService := NewLedgerService(mockLedgerRepo, mockUserRepo)

	mockUserRepo.On("GetUserCredentials", mock.Anything, "recipient").Return("", "", nil)

	err := ledgerService.SendMoney(context.Background(), "sender", "recipient", 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipient not found")
}

func TestGetUserTransactions_Success(t *testing.T) {
    mockLedgerRepo := new(MockLedgerRepo)
    mockUserRepo := new(MockUserRepo)
    ledgerService := NewLedgerService(mockLedgerRepo, mockUserRepo)

    transactions := []*models.Ledger{
        {ID: 1, MovementType: "transfer_in"},
        {ID: 2, MovementType: "transfer_out"},
        {ID: 3, MovementType: "transfer_in"},
    }

    mockLedgerRepo.On("GetUserTransactions", mock.Anything, "user-id", 100, 0).Return(transactions, nil)

    inTx, outTx, err := ledgerService.GetUserTransactions(context.Background(), "user-id")
    assert.NoError(t, err)
    assert.Len(t, inTx, 2)
    assert.Len(t, outTx, 1)

    mockLedgerRepo.AssertExpectations(t)
}

func TestGetUserTransactions_Error(t *testing.T) {
	mockLedgerRepo := new(MockLedgerRepo)
	mockUserRepo := new(MockUserRepo)
	ledgerService := NewLedgerService(mockLedgerRepo, mockUserRepo)

	mockLedgerRepo.On("GetUserTransactions", mock.Anything, "user-id", 100, 0).Return(nil, errors.New("DB error"))

	inTx, outTx, err := ledgerService.GetUserTransactions(context.Background(), "user-id")
	assert.Error(t, err)
	assert.Nil(t, inTx)
	assert.Nil(t, outTx)

	mockLedgerRepo.AssertExpectations(t)
}
