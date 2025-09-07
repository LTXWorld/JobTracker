package database

import (
	"database/sql"
	"fmt"
	"jobView-backend/internal/config"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func New(cfg *config.DatabaseConfig) (*DB, error) {
	// 使用PostgreSQL URL格式确保正确连接
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database with DSN %s: %w", dsn, err)
	}

	return &DB{db}, nil
}