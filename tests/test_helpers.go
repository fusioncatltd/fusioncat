package tests

import (
	"github.com/fusioncatltd/fusioncat/db"
	"os"
	"path/filepath"
	"testing"
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
	database := db.GetDB()
	
	// Get all table names except schema_migrations
	var tableNames []string
	database.Raw(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename != 'schema_migrations'
	`).Scan(&tableNames)
	
	// Truncate each table with CASCADE to handle foreign keys
	for _, tableName := range tableNames {
		if err := database.Exec("TRUNCATE TABLE " + tableName + " CASCADE").Error; err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", tableName, err)
		}
	}
}