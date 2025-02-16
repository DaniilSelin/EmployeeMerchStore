package service

import (
	"context"
	"fmt"
    "time"

	"EmployeeMerchStore/internal/repository"
	"EmployeeMerchStore/internal/models"
    "EmployeeMerchStore/internal/cache"
	"EmployeeMerchStore/config"

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt"
    "golang.org/x/crypto/bcrypt"
)

type UserService struct {
    config   *config.Config
    userRepo repository.UserRepositoryInterface
    cache    *cache.Cache 
}

func NewUserService(userRepo repository.UserRepositoryInterface, config *config.Config) *UserService {
    c := cache.NewCache()

    // Запуск горутины для очистки кэша
    go c.СleanupExpiredItems()

    return &UserService{
        userRepo: userRepo,
        config: config,
        cache:    c,
    }
}

func (us *UserService) CreateUser(ctx context.Context, username, password string) (string, error) {
    id := uuid.New().String()

    // Хешируем пароль
    hashPswd, err := us.CreateHash(password)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }

    // Создаём пользователя с балансом 1000 (по тз)
    if err := us.userRepo.CreateUser(ctx, id, username, hashPswd, 1000); err != nil {
        return "", fmt.Errorf("failed to create user: %w", err)
    }

    // Генерируем JWT для нового пользователя
    token, err := us.GenerateJWT(id)
    if err != nil {
        return "", fmt.Errorf("failed to create auth token: %w", err)
    }

    // Кэшируем токен
    cacheKey := "auth:" + username + ":" + password
    us.cache.Set(cacheKey, token, 10*time.Minute)

    return token, nil
}

func (us *UserService) GetInfo(ctx context.Context, token string, ps *PurchasesService, ls *LedgerService) (int, []*models.UserMerch, []*models.Ledger, []*models.Ledger, error) {
    userID, err := us.DecodeToken(token)
    if err != nil {
        return 0, nil, nil, nil, fmt.Errorf("failed to decode auth token: %w", err)
    }

    balance, err := us.GetBalance(ctx, userID)
    if err != nil {
        return 0, nil, nil, nil, fmt.Errorf("failed to get balance: %w", err)
    }

    merchList, err := ps.GetUserMerch(ctx, userID)
    if err != nil {
        return 0, nil, nil, nil, fmt.Errorf("failed to get user merch: %w", err)
    }

    transactionsIn, transactionsOut, err := ls.GetUserTransactions(ctx, userID)
    if err != nil {
        return 0, nil, nil, nil, fmt.Errorf("failed to get user transactions: %w", err)
    }

    return balance, merchList, transactionsIn, transactionsOut, nil
}


func (us *UserService) GetBalance(ctx context.Context, id string) (int, error) {
    balance, err := us.userRepo.GetBalance(ctx, id)

	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (us *UserService) Auth(ctx context.Context, username, password string) (string, error) {
    // Проверяем кэш 
    cacheKey := "auth:" + username + ":" + password
    if tokenCached, found := us.cache.Get(cacheKey); found {
        if tokenStr, ok := tokenCached.(string); ok {
            return tokenStr, nil // Возвращаем токен из кэша
        }
    }

    userID, storedHash, err := us.userRepo.GetUserCredentials(ctx, username)
    if err != nil ||  userID == "" {
        return "", fmt.Errorf("failed to get user credentials: %w", err)
    }
    
    // Сравниваем хэш с предоставленным паролем
    if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
        return "", fmt.Errorf("invalid credentials: %w", err)
    }
    
    token, err := us.GenerateJWT(userID)
    if err != nil {
        return "", fmt.Errorf("failed to generate JWT: %w", err)
    }

    us.cache.Set(cacheKey, token, 10*time.Minute)
    
    return token, nil
}

func (us *UserService) DecodeToken(tokenStr string) (string, error) {
    jwtKey := []byte(us.config.Jwt.SecretKey)
    claims := &models.Claims{}

    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })
    if err != nil || !token.Valid {
        return "", fmt.Errorf("invalid token")
    }
    return claims.UserID, nil
}

func (us *UserService) CreateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to create hash: %w", err)
	}
	return string(hash), nil
}

func (us *UserService) GenerateJWT(id string) (string, error) {
	var jwtKey = []byte(us.config.Jwt.SecretKey)

    expirationTime := time.Now().Add(time.Duration(us.config.Jwt.Expiration) * time.Minute)

	claims := &models.Claims{
        UserID:   id,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        return "", err
    }
    return tokenString, nil
}