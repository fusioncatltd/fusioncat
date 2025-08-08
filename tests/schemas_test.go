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

func TestSchemasEndpoints(t *testing.T) {
	h := os.Getenv("TESTSERVER_URL")
	e := httpexpect.Default(t, h)

	// Create a user
	userEmail := fmt.Sprintf("test-schemas-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	userPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    userEmail,
		Password: "TestPassword123",
	}

	// Sign up and get bearer token
	signUpResponse := e.POST("/v1/public/users").
		WithJSON(userPayload).
		Expect().
		Status(http.StatusOK)

	bearerToken := signUpResponse.Raw().Header.Get("Authorization")
	require.NotEmpty(t, bearerToken, "Bearer token should not be empty")

	// Create a project
	projectName := fmt.Sprintf("SchemasTestProject%d", time.Now().UnixNano())
	projectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        projectName,
		Description: "Project for testing schemas",
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
	projectID := createdProject.ID

	// Test 1: Get schemas list - should be empty initially
	schemasListResponse := e.GET("/v1/protected/projects/"+projectID+"/schemas").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusOK)

	var initialSchemas []logic.SchemaDBSerializerStruct
	rawSchemasReader := schemasListResponse.Raw().Body
	defer rawSchemasReader.Close()
	rawSchemasBytes, _ := io.ReadAll(rawSchemasReader)

	require.NoError(t, json.Unmarshal(rawSchemasBytes, &initialSchemas))
	require.Empty(t, initialSchemas, "Project should have no schemas initially")

	// Test 2: Create a schema with valid JSON schema
	validSchemaContent, err := ReadTestFileString("jsonschemas/validSchema1.json")
	require.NoError(t, err, "Should be able to read valid schema file")

	schemaName := fmt.Sprintf("TestSchema%d", time.Now().UnixNano())
	createSchemaPayload := input_contracts.CreateSchemaApiInputContract{
		Name:        schemaName,
		Description: "Test schema for person data",
		Type:        "jsonschema",
		Schema:      validSchemaContent,
	}

	createSchemaResponse := e.POST("/v1/protected/projects/"+projectID+"/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(createSchemaPayload).
		Expect().
		Status(http.StatusOK)

	var createdSchema logic.SchemaDBSerializerStruct
	rawCreatedSchemaReader := createSchemaResponse.Raw().Body
	defer rawCreatedSchemaReader.Close()
	rawCreatedSchemaBytes, _ := io.ReadAll(rawCreatedSchemaReader)

	require.NoError(t, json.Unmarshal(rawCreatedSchemaBytes, &createdSchema))
	require.Equal(t, schemaName, createdSchema.Name)
	require.Equal(t, "Test schema for person data", createdSchema.Description)
	require.Equal(t, "jsonschema", createdSchema.Type)
	require.Equal(t, projectID, createdSchema.ProjectID)
	require.JSONEq(t, validSchemaContent, createdSchema.Schema)

	// Test 3: Get schemas list - should now have one schema
	schemasListAfterCreateResponse := e.GET("/v1/protected/projects/"+projectID+"/schemas").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusOK)

	var schemasAfterCreate []logic.SchemaDBSerializerStruct
	rawSchemasAfterReader := schemasListAfterCreateResponse.Raw().Body
	defer rawSchemasAfterReader.Close()
	rawSchemasAfterBytes, _ := io.ReadAll(rawSchemasAfterReader)

	require.NoError(t, json.Unmarshal(rawSchemasAfterBytes, &schemasAfterCreate))
	require.Len(t, schemasAfterCreate, 1, "Project should have one schema after creation")
	require.Equal(t, schemaName, schemasAfterCreate[0].Name)

	// Test 4: Try to create a schema with duplicate name (should fail with 409)
	duplicateSchemaPayload := input_contracts.CreateSchemaApiInputContract{
		Name:        schemaName, // Same name as before
		Description: "Duplicate schema attempt",
		Type:        "jsonschema",
		Schema:      validSchemaContent,
	}

	_ = e.POST("/v1/protected/projects/"+projectID+"/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(duplicateSchemaPayload).
		Expect().
		Status(http.StatusConflict)

	// Test 5: Try to create a schema with invalid JSON schema
	invalidSchemaContent, err := ReadTestFileString("jsonschemas/invalidSchema1InvalidType.json")
	require.NoError(t, err, "Should be able to read invalid schema file")

	invalidSchemaPayload := input_contracts.CreateSchemaApiInputContract{
		Name:        "InvalidSchema",
		Description: "Schema with invalid content",
		Type:        "jsonschema",
		Schema:      invalidSchemaContent,
	}

	_ = e.POST("/v1/protected/projects/"+projectID+"/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(invalidSchemaPayload).
		Expect().
		Status(http.StatusUnprocessableEntity)

	// Test 6: Try to create a schema in non-existent project
	_ = e.POST("/v1/protected/projects/00000000-0000-0000-0000-000000000000/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(createSchemaPayload).
		Expect().
		Status(http.StatusNotFound)

	// Test 7: Try to get schemas from non-existent project
	_ = e.GET("/v1/protected/projects/00000000-0000-0000-0000-000000000000/schemas").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusNotFound)

	// Test 8: Create a second valid schema
	secondSchemaName := fmt.Sprintf("SecondSchema%d", time.Now().UnixNano())
	secondSchemaPayload := input_contracts.CreateSchemaApiInputContract{
		Name:        secondSchemaName,
		Description: "Second test schema",
		Type:        "jsonschema",
		Schema:      validSchemaContent,
	}

	secondSchemaResponse := e.POST("/v1/protected/projects/"+projectID+"/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(secondSchemaPayload).
		Expect().
		Status(http.StatusOK)

	var secondCreatedSchema logic.SchemaDBSerializerStruct
	rawSecondSchemaReader := secondSchemaResponse.Raw().Body
	defer rawSecondSchemaReader.Close()
	rawSecondSchemaBytes, _ := io.ReadAll(rawSecondSchemaReader)

	require.NoError(t, json.Unmarshal(rawSecondSchemaBytes, &secondCreatedSchema))
	require.Equal(t, secondSchemaName, secondCreatedSchema.Name)

	// Test 9: Get schemas list - should now have two schemas
	finalSchemasListResponse := e.GET("/v1/protected/projects/"+projectID+"/schemas").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusOK)

	var finalSchemas []logic.SchemaDBSerializerStruct
	rawFinalSchemasReader := finalSchemasListResponse.Raw().Body
	defer rawFinalSchemasReader.Close()
	rawFinalSchemasBytes, _ := io.ReadAll(rawFinalSchemasReader)

	require.NoError(t, json.Unmarshal(rawFinalSchemasBytes, &finalSchemas))
	require.Len(t, finalSchemas, 2, "Project should have two schemas after creating second one")

	// Verify both schemas are in the list
	schemaNames := make(map[string]bool)
	for _, schema := range finalSchemas {
		schemaNames[schema.Name] = true
	}
	require.True(t, schemaNames[schemaName], "First schema should be in the list")
	require.True(t, schemaNames[secondSchemaName], "Second schema should be in the list")

	// Test 10: Test with invalid schema name (contains special characters)
	invalidNamePayload := input_contracts.CreateSchemaApiInputContract{
		Name:        "Invalid-Schema-Name!",
		Description: "Schema with invalid name",
		Type:        "jsonschema",
		Schema:      validSchemaContent,
	}

	_ = e.POST("/v1/protected/projects/"+projectID+"/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(invalidNamePayload).
		Expect().
		Status(http.StatusUnprocessableEntity)

	// Test 11: Test with valid alphanumeric name with underscore
	validNamePayload := input_contracts.CreateSchemaApiInputContract{
		Name:        "Valid_Schema_123",
		Description: "Schema with valid name containing underscore",
		Type:        "jsonschema",
		Schema:      validSchemaContent,
	}

	validNameResponse := e.POST("/v1/protected/projects/"+projectID+"/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(validNamePayload).
		Expect().
		Status(http.StatusOK)

	var validNameSchema logic.SchemaDBSerializerStruct
	rawValidNameReader := validNameResponse.Raw().Body
	defer rawValidNameReader.Close()
	rawValidNameBytes, _ := io.ReadAll(rawValidNameReader)

	require.NoError(t, json.Unmarshal(rawValidNameBytes, &validNameSchema))
	require.Equal(t, "Valid_Schema_123", validNameSchema.Name)

	// Test 12: Get single schema by ID
	singleSchemaResponse := e.GET("/v1/protected/schemas/"+createdSchema.ID).
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusOK)

	var singleSchema logic.SchemaDBSerializerStruct
	rawSingleSchemaReader := singleSchemaResponse.Raw().Body
	defer rawSingleSchemaReader.Close()
	rawSingleSchemaBytes, _ := io.ReadAll(rawSingleSchemaReader)

	require.NoError(t, json.Unmarshal(rawSingleSchemaBytes, &singleSchema))
	require.Equal(t, createdSchema.ID, singleSchema.ID)
	require.Equal(t, schemaName, singleSchema.Name)
	require.Equal(t, "Test schema for person data", singleSchema.Description)
	require.Equal(t, 1, singleSchema.Version)
	require.JSONEq(t, validSchemaContent, singleSchema.Schema)

	// Test 13: Get non-existent schema by ID (should return 404)
	_ = e.GET("/v1/protected/schemas/00000000-0000-0000-0000-000000000000").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusNotFound)

	// Test 14: Modify schema (create new version)
	// Create a modified schema content
	modifiedSchemaContent := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title": "Person",
		"type": "object",
		"properties": {
			"firstName": {
				"type": "string",
				"description": "The person's first name."
			},
			"lastName": {
				"type": "string",
				"description": "The person's last name."
			},
			"age": {
				"description": "Age in years which must be a non-negative integer.",
				"type": "integer",
				"minimum": 0
			},
			"email": {
				"type": "string",
				"format": "email",
				"description": "The person's email address."
			}
		},
		"required": ["firstName", "lastName", "email"]
	}`

	modifySchemaPayload := input_contracts.ModifySchemaApiInputContract{
		Schema: modifiedSchemaContent,
	}

	modifySchemaResponse := e.PUT("/v1/protected/schemas/"+createdSchema.ID).
		WithHeader("Authorization", bearerToken).
		WithJSON(modifySchemaPayload).
		Expect().
		Status(http.StatusOK)

	var modifiedSchema logic.SchemaDBSerializerStruct
	rawModifiedSchemaReader := modifySchemaResponse.Raw().Body
	defer rawModifiedSchemaReader.Close()
	rawModifiedSchemaBytes, _ := io.ReadAll(rawModifiedSchemaReader)

	require.NoError(t, json.Unmarshal(rawModifiedSchemaBytes, &modifiedSchema))
	require.Equal(t, createdSchema.ID, modifiedSchema.ID)
	require.Equal(t, schemaName, modifiedSchema.Name)
	require.Equal(t, 2, modifiedSchema.Version, "Schema version should be incremented to 2")
	require.JSONEq(t, modifiedSchemaContent, modifiedSchema.Schema)

	// Test 15: Get the modified schema by ID to verify changes persist
	getModifiedSchemaResponse := e.GET("/v1/protected/schemas/"+createdSchema.ID).
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusOK)

	var retrievedModifiedSchema logic.SchemaDBSerializerStruct
	rawRetrievedModifiedReader := getModifiedSchemaResponse.Raw().Body
	defer rawRetrievedModifiedReader.Close()
	rawRetrievedModifiedBytes, _ := io.ReadAll(rawRetrievedModifiedReader)

	require.NoError(t, json.Unmarshal(rawRetrievedModifiedBytes, &retrievedModifiedSchema))
	require.Equal(t, 2, retrievedModifiedSchema.Version, "Retrieved schema should show version 2")
	require.JSONEq(t, modifiedSchemaContent, retrievedModifiedSchema.Schema)

	// Test 16: Try to modify non-existent schema (should return 404)
	_ = e.PUT("/v1/protected/schemas/00000000-0000-0000-0000-000000000000").
		WithHeader("Authorization", bearerToken).
		WithJSON(modifySchemaPayload).
		Expect().
		Status(http.StatusNotFound)

	// Test 17: Try to modify schema with invalid JSON schema
	invalidModifyPayload := input_contracts.ModifySchemaApiInputContract{
		Schema: invalidSchemaContent,
	}

	_ = e.PUT("/v1/protected/schemas/"+createdSchema.ID).
		WithHeader("Authorization", bearerToken).
		WithJSON(invalidModifyPayload).
		Expect().
		Status(http.StatusUnprocessableEntity)

	// Test 18: Modify schema again to create version 3
	anotherModifiedSchema, err := ReadTestFileString("jsonschemas/validSchema2.json")
	require.NoError(t, err, "Should be able to read person schema with validation file")

	anotherModifyPayload := input_contracts.ModifySchemaApiInputContract{
		Schema: anotherModifiedSchema,
	}

	finalModifyResponse := e.PUT("/v1/protected/schemas/"+createdSchema.ID).
		WithHeader("Authorization", bearerToken).
		WithJSON(anotherModifyPayload).
		Expect().
		Status(http.StatusOK)

	var finalModifiedSchema logic.SchemaDBSerializerStruct
	rawFinalModifiedReader := finalModifyResponse.Raw().Body
	defer rawFinalModifiedReader.Close()
	rawFinalModifiedBytes, _ := io.ReadAll(rawFinalModifiedReader)

	require.NoError(t, json.Unmarshal(rawFinalModifiedBytes, &finalModifiedSchema))
	require.Equal(t, 3, finalModifiedSchema.Version, "Schema version should be incremented to 3")
	require.JSONEq(t, anotherModifiedSchema, finalModifiedSchema.Schema)

	// Test 19: Get another user's schema (second user gets first user's schema)
	secondUserEmail := fmt.Sprintf("test-schemas-user2-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	secondUserPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    secondUserEmail,
		Password: "SecondUserPass123",
	}

	secondUserSignUpResponse := e.POST("/v1/public/users").
		WithJSON(secondUserPayload).
		Expect().
		Status(http.StatusOK)

	secondUserBearer := secondUserSignUpResponse.Raw().Header.Get("Authorization")

	// Second user should be able to get the schema since they have access to the project
	secondUserGetSchemaResponse := e.GET("/v1/protected/schemas/"+createdSchema.ID).
		WithHeader("Authorization", secondUserBearer).
		Expect().
		Status(http.StatusOK)

	var secondUserSchema logic.SchemaDBSerializerStruct
	rawSecondUserSchemaReader := secondUserGetSchemaResponse.Raw().Body
	defer rawSecondUserSchemaReader.Close()
	rawSecondUserSchemaBytes, _ := io.ReadAll(rawSecondUserSchemaReader)

	require.NoError(t, json.Unmarshal(rawSecondUserSchemaBytes, &secondUserSchema))
	require.Equal(t, createdSchema.ID, secondUserSchema.ID)

	// Second user should be able to modify the schema too
	secondUserModifyPayload := input_contracts.ModifySchemaApiInputContract{
		Schema: validSchemaContent,
	}

	secondUserModifyResponse := e.PUT("/v1/protected/schemas/"+createdSchema.ID).
		WithHeader("Authorization", secondUserBearer).
		WithJSON(secondUserModifyPayload).
		Expect().
		Status(http.StatusOK)

	var secondUserModifiedSchema logic.SchemaDBSerializerStruct
	rawSecondUserModifiedReader := secondUserModifyResponse.Raw().Body
	defer rawSecondUserModifiedReader.Close()
	rawSecondUserModifiedBytes, _ := io.ReadAll(rawSecondUserModifiedReader)

	require.NoError(t, json.Unmarshal(rawSecondUserModifiedBytes, &secondUserModifiedSchema))
	require.Equal(t, 4, secondUserModifiedSchema.Version, "Schema version should be incremented to 4 after second user's modification")
}
