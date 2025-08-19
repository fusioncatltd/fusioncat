package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

// Helper function to load YAML file
func loadYAMLFile(t *testing.T, filename string) string {
	filePath := filepath.Join("testfiles", "imports", filename)
	content, err := os.ReadFile(filePath)
	require.NoError(t, err, "Failed to read YAML file: %s", filename)
	return string(content)
}

func TestProjectImportsAndValidation(t *testing.T) {
	// Clean database before running test
	CleanDatabase(t)

	h := os.Getenv("TESTSERVER_URL")
	e := httpexpect.Default(t, h)

	// Create a user and get bearer token
	userEmail := fmt.Sprintf("test-email-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	userPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    userEmail,
		Password: "123456789",
	}

	userSignUpResponse := e.POST("/v1/public/users").
		WithJSON(userPayload).
		Expect().
		Status(http.StatusOK)

	userSignUpResponse.Header("Authorization").NotEmpty()
	userBearer := userSignUpResponse.Raw().Header.Get("Authorization")

	// Create a project
	projectName := fmt.Sprintf("TestProject%d", time.Now().UnixNano())
	projectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        projectName,
		Description: "Test project for imports",
	}

	projectResponse := e.POST("/v1/protected/projects").
		WithHeader("Authorization", userBearer).
		WithJSON(projectPayload).
		Expect().
		Status(http.StatusOK)

	var createdProject logic.ProjectDBSerializerStruct
	rawProjectReader := projectResponse.Raw().Body
	defer rawProjectReader.Close()
	rawProjectBytes, _ := io.ReadAll(rawProjectReader)

	require.NoError(t, json.Unmarshal(rawProjectBytes, &createdProject))

	// Test 1: Validate a valid YAML file
	validYAML := loadYAMLFile(t, "valid_basic.yaml")

	validatePayload := input_contracts.ImportFileInputContract{
		YAML: validYAML,
	}

	// Validate the YAML
	validateResponse := e.POST("/v1/protected/projects/"+createdProject.ID+"/imports/validator").
		WithHeader("Authorization", userBearer).
		WithJSON(validatePayload).
		Expect()

	// Check response status and handle both success and error cases
	if validateResponse.Raw().StatusCode != http.StatusOK {
		// If validation failed, print the errors for debugging
		var errorResult map[string]interface{}
		rawErrorReader := validateResponse.Raw().Body
		defer rawErrorReader.Close()
		rawErrorBytes, _ := io.ReadAll(rawErrorReader)
		json.Unmarshal(rawErrorBytes, &errorResult)
		t.Logf("Validation errors: %+v", errorResult)
		require.Equal(t, http.StatusOK, validateResponse.Raw().StatusCode, "Validation should succeed")
	}

	var validateResult map[string]string
	rawValidateReader := validateResponse.Raw().Body
	defer rawValidateReader.Close()
	rawValidateBytes, _ := io.ReadAll(rawValidateReader)

	require.NoError(t, json.Unmarshal(rawValidateBytes, &validateResult))
	require.Equal(t, "YAML is valid", validateResult["message"])

	// Test 2: Import the valid YAML
	importResponse := e.POST("/v1/protected/projects/"+createdProject.ID+"/imports").
		WithHeader("Authorization", userBearer).
		WithJSON(validatePayload).
		Expect().
		Status(http.StatusOK)

	var importResult map[string]string
	rawImportReader := importResponse.Raw().Body
	defer rawImportReader.Close()
	rawImportBytes, _ := io.ReadAll(rawImportReader)

	require.NoError(t, json.Unmarshal(rawImportBytes, &importResult))
	require.Equal(t, "Import completed successfully", importResult["message"])

	// Verify imported data - check servers
	serversResponse := e.GET("/v1/protected/projects/"+createdProject.ID+"/servers").
		WithHeader("Authorization", userBearer).
		Expect().
		Status(http.StatusOK)

	var servers []logic.ServerDBSerializerStruct
	rawServersReader := serversResponse.Raw().Body
	defer rawServersReader.Close()
	rawServersBytes, _ := io.ReadAll(rawServersReader)

	require.NoError(t, json.Unmarshal(rawServersBytes, &servers))
	require.Len(t, servers, 1)
	require.Equal(t, "kafka_server", servers[0].Name)
	require.Equal(t, "kafka", servers[0].Protocol)

	// Verify imported data - check apps
	appsResponse := e.GET("/v1/protected/projects/"+createdProject.ID+"/apps").
		WithHeader("Authorization", userBearer).
		Expect().
		Status(http.StatusOK)

	var apps []logic.AppDBSerializerStruct
	rawAppsReader := appsResponse.Raw().Body
	defer rawAppsReader.Close()
	rawAppsBytes, _ := io.ReadAll(rawAppsReader)

	require.NoError(t, json.Unmarshal(rawAppsBytes, &apps))
	require.Len(t, apps, 1)
	require.Equal(t, "UserApp", apps[0].Name)

	// Test 3: Validate invalid YAML - missing version
	invalidYAML1 := loadYAMLFile(t, "invalid_missing_version.yaml")

	invalidPayload1 := input_contracts.ImportFileInputContract{
		YAML: invalidYAML1,
	}

	validateInvalidResponse1 := e.POST("/v1/protected/projects/"+createdProject.ID+"/imports/validator").
		WithHeader("Authorization", userBearer).
		WithJSON(invalidPayload1).
		Expect().
		Status(http.StatusConflict)

	var validateInvalidResult1 map[string]interface{}
	rawValidateInvalidReader1 := validateInvalidResponse1.Raw().Body
	defer rawValidateInvalidReader1.Close()
	rawValidateInvalidBytes1, _ := io.ReadAll(rawValidateInvalidReader1)

	require.NoError(t, json.Unmarshal(rawValidateInvalidBytes1, &validateInvalidResult1))
	require.Contains(t, validateInvalidResult1, "errors")

	// Test 4: Validate invalid YAML - invalid JSON schema
	invalidYAML2 := loadYAMLFile(t, "invalid_bad_schema.yaml")

	invalidPayload2 := input_contracts.ImportFileInputContract{
		YAML: invalidYAML2,
	}

	validateInvalidResponse2 := e.POST("/v1/protected/projects/"+createdProject.ID+"/imports/validator").
		WithHeader("Authorization", userBearer).
		WithJSON(invalidPayload2).
		Expect().
		Status(http.StatusConflict)

	var validateInvalidResult2 map[string]interface{}
	rawValidateInvalidReader2 := validateInvalidResponse2.Raw().Body
	defer rawValidateInvalidReader2.Close()
	rawValidateInvalidBytes2, _ := io.ReadAll(rawValidateInvalidReader2)

	require.NoError(t, json.Unmarshal(rawValidateInvalidBytes2, &validateInvalidResult2))
	require.Contains(t, validateInvalidResult2, "errors")

	// Test 5: Try to import with duplicate names (should fail validation)
	duplicateYAML := loadYAMLFile(t, "duplicate_server.yaml")

	duplicatePayload := input_contracts.ImportFileInputContract{
		YAML: duplicateYAML,
	}

	validateDuplicateResponse := e.POST("/v1/protected/projects/"+createdProject.ID+"/imports/validator").
		WithHeader("Authorization", userBearer).
		WithJSON(duplicatePayload).
		Expect().
		Status(http.StatusConflict)

	var validateDuplicateResult map[string]interface{}
	rawValidateDuplicateReader := validateDuplicateResponse.Raw().Body
	defer rawValidateDuplicateReader.Close()
	rawValidateDuplicateBytes, _ := io.ReadAll(rawValidateDuplicateReader)

	require.NoError(t, json.Unmarshal(rawValidateDuplicateBytes, &validateDuplicateResult))
	errors, ok := validateDuplicateResult["errors"].([]interface{})
	require.True(t, ok)
	require.Greater(t, len(errors), 0)
	// Check that at least one error mentions the duplicate server
	foundDuplicateError := false
	for _, err := range errors {
		if errStr, ok := err.(string); ok {
			if contains(errStr, "kafka_server") && contains(errStr, "already exists") {
				foundDuplicateError = true
				break
			}
		}
	}
	require.True(t, foundDuplicateError, "Should have found error about duplicate server name")

	// Test 6: Validate YAML with invalid resource reference
	invalidResourceYAML := loadYAMLFile(t, "invalid_missing_message.yaml")

	invalidResourcePayload := input_contracts.ImportFileInputContract{
		YAML: invalidResourceYAML,
	}

	validateInvalidResourceResponse := e.POST("/v1/protected/projects/"+createdProject.ID+"/imports/validator").
		WithHeader("Authorization", userBearer).
		WithJSON(invalidResourcePayload).
		Expect().
		Status(http.StatusConflict)

	var validateInvalidResourceResult map[string]interface{}
	rawValidateInvalidResourceReader := validateInvalidResourceResponse.Raw().Body
	defer rawValidateInvalidResourceReader.Close()
	rawValidateInvalidResourceBytes, _ := io.ReadAll(rawValidateInvalidResourceReader)

	require.NoError(t, json.Unmarshal(rawValidateInvalidResourceBytes, &validateInvalidResourceResult))
	require.Contains(t, validateInvalidResourceResult, "errors")

	// Test 7: Test validation without authentication
	_ = e.POST("/v1/protected/projects/" + createdProject.ID + "/imports/validator").
		WithJSON(validatePayload).
		Expect().
		Status(http.StatusUnauthorized)

	// Test 8: Test import without authentication
	_ = e.POST("/v1/protected/projects/" + createdProject.ID + "/imports").
		WithJSON(validatePayload).
		Expect().
		Status(http.StatusUnauthorized)

	// Test 9: Test with non-existent project
	_ = e.POST("/v1/protected/projects/00000000-0000-0000-0000-000000000000/imports/validator").
		WithHeader("Authorization", userBearer).
		WithJSON(validatePayload).
		Expect().
		Status(http.StatusNotFound)

	// Test 10: Test with more complex valid YAML file
	complexYAML := loadYAMLFile(t, "validImportReworked2.yaml")

	complexPayload := input_contracts.ImportFileInputContract{
		YAML: complexYAML,
	}

	// Validate the complex YAML
	validateComplexResponse := e.POST("/v1/protected/projects/"+createdProject.ID+"/imports/validator").
		WithHeader("Authorization", userBearer).
		WithJSON(complexPayload).
		Expect()

	// Check response and handle potential errors
	if validateComplexResponse.Raw().StatusCode != http.StatusOK {
		// If validation failed, print the errors for debugging
		var errorResult map[string]interface{}
		rawErrorReader := validateComplexResponse.Raw().Body
		defer rawErrorReader.Close()
		rawErrorBytes, _ := io.ReadAll(rawErrorReader)
		json.Unmarshal(rawErrorBytes, &errorResult)
		t.Logf("Complex YAML validation errors: %+v", errorResult)
		require.Equal(t, http.StatusOK, validateComplexResponse.Raw().StatusCode, "Complex YAML validation should succeed")
	}

	var validateComplexResult map[string]string
	rawValidateComplexReader := validateComplexResponse.Raw().Body
	defer rawValidateComplexReader.Close()
	rawValidateComplexBytes, _ := io.ReadAll(rawValidateComplexReader)

	require.NoError(t, json.Unmarshal(rawValidateComplexBytes, &validateComplexResult))
	require.Equal(t, "YAML is valid", validateComplexResult["message"])
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) == 0 || (len(substr) > 0 && s[0:len(substr)] == substr) || contains(s[1:], substr))
}
