package main

import (
	"context"
	"log"

	"EmployeeMerchStore/config"
	"EmployeeMerchStore/internal/database"
)

func main() {
// 1. Загружаем конфиг
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 2. Подключаемся к БД
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
}