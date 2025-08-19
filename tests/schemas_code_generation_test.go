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

func TestSchemaCodeGeneration(t *testing.T) {
	// Clean database before running test
	CleanDatabase(t)

	h := os.Getenv("TESTSERVER_URL")
	e := httpexpect.Default(t, h)

	// Create test user
	testEmail := fmt.Sprintf("test-codegen-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
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
	projectName := fmt.Sprintf("CodeGenTestProject%d", time.Now().UnixNano())
	projectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        projectName,
		Description: "Test project for schema code generation",
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

	// Create a test schema
	schemaName := "user_profile"
	schemaContent := `{"$schema":"https://json-schema.org/draft/2020-12/schema","title":"UserProfile","type":"object","properties":{"id":{"type":"string"},"name":{"type":"string"},"email":{"type":"string"},"age":{"type":"integer"},"is_active":{"type":"boolean"}},"required":["id","name","email"]}`

	schemaPayload := input_contracts.CreateSchemaApiInputContract{
		Name:        schemaName,
		Description: "Test user profile schema",
		Schema:      schemaContent,
		Type:        "jsonschema",
	}

	schemaResponse := e.POST("/v1/protected/projects/"+createdProject.ID+"/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(schemaPayload).
		Expect()

	// Check the response and print error if 422
	rawSchemaReader := schemaResponse.Raw().Body
	defer rawSchemaReader.Close()
	rawSchemaBytes, _ := io.ReadAll(rawSchemaReader)

	if schemaResponse.Raw().StatusCode != http.StatusOK {
		t.Logf("Schema creation failed with status %d: %s", schemaResponse.Raw().StatusCode, string(rawSchemaBytes))
		require.Equal(t, http.StatusOK, schemaResponse.Raw().StatusCode, "Schema creation should succeed")
	}

	var createdSchema logic.SchemaDBSerializerStruct
	require.NoError(t, json.Unmarshal(rawSchemaBytes, &createdSchema))

	// Test Go code generation
	t.Run("Generate Go code", func(t *testing.T) {
		codeResponse := e.GET("/v1/protected/schemas/"+createdSchema.ID+"/code/go").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusOK)

		codeResponse.Header("Content-Disposition").Contains("generated_code.go")

		generatedCode := codeResponse.Body().Raw()

		// Check that the generated Go code contains the expected struct with capitalized name
		require.Contains(t, generatedCode, "type UserProfile struct", "Generated Go code should contain UserProfile struct with capital letters")
		require.Contains(t, generatedCode, "ID", "Should contain ID field")
		require.Contains(t, generatedCode, "Name", "Should contain Name field")
		require.Contains(t, generatedCode, "Email", "Should contain Email field")
		require.Contains(t, generatedCode, "Age", "Should contain Age field")
		require.Contains(t, generatedCode, "IsActive", "Should contain IsActive field")
		require.Contains(t, generatedCode, `json:"id"`, "Should have json tag for id")
		require.Contains(t, generatedCode, `json:"email"`, "Should have json tag for email")

		// Ensure the struct name is capitalized
		require.NotContains(t, generatedCode, "type userProfile struct", "Struct name should be capitalized")
		require.NotContains(t, generatedCode, "type user_profile struct", "Struct name should be in CamelCase, not snake_case")
	})

	// Test Python code generation
	t.Run("Generate Python code", func(t *testing.T) {
		codeResponse := e.GET("/v1/protected/schemas/"+createdSchema.ID+"/code/python").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusOK)

		codeResponse.Header("Content-Disposition").Contains("generated_code.py")

		generatedCode := codeResponse.Body().Raw()

		// Check that the generated Python code contains expected content
		require.Contains(t, generatedCode, "class", "Generated Python code should contain a class")
		// Python code structure may vary based on quicktype version
	})

	// Test TypeScript code generation
	t.Run("Generate TypeScript code", func(t *testing.T) {
		codeResponse := e.GET("/v1/protected/schemas/"+createdSchema.ID+"/code/typescript").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusOK)

		codeResponse.Header("Content-Disposition").Contains("generated_code.ts")

		generatedCode := codeResponse.Body().Raw()

		// Check that the generated TypeScript code contains expected content
		require.Contains(t, generatedCode, "interface", "Generated TypeScript code should contain an interface")
	})

	// Test Java code generation
	t.Run("Generate Java code", func(t *testing.T) {
		codeResponse := e.GET("/v1/protected/schemas/"+createdSchema.ID+"/code/java").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusOK)

		codeResponse.Header("Content-Disposition").Contains("generated_code.java")

		generatedCode := codeResponse.Body().Raw()

		// Check that the generated Java code contains expected content
		require.Contains(t, generatedCode, "class", "Generated Java code should contain a class")
	})

	// Test invalid language
	t.Run("Invalid language", func(t *testing.T) {
		e.GET("/v1/protected/schemas/"+createdSchema.ID+"/code/ruby").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ValueEqual("error", "Invalid programming language. Supported languages are: typescript, java, go, python")
	})

	// Test non-existent schema
	t.Run("Non-existent schema", func(t *testing.T) {
		e.GET("/v1/protected/schemas/00000000-0000-0000-0000-000000000000/code/go").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object().
			ValueEqual("error", "Schema not found")
	})

	// Test schema with snake_case name to ensure proper capitalization
	t.Run("Snake case schema name", func(t *testing.T) {
		snakeSchemaName := "order_item_detail"
		snakeSchemaContent := `{"$schema":"https://json-schema.org/draft/2020-12/schema","title":"OrderItemDetail","type":"object","properties":{"order_id":{"type":"string"},"item_name":{"type":"string"}}}`

		snakeSchemaPayload := input_contracts.CreateSchemaApiInputContract{
			Name:        snakeSchemaName,
			Description: "Test schema with snake_case name",
			Schema:      snakeSchemaContent,
			Type:        "jsonschema",
		}

		snakeSchemaResponse := e.POST("/v1/protected/projects/"+createdProject.ID+"/schemas").
			WithHeader("Authorization", bearerToken).
			WithJSON(snakeSchemaPayload).
			Expect().
			Status(http.StatusOK)

		var snakeSchema logic.SchemaDBSerializerStruct
		rawSnakeReader := snakeSchemaResponse.Raw().Body
		defer rawSnakeReader.Close()
		rawSnakeBytes, _ := io.ReadAll(rawSnakeReader)
		require.NoError(t, json.Unmarshal(rawSnakeBytes, &snakeSchema))

		codeResponse := e.GET("/v1/protected/schemas/"+snakeSchema.ID+"/code/go").
			WithHeader("Authorization", bearerToken).
			Expect().
			Status(http.StatusOK)

		generatedCode := codeResponse.Body().Raw()

		// Check that snake_case name is converted to CamelCase with capital first letter
		require.Contains(t, generatedCode, "type OrderItemDetail struct", "Snake_case should be converted to CamelCase with capital first letter")
		require.NotContains(t, generatedCode, "type order_item_detail struct", "Should not contain snake_case struct name")
		require.NotContains(t, generatedCode, "type orderItemDetail struct", "First letter should be capitalized")
	})
}
