package tests

import (
	"encoding/json"
	"fmt"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestAppsEndpoints(t *testing.T) {
	// Clean database before running test
	CleanDatabase(t)
	
	h := os.Getenv("TESTSERVER_URL")
	e := httpexpect.Default(t, h)

	// Create first user
	firstUserEmail := fmt.Sprintf("test-email-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	firstUserPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    firstUserEmail,
		Password: "123456789",
	}

	// Create first user and get bearer token
	firstUserSignUpResponse := e.POST("/v1/public/users").
		WithJSON(firstUserPayload).
		Expect().
		Status(http.StatusOK)

	firstUserSignUpResponse.Header("Authorization").NotEmpty()
	firstUserBearer := firstUserSignUpResponse.Raw().Header.Get("Authorization")

	// Create a project first
	projectName := fmt.Sprintf("TestProject%d", time.Now().UnixNano())
	projectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        projectName,
		Description: "Test project for apps",
	}

	projectResponse := e.POST("/v1/protected/projects").
		WithHeader("Authorization", firstUserBearer).
		WithJSON(projectPayload).
		Expect().
		Status(http.StatusOK)

	var createdProject logic.ProjectDBSerializerStruct
	rawProjectReader := projectResponse.Raw().Body
	defer rawProjectReader.Close()
	rawProjectBytes, _ := io.ReadAll(rawProjectReader)

	require.NoError(t, json.Unmarshal(rawProjectBytes, &createdProject))

	// Test creating an app
	appName := fmt.Sprintf("TestApp%d", time.Now().UnixNano())
	appPayload := input_contracts.CreateAppApiInputContract{
		Name:        appName,
		Description: "Test app description",
	}

	createAppResponse := e.POST("/v1/protected/projects/" + createdProject.ID + "/apps").
		WithHeader("Authorization", firstUserBearer).
		WithJSON(appPayload).
		Expect().
		Status(http.StatusOK)

	var createdApp logic.AppDBSerializerStruct
	rawAppReader := createAppResponse.Raw().Body
	defer rawAppReader.Close()
	rawAppBytes, _ := io.ReadAll(rawAppReader)

	require.NoError(t, json.Unmarshal(rawAppBytes, &createdApp))
	require.Equal(t, appName, createdApp.Name)
	require.Equal(t, "Test app description", createdApp.Description)
	require.Equal(t, createdProject.ID, createdApp.ProjectID)

	// Test getting all apps for a project
	getAppsResponse := e.GET("/v1/protected/projects/" + createdProject.ID + "/apps").
		WithHeader("Authorization", firstUserBearer).
		Expect().
		Status(http.StatusOK)

	var apps []logic.AppDBSerializerStruct
	rawAppsReader := getAppsResponse.Raw().Body
	defer rawAppsReader.Close()
	rawAppsBytes, _ := io.ReadAll(rawAppsReader)

	require.NoError(t, json.Unmarshal(rawAppsBytes, &apps))
	require.Len(t, apps, 1)
	require.Equal(t, createdApp.ID, apps[0].ID)
	require.Equal(t, appName, apps[0].Name)

	// Test creating an app with duplicate name (should fail)
	duplicateAppPayload := input_contracts.CreateAppApiInputContract{
		Name:        appName,
		Description: "Duplicate app",
	}

	_ = e.POST("/v1/protected/projects/" + createdProject.ID + "/apps").
		WithHeader("Authorization", firstUserBearer).
		WithJSON(duplicateAppPayload).
		Expect().
		Status(http.StatusConflict)

	// Test creating an app for non-existent project
	_ = e.POST("/v1/protected/projects/00000000-0000-0000-0000-000000000000/apps").
		WithHeader("Authorization", firstUserBearer).
		WithJSON(appPayload).
		Expect().
		Status(http.StatusNotFound)

	// Test getting apps for non-existent project
	_ = e.GET("/v1/protected/projects/00000000-0000-0000-0000-000000000000/apps").
		WithHeader("Authorization", firstUserBearer).
		Expect().
		Status(http.StatusNotFound)

	// Test validation errors
	invalidAppPayload := input_contracts.CreateAppApiInputContract{
		Name:        "", // Empty name should fail validation
		Description: "Invalid app",
	}

	_ = e.POST("/v1/protected/projects/" + createdProject.ID + "/apps").
		WithHeader("Authorization", firstUserBearer).
		WithJSON(invalidAppPayload).
		Expect().
		Status(http.StatusUnprocessableEntity)

	// Create second app with different name
	secondAppName := fmt.Sprintf("SecondTestApp%d", time.Now().UnixNano())
	secondAppPayload := input_contracts.CreateAppApiInputContract{
		Name:        secondAppName,
		Description: "Second test app",
	}

	secondAppResponse := e.POST("/v1/protected/projects/" + createdProject.ID + "/apps").
		WithHeader("Authorization", firstUserBearer).
		WithJSON(secondAppPayload).
		Expect().
		Status(http.StatusOK)

	var secondApp logic.AppDBSerializerStruct
	rawSecondAppReader := secondAppResponse.Raw().Body
	defer rawSecondAppReader.Close()
	rawSecondAppBytes, _ := io.ReadAll(rawSecondAppReader)

	require.NoError(t, json.Unmarshal(rawSecondAppBytes, &secondApp))
	require.Equal(t, secondAppName, secondApp.Name)

	// Verify both apps are returned in the list
	allAppsResponse := e.GET("/v1/protected/projects/" + createdProject.ID + "/apps").
		WithHeader("Authorization", firstUserBearer).
		Expect().
		Status(http.StatusOK)

	var allApps []logic.AppDBSerializerStruct
	rawAllAppsReader := allAppsResponse.Raw().Body
	defer rawAllAppsReader.Close()
	rawAllAppsBytes, _ := io.ReadAll(rawAllAppsReader)

	require.NoError(t, json.Unmarshal(rawAllAppsBytes, &allApps))
	require.Len(t, allApps, 2)

	// Test accessing apps without authentication
	_ = e.POST("/v1/protected/projects/" + createdProject.ID + "/apps").
		WithJSON(appPayload).
		Expect().
		Status(http.StatusUnauthorized)

	_ = e.GET("/v1/protected/projects/" + createdProject.ID + "/apps").
		Expect().
		Status(http.StatusUnauthorized)
}