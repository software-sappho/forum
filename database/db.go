package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
	var err error

	DB, err = sql.Open("sqlite3", "forum.db")
	if err != nil {
		return err
	}

	// Check if the database connection works
	if err = DB.Ping(); err != nil {
		return err
	}

	// Find all .sql files in the SQL/ folder
	files, err := filepath.Glob("sql/*.sql")
	if err != nil {
		return err
	}

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %v", file, err)
		}

		sqlStmt := string(sqlBytes)

		if _, err := DB.Exec(sqlStmt); err != nil {
			return fmt.Errorf("failed to execute %s: %v", file, err)
		}
	}

	return nil
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	return DB.Exec(query, args...)
}

func Prepare(query string) (*sql.Stmt, error) {
	return DB.Prepare(query)
}
