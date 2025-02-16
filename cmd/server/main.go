package main

import (
	"log"
	"net/http"
	"context"
	"EmployeeMerchStore/api"
	"EmployeeMerchStore/config"
	"EmployeeMerchStore/internal/database"
	"EmployeeMerchStore/internal/repository"
	"EmployeeMerchStore/internal/service"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Подключаемся к БД
	dbPool, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbPool.Close()

	// 3. Запускаем миграции
	ctx := context.Background()
	err = database.RunMigrations(ctx, dbPool)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Создаем репозитории
	userRepo := repository.NewUserRepository(dbPool)
	purchasesRepo := repository.NewPurchasesRepository(dbPool)
	ledgerRepo := repository.NewLedgerRepository(dbPool)
	// merchrRepo := repository.NewMerchRepository(dbPool)

	// Создаем сервисы
	userService := service.NewUserService(userRepo, cfg)
	purchasesService := service.NewPurchasesService(purchasesRepo, userRepo)
	ledgerService := service.NewLedgerService(ledgerRepo, userRepo)

	// Создаем хэндлер
	handler := api.NewHandler(userService, purchasesService, ledgerService)

	// Создаем роутер
	router := api.RegisterRoutes(handler)

	// Запускаем сервер
	log.Fatal(http.ListenAndServe(":8080", router))
}
