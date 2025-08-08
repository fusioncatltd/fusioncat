package tests

import (
	"os"
	"path/filepath"
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