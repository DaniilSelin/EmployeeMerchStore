package service

import (
    "context"
    "errors"
    "testing"

    "EmployeeMerchStore/config"
    "github.com/stretchr/testify/assert"
    "golang.org/x/crypto/bcrypt"
    "github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) GetUserCredentials(ctx context.Context, username string) (string, string, error) {
    args := m.Called(ctx, username)
    return args.String(0), args.String(1), args.Error(2)
}

func (m *MockUserRepo) GetBalance(ctx context.Context, id string) (int, error) {
    _ = m.Called(ctx, id)
    return 1000, nil
}

func (m *MockUserRepo) CreateUser(ctx context.Context, id, username, hashPswd string, balance float64) error {
    args := m.Called(ctx, id, username, hashPswd, balance)
    return args.Error(0)
}

func TestCreateUser(t *testing.T) {
    mockRepo := &MockUserRepo{}
    cfg := &config.Config{}
    userService := NewUserService(mockRepo, cfg)


    mockRepo.On("CreateUser", mock.Anything, mock.Anything, "testuser", mock.Anything, 1000).Return(nil)

    token, err := userService.CreateUser(context.Background(), "testuser", "password123")
    assert.NoError(t, err)
    assert.NotEmpty(t, token)

    mockRepo.AssertExpectations(t)
}

func TestAuth_Success(t *testing.T) {
    mockRepo := &MockUserRepo{}
    cfg := &config.Config{}
    userService := NewUserService(mockRepo, cfg)

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
    mockRepo.On("GetUserCredentials", mock.Anything, "testuser").Return("user-id", string(hashedPassword), nil)

    token, err := userService.Auth(context.Background(), "testuser", "password123")
    assert.NoError(t, err)
    assert.NotEmpty(t, token)

    mockRepo.AssertExpectations(t)
}

func TestAuth_InvalidPassword(t *testing.T) {
    mockRepo := &MockUserRepo{}
    cfg := &config.Config{}
    userService := NewUserService(mockRepo, cfg)

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
    mockRepo.On("GetUserCredentials", mock.Anything, "testuser").Return("user-id", string(hashedPassword), nil)

    token, err := userService.Auth(context.Background(), "testuser", "wrongpassword")
    assert.Error(t, err)
    assert.Empty(t, token)

    mockRepo.AssertExpectations(t)
}

func TestGetBalance_Success(t *testing.T) {
    mockRepo := &MockUserRepo{}
    cfg := &config.Config{}
    userService := NewUserService(mockRepo, cfg)

    mockRepo.On("GetBalance", mock.Anything, "user-id").Return(500, nil)

    balance, err := userService.GetBalance(context.Background(), "user-id")
    assert.NoError(t, err)
    assert.Equal(t, 500, balance)

    mockRepo.AssertExpectations(t)
}

func TestGetBalance_Error(t *testing.T) {
    mockRepo := &MockUserRepo{}
    cfg := &config.Config{}
    userService := NewUserService(mockRepo, cfg)

    mockRepo.On("GetBalance", mock.Anything, "user-id").Return(0, errors.New("DB error"))

    balance, err := userService.GetBalance(context.Background(), "user-id")
    assert.Error(t, err)
    assert.Equal(t, 0, balance)

    mockRepo.AssertExpectations(t)
}

func TestGenerateJWT_Success(t *testing.T) {
    cfg := &config.Config{
        Jwt: config.JwtConfig{SecretKey: "test-secret"},
    }
    userService := NewUserService(nil, cfg)

    token, err := userService.GenerateJWT("user-id")
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
}

