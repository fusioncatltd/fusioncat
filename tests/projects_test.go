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

func TestProjectsEndpoints(t *testing.T) {
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

	// First user sees no projects initially
	firstUserProjectsResponse := e.GET("/v1/protected/projects").
		WithHeader("Authorization", firstUserBearer).
		Expect().
		Status(http.StatusOK)

	var firstUserProjects []logic.ProjectDBSerializerStruct
	rawFirstUserProjectsReader := firstUserProjectsResponse.Raw().Body
	defer rawFirstUserProjectsReader.Close()
	rawFirstUserProjectsBytes, _ := io.ReadAll(rawFirstUserProjectsReader)

	require.NoError(t, json.Unmarshal(rawFirstUserProjectsBytes, &firstUserProjects))
	require.Empty(t, firstUserProjects, "First user should initially have no projects")

	// First user creates a project
	firstProjectName := fmt.Sprintf("TestProject%d", time.Now().UnixNano())
	firstProjectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        firstProjectName,
		Description: "Test project description",
	}

	firstProjectResponse := e.POST("/v1/protected/projects").
		WithHeader("Authorization", firstUserBearer).
		WithJSON(firstProjectPayload).
		Expect().
		Status(http.StatusOK)

	var createdFirstProject logic.ProjectDBSerializerStruct
	rawFirstProjectReader := firstProjectResponse.Raw().Body
	defer rawFirstProjectReader.Close()
	rawFirstProjectBytes, _ := io.ReadAll(rawFirstProjectReader)

	require.NoError(t, json.Unmarshal(rawFirstProjectBytes, &createdFirstProject))
	require.Equal(t, firstProjectName, createdFirstProject.Name)
	require.Equal(t, "Test project description", createdFirstProject.Description)

	// Test getting a single project by ID
	singleProjectResponse := e.GET("/v1/protected/projects/" + createdFirstProject.ID).
		WithHeader("Authorization", firstUserBearer).
		Expect().
		Status(http.StatusOK)

	var singleProject logic.ProjectDBSerializerStruct
	rawSingleProjectReader := singleProjectResponse.Raw().Body
	defer rawSingleProjectReader.Close()
	rawSingleProjectBytes, _ := io.ReadAll(rawSingleProjectReader)

	require.NoError(t, json.Unmarshal(rawSingleProjectBytes, &singleProject))
	require.Equal(t, createdFirstProject.ID, singleProject.ID)
	require.Equal(t, firstProjectName, singleProject.Name)
	require.Equal(t, "Test project description", singleProject.Description)

	// Test getting non-existent project returns 404
	_ = e.GET("/v1/protected/projects/00000000-0000-0000-0000-000000000000").
		WithHeader("Authorization", firstUserBearer).
		Expect().
		Status(http.StatusNotFound)

	// First user now sees one project in the list
	firstUserProjectsAfterCreate := e.GET("/v1/protected/projects").
		WithHeader("Authorization", firstUserBearer).
		Expect().
		Status(http.StatusOK)

	var firstUserProjectsListAfterCreate []logic.ProjectDBSerializerStruct
	rawFirstUserProjectsAfterReader := firstUserProjectsAfterCreate.Raw().Body
	defer rawFirstUserProjectsAfterReader.Close()
	rawFirstUserProjectsAfterBytes, _ := io.ReadAll(rawFirstUserProjectsAfterReader)

	require.NoError(t, json.Unmarshal(rawFirstUserProjectsAfterBytes, &firstUserProjectsListAfterCreate))
	require.Len(t, firstUserProjectsListAfterCreate, 1, "First user should have one project after creation")
	require.Equal(t, firstProjectName, firstUserProjectsListAfterCreate[0].Name)

	// Create second user
	secondUserEmail := fmt.Sprintf("test-email-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	secondUserPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    secondUserEmail,
		Password: "987654321",
	}

	secondUserSignUpResponse := e.POST("/v1/public/users").
		WithJSON(secondUserPayload).
		Expect().
		Status(http.StatusOK)

	secondUserBearer := secondUserSignUpResponse.Raw().Header.Get("Authorization")

	// Second user sees the previously created project
	secondUserProjectsResponse := e.GET("/v1/protected/projects").
		WithHeader("Authorization", secondUserBearer).
		Expect().
		Status(http.StatusOK)

	var secondUserProjects []logic.ProjectDBSerializerStruct
	rawSecondUserProjectsReader := secondUserProjectsResponse.Raw().Body
	defer rawSecondUserProjectsReader.Close()
	rawSecondUserProjectsBytes, _ := io.ReadAll(rawSecondUserProjectsReader)

	require.NoError(t, json.Unmarshal(rawSecondUserProjectsBytes, &secondUserProjects))
	require.Len(t, secondUserProjects, 1, "Second user should see the previously created project")
	require.Equal(t, firstProjectName, secondUserProjects[0].Name)

	// Second user creates a new project
	secondProjectName := fmt.Sprintf("SecondProject%d", time.Now().UnixNano())
	secondProjectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        secondProjectName,
		Description: "Second test project",
	}

	secondProjectResponse := e.POST("/v1/protected/projects").
		WithHeader("Authorization", secondUserBearer).
		WithJSON(secondProjectPayload).
		Expect().
		Status(http.StatusOK)

	var createdSecondProject logic.ProjectDBSerializerStruct
	rawSecondProjectReader := secondProjectResponse.Raw().Body
	defer rawSecondProjectReader.Close()
	rawSecondProjectBytes, _ := io.ReadAll(rawSecondProjectReader)

	require.NoError(t, json.Unmarshal(rawSecondProjectBytes, &createdSecondProject))
	require.Equal(t, secondProjectName, createdSecondProject.Name)

	// Second user can also get the first project by ID
	secondUserSingleProjectResponse := e.GET("/v1/protected/projects/" + createdFirstProject.ID).
		WithHeader("Authorization", secondUserBearer).
		Expect().
		Status(http.StatusOK)

	var secondUserSingleProject logic.ProjectDBSerializerStruct
	rawSecondUserSingleProjectReader := secondUserSingleProjectResponse.Raw().Body
	defer rawSecondUserSingleProjectReader.Close()
	rawSecondUserSingleProjectBytes, _ := io.ReadAll(rawSecondUserSingleProjectReader)

	require.NoError(t, json.Unmarshal(rawSecondUserSingleProjectBytes, &secondUserSingleProject))
	require.Equal(t, createdFirstProject.ID, secondUserSingleProject.ID)
	require.Equal(t, firstProjectName, secondUserSingleProject.Name)

	// Second user now sees 2 projects
	secondUserProjectsAfterCreate := e.GET("/v1/protected/projects").
		WithHeader("Authorization", secondUserBearer).
		Expect().
		Status(http.StatusOK)

	var secondUserProjectsListAfterCreate []logic.ProjectDBSerializerStruct
	rawSecondUserProjectsAfterReader := secondUserProjectsAfterCreate.Raw().Body
	defer rawSecondUserProjectsAfterReader.Close()
	rawSecondUserProjectsAfterBytes, _ := io.ReadAll(rawSecondUserProjectsAfterReader)

	require.NoError(t, json.Unmarshal(rawSecondUserProjectsAfterBytes, &secondUserProjectsListAfterCreate))
	require.Len(t, secondUserProjectsListAfterCreate, 2, "Second user should see 2 projects after creating one")

	// Second user tries to create a project with existing name (should get 409)
	duplicateProjectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        firstProjectName, // Using the name of the first project
		Description: "Duplicate project attempt",
	}

	_ = e.POST("/v1/protected/projects").
		WithHeader("Authorization", secondUserBearer).
		WithJSON(duplicateProjectPayload).
		Expect().
		Status(http.StatusConflict)

	// Test validation: only alphanumeric symbols allowed for project names
	invalidNamePayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        "Invalid-Project-Name!", // Contains special characters
		Description: "Testing invalid name",
	}

	_ = e.POST("/v1/protected/projects").
		WithHeader("Authorization", secondUserBearer).
		WithJSON(invalidNamePayload).
		Expect().
		Status(http.StatusUnprocessableEntity)

	// Test another invalid name with spaces
	invalidNameWithSpacesPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        "Project With Spaces",
		Description: "Testing name with spaces",
	}

	_ = e.POST("/v1/protected/projects").
		WithHeader("Authorization", secondUserBearer).
		WithJSON(invalidNameWithSpacesPayload).
		Expect().
		Status(http.StatusUnprocessableEntity)

	// Test valid alphanumeric name
	validAlphanumericPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        "ValidProject123",
		Description: "Testing valid alphanumeric name",
	}

	validNameResponse := e.POST("/v1/protected/projects").
		WithHeader("Authorization", secondUserBearer).
		WithJSON(validAlphanumericPayload).
		Expect().
		Status(http.StatusOK)

	var validNameProject logic.ProjectDBSerializerStruct
	rawValidNameReader := validNameResponse.Raw().Body
	defer rawValidNameReader.Close()
	rawValidNameBytes, _ := io.ReadAll(rawValidNameReader)

	require.NoError(t, json.Unmarshal(rawValidNameBytes, &validNameProject))
	require.Equal(t, "ValidProject123", validNameProject.Name)
}
