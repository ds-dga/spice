package web

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

// WebApp struct
type WebApp struct {
	pdb      *sql.DB
	surveyDB *sql.DB
	basePath string
	secret   []byte
}

// NewWebApp to initializes WebApp
func NewApp() (*WebApp, error) {
	postgresURI := os.Getenv("POSTGRES_URI")
	surveyPostgresURI := os.Getenv("SURVEY_POSTGRES_URI")
	if postgresURI == "" {
		postgresURI = "postgres://sipp11:banshee10@localhost/hailing?sslmode=verify-full"
	}
	if surveyPostgresURI == "" {
		surveyPostgresURI = "postgres://sipp11:banshee10@localhost/survey?sslmode=verify-full"
	}
	psqlDB, err := sql.Open("postgres", postgresURI)
	if err != nil {
		return nil, err
	}
	surveyDB, err := sql.Open("postgres", surveyPostgresURI)
	if err != nil {
		return nil, err
	}
	basePath := os.Getenv("BASE_DIR")
	if basePath == "" {
		ex, err := os.Executable()
		if err != nil {
			return nil, err
		}
		basePath = filepath.Dir(ex)
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = `aeaf43eebaadc614a19bf123dab7ce4924506682efd544a27f6e207f2579b5aa8b47b4a7cee4991a5b655eaff2550a40d4c0bc79ad01462eee8fa711f7199db1`
	}

	log.Printf("[web] db=%v", postgresURI)
	return &WebApp{
		pdb:      psqlDB,
		surveyDB: surveyDB,
		basePath: basePath,
		secret:   []byte(secretKey),
	}, nil
}

// CheckOrCreateDir is a shortcut util
func CheckOrCreateDir(fp ...string) (string, error) {
	targetDir := filepath.Join(fp...)
	fInfo, err := os.Lstat(targetDir)
	if err != nil || !fInfo.IsDir() {
		if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return targetDir, err
		}
	}
	return targetDir, nil
}
