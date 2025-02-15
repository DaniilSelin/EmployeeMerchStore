package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func RunMigrations(ctx context.Context, conn *pgxpool.Pool) error {
	// Читаем SQL файлы миграций
	files := []string{
		"internal/database/migrations/create_user.sql",
		"internal/database/migrations/create_merch.sql",
		"internal/database/migrations/create_ledger.sql",
		"internal/database/migrations/create_idx_lastLedger.sql",
		"internal/database/migrations/create_purchases.sql",
	}
	for _, file := range files {
		// Читаем содержимое файла
		sqlContent, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read SQL file %s: %w", file, err)
		}

		// Выполняем SQL запрос
		_, err = conn.Exec(ctx, string(sqlContent)) // Передаем ctx
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		log.Printf("Successfully executed migration: %s", file)
	}

	return nil
}
