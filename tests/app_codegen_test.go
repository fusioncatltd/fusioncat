package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

func TestAppCodeGenerationWithImport(t *testing.T) {
	// Clean database before running test
	CleanDatabase(t)

	h := os.Getenv("TESTSERVER_URL")
	e := httpexpect.Default(t, h)

	// Create user and get bearer token
	userEmail := fmt.Sprintf("test-email-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	userPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    userEmail,
		Password: "123456789",
	}

	userSignUpResponse := e.POST("/v1/public/users").
		WithJSON(userPayload).
		Expect().
		Status(http.StatusOK)

	userBearer := userSignUpResponse.Raw().Header.Get("Authorization")

	// Create a project
	projectName := fmt.Sprintf("CodeGenProject%d", time.Now().UnixNano())
	projectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        projectName,
		Description: "Test project for code generation",
	}

	projectResponse := e.POST("/v1/protected/projects").
		WithHeader("Authorization", userBearer).
		WithJSON(projectPayload).
		Expect().
		Status(http.StatusOK)

	var project logic.ProjectDBSerializerStruct
	rawProjectReader := projectResponse.Raw().Body
	defer rawProjectReader.Close()
	rawProjectBytes, _ := io.ReadAll(rawProjectReader)
	require.NoError(t, json.Unmarshal(rawProjectBytes, &project))

	// Read a complete project structure from test file
	yamlBytes, err := os.ReadFile("testfiles/imports/validImportReworked2.yaml")
	require.NoError(t, err)
	yamlContent := string(yamlBytes)

	// Import the YAML to create everything
	importPayload := input_contracts.ImportFileInputContract{
		YAML: yamlContent,
	}

	_ = e.POST("/v1/protected/projects/"+project.ID+"/imports").
		WithHeader("Authorization", userBearer).
		WithJSON(importPayload).
		Expect().
		Status(http.StatusOK)

	// Get the app that was created
	appsResponse := e.GET("/v1/protected/projects/"+project.ID+"/apps").
		WithHeader("Authorization", userBearer).
		Expect().
		Status(http.StatusOK)

	var apps []logic.AppDBSerializerStruct
	rawAppsReader := appsResponse.Raw().Body
	defer rawAppsReader.Close()
	rawAppsBytes, _ := io.ReadAll(rawAppsReader)
	require.NoError(t, json.Unmarshal(rawAppsBytes, &apps))
	require.True(t, len(apps) > 0, "At least one app should be created")

	// Find the payment_processor app which has both sends and receives
	var createdApp logic.AppDBSerializerStruct
	for _, app := range apps {
		if app.Name == "payment_processor" {
			createdApp = app
			break
		}
	}
	require.NotEmpty(t, createdApp.ID, "payment_processor app should exist")

	// Test getting app usage
	usageResponse := e.GET("/v1/protected/apps/"+createdApp.ID+"/usage").
		WithHeader("Authorization", userBearer).
		Expect().
		Status(http.StatusOK)

	var usage logic.AppUsageMatrixResponse
	rawUsageReader := usageResponse.Raw().Body
	defer rawUsageReader.Close()
	rawUsageBytes, _ := io.ReadAll(rawUsageReader)
	require.NoError(t, json.Unmarshal(rawUsageBytes, &usage))

	// Verify usage matrix has connections
	require.True(t, len(usage.Sends) > 0, "App should have sends")
	require.True(t, len(usage.Receives) > 0, "App should have receives")

	// Test code generation endpoint
	codeResponse := e.GET("/v1/protected/apps/"+createdApp.ID+"/code/go").
		WithHeader("Authorization", userBearer).
		Expect()

	// Check if we got a successful response
	status := codeResponse.Raw().StatusCode
	if status != http.StatusOK {
		t.Logf("Unexpected status code: %d", status)
		body, _ := io.ReadAll(codeResponse.Raw().Body)
		t.Logf("Response body: %s", string(body))
	}
	require.Equal(t, http.StatusOK, status)

	generatedCode := codeResponse.Body().Raw()

	// Debug: Print the generated code if it's empty or short
	if len(generatedCode) < 100 {
		t.Logf("Generated code length: %d", len(generatedCode))
		t.Logf("Generated code: %q", generatedCode)
	}

	// Verify the generated code is not empty
	require.NotEmpty(t, generatedCode, "Generated code should not be empty")
	require.True(t, len(generatedCode) > 100, "Generated code should be substantial")

	// Verify the generated code contains expected elements
	require.Contains(t, generatedCode, "package fusioncat")

	// Verify interfaces
	require.Contains(t, generatedCode, "type Message interface")
	require.Contains(t, generatedCode, "type Schema interface")
	require.Contains(t, generatedCode, "type Resource interface")

	// Print the generated code for inspection
	t.Logf("Generated Go code:\n%s", generatedCode)
	
	// Basic check that some content was generated
	// The exact content may vary based on templates
	require.True(t,
		strings.Contains(generatedCode, "func") ||
			strings.Contains(generatedCode, "type") ||
			strings.Contains(generatedCode, "struct"),
		"Generated code should contain Go code structures")
	
	// Verify URIs are correctly formatted
	require.True(t,
		strings.Contains(generatedCode, "async+"),
		"Generated code should contain properly formatted async URIs")

	// Test invalid language
	_ = e.GET("/v1/protected/apps/"+createdApp.ID+"/code/python").
		WithHeader("Authorization", userBearer).
		Expect().
		Status(http.StatusBadRequest)

	// Test non-existent app
	_ = e.GET("/v1/protected/apps/00000000-0000-0000-0000-000000000000/code/go").
		WithHeader("Authorization", userBearer).
		Expect().
		Status(http.StatusNotFound)

	// Test without authentication
	_ = e.GET("/v1/protected/apps/" + createdApp.ID + "/code/go").
		Expect().
		Status(http.StatusUnauthorized)
}
