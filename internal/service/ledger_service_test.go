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

    mockUserRepo.On("GetUserCredentials", mock.Anything, "recipient").
        Return("user-id-2", "some-pass", nil).Once()
    // Ожидаем один вызов GetBalance для "sender"
    mockUserRepo.On("GetBalance", mock.Anything, "sender").
        Return(100, nil).Once()
    // Ожидаем вызов SendMoney с суммой 50
    mockLedgerRepo.On("SendMoney", mock.Anything, "sender", "user-id-2", 50).
        Return(nil).Once()

    err := ledgerService.SendMoney(context.Background(), "sender", "recipient", 50)
    assert.NoError(t, err)

    mockUserRepo.AssertExpectations(t)
    mockLedgerRepo.AssertExpectations(t)
}

func TestSendMoney_InsufficientFunds(t *testing.T) {
    mockLedgerRepo := new(MockLedgerRepo)
    mockUserRepo := new(MockUserRepo)
    ledgerService := NewLedgerService(mockLedgerRepo, mockUserRepo)

    mockUserRepo.On("GetUserCredentials", mock.Anything, "recipient").
        Return("recipientID", "some-pass", nil).Once()
    mockUserRepo.On("GetBalance", mock.Anything, "sender").
        Return(100, nil).Once()

    // Пытаемся перевести 150, что больше баланса
    err := ledgerService.SendMoney(context.Background(), "sender", "recipient", 150)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient balance")

    mockLedgerRepo.AssertNotCalled(t, "SendMoney", mock.Anything, "sender", mock.Anything, mock.Anything)
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

    mockUserRepo.On("GetUserCredentials", mock.Anything, "recipient").
        Return("", "", errors.New("user not found")).Once()

    err := ledgerService.SendMoney(context.Background(), "sender", "recipient", 50)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to get recipient id for username")

    mockLedgerRepo.AssertNotCalled(t, "SendMoney", mock.Anything, "sender", mock.Anything, mock.Anything)

    mockUserRepo.AssertNotCalled(t, "GetBalance", mock.Anything, "sender")

    mockUserRepo.AssertExpectations(t)
    mockLedgerRepo.AssertExpectations(t)
}



func TestGetUserTransactions_Success(t *testing.T) {
    mockLedgerRepo := new(MockLedgerRepo)
    ledgerService := NewLedgerService(mockLedgerRepo, nil)

    transactions := []models.Ledger{
        {ID: 1, MovementType: "transfer_in"},
        {ID: 2, MovementType: "transfer_out"},
        {ID: 3, MovementType: "transfer_in"},
    }

    mockLedgerRepo.On("GetUserTransactions", mock.Anything, "user-id", 100, 0).
        Return(&transactions, nil).Once()

    inTx, outTx, err := ledgerService.GetUserTransactions(context.Background(), "user-id")
    assert.NoError(t, err)
    assert.Len(t, inTx, 2)
    assert.Len(t, outTx, 1)

    mockLedgerRepo.AssertExpectations(t)
}

func TestGetUserTransactions_Error(t *testing.T) {
    mockLedgerRepo := new(MockLedgerRepo)
    ledgerService := NewLedgerService(mockLedgerRepo, nil)

    mockLedgerRepo.On("GetUserTransactions", mock.Anything, "user-id", 100, 0).
        Return(nil, errors.New("DB error")).Once()

    inTx, outTx, err := ledgerService.GetUserTransactions(context.Background(), "user-id")
    assert.Error(t, err)
    assert.Nil(t, inTx)
    assert.Nil(t, outTx)

    mockLedgerRepo.AssertExpectations(t)
}
