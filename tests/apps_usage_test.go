package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

func TestGetAppUsageV1(t *testing.T) {
	// Clean database before running test
	CleanDatabase(t)
	
	h := os.Getenv("TESTSERVER_URL")
	e := httpexpect.Default(t, h)
	
	// Create test user
	testEmail := fmt.Sprintf("test-email-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	userPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    testEmail,
		Password: "123456789",
	}
	
	// Create user and get bearer token
	signUpResponse := e.POST("/v1/public/users").
		WithJSON(userPayload).
		Expect().
		Status(http.StatusOK)
	
	signUpResponse.Header("Authorization").NotEmpty()
	bearerToken := signUpResponse.Raw().Header.Get("Authorization")
	
	// Create a project
	projectName := fmt.Sprintf("TestProject%d", time.Now().UnixNano())
	projectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        projectName,
		Description: "Test project for app usage",
	}
	
	projectResponse := e.POST("/v1/protected/projects").
		WithHeader("Authorization", bearerToken).
		WithJSON(projectPayload).
		Expect().
		Status(http.StatusOK)
	
	var createdProject logic.ProjectDBSerializerStruct
	rawProjectReader := projectResponse.Raw().Body
	defer rawProjectReader.Close()
	rawProjectBytes, _ := io.ReadAll(rawProjectReader)
	require.NoError(t, json.Unmarshal(rawProjectBytes, &createdProject))
	
	// Create an app
	appName := fmt.Sprintf("TestApp%d", time.Now().UnixNano())
	appPayload := input_contracts.CreateAppApiInputContract{
		Name:        appName,
		Description: "Test app for usage",
	}
	
	createAppResponse := e.POST("/v1/protected/projects/" + createdProject.ID + "/apps").
		WithHeader("Authorization", bearerToken).
		WithJSON(appPayload).
		Expect().
		Status(http.StatusOK)
	
	var createdApp logic.AppDBSerializerStruct
	rawAppReader := createAppResponse.Raw().Body
	defer rawAppReader.Close()
	rawAppBytes, _ := io.ReadAll(rawAppReader)
	require.NoError(t, json.Unmarshal(rawAppBytes, &createdApp))
	
	// Test getting app usage (should be empty initially)
	usageResponse := e.GET("/v1/protected/apps/" + createdApp.ID + "/usage").
		WithHeader("Authorization", bearerToken).
		Expect()
	
	// Print the response for debugging
	rawUsageReader := usageResponse.Raw().Body
	defer rawUsageReader.Close()
	rawUsageBytes, _ := io.ReadAll(rawUsageReader)
	
	// Check status first
	if usageResponse.Raw().StatusCode != http.StatusOK {
		t.Logf("Error response: %s", string(rawUsageBytes))
		require.Equal(t, http.StatusOK, usageResponse.Raw().StatusCode, "Expected 200 OK but got %d", usageResponse.Raw().StatusCode)
	}
	
	var appUsage logic.AppUsageMatrixResponse
	require.NoError(t, json.Unmarshal(rawUsageBytes, &appUsage))
	
	require.Equal(t, createdApp.ID, appUsage.AppID)
	require.Empty(t, appUsage.Sends)
	require.Empty(t, appUsage.Receives)
	
	// Test with non-existent app
	e.GET("/v1/protected/apps/00000000-0000-0000-0000-000000000000/usage").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusNotFound)
	
	// Test without auth
	e.GET("/v1/protected/apps/" + createdApp.ID + "/usage").
		Expect().
		Status(http.StatusUnauthorized)
}