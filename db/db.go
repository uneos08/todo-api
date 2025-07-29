// db/store.go
package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	DB *sqlx.DB
}

func NewPostgresStore(connStr string) *PostgresStore {
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	return &PostgresStore{DB: db}
}
