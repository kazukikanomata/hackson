package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func InitDB(ctx context.Context) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("POSTGRES_DB"),
	)

	dbpool, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("DBプールの作成に失敗しました: %v", err)
	}

	if err := dbpool.PingContext(ctx); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("DB接続確認に失敗しました: %v", err)
	}

	fmt.Println("PostgreSQLに接続しました。")
	return dbpool, nil
}
