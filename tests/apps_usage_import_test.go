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

func TestAppUsageWithImportedProject(t *testing.T) {
	// Clean database before running test
	CleanDatabase(t)
	
	h := os.Getenv("TESTSERVER_URL")
	e := httpexpect.Default(t, h)
	
	// Create test user
	testEmail := fmt.Sprintf("test-usage-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
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
	projectName := fmt.Sprintf("UsageTestProject%d", time.Now().UnixNano())
	projectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        projectName,
		Description: "Test project for app usage with imports",
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
	
	// Load the import file with complex connections
	yamlContent := loadYAMLFile(t, "validImportReworked2.yaml")
	
	// Import the project structure
	importPayload := input_contracts.ImportFileInputContract{
		YAML: yamlContent,
	}
	
	importResponse := e.POST("/v1/protected/projects/" + createdProject.ID + "/imports").
		WithHeader("Authorization", bearerToken).
		WithJSON(importPayload).
		Expect()
	
	// Check the response
	rawImportReader := importResponse.Raw().Body
	defer rawImportReader.Close()
	rawImportBytes, _ := io.ReadAll(rawImportReader)
	
	if importResponse.Raw().StatusCode != http.StatusOK {
		t.Logf("Import failed with status %d: %s", importResponse.Raw().StatusCode, string(rawImportBytes))
		require.Equal(t, http.StatusOK, importResponse.Raw().StatusCode, "Import should succeed")
	}
	
	var importResult map[string]interface{}
	require.NoError(t, json.Unmarshal(rawImportBytes, &importResult))
	require.Equal(t, "Import completed successfully", importResult["message"])
	
	// Get list of all apps in the project
	appsResponse := e.GET("/v1/protected/projects/" + createdProject.ID + "/apps").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusOK)
	
	var apps []logic.AppDBSerializerStruct
	rawAppsReader := appsResponse.Raw().Body
	defer rawAppsReader.Close()
	rawAppsBytes, _ := io.ReadAll(rawAppsReader)
	require.NoError(t, json.Unmarshal(rawAppsBytes, &apps))
	
	// We should have imported 6 apps from the YAML file
	require.Len(t, apps, 6, "Should have imported 6 apps")
	
	// Create a map of app names to IDs for easier testing
	appNameToID := make(map[string]string)
	for _, app := range apps {
		appNameToID[app.Name] = app.ID
	}
	
	// Test main_backend app usage (sends 3 messages, receives 0)
	t.Run("main_backend usage", func(t *testing.T) {
		appID := appNameToID["main_backend"]
		require.NotEmpty(t, appID, "main_backend app should exist")
		
		usageResponse := e.GET("/v1/protected/apps/" + appID + "/usage").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusOK)
		
		var usage logic.AppUsageMatrixResponse
		rawUsageReader := usageResponse.Raw().Body
		defer rawUsageReader.Close()
		rawUsageBytes, _ := io.ReadAll(rawUsageReader)
		require.NoError(t, json.Unmarshal(rawUsageBytes, &usage))
		
		require.Equal(t, appID, usage.AppID)
		require.Len(t, usage.Sends, 3, "main_backend should send 3 messages")
		require.Empty(t, usage.Receives, "main_backend should not receive any messages")
		
		// Check that it sends the correct messages
		sentMessages := make(map[string]bool)
		for _, send := range usage.Sends {
			sentMessages[send.Message.Name] = true
		}
		require.True(t, sentMessages["send_email"], "Should send send_email message")
		require.True(t, sentMessages["track_user_event"], "Should send track_user_event message")
		require.True(t, sentMessages["process_payment"], "Should send process_payment message")
	})
	
	// Test mailer app usage (receives 1 message, sends 1)
	t.Run("mailer usage", func(t *testing.T) {
		appID := appNameToID["mailer"]
		require.NotEmpty(t, appID, "mailer app should exist")
		
		usageResponse := e.GET("/v1/protected/apps/" + appID + "/usage").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusOK)
		
		var usage logic.AppUsageMatrixResponse
		rawUsageReader := usageResponse.Raw().Body
		defer rawUsageReader.Close()
		rawUsageBytes, _ := io.ReadAll(rawUsageReader)
		require.NoError(t, json.Unmarshal(rawUsageBytes, &usage))
		
		require.Equal(t, appID, usage.AppID)
		require.Len(t, usage.Receives, 1, "mailer should receive 1 message")
		require.Len(t, usage.Sends, 1, "mailer should send 1 message")
		
		// Check specific messages
		require.Equal(t, "send_email", usage.Receives[0].Message.Name, "Should receive send_email")
		require.Equal(t, "record_analytics", usage.Sends[0].Message.Name, "Should send record_analytics")
	})
	
	// Test payment_processor app usage (receives 1, sends 2)
	t.Run("payment_processor usage", func(t *testing.T) {
		appID := appNameToID["payment_processor"]
		require.NotEmpty(t, appID, "payment_processor app should exist")
		
		usageResponse := e.GET("/v1/protected/apps/" + appID + "/usage").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusOK)
		
		var usage logic.AppUsageMatrixResponse
		rawUsageReader := usageResponse.Raw().Body
		defer rawUsageReader.Close()
		rawUsageBytes, _ := io.ReadAll(rawUsageReader)
		require.NoError(t, json.Unmarshal(rawUsageBytes, &usage))
		
		require.Equal(t, appID, usage.AppID)
		require.Len(t, usage.Receives, 1, "payment_processor should receive 1 message")
		require.Len(t, usage.Sends, 2, "payment_processor should send 2 messages")
		
		// Check receives
		require.Equal(t, "process_payment", usage.Receives[0].Message.Name)
		
		// Check sends
		sentMessages := make(map[string]bool)
		for _, send := range usage.Sends {
			sentMessages[send.Message.Name] = true
		}
		require.True(t, sentMessages["track_user_event"])
		require.True(t, sentMessages["record_analytics"])
	})
	
	// Test analytics_service app usage (receives 1, sends 2)
	t.Run("analytics_service usage", func(t *testing.T) {
		appID := appNameToID["analytics_service"]
		require.NotEmpty(t, appID, "analytics_service app should exist")
		
		usageResponse := e.GET("/v1/protected/apps/" + appID + "/usage").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusOK)
		
		var usage logic.AppUsageMatrixResponse
		rawUsageReader := usageResponse.Raw().Body
		defer rawUsageReader.Close()
		rawUsageBytes, _ := io.ReadAll(rawUsageReader)
		require.NoError(t, json.Unmarshal(rawUsageBytes, &usage))
		
		require.Equal(t, appID, usage.AppID)
		require.Len(t, usage.Receives, 1, "analytics_service should receive 1 message")
		require.Len(t, usage.Sends, 2, "analytics_service should send 2 messages")
		
		// Check that resources and servers are properly populated
		for _, receive := range usage.Receives {
			require.NotNil(t, receive.Resource, "Resource should not be nil")
			require.NotNil(t, receive.Server, "Server should not be nil")
			require.NotNil(t, receive.Message, "Message should not be nil")
			
			// Verify resource details
			require.NotEmpty(t, receive.Resource.Name, "Resource name should not be empty")
			require.NotEmpty(t, receive.Resource.ResourceType, "Resource type should not be empty")
			
			// Verify server details
			require.NotEmpty(t, receive.Server.Name, "Server name should not be empty")
			require.NotEmpty(t, receive.Server.Protocol, "Server protocol should not be empty")
		}
		
		for _, send := range usage.Sends {
			require.NotNil(t, send.Resource, "Resource should not be nil")
			require.NotNil(t, send.Server, "Server should not be nil")
			require.NotNil(t, send.Message, "Message should not be nil")
		}
	})
	
	// Test that non-existent app returns 404
	t.Run("non-existent app", func(t *testing.T) {
		e.GET("/v1/protected/apps/00000000-0000-0000-0000-000000000000/usage").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusNotFound)
	})
}