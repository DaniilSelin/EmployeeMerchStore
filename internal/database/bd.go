package database

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v4/pgxpool"
    "EmployeeMerchStore/config"
)

// InitDB инициализирует подключение к базе данных
func InitDB(cfg *config.Config) (*pgxpool.Pool, error) {
    dbURL := fmt.Sprintf(
        "postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Dbname,
        cfg.Database.Sslmode,
        cfg.Database.Schema,
    )

    pool, err := pgxpool.Connect(context.Background(), dbURL)
    if err != nil {
        return nil, fmt.Errorf("unable to connect to database: %w", err)
    }

    log.Println("Connected to database")
    return pool, nil
}
