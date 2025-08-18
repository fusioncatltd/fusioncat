package tests

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

// ReadTestFile reads the content of a file from the testfiles directory.
// This function is meant to be used in tests only.
// The filePath parameter should be a relative path within the testfiles directory.
// Example: ReadTestFile("jsonschemas/validSchema1.json")
func ReadTestFile(filePath string) ([]byte, error) {
	// Get the absolute path to the testfiles directory
	// This assumes the function is called from within the tests package
	fullPath := filepath.Join("testfiles", filePath)
	
	// Read and return the file content
	return os.ReadFile(fullPath)
}

// ReadTestFileString is a convenience wrapper that returns the file content as a string
func ReadTestFileString(filePath string) (string, error) {
	content, err := ReadTestFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// CleanDatabase truncates all tables except schema_migrations
// This should be called at the beginning of each test to ensure a clean state
func CleanDatabase(t *testing.T) {
	// Load .env file from parent directory
	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		t.Logf("Warning: Could not load .env file from %s: %v", envPath, err)
	}
	
	// Get database connection parameters from environment
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	dbname := os.Getenv("PG_DB_NAME")
	sslmode := os.Getenv("PG_SSLMODE")
	
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "disable"
	}
	
	// Connect to database directly
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// Get all table names except schema_migrations
	rows, err := db.Query(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename != 'schema_migrations'
	`)
	if err != nil {
		t.Fatalf("Failed to query table names: %v", err)
	}
	defer rows.Close()
	
	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			t.Logf("Warning: Failed to scan table name: %v", err)
			continue
		}
		tableNames = append(tableNames, tableName)
	}
	
	// Truncate each table with CASCADE to handle foreign keys
	for _, tableName := range tableNames {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)
		if _, err := db.Exec(query); err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", tableName, err)
		}
	}
}