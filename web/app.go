package web

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// WebApp struct
type WebApp struct {
	pdb *sql.DB
}

// NewWebApp to initializes WebApp
func NewApp() (*WebApp, error) {
	postgresURI := os.Getenv("POSTGRES_URI")
	if postgresURI == "" {
		postgresURI = "postgres://sipp11:banshee10@localhost/hailing?sslmode=verify-full"
	}
	psqlDB, err := sql.Open("postgres", postgresURI)
	if err != nil {
		return nil, err
	}

	log.Printf("[web] db=%v", postgresURI)
	return &WebApp{
		pdb: psqlDB,
	}, nil
}
