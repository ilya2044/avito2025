package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func NewDBFromEnv() (*sql.DB, error) {
	host := env("DB_HOST", "db")
	port := env("DB_PORT", "5432")
	user := env("DB_USER", "postgres")
	pass := env("DB_PASS", "postgres")
	name := env("DB_NAME", "prdb")
	ssl := env("DB_SSLMODE", "disable")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, name, ssl)
	return sql.Open("postgres", dsn)
}

func env(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
